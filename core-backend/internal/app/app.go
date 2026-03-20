package app

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"dimplom/internal/answers"
	"dimplom/internal/auth"
	"dimplom/internal/config"
	"dimplom/internal/examinations"
	httpserver "dimplom/internal/http"
	appmigrations "dimplom/internal/migrations"
	"dimplom/internal/postgres"
	"dimplom/internal/questionnaires"
	"dimplom/internal/specialists"
	"dimplom/internal/storage"
)

type App struct {
	server *http.Server
	db     *postgres.Client
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

	tokenManager := auth.NewJWTManager(cfg.JWTIssuer, cfg.JWTAccessSecret, cfg.JWTAccessTTL)
	authRepository := auth.NewRepository(db.Pool())
	authService := auth.NewService(authRepository, tokenManager)

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
	examinationsService := examinations.NewService(examinations.NewRepository(db.Pool()))
	questionnairesService := questionnaires.NewService(questionnaires.NewRepository(db.Pool()))
	answersService := answers.NewService(answers.NewRepository(queries), s3Client)

	handler := httpserver.NewRouter(httpserver.Dependencies{
		DB:             db,
		AuthService:    authService,
		AuthTokens:     tokenManager,
		Specialists:    specialistsService,
		Examinations:   examinationsService,
		Questionnaires: questionnairesService,
		Answers:        answersService,
		AllowedOrigins: cfg.AllowedOrigins,
		MaxUploadSize:  cfg.MaxUploadSizeBytes,
	})

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	return &App{
		server: server,
		db:     db,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	errCh := make(chan error, 1)

	go func() {
		errCh <- a.server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		return a.server.Shutdown(shutdownCtx)
	case err := <-errCh:
		return err
	}
}

func (a *App) Close() {
	if a.db != nil {
		a.db.Close()
	}
}
