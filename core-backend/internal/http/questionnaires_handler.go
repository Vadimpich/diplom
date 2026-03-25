package http

import (
	"net/http"

	"diplom/internal/questionnaires"
)

type QuestionnairesHandler struct {
	service *questionnaires.Service
}

type questionnaireQuestionRequest struct {
	Text string `json:"text"`
}

type questionnaireRequest struct {
	Title       string                         `json:"title"`
	Description *string                        `json:"description"`
	IsActive    bool                           `json:"is_active"`
	Questions   []questionnaireQuestionRequest `json:"questions"`
}

type questionnairesResponse struct {
	Items []questionnaires.Questionnaire `json:"items"`
}

func (h QuestionnairesHandler) List(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.List(r.Context())
	if err != nil {
		status, message := mapDomainError(err)
		writeError(w, status, message)
		return
	}

	writeJSON(w, http.StatusOK, questionnairesResponse{Items: items})
}

func (h QuestionnairesHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims, ok := ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	var request questionnaireRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	item, err := h.service.Create(r.Context(), questionnaires.CreateInput{
		Title:        request.Title,
		Description:  request.Description,
		IsActive:     request.IsActive,
		Questions:    mapQuestionInputs(request.Questions),
		EditorUserID: claims.UserID,
	})
	if err != nil {
		status, message := mapDomainError(err)
		writeError(w, status, message)
		return
	}

	writeJSON(w, http.StatusCreated, item)
}

func (h QuestionnairesHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := parseInt64Param(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid questionnaire id")
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

func (h QuestionnairesHandler) Update(w http.ResponseWriter, r *http.Request) {
	claims, ok := ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	id, err := parseInt64Param(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid questionnaire id")
		return
	}

	var request questionnaireRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	item, err := h.service.Update(r.Context(), questionnaires.UpdateInput{
		ID:           id,
		Title:        request.Title,
		Description:  request.Description,
		IsActive:     request.IsActive,
		Questions:    mapQuestionInputs(request.Questions),
		EditorUserID: claims.UserID,
	})
	if err != nil {
		status, message := mapDomainError(err)
		writeError(w, status, message)
		return
	}

	writeJSON(w, http.StatusOK, item)
}

func mapQuestionInputs(items []questionnaireQuestionRequest) []questionnaires.QuestionInput {
	result := make([]questionnaires.QuestionInput, 0, len(items))
	for _, item := range items {
		result = append(result, questionnaires.QuestionInput{Text: item.Text})
	}
	return result
}
