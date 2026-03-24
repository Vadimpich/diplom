package observability

import (
	"log/slog"
	"net/http"
	"time"

	"diplom/internal/audit"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func HTTPMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return HTTPMiddlewareWithMetrics(logger, nil)
}

func HTTPMiddlewareWithMetrics(logger *slog.Logger, metrics *MetricsRegistry) func(http.Handler) http.Handler {
	if logger == nil {
		logger = NewLogger("INFO", nil)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get("X-Request-Id")
			if requestID == "" {
				requestID = chimiddleware.GetReqID(r.Context())
			}
			trace := BuildTraceContext(requestID, r.Header.Get("traceparent"), r.Header.Get("tracestate"))

			ctx := WithTraceContext(r.Context(), trace)
			meta := audit.MetadataFromContext(ctx)
			meta.RequestID = trace.RequestID
			meta.TraceID = trace.TraceID
			meta.TraceParent = trace.TraceParent
			meta.TraceState = trace.TraceState
			ctx = audit.WithMetadata(ctx, meta)

			w.Header().Set("X-Request-Id", trace.RequestID)
			w.Header().Set("traceparent", trace.TraceParent)
			if trace.TraceState != "" {
				w.Header().Set("tracestate", trace.TraceState)
			}

			recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			startedAt := time.Now()
			next.ServeHTTP(recorder, r.WithContext(ctx))
			if metrics != nil {
				metrics.RecordHTTPRequest(r.URL.Path, r.Method, recorder.status, time.Since(startedAt))
			}

			logger.Info("http_request",
				slog.String("request_id", trace.RequestID),
				slog.String("trace_id", trace.TraceID),
				slog.String("traceparent", trace.TraceParent),
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", recorder.status),
				slog.Duration("duration", time.Since(startedAt)),
			)
		})
	}
}
