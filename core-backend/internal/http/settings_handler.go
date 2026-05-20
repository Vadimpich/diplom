package http

import (
	"net/http"

	"diplom/internal/settings"
)

type SettingsHandler struct {
	service *settings.Service
}

func (h SettingsHandler) Get(w http.ResponseWriter, r *http.Request) {
	item, err := h.service.Get(r.Context())
	if err != nil {
		status, message := mapDomainError(err)
		writeError(w, status, message)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h SettingsHandler) Update(w http.ResponseWriter, r *http.Request) {
	var input settings.UpdateInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	item, err := h.service.Update(r.Context(), input)
	if err != nil {
		status, message := mapDomainError(err)
		writeError(w, status, message)
		return
	}
	writeJSON(w, http.StatusOK, item)
}
