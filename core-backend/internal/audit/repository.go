package audit

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	sqlcdb "diplom/db/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	Append(context.Context, Event) error
	List(context.Context, ListFilter) ([]Event, error)
}

type SQLRepository struct {
	pool    *pgxpool.Pool
	queries *sqlcdb.Queries
}

func NewRepository(pool *pgxpool.Pool) *SQLRepository {
	return &SQLRepository{
		pool:    pool,
		queries: sqlcdb.New(pool),
	}
}

func (r *SQLRepository) Append(ctx context.Context, event Event) error {
	payload := event.Payload
	if len(payload) == 0 {
		payload = json.RawMessage(`{}`)
	}

	_, err := r.queries.CreateAuditLog(ctx, sqlcdb.CreateAuditLogParams{
		EventType:          event.Type,
		EventKey:           textValue(event.Key),
		Outcome:            event.Outcome,
		HappenedAt:         pgtype.Timestamptz{Time: event.HappenedAt.UTC(), Valid: true},
		RequestID:          textValue(event.RequestID),
		TraceID:            textValue(event.TraceID),
		Traceparent:        textValue(event.TraceParent),
		Tracestate:         textValue(event.TraceState),
		CorrelationID:      textValue(event.CorrelationID),
		ActorUserID:        int8Value(event.Actor.UserID),
		ActorLogin:         textValue(event.Actor.Login),
		ActorRoleSlug:      textValue(event.Actor.RoleSlug),
		ActorIp:            textValue(event.Actor.IP),
		ActorUserAgent:     textValue(event.Actor.UserAgent),
		ResourceKind:       event.Resource.Kind,
		ResourceID:         nullableInt8Value(event.Resource.ID),
		ExaminationID:      int8Value(event.DomainRefs.ExaminationID),
		SpecialistID:       int8Value(event.DomainRefs.SpecialistID),
		QuestionnaireID:    int8Value(event.DomainRefs.QuestionnaireID),
		Channel:            textPtrValue(event.DomainRefs.Channel),
		DecisionSnapshotID: int8Value(event.DomainRefs.DecisionSnapshotID),
		Payload:            payload,
	})
	if err != nil && !isNoRowsConflict(err) {
		return err
	}
	return nil
}

func (r *SQLRepository) List(ctx context.Context, filter ListFilter) ([]Event, error) {
	limit := filter.Limit
	if limit <= 0 {
		limit = 100
	}
	fromAt := filter.From
	if fromAt.IsZero() {
		fromAt = time.Unix(0, 0).UTC()
	}
	toAt := filter.To
	if toAt.IsZero() {
		toAt = time.Now().UTC().Add(time.Hour)
	}

	rows, err := r.queries.ListAuditLogs(ctx, sqlcdb.ListAuditLogsParams{
		EventType:     textPtrArg(filter.EventType),
		ResourceKind:  textPtrArg(filter.ResourceKind),
		ResourceID:    int8PtrArg(filter.ResourceID),
		ActorUserID:   int8PtrArg(filter.ActorUserID),
		ExaminationID: int8PtrArg(filter.ExaminationID),
		FromAt:        pgtype.Timestamptz{Time: fromAt.UTC(), Valid: true},
		ToAt:          pgtype.Timestamptz{Time: toAt.UTC(), Valid: true},
		LimitCount:    limit,
	})
	if err != nil {
		return nil, err
	}

	result := make([]Event, 0, len(rows))
	for _, row := range rows {
		result = append(result, mapEvent(row))
	}
	return result, nil
}

func mapEvent(row sqlcdb.AuditLog) Event {
	return Event{
		ID:            row.ID,
		Type:          row.EventType,
		Key:           textOrEmpty(row.EventKey),
		Outcome:       row.Outcome,
		HappenedAt:    row.HappenedAt.Time,
		RequestID:     textOrEmpty(row.RequestID),
		TraceID:       textOrEmpty(row.TraceID),
		TraceParent:   textOrEmpty(row.Traceparent),
		TraceState:    textOrEmpty(row.Tracestate),
		CorrelationID: textOrEmpty(row.CorrelationID),
		Actor: Actor{
			UserID:    nullableInt64(row.ActorUserID),
			Login:     textOrEmpty(row.ActorLogin),
			RoleSlug:  textOrEmpty(row.ActorRoleSlug),
			IP:        textOrEmpty(row.ActorIp),
			UserAgent: textOrEmpty(row.ActorUserAgent),
		},
		Resource: ResourceRef{
			Kind: row.ResourceKind,
			ID:   int64OrZero(row.ResourceID),
		},
		DomainRefs: DomainRefs{
			ExaminationID:      nullableInt64(row.ExaminationID),
			SpecialistID:       nullableInt64(row.SpecialistID),
			QuestionnaireID:    nullableInt64(row.QuestionnaireID),
			Channel:            nullableText(row.Channel),
			DecisionSnapshotID: nullableInt64(row.DecisionSnapshotID),
		},
		Payload: append(json.RawMessage(nil), row.Payload...),
	}
}

func isNoRowsConflict(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}

func textValue(value string) pgtype.Text {
	if value == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: value, Valid: true}
}

func textPtrValue(value *string) pgtype.Text {
	if value == nil || *value == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *value, Valid: true}
}

func textPtrArg(value *string) pgtype.Text {
	if value == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *value, Valid: true}
}

func textOrEmpty(value pgtype.Text) string {
	if !value.Valid {
		return ""
	}
	return value.String
}

func nullableText(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	result := value.String
	return &result
}

func int8Value(value *int64) pgtype.Int8 {
	if value == nil {
		return pgtype.Int8{}
	}
	return pgtype.Int8{Int64: *value, Valid: true}
}

func nullableInt8Value(value int64) pgtype.Int8 {
	if value == 0 {
		return pgtype.Int8{}
	}
	return pgtype.Int8{Int64: value, Valid: true}
}

func int8PtrArg(value *int64) pgtype.Int8 {
	if value == nil {
		return pgtype.Int8{}
	}
	return pgtype.Int8{Int64: *value, Valid: true}
}

func nullableInt64(value pgtype.Int8) *int64 {
	if !value.Valid {
		return nil
	}
	result := value.Int64
	return &result
}

func int64OrZero(value pgtype.Int8) int64 {
	if !value.Valid {
		return 0
	}
	return value.Int64
}
