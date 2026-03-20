package http

import (
	"net/http"

	"dimplom/internal/examinations"
)

type ExaminationsHandler struct {
	service *examinations.Service
}

type createExaminationRequest struct {
	SpecialistID    int64  `json:"specialist_id"`
	QuestionnaireID *int64 `json:"questionnaire_id"`
}

type examinationsResponse struct {
	Items []examinations.Examination `json:"items"`
}

func (h ExaminationsHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims, ok := ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	var request createExaminationRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	item, err := h.service.Create(r.Context(), examinations.CreateInput{
		SpecialistID:    request.SpecialistID,
		CreatedByUserID: claims.UserID,
		QuestionnaireID: request.QuestionnaireID,
	})
	if err != nil {
		status, message := mapDomainError(err)
		writeError(w, status, message)
		return
	}

	writeJSON(w, http.StatusCreated, item)
}

func (h ExaminationsHandler) List(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.List(r.Context())
	if err != nil {
		status, message := mapDomainError(err)
		writeError(w, status, message)
		return
	}

	writeJSON(w, http.StatusOK, examinationsResponse{Items: items})
}

func (h ExaminationsHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := parseInt64Param(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid examination id")
		return
	}

	item, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		status, message := mapDomainError(err)
		writeError(w, status, message)
		return
	}

	writeJSON(w, http.StatusOK, item)
}

func (h ExaminationsHandler) ListBySpecialist(w http.ResponseWriter, r *http.Request) {
	specialistID, err := parseInt64Param(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid specialist id")
		return
	}

	items, err := h.service.ListBySpecialistID(r.Context(), specialistID)
	if err != nil {
		status, message := mapDomainError(err)
		writeError(w, status, message)
		return
	}

	writeJSON(w, http.StatusOK, examinationsResponse{Items: items})
}

func (h ExaminationsHandler) Start(w http.ResponseWriter, r *http.Request) {
	id, err := parseInt64Param(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid examination id")
		return
	}

	item, err := h.service.Start(r.Context(), id)
	if err != nil {
		status, message := mapDomainError(err)
		writeError(w, status, message)
		return
	}

	writeJSON(w, http.StatusOK, item)
}

func (h ExaminationsHandler) Finish(w http.ResponseWriter, r *http.Request) {
	id, err := parseInt64Param(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid examination id")
		return
	}

	item, err := h.service.Finish(r.Context(), id)
	if err != nil {
		status, message := mapDomainError(err)
		writeError(w, status, message)
		return
	}

	writeJSON(w, http.StatusOK, item)
}
