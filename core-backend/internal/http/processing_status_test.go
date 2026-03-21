package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"dimplom/internal/auth"
	"dimplom/internal/processing"
)

func TestProcessingStatusEndpoint(t *testing.T) {
	now := time.Unix(1_742_550_000, 0).UTC()
	brokerMessageID := "msg-100-text-1"
	brokerCorrelationID := "exam-100-text-v1"
	response := processing.ProcessingStatusResponse{
		ExaminationID:    100,
		Status:           "processing",
		MessageVersion:   processing.MessageVersionV1,
		ChannelsTotal:    3,
		ChannelsComplete: 1,
		Terminal:         false,
		StartedAt:        &now,
		UpdatedAt:        now,
		Channels: []processing.ChannelStatusDTO{
			{
				Channel:             processing.ChannelText,
				Status:              "succeeded",
				AttemptCount:        1,
				MaxAttempts:         3,
				MessageVersion:      processing.MessageVersionV1,
				QueuedAt:            &now,
				StartedAt:           &now,
				FinishedAt:          &now,
				BrokerMessageID:     &brokerMessageID,
				BrokerCorrelationID: &brokerCorrelationID,
			},
		},
	}

	payload, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("marshal processing status response: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("decode processing status response: %v", err)
	}

	if decoded["status"] != "processing" {
		t.Fatalf("expected DTO to expose processing status, got %#v", decoded["status"])
	}
	if _, ok := decoded["channels"]; !ok {
		t.Fatal("expected DTO to expose per-channel runtime state")
	}

	router := NewRouter(Dependencies{
		AuthTokens: processingStatusTokenStub{},
	})

	req := httptest.NewRequest(http.MethodGet, "/examinations/100/processing-status", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected GET /examinations/{id}/processing-status to be available, got status %d", rec.Code)
	}
}

type processingStatusTokenStub struct{}

func (processingStatusTokenStub) Issue(auth.User) (string, int64, error) {
	return "token", 900, nil
}

func (processingStatusTokenStub) Parse(string) (auth.Claims, error) {
	return auth.Claims{
		UserID:   1,
		Login:    "operator",
		RoleSlug: "operator",
	}, nil
}
