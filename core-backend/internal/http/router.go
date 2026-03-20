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
	"dimplom/internal/questionnaires"
	"dimplom/internal/specialists"
)

type Dependencies struct {
	DB             *postgres.Client
	AuthService    *auth.Service
	AuthTokens     auth.TokenManager
	Specialists    *specialists.Service
	Examinations   *examinations.Service
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
	questionnairesHandler := QuestionnairesHandler{service: deps.Questionnaires}
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
		private.Get("/users", authHandler.ListUsers)
		private.Get("/users/{id}", authHandler.GetUserByID)
		private.Post("/users", authHandler.CreateUser)
		private.Put("/users/{id}", authHandler.UpdateUser)
		private.Post("/specialists", specialistsHandler.Create)
		private.Get("/specialists", specialistsHandler.List)
		private.Get("/specialists/{id}", specialistsHandler.GetByID)
		private.Put("/specialists/{id}", specialistsHandler.Update)
		private.Delete("/specialists/{id}", specialistsHandler.Delete)
		private.Get("/specialists/{id}/examinations", examinationsHandler.ListBySpecialist)
		private.Get("/examinations", examinationsHandler.List)
		private.Post("/examinations", examinationsHandler.Create)
		private.Get("/examinations/{id}", examinationsHandler.GetByID)
		private.Post("/examinations/{id}/start", examinationsHandler.Start)
		private.Post("/examinations/{id}/finish", examinationsHandler.Finish)
		private.Get("/questionnaires", questionnairesHandler.List)
		private.Post("/questionnaires", questionnairesHandler.Create)
		private.Get("/questionnaires/{id}", questionnairesHandler.GetByID)
		private.Put("/questionnaires/{id}", questionnairesHandler.Update)
		private.Post("/answers", answersHandler.Create)
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
