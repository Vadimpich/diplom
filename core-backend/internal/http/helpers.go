package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"dimplom/internal/answers"
	"dimplom/internal/auth"
	"dimplom/internal/examinations"
	"dimplom/internal/questionnaires"
	"dimplom/internal/repository"
	"dimplom/internal/specialists"
)

type errorResponse struct {
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, errorResponse{Error: message})
}

func decodeJSON(r *http.Request, dest any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(dest)
}

func parseInt64Param(r *http.Request, key string) (int64, error) {
	return strconv.ParseInt(chi.URLParam(r, key), 10, 64)
}

func mapDomainError(err error) (int, string) {
	switch {
	case errors.Is(err, auth.ErrInvalidCredentials):
		return http.StatusUnauthorized, "invalid credentials"
	case errors.Is(err, auth.ErrInactiveUser):
		return http.StatusForbidden, "user is inactive"
	case errors.Is(err, auth.ErrInvalidInput):
		return http.StatusBadRequest, "invalid auth payload"
	case errors.Is(err, auth.ErrInvalidToken):
		return http.StatusUnauthorized, "invalid token"
	case errors.Is(err, auth.ErrInvalidSession):
		return http.StatusUnauthorized, "invalid refresh session"
	case errors.Is(err, specialists.ErrInvalidInput):
		return http.StatusBadRequest, "invalid specialist payload"
	case errors.Is(err, examinations.ErrInvalidInput):
		return http.StatusBadRequest, "invalid examination payload"
	case errors.Is(err, examinations.ErrInvalidTransition):
		return http.StatusConflict, "invalid examination status transition"
	case errors.Is(err, examinations.ErrAnswersIncomplete):
		return http.StatusConflict, "examination answers are incomplete"
	case errors.Is(err, questionnaires.ErrInvalidInput):
		return http.StatusBadRequest, "invalid questionnaire payload"
	case errors.Is(err, answers.ErrInvalidInput):
		return http.StatusBadRequest, "invalid answer payload"
	case errors.Is(err, answers.ErrExaminationNotReady):
		return http.StatusConflict, "examination is not collecting answers"
	case errors.Is(err, repository.ErrConflict):
		return http.StatusConflict, "resource conflict"
	case errors.Is(err, repository.ErrNotFound):
		return http.StatusNotFound, "resource not found"
	default:
		return http.StatusInternalServerError, "internal server error"
	}
}

func clientIP(r *http.Request) string {
	for _, header := range []string{"X-Forwarded-For", "X-Real-IP"} {
		value := strings.TrimSpace(r.Header.Get(header))
		if value == "" {
			continue
		}
		if header == "X-Forwarded-For" {
			value = strings.TrimSpace(strings.Split(value, ",")[0])
		}
		if value != "" {
			return value
		}
	}
	return r.RemoteAddr
}
