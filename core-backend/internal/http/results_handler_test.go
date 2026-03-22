package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"dimplom/internal/auth"
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
