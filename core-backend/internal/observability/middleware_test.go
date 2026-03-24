package observability

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTraceContextMiddleware(t *testing.T) {
	logger := NewLogger("DEBUG", &bytes.Buffer{})
	handler := HTTPMiddleware(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		trace := TraceFromContext(r.Context())
		if trace.RequestID == "" || trace.TraceID == "" || trace.TraceParent == "" {
			t.Fatal("expected trace context in request context")
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set("traceparent", "00-8ec8c1b6409f4a6cb80cfcb4f74aa98c-5d7c1f97db7840b3-01")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Header().Get("X-Request-Id") == "" || rec.Header().Get("traceparent") == "" {
		t.Fatal("expected response headers to expose request and trace identifiers")
	}
}

func TestRequestLoggingEmitsStructuredFields(t *testing.T) {
	buffer := &bytes.Buffer{}
	logger := NewLogger("INFO", buffer)
	handler := HTTPMiddleware(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	}))

	req := httptest.NewRequest(http.MethodPost, "/examinations/1/finish", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	output := buffer.String()
	for _, fragment := range []string{`"msg":"http_request"`, `"request_id"`, `"trace_id"`, `"method":"POST"`, `"status":202`} {
		if !bytes.Contains([]byte(output), []byte(fragment)) {
			t.Fatalf("expected structured log fragment %s in %s", fragment, output)
		}
	}
}

func TestRequestContextIncludesTraceAndRequestIDs(t *testing.T) {
	trace := BuildTraceContext("req-100", "00-8ec8c1b6409f4a6cb80cfcb4f74aa98c-5d7c1f97db7840b3-01", "")
	ctx := WithTraceContext(context.Background(), trace)

	got := TraceFromContext(ctx)
	if got.RequestID != "req-100" || got.TraceParent == "" {
		t.Fatalf("expected trace metadata round-trip, got %#v", got)
	}
}
