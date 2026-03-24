package http

import (
	"mime/multipart"
	"net/http"
	"strconv"

	"diplom/internal/answers"
)

type AnswersHandler struct {
	service       *answers.Service
	maxUploadSize int64
}

func (h AnswersHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims, ok := ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, h.maxUploadSize)
	if err := r.ParseMultipartForm(h.maxUploadSize); err != nil {
		writeError(w, http.StatusBadRequest, "invalid multipart form")
		return
	}

	examinationID, err := strconv.ParseInt(r.FormValue("examination_id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid examination_id")
		return
	}
	examinationQuestionID, err := strconv.ParseInt(r.FormValue("examination_question_id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid examination_question_id")
		return
	}
	specialistID, err := strconv.ParseInt(r.FormValue("specialist_id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid specialist_id")
		return
	}

	text := r.FormValue("text")
	file, header, err := r.FormFile("audio")
	if err != nil {
		writeError(w, http.StatusBadRequest, "audio file is required")
		return
	}
	defer file.Close()

	item, err := h.service.Create(r.Context(), answers.CreateInput{
		ExaminationID:         examinationID,
		ExaminationQuestionID: examinationQuestionID,
		SpecialistID:          specialistID,
		CreatedByUserID:       claims.UserID,
		Text:                  text,
		FileName:              header.Filename,
		ContentType:           fileContentType(header),
		Size:                  header.Size,
		Content:               file,
	})
	if err != nil {
		status, message := mapDomainError(err)
		writeError(w, status, message)
		return
	}

	writeJSON(w, http.StatusCreated, item)
}

func fileContentType(header *multipart.FileHeader) string {
	if header.Header.Get("Content-Type") == "" {
		return "application/octet-stream"
	}
	return header.Header.Get("Content-Type")
}
