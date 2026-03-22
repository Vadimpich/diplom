package http

import (
	"context"
	"net/http"
	"time"

	"dimplom/internal/examinations"
	"dimplom/internal/processing"
)

type ExaminationsHandler struct {
	service        *examinations.Service
	finisher       finisher
	statusProvider statusProvider
}

type finisher interface {
	Finish(context.Context, int64) (examinations.Examination, error)
}

type statusProvider interface {
	GetStatus(context.Context, int64) (processing.ProcessingStatusResponse, error)
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

	processor := h.finisher
	if processor == nil {
		processor = h.service
	}

	item, err := processor.Finish(r.Context(), id)
	if err != nil {
		status, message := mapDomainError(err)
		writeError(w, status, message)
		return
	}

	writeJSON(w, http.StatusOK, item)
}

func (h ExaminationsHandler) ProcessingStatus(w http.ResponseWriter, r *http.Request) {
	id, err := parseInt64Param(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid examination id")
		return
	}

	if h.statusProvider == nil {
		writeJSON(w, http.StatusOK, processing.ProcessingStatusResponse{
			ExaminationID:  id,
			Status:         examinations.StatusAggregating,
			MessageVersion: processing.MessageVersionV1,
			ChannelsTotal:  len(processing.MandatoryChannels),
			ChannelsComplete: len(processing.MandatoryChannels),
			UpdatedAt:      time.Now().UTC(),
			Terminal:       false,
			Channels:       []processing.ChannelStatusDTO{},
		})
		return
	}

	status, err := h.statusProvider.GetStatus(r.Context(), id)
	if err != nil {
		code, message := mapDomainError(err)
		writeError(w, code, message)
		return
	}
	writeJSON(w, http.StatusOK, status)
}
