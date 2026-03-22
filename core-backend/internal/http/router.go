package http

import (
	"log"
	nethttp "net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"dimplom/internal/answers"
	"dimplom/internal/auth"
	"dimplom/internal/examinations"
	"dimplom/internal/postgres"
	"dimplom/internal/processing"
	"dimplom/internal/questionnaires"
	"dimplom/internal/results"
	"dimplom/internal/specialists"
)

type Dependencies struct {
	DB             *postgres.Client
	AuthService    *auth.Service
	AuthTokens     auth.TokenManager
	Specialists    *specialists.Service
	Examinations   *examinations.Service
	Processing     *processing.Service
	Results        *results.Service
	Questionnaires *questionnaires.Service
	Answers        *answers.Service
	AllowedOrigins []string
	MaxUploadSize  int64
}

func NewRouter(deps Dependencies) nethttp.Handler {
	router := chi.NewRouter()

	router.Use(CORSMiddleware(CORSConfig{
		AllowedOrigins: deps.AllowedOrigins,
	}))
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(15 * time.Second))
	router.Use(requestLogger)

	handler := Handler{
		db: deps.DB,
	}
	authHandler := AuthHandler{service: deps.AuthService}
	specialistsHandler := SpecialistsHandler{service: deps.Specialists}
	examinationsHandler := ExaminationsHandler{service: deps.Examinations}
	if deps.Processing != nil {
		examinationsHandler.finisher = deps.Processing
		examinationsHandler.statusProvider = deps.Processing
	}
	questionnairesHandler := QuestionnairesHandler{service: deps.Questionnaires}
	resultsHandler := ResultsHandler{}
	if deps.Results != nil {
		resultsHandler.service = deps.Results
	}
	answersHandler := AnswersHandler{
		service:       deps.Answers,
		maxUploadSize: deps.MaxUploadSize,
	}

	router.Get("/health", handler.Health)
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
			admin.Get("/questionnaires", questionnairesHandler.List)
			admin.Post("/questionnaires", questionnairesHandler.Create)
			admin.Get("/questionnaires/{id}", questionnairesHandler.GetByID)
			admin.Put("/questionnaires/{id}", questionnairesHandler.Update)
		})

		private.Group(func(operatorOrAdmin chi.Router) {
			operatorOrAdmin.Use(RequireRoles("operator", "admin"))
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

func requestLogger(next nethttp.Handler) nethttp.Handler {
	return nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("http request method=%s path=%s duration=%s", r.Method, r.URL.Path, time.Since(start))
	})
}
