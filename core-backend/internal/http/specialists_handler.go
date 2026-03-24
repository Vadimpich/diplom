package http

import (
	"net/http"

	"diplom/internal/specialists"
)

type SpecialistsHandler struct {
	service *specialists.Service
}

type specialistRequest struct {
	FullName        string  `json:"full_name"`
	PersonnelNumber *string `json:"personnel_number"`
}

func (h SpecialistsHandler) Create(w http.ResponseWriter, r *http.Request) {
	var request specialistRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	item, err := h.service.Create(r.Context(), specialists.CreateInput{
		FullName:        request.FullName,
		PersonnelNumber: request.PersonnelNumber,
	})
	if err != nil {
		status, message := mapDomainError(err)
		writeError(w, status, message)
		return
	}

	writeJSON(w, http.StatusCreated, item)
}

func (h SpecialistsHandler) List(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.List(r.Context())
	if err != nil {
		status, message := mapDomainError(err)
		writeError(w, status, message)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h SpecialistsHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := parseInt64Param(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid specialist id")
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

func (h SpecialistsHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := parseInt64Param(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid specialist id")
		return
	}

	var request specialistRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	item, err := h.service.Update(r.Context(), specialists.UpdateInput{
		ID:              id,
		FullName:        request.FullName,
		PersonnelNumber: request.PersonnelNumber,
	})
	if err != nil {
		status, message := mapDomainError(err)
		writeError(w, status, message)
		return
	}

	writeJSON(w, http.StatusOK, item)
}

func (h SpecialistsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseInt64Param(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid specialist id")
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		status, message := mapDomainError(err)
		writeError(w, status, message)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
