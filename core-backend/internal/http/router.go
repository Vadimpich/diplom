package http

import (
	"context"
	"log/slog"
	nethttp "net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"diplom/internal/answers"
	"diplom/internal/audit"
	"diplom/internal/auth"
	"diplom/internal/examinations"
	"diplom/internal/observability"
	"diplom/internal/postgres"
	"diplom/internal/processing"
	"diplom/internal/questionnaires"
	"diplom/internal/results"
	"diplom/internal/settings"
	"diplom/internal/specialists"
)

type Dependencies struct {
	DB              *postgres.Client
	AuthService     *auth.Service
	AuthTokens      auth.TokenManager
	Specialists     *specialists.Service
	Examinations    *examinations.Service
	Processing      *processing.Service
	Results         *results.Service
	Questionnaires  *questionnaires.Service
	Answers         *answers.Service
	Audit           *audit.Service
	Settings        *settings.Service
	AllowedOrigins  []string
	MaxUploadSize   int64
	Logger          *slog.Logger
	ReadinessChecks map[string]func(context.Context) error
	Metrics         *observability.MetricsRegistry
}

func NewRouter(deps Dependencies) nethttp.Handler {
	if deps.Metrics == nil {
		deps.Metrics = observability.NewMetricsRegistry()
	}
	router := chi.NewRouter()

	router.Use(CORSMiddleware(CORSConfig{
		AllowedOrigins: deps.AllowedOrigins,
	}))
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(15 * time.Second))
	router.Use(observability.HTTPMiddlewareWithMetrics(deps.Logger, deps.Metrics))

	handler := Handler{
		db:              deps.DB,
		readinessChecks: deps.ReadinessChecks,
		metrics:         deps.Metrics,
	}
	authHandler := AuthHandler{service: deps.AuthService}
	specialistsHandler := SpecialistsHandler{service: deps.Specialists}
	examinationsHandler := ExaminationsHandler{service: deps.Examinations}
	if deps.Processing != nil {
		examinationsHandler.finisher = deps.Processing
		examinationsHandler.statusProvider = deps.Processing
	}
	questionnairesHandler := QuestionnairesHandler{service: deps.Questionnaires}
	auditHandler := AuditHandler{service: deps.Audit}
	settingsHandler := SettingsHandler{service: deps.Settings}
	resultsHandler := ResultsHandler{}
	if deps.Results != nil {
		resultsHandler.service = deps.Results
	}
	answersHandler := AnswersHandler{
		service:       deps.Answers,
		maxUploadSize: deps.MaxUploadSize,
	}

	router.Get("/health", handler.Health)
	router.Get("/ready", handler.Ready)
	router.Get("/metrics", handler.Metrics)
	router.Post("/auth/login", authHandler.Login)
	router.Post("/auth/refresh", authHandler.Refresh)
	router.Post("/auth/logout", authHandler.Logout)

	router.Group(func(private chi.Router) {
		private.Use(AuthMiddleware(deps.AuthTokens))

		private.Get("/me", authHandler.Me)

		private.Group(func(admin chi.Router) {
			admin.Use(RequireRoles("admin"))
			admin.Get("/users", authHandler.ListUsers)
			admin.Get("/users/{id}", authHandler.GetUserByID)
			admin.Post("/users", authHandler.CreateUser)
			admin.Put("/users/{id}", authHandler.UpdateUser)
			admin.Get("/audit/events", auditHandler.List)
			admin.Get("/settings", settingsHandler.Get)
			admin.Put("/settings", settingsHandler.Update)
			admin.Post("/questionnaires", questionnairesHandler.Create)
			admin.Get("/questionnaires/{id}", questionnairesHandler.GetByID)
			admin.Put("/questionnaires/{id}", questionnairesHandler.Update)
		})

		private.Group(func(operatorOrAdmin chi.Router) {
			operatorOrAdmin.Use(RequireRoles("operator", "admin"))
			operatorOrAdmin.Get("/questionnaires", questionnairesHandler.List)
			operatorOrAdmin.Post("/specialists", specialistsHandler.Create)
			operatorOrAdmin.Get("/specialists", specialistsHandler.List)
			operatorOrAdmin.Get("/specialists/{id}", specialistsHandler.GetByID)
			operatorOrAdmin.Put("/specialists/{id}", specialistsHandler.Update)
			operatorOrAdmin.Delete("/specialists/{id}", specialistsHandler.Delete)
			operatorOrAdmin.Get("/specialists/{id}/examinations", examinationsHandler.ListBySpecialist)
			operatorOrAdmin.Get("/specialists/{id}/result-history", resultsHandler.GetSpecialistHistory)
			operatorOrAdmin.Get("/examinations", examinationsHandler.List)
			operatorOrAdmin.Post("/examinations", examinationsHandler.Create)
			operatorOrAdmin.Get("/examinations/{id}", examinationsHandler.GetByID)
			operatorOrAdmin.Get("/examinations/{id}/processing-status", examinationsHandler.ProcessingStatus)
			operatorOrAdmin.Get("/examinations/{id}/result", resultsHandler.GetExaminationResult)
			operatorOrAdmin.Post("/examinations/{id}/start", examinationsHandler.Start)
			operatorOrAdmin.Post("/examinations/{id}/finish", examinationsHandler.Finish)
			operatorOrAdmin.Post("/answers", answersHandler.Create)
		})
	})

	return router
}
