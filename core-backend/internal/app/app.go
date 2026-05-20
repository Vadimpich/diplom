package app

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"diplom/internal/aggregation"
	"diplom/internal/answers"
	"diplom/internal/audit"
	"diplom/internal/auth"
	"diplom/internal/baselineclient"
	"diplom/internal/channelresults"
	"diplom/internal/config"
	"diplom/internal/decision"
	"diplom/internal/examinations"
	httpserver "diplom/internal/http"
	"diplom/internal/kesmi"
	appmigrations "diplom/internal/migrations"
	"diplom/internal/observability"
	"diplom/internal/postgres"
	"diplom/internal/processing"
	"diplom/internal/questionnaires"
	"diplom/internal/results"
	"diplom/internal/settings"
	"diplom/internal/specialists"
	"diplom/internal/storage"
)

type App struct {
	server          *http.Server
	db              *postgres.Client
	publisher       *processing.Relay
	resultsConsumer *processing.ResultsConsumer
	decisionRelay   *decision.Relay
	logger          *slog.Logger
}

func New(ctx context.Context, cfg config.Config) (*App, error) {
	db, err := postgres.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	if err := appmigrations.Apply(ctx, db.Pool(), cfg.MigrationsDir); err != nil {
		db.Close()
		return nil, fmt.Errorf("apply migrations: %w", err)
	}

	queries := db.Queries()
	auditService := audit.NewService(audit.NewRepository(db.Pool()))

	tokenManager := auth.NewJWTManager(cfg.JWTIssuer, cfg.JWTAccessSecret, cfg.JWTAccessTTL)
	authRepository := auth.NewRepository(db.Pool())
	authService := auth.NewService(authRepository, tokenManager, auditService)

	if err := authService.EnsureInitialUser(ctx, auth.BootstrapConfig{
		Login:    cfg.InitialUserLogin,
		Password: cfg.InitialUserPassword,
		RoleSlug: cfg.InitialUserRole,
	}); err != nil {
		db.Close()
		return nil, fmt.Errorf("bootstrap initial user: %w", err)
	}

	s3Client, err := storage.NewS3(ctx, storage.Config{
		Endpoint:        cfg.S3Endpoint,
		AccessKeyID:     cfg.S3AccessKeyID,
		SecretAccessKey: cfg.S3SecretAccessKey,
		Bucket:          cfg.S3Bucket,
		UseSSL:          cfg.S3UseSSL,
	})
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("init s3 client: %w", err)
	}

	specialistsService := specialists.NewService(specialists.NewRepository(queries))
	examinationsService := examinations.NewService(examinations.NewRepository(db.Pool()), auditService)
	processingRepository := processing.NewRepository(db.Pool(), cfg.S3Bucket)
	processingService := processing.NewService(processingRepository, auditService)
	baselineService := baselineclient.New(cfg.BaselineBaseURL, cfg.BaselineTimeout)
	settingsService := settings.NewService(settings.NewRepository(db.Pool()), settings.Defaults{
		AudioRetentionTTLDays: cfg.AudioRetentionTTLDays,
		ProcessingMaxAttempts: cfg.OutboxMaxAttempts,
		KESMIMaxRetries:       cfg.KESMIMaxRetries,
	}, auditService)
	if _, err := settingsService.EnsureDefaults(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("ensure system settings: %w", err)
	}
	processingRepository = processing.NewRepository(db.Pool(), cfg.S3Bucket, settingsService)
	decisionRepository := decision.NewRepository(db.Pool())
	kesmiClient := kesmi.NewClient(cfg.KESMIBaseURL, cfg.KESMIModelID, cfg.KESMITimeout)
	decisionService := decision.NewService(decisionRepository, kesmiClient, decision.Config{
		MaxAttempts: cfg.KESMIMaxRetries,
	}, auditService).WithMaxAttemptsProvider(settingsService)
	decisionRelay := decision.NewRelay(decisionRepository, decisionService, cfg.KESMIRetryBackoff)
	aggregationRepository := aggregation.NewRepository(db.Pool())
	aggregationService := aggregation.NewService(
		aggregationRepository,
		baselineService,
		decisionService,
		cfg.BaselineReference,
		cfg.BaselineAlgorithm,
	)
	channelResultsRepository := channelresults.NewRepository(db.Pool())
	channelResultsHandler := channelresults.NewHandler(channelResultsRepository, aggregationService, auditService)
	processingRelay := processing.NewRelay(processingRepository, processing.RelayConfig{
		BrokerURL:    cfg.RabbitMQURL,
		PollInterval: cfg.OutboxPollInterval,
		MaxAttempts:  cfg.OutboxMaxAttempts,
	})
	resultsConsumer := processing.NewResultsConsumer(cfg.RabbitMQURL, channelResultsHandler)
	questionnairesService := questionnaires.NewService(questionnaires.NewRepository(db.Pool()), auditService)
	answersService := answers.NewService(answers.NewRepository(queries), s3Client)
	resultsService := results.NewService(results.NewRepository(db.Pool()))
	logger := observability.NewLogger(cfg.LogLevel, os.Stdout)
	metrics := observability.NewMetricsRegistry()

	handler := httpserver.NewRouter(httpserver.Dependencies{
		DB:             db,
		AuthService:    authService,
		AuthTokens:     tokenManager,
		Specialists:    specialistsService,
		Examinations:   examinationsService,
		Processing:     processingService,
		Results:        resultsService,
		Questionnaires: questionnairesService,
		Answers:        answersService,
		Audit:          auditService,
		Settings:       settingsService,
		AllowedOrigins: cfg.AllowedOrigins,
		MaxUploadSize:  cfg.MaxUploadSizeBytes,
		Logger:         logger,
		Metrics:        metrics,
		ReadinessChecks: map[string]func(context.Context) error{
			"postgres": func(ctx context.Context) error {
				return db.Ping(ctx)
			},
			"rabbitmq": func(ctx context.Context) error {
				return dialTCP(ctx, cfg.RabbitMQURL, "5672")
			},
			"minio": func(ctx context.Context) error {
				return httpProbe(ctx, strings.TrimRight(cfg.S3Endpoint, "/")+"/minio/health/live")
			},
			"baseline": func(ctx context.Context) error {
				return httpProbe(ctx, strings.TrimRight(cfg.BaselineBaseURL, "/")+"/health")
			},
			"kesmi": func(ctx context.Context) error {
				return kesmiClient.Models(ctx)
			},
		},
	})

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	return &App{
		server:          server,
		db:              db,
		publisher:       processingRelay,
		resultsConsumer: resultsConsumer,
		decisionRelay:   decisionRelay,
		logger:          logger,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	errCh := make(chan error, 1)

	go func() {
		errCh <- a.server.ListenAndServe()
	}()
	if a.publisher != nil {
		go a.publisher.Run(ctx)
	}
	if a.resultsConsumer != nil {
		go a.resultsConsumer.Run(ctx)
	}
	if a.decisionRelay != nil {
		go a.decisionRelay.Run(ctx)
	}

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		return a.server.Shutdown(shutdownCtx)
	case err := <-errCh:
		if err != nil && err != http.ErrServerClosed && a.logger != nil {
			a.logger.Error("http_server_stopped", slog.String("error", err.Error()))
		}
		return err
	}
}

func (a *App) Close() {
	if a.db != nil {
		a.db.Close()
	}
}

func dialTCP(ctx context.Context, rawURL, fallbackPort string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Host == "" {
		host := rawURL
		if strings.Contains(host, "@") {
			parts := strings.SplitN(host, "@", 2)
			host = parts[1]
		}
		if !strings.Contains(host, ":") {
			host = net.JoinHostPort(host, fallbackPort)
		}
		conn, dialErr := (&net.Dialer{}).DialContext(ctx, "tcp", host)
		if dialErr != nil {
			return dialErr
		}
		return conn.Close()
	}
	host := parsed.Host
	if !strings.Contains(host, ":") {
		host = net.JoinHostPort(host, fallbackPort)
	}
	conn, err := (&net.Dialer{}).DialContext(ctx, "tcp", host)
	if err != nil {
		return err
	}
	return conn.Close()
}

func httpProbe(ctx context.Context, target string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("status %d", resp.StatusCode)
	}
	return nil
}
