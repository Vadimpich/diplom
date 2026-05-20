package http

import (
	"net/http"

	"diplom/internal/audit"
	"diplom/internal/auth"
)

type AuthHandler struct {
	service *auth.Service
}

type loginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type createUserRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type updateUserRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	Role     string `json:"role"`
	IsActive bool   `json:"is_active"`
}

type usersResponse struct {
	Items []auth.User `json:"items"`
}

func (h AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var request loginRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	ctx := audit.WithMetadata(r.Context(), audit.Metadata{
		Actor: audit.Actor{
			Login:     request.Login,
			IP:        clientIP(r),
			UserAgent: r.UserAgent(),
		},
	})

	result, err := h.service.Login(ctx, auth.LoginInput{
		Login:     request.Login,
		Password:  request.Password,
		IP:        clientIP(r),
		UserAgent: r.UserAgent(),
	})
	if err != nil {
		status, message := mapDomainError(err)
		writeError(w, status, message)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (h AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var request refreshRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.service.Refresh(r.Context(), auth.RefreshInput{
		RefreshToken: request.RefreshToken,
		IP:           clientIP(r),
		UserAgent:    r.UserAgent(),
	})
	if err != nil {
		status, message := mapDomainError(err)
		writeError(w, status, message)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (h AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var request refreshRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.service.Logout(r.Context(), auth.LogoutInput{
		RefreshToken: request.RefreshToken,
	}); err != nil {
		status, message := mapDomainError(err)
		writeError(w, status, message)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	claims, ok := ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	user, err := h.service.Me(r.Context(), claims.UserID)
	if err != nil {
		status, message := mapDomainError(err)
		writeError(w, status, message)
		return
	}

	writeJSON(w, http.StatusOK, user)
}

func (h AuthHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var request createUserRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := h.service.CreateUser(r.Context(), auth.CreateUserInput{
		Login:    request.Login,
		Password: request.Password,
		RoleSlug: request.Role,
	})
	if err != nil {
		status, message := mapDomainError(err)
		writeError(w, status, message)
		return
	}

	writeJSON(w, http.StatusCreated, user)
}

func (h AuthHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.ListUsers(r.Context())
	if err != nil {
		status, message := mapDomainError(err)
		writeError(w, status, message)
		return
	}

	writeJSON(w, http.StatusOK, usersResponse{Items: items})
}

func (h AuthHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	id, err := parseInt64Param(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	user, err := h.service.GetUserByID(r.Context(), id)
	if err != nil {
		status, message := mapDomainError(err)
		writeError(w, status, message)
		return
	}

	writeJSON(w, http.StatusOK, user)
}

func (h AuthHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id, err := parseInt64Param(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	var request updateUserRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := h.service.UpdateUser(r.Context(), auth.UpdateUserInput{
		ID:       id,
		Login:    request.Login,
		Password: request.Password,
		RoleSlug: request.Role,
		IsActive: request.IsActive,
	})
	if err != nil {
		status, message := mapDomainError(err)
		writeError(w, status, message)
		return
	}

	writeJSON(w, http.StatusOK, user)
}
