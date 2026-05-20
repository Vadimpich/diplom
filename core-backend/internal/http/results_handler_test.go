package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"diplom/internal/auth"
	"diplom/internal/results"
)

func TestProcessingStatusShowsAggregating(t *testing.T) {
	router := NewRouter(Dependencies{
		AuthTokens: resultsTokenStub{},
	})

	req := httptest.NewRequest(http.MethodGet, "/examinations/100/processing-status", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected endpoint availability for aggregating contract test, got %d", rec.Code)
	}

	body := rec.Body.String()
	if body == "" || !containsAll(body, `"status":"aggregating"`, `"terminal":false`) {
		t.Fatalf("expected processing-status contract to expose aggregating as non-terminal, got %s", body)
	}
}

func TestSpecialistResultHistoryEndpoint(t *testing.T) {
	router := NewRouter(Dependencies{
		AuthTokens: resultsTokenStub{},
	})

	req := httptest.NewRequest(http.MethodGet, "/specialists/55/result-history", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected GET /specialists/{id}/result-history to return canonical aggregated trend history, got status %d", rec.Code)
	}

	body := rec.Body.String()
	if !containsAll(body, `"specialist_id":55`, `"key_metrics"`, `"baseline_snapshot"`) {
		t.Fatalf("expected specialist result-history payload with baseline snapshot and key-metric dynamics, got %s", body)
	}
}

func TestExaminationResultIncludesDecisionBlock(t *testing.T) {
	router := NewRouter(Dependencies{
		AuthTokens: resultsTokenStub{},
	})

	req := httptest.NewRequest(http.MethodGet, "/examinations/100/result", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected GET /examinations/{id}/result availability, got %d", rec.Code)
	}

	body := rec.Body.String()
	if !containsAll(body, `"decision"`, `"state":"pending"`, `"recommendation":"unavailable"`, `"correlation_id"`, `"channel_reports"`) {
		t.Fatalf("expected result contract to expose normalized decision block, got %s", body)
	}
}

func TestDecisionFailureDiagnostics(t *testing.T) {
	router := NewRouter(Dependencies{
		AuthTokens: resultsTokenStub{},
		Results:    &results.Service{},
	})
	router = NewRouter(Dependencies{
		AuthTokens: resultsTokenStub{},
		Results:    results.NewService(resultsRepoStubForHTTP{result: results.ExaminationResultResponse{ExaminationID: 100, Status: "completed", Decision: results.DecisionResultView{State: "business_error", Recommendation: "unavailable", Message: "analysis_not_implemented_yet", CorrelationID: "exam-100-kesmi-1", AttemptCount: 1, MaxAttempts: 2, Diagnostics: results.DecisionDiagnosticsView{ErrorClass: stringPtr("business"), ErrorCode: stringPtr("unknown_model"), ErrorMessage: stringPtr("unknown model"), Retryable: false}, RawResponseAvailable: true}}}),
	})
	req := httptest.NewRequest(http.MethodGet, "/examinations/100/result", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	body := rec.Body.String()
	if !containsAll(body, `"diagnostics"`, `"error_class"`, `"raw_response_available"`) {
		t.Fatalf("expected result contract to surface normalized diagnostics fields, got %s", body)
	}
}

func TestCompletedResultIncludesPlaceholderDecision(t *testing.T) {
	router := NewRouter(Dependencies{
		AuthTokens: resultsTokenStub{},
		Results:    results.NewService(resultsRepoStubForHTTP{result: results.ExaminationResultResponse{ExaminationID: 100, Status: "completed", Decision: results.DecisionResultView{State: "succeeded", Recommendation: "unavailable", Message: "analysis_not_implemented_yet", CorrelationID: "exam-100-kesmi-1", AttemptCount: 1, MaxAttempts: 2, RawResponseAvailable: true}}}),
	})
	req := httptest.NewRequest(http.MethodGet, "/examinations/100/result", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	body := rec.Body.String()
	if !containsAll(body, `"status":"completed"`, `"message":"analysis_not_implemented_yet"`) {
		t.Fatalf("expected completed result to keep honest placeholder decision semantics, got %s", body)
	}
}

func containsAll(body string, fragments ...string) bool {
	for _, fragment := range fragments {
		if !contains(body, fragment) {
			return false
		}
	}
	return true
}

func contains(body, fragment string) bool {
	return len(fragment) == 0 || (len(body) >= len(fragment) && index(body, fragment) >= 0)
}

func index(s, sep string) int {
	n := len(sep)
	if n == 0 {
		return 0
	}
	for i := 0; i+n <= len(s); i++ {
		if s[i:i+n] == sep {
			return i
		}
	}
	return -1
}

type resultsTokenStub struct{}

func (resultsTokenStub) Issue(auth.User) (string, int64, error) {
	return "token", 900, nil
}

func (resultsTokenStub) Parse(string) (auth.Claims, error) {
	return auth.Claims{
		UserID:   1,
		Login:    "operator",
		RoleSlug: "operator",
	}, nil
}

type resultsRepoStubForHTTP struct {
	result results.ExaminationResultResponse
}

func (s resultsRepoStubForHTTP) GetExaminationResult(context.Context, int64) (results.ExaminationResultResponse, error) {
	return s.result, nil
}

func (s resultsRepoStubForHTTP) GetSpecialistHistory(context.Context, int64) (results.SpecialistHistoryResponse, error) {
	return results.SpecialistHistoryResponse{}, nil
}

func stringPtr(value string) *string {
	return &value
}
