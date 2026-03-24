package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"diplom/internal/auth"
)

func TestHealthEndpointIsCheap(t *testing.T) {
	checker := &observabilityDBStub{}
	handler := Handler{db: checker}

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	handler.Health(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected /health to stay cheap and return 200 when process is alive, got %d", rec.Code)
	}
	if checker.pingCalls != 0 {
		t.Fatalf("expected /health to avoid dependency fan-in, got %d dependency checks", checker.pingCalls)
	}
}

func TestReadinessDegradesOnDependencyFailure(t *testing.T) {
	router := NewRouter(Dependencies{
		AuthTokens: observabilityTokenStub{},
	})

	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected /ready to exist and degrade while dependencies are unavailable, got %d", rec.Code)
	}
}

func TestMetricsEndpointExposesLowCardinalityFamilies(t *testing.T) {
	router := NewRouter(Dependencies{
		AuthTokens: observabilityTokenStub{},
	})

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected /metrics availability, got %d", rec.Code)
	}
	body := rec.Body.String()
	if !containsAll(body, "diplom_http_requests_total", "diplom_dependency_up") {
		t.Fatalf("expected prometheus metric families, got %s", body)
	}
	if contains(body, "examination_id") || contains(body, "user_id") || contains(body, "correlation_id") {
		t.Fatalf("expected low-cardinality metrics without domain identifiers, got %s", body)
	}
}

func TestRequestContextIncludesTraceAndRequestIDs(t *testing.T) {
	router := NewRouter(Dependencies{
		AuthTokens: observabilityTokenStub{},
	})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set("traceparent", "00-8ec8c1b6409f4a6cb80cfcb4f74aa98c-5d7c1f97db7840b3-01")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if got := rec.Header().Get("X-Request-Id"); got == "" {
		t.Fatal("expected request id header to be echoed/generated")
	}
	if got := rec.Header().Get("traceparent"); got == "" {
		t.Fatal("expected traceparent response header for continued trace context")
	}
}

type observabilityDBStub struct {
	pingCalls int
}

func (s *observabilityDBStub) Ping(context.Context) error {
	s.pingCalls++
	return nil
}

type observabilityTokenStub struct{}

func (observabilityTokenStub) Issue(auth.User) (string, int64, error) {
	return "token", 900, nil
}

func (observabilityTokenStub) Parse(string) (auth.Claims, error) {
	return auth.Claims{UserID: 1, Login: "operator", RoleSlug: "operator"}, nil
}
