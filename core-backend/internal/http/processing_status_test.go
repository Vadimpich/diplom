package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"diplom/internal/auth"
	"diplom/internal/examinations"
	"diplom/internal/processing"
)

func TestProcessingStatusEndpoint(t *testing.T) {
	now := time.Unix(1_742_550_000, 0).UTC()
	brokerMessageID := "msg-100-text-1"
	brokerCorrelationID := "exam-100-text-v1"
	router := NewRouter(Dependencies{
		AuthTokens: processingStatusTokenStub{},
		Processing: processing.NewService(&processingStatusRepoStub{
			response: processing.ProcessingStatusResponse{
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
					{
						Channel:        processing.ChannelAcoustic,
						Status:         "processing",
						AttemptCount:   1,
						MaxAttempts:    3,
						MessageVersion: processing.MessageVersionV1,
						QueuedAt:       &now,
						StartedAt:      &now,
					},
					{
						Channel:        processing.ChannelParalinguistic,
						Status:         "queued",
						AttemptCount:   0,
						MaxAttempts:    3,
						MessageVersion: processing.MessageVersionV1,
						QueuedAt:       &now,
					},
				},
			},
		}),
	})

	req := httptest.NewRequest(http.MethodGet, "/examinations/100/processing-status", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected GET /examinations/{id}/processing-status to be available, got status %d", rec.Code)
	}

	var payload processing.ProcessingStatusResponse
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode processing status body: %v", err)
	}
	if payload.Status != "processing" {
		t.Fatalf("expected processing status, got %q", payload.Status)
	}
	if payload.ChannelsTotal != 3 || len(payload.Channels) != 3 {
		t.Fatalf("expected three mandatory channels in response, got total=%d len=%d", payload.ChannelsTotal, len(payload.Channels))
	}
	if payload.Terminal {
		t.Fatal("expected active pipeline to be non-terminal")
	}
}

func TestProcessingStatusEndpointReturnsTerminalError(t *testing.T) {
	now := time.Unix(1_742_550_100, 0).UTC()
	failedAt := now
	errorCode := "invalid_payload"
	errorMessage := "unsupported document"
	router := NewRouter(Dependencies{
		AuthTokens: processingStatusTokenStub{},
		Processing: processing.NewService(&processingStatusRepoStub{
			response: processing.ProcessingStatusResponse{
				ExaminationID:    100,
				Status:           examinations.StatusFailed,
				MessageVersion:   processing.MessageVersionV1,
				ChannelsTotal:    3,
				ChannelsComplete: 3,
				Terminal:         true,
				UpdatedAt:        now,
				FailedAt:         &failedAt,
				Channels: []processing.ChannelStatusDTO{
					{
						Channel:          processing.ChannelText,
						Status:           "exhausted",
						AttemptCount:     3,
						MaxAttempts:      3,
						MessageVersion:   processing.MessageVersionV1,
						FinishedAt:       &failedAt,
						LastErrorCode:    &errorCode,
						LastErrorMessage: &errorMessage,
					},
				},
			},
		}),
	})

	req := httptest.NewRequest(http.MethodGet, "/examinations/100/processing-status", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "\"status\":\"failed\"") {
		t.Fatalf("expected terminal failed status in payload, got %s", body)
	}
	if !strings.Contains(body, "\"last_error_code\":\"invalid_payload\"") {
		t.Fatalf("expected failing channel error details in payload, got %s", body)
	}
}

func TestProcessingStatusEndpointRejectsUnavailablePipeline(t *testing.T) {
	router := NewRouter(Dependencies{
		AuthTokens: processingStatusTokenStub{},
		Processing: processing.NewService(&processingStatusRepoStub{
			statusErr: processing.ErrProcessingStatusUnavailable,
		}),
	})

	req := httptest.NewRequest(http.MethodGet, "/examinations/100/processing-status", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409 when examination has not entered processing pipeline, got %d", rec.Code)
	}
}

type processingStatusRepoStub struct {
	response  processing.ProcessingStatusResponse
	statusErr error
}

func (s *processingStatusRepoStub) FinishLaunch(context.Context, int64) (examinations.Examination, []processing.ProcessingCommandEnvelope, error) {
	return examinations.Examination{}, nil, nil
}

func (s *processingStatusRepoStub) GetProcessingStatus(context.Context, int64) (processing.ProcessingStatusResponse, error) {
	if s.statusErr != nil {
		return processing.ProcessingStatusResponse{}, s.statusErr
	}
	return s.response, nil
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
