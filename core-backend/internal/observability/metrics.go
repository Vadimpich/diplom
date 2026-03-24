package observability

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

type MetricsRegistry struct {
	mu                  sync.Mutex
	httpRequestsTotal   int64
	httpRequestDuration float64
	dependencyStatus    map[string]float64
}

func NewMetricsRegistry() *MetricsRegistry {
	return &MetricsRegistry{dependencyStatus: map[string]float64{}}
}

func (r *MetricsRegistry) RecordHTTPRequest(_ string, _ string, _ int, duration time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.httpRequestsTotal++
	r.httpRequestDuration += duration.Seconds()
}

func (r *MetricsRegistry) SetDependencyUp(name string, up bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if up {
		r.dependencyStatus[name] = 1
		return
	}
	r.dependencyStatus[name] = 0
}

func (r *MetricsRegistry) Render() string {
	r.mu.Lock()
	defer r.mu.Unlock()

	var builder strings.Builder
	builder.WriteString("# TYPE diplom_http_requests_total counter\n")
	builder.WriteString(fmt.Sprintf("diplom_http_requests_total %d\n", r.httpRequestsTotal))
	builder.WriteString("# TYPE diplom_http_request_duration_seconds histogram\n")
	builder.WriteString(fmt.Sprintf("diplom_http_request_duration_seconds_sum %f\n", r.httpRequestDuration))
	builder.WriteString(fmt.Sprintf("diplom_http_request_duration_seconds_count %d\n", r.httpRequestsTotal))
	builder.WriteString("# TYPE diplom_dependency_up gauge\n")
	for name, value := range r.dependencyStatus {
		builder.WriteString(fmt.Sprintf("diplom_dependency_up{dependency=%q} %v\n", name, value))
	}
	return builder.String()
}
