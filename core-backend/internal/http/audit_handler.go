package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"diplom/internal/audit"
)

const adminAuditTimeLayout = time.RFC3339

type AuditHandler struct {
	service *audit.Service
}

type auditEventsResponse struct {
	Items []auditEventResponse `json:"items"`
}

type auditActorResponse struct {
	UserID    *int64  `json:"user_id"`
	Login     string  `json:"login"`
	RoleSlug  string  `json:"role_slug"`
	IP        string  `json:"ip"`
	UserAgent string  `json:"user_agent"`
}

type auditResourceResponse struct {
	Kind string `json:"kind"`
	ID   int64  `json:"id"`
}

type auditDomainRefsResponse struct {
	ExaminationID      *int64  `json:"examination_id"`
	SpecialistID       *int64  `json:"specialist_id"`
	QuestionnaireID    *int64  `json:"questionnaire_id"`
	Channel            *string `json:"channel"`
	DecisionSnapshotID *int64  `json:"decision_snapshot_id"`
}

type auditEventResponse struct {
	ID            int64                   `json:"id"`
	EventType     string                  `json:"event_type"`
	EventKey      string                  `json:"event_key"`
	Outcome       string                  `json:"outcome"`
	HappenedAt    time.Time               `json:"happened_at"`
	RequestID     string                  `json:"request_id"`
	TraceID       string                  `json:"trace_id"`
	TraceParent   string                  `json:"traceparent"`
	TraceState    string                  `json:"tracestate"`
	CorrelationID string                  `json:"correlation_id"`
	Actor         *auditActorResponse     `json:"actor"`
	Resource      *auditResourceResponse  `json:"resource"`
	DomainRefs    auditDomainRefsResponse `json:"domain_refs"`
	Payload       json.RawMessage         `json:"payload"`
}

func (h AuditHandler) List(w http.ResponseWriter, r *http.Request) {
	filter, err := parseAuditListFilter(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	items, err := h.service.List(r.Context(), filter)
	if err != nil {
		status, message := mapDomainError(err)
		writeError(w, status, message)
		return
	}

	writeJSON(w, http.StatusOK, auditEventsResponse{Items: mapAuditEvents(items)})
}

func mapAuditEvents(items []audit.Event) []auditEventResponse {
	result := make([]auditEventResponse, 0, len(items))
	for _, item := range items {
		result = append(result, mapAuditEvent(item))
	}
	return result
}

func mapAuditEvent(item audit.Event) auditEventResponse {
	response := auditEventResponse{
		ID:            item.ID,
		EventType:     item.Type,
		EventKey:      item.Key,
		Outcome:       item.Outcome,
		HappenedAt:    item.HappenedAt,
		RequestID:     item.RequestID,
		TraceID:       item.TraceID,
		TraceParent:   item.TraceParent,
		TraceState:    item.TraceState,
		CorrelationID: item.CorrelationID,
		DomainRefs: auditDomainRefsResponse{
			ExaminationID:      item.DomainRefs.ExaminationID,
			SpecialistID:       item.DomainRefs.SpecialistID,
			QuestionnaireID:    item.DomainRefs.QuestionnaireID,
			Channel:            item.DomainRefs.Channel,
			DecisionSnapshotID: item.DomainRefs.DecisionSnapshotID,
		},
		Payload: item.Payload,
	}

	if item.Actor.UserID != nil || item.Actor.Login != "" || item.Actor.RoleSlug != "" || item.Actor.IP != "" || item.Actor.UserAgent != "" {
		response.Actor = &auditActorResponse{
			UserID:    item.Actor.UserID,
			Login:     item.Actor.Login,
			RoleSlug:  item.Actor.RoleSlug,
			IP:        item.Actor.IP,
			UserAgent: item.Actor.UserAgent,
		}
	}

	if item.Resource.Kind != "" || item.Resource.ID != 0 {
		response.Resource = &auditResourceResponse{
			Kind: item.Resource.Kind,
			ID:   item.Resource.ID,
		}
	}

	return response
}

func parseAuditListFilter(r *http.Request) (audit.ListFilter, error) {
	query := r.URL.Query()
	filter := audit.ListFilter{}

	if value := strings.TrimSpace(query.Get("event_type")); value != "" {
		filter.EventType = &value
	}
	if value := strings.TrimSpace(query.Get("resource_kind")); value != "" {
		filter.ResourceKind = &value
	}
	if value := strings.TrimSpace(query.Get("resource_id")); value != "" {
		resourceID, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return audit.ListFilter{}, errInvalid("invalid resource_id")
		}
		filter.ResourceID = &resourceID
	}
	if filter.ResourceID != nil && filter.ResourceKind == nil {
		return audit.ListFilter{}, errInvalid("resource_kind is required when resource_id is provided")
	}
	if value := strings.TrimSpace(query.Get("from")); value != "" {
		fromAt, err := time.Parse(adminAuditTimeLayout, value)
		if err != nil {
			return audit.ListFilter{}, errInvalid("invalid from")
		}
		filter.From = fromAt
	}
	if value := strings.TrimSpace(query.Get("to")); value != "" {
		toAt, err := time.Parse(adminAuditTimeLayout, value)
		if err != nil {
			return audit.ListFilter{}, errInvalid("invalid to")
		}
		filter.To = toAt
	}
	if !filter.From.IsZero() && !filter.To.IsZero() && filter.From.After(filter.To) {
		return audit.ListFilter{}, errInvalid("from must be before or equal to to")
	}
	if value := strings.TrimSpace(query.Get("limit")); value != "" {
		limit, err := strconv.ParseInt(value, 10, 32)
		if err != nil || limit <= 0 || limit > 200 {
			return audit.ListFilter{}, errInvalid("invalid limit")
		}
		filter.Limit = int32(limit)
	}

	return filter, nil
}

type invalidQueryError struct {
	message string
}

func (e invalidQueryError) Error() string {
	return e.message
}

func errInvalid(message string) error {
	return invalidQueryError{message: message}
}
