package observability

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"strings"
)

type TraceContext struct {
	RequestID   string
	TraceID     string
	TraceParent string
	TraceState  string
}

type contextKey string

const traceContextKey contextKey = "observability_trace_context"

func WithTraceContext(ctx context.Context, trace TraceContext) context.Context {
	return context.WithValue(ctx, traceContextKey, trace)
}

func TraceFromContext(ctx context.Context) TraceContext {
	trace, _ := ctx.Value(traceContextKey).(TraceContext)
	return trace
}

func BuildTraceContext(requestID, traceParent, traceState string) TraceContext {
	requestID = strings.TrimSpace(requestID)
	if requestID == "" {
		requestID = randomHex(8)
	}

	traceParent = strings.TrimSpace(traceParent)
	if traceParent == "" {
		traceID := randomHex(16)
		spanID := randomHex(8)
		traceParent = "00-" + traceID + "-" + spanID + "-01"
		return TraceContext{
			RequestID:   requestID,
			TraceID:     traceID,
			TraceParent: traceParent,
			TraceState:  strings.TrimSpace(traceState),
		}
	}

	traceID := parseTraceID(traceParent)
	if traceID == "" {
		traceID = randomHex(16)
		spanID := randomHex(8)
		traceParent = "00-" + traceID + "-" + spanID + "-01"
	}

	return TraceContext{
		RequestID:   requestID,
		TraceID:     traceID,
		TraceParent: traceParent,
		TraceState:  strings.TrimSpace(traceState),
	}
}

func parseTraceID(traceParent string) string {
	parts := strings.Split(traceParent, "-")
	if len(parts) != 4 || len(parts[1]) != 32 {
		return ""
	}
	return parts[1]
}

func randomHex(size int) string {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		for i := range buf {
			buf[i] = byte(i + 1)
		}
	}
	return hex.EncodeToString(buf)
}
