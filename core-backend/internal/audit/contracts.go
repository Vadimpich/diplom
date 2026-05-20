package audit

import (
	"context"
	"encoding/json"
	"time"
)

const (
	OutcomeSucceeded = "succeeded"
	OutcomeFailed    = "failed"
	OutcomeRejected  = "rejected"

	EventTypeAuthLogin            = "auth.login"
	EventTypeAuthLoginFailed      = "auth.login_failed"
	EventTypeAdminUserCreated     = "admin.user_created"
	EventTypeAdminUserUpdated     = "admin.user_updated"
	EventTypeAdminSettingsUpdated = "admin.settings_updated"
	EventTypeQuestionnaireCreated = "admin.questionnaire_created"
	EventTypeQuestionnaireUpdated = "admin.questionnaire_updated"
	EventTypeExaminationCreated   = "examination.created"
	EventTypeExaminationStarted   = "examination.started"
	EventTypeExaminationFinished  = "examination.finished"
	EventTypeProcessingLaunch     = "processing.launch"
	EventTypeResultReceived       = "processing.result_received"
	EventTypeAggregationCompleted = "aggregation.completed"
	EventTypeDecisionCompleted    = "decision.completed"
	EventTypeDecisionFailed       = "decision.failed"
)

type Actor struct {
	UserID    *int64
	Login     string
	RoleSlug  string
	IP        string
	UserAgent string
}

type ResourceRef struct {
	Kind string
	ID   int64
}

type DomainRefs struct {
	ExaminationID      *int64
	SpecialistID       *int64
	QuestionnaireID    *int64
	Channel            *string
	DecisionSnapshotID *int64
}

type Event struct {
	ID            int64
	Type          string
	Key           string
	Outcome       string
	HappenedAt    time.Time
	RequestID     string
	TraceID       string
	TraceParent   string
	TraceState    string
	CorrelationID string
	Actor         Actor
	Resource      ResourceRef
	DomainRefs    DomainRefs
	Payload       json.RawMessage
}

type ListFilter struct {
	EventType     *string
	ResourceKind  *string
	ResourceID    *int64
	ActorUserID   *int64
	ExaminationID *int64
	From          time.Time
	To            time.Time
	Limit         int32
}

type Metadata struct {
	RequestID     string
	TraceID       string
	TraceParent   string
	TraceState    string
	CorrelationID string
	Actor         Actor
}

type contextKey string

const metadataContextKey contextKey = "audit_metadata"

func WithMetadata(ctx context.Context, meta Metadata) context.Context {
	return context.WithValue(ctx, metadataContextKey, meta)
}

func MetadataFromContext(ctx context.Context) Metadata {
	meta, _ := ctx.Value(metadataContextKey).(Metadata)
	return meta
}
