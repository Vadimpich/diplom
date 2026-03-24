package baselineclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"diplom/internal/aggregation"
	"diplom/internal/observability"
)

var ErrUnavailable = errors.New("baselineclient: unavailable")
var ErrInvalidResponse = errors.New("baselineclient: invalid response")

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func New(baseURL string, timeout time.Duration) *Client {
	baseURL = strings.TrimRight(baseURL, "/")
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *Client) Calculate(ctx context.Context, request aggregation.BaselineRequest) (aggregation.BaselineResponse, error) {
	body, err := json.Marshal(request)
	if err != nil {
		return aggregation.BaselineResponse{}, err
	}
	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/baseline/calculate", bytes.NewReader(body))
	if err != nil {
		return aggregation.BaselineResponse{}, err
	}
	httpRequest.Header.Set("Content-Type", "application/json")
	trace := observability.TraceFromContext(ctx)
	if trace.RequestID != "" {
		httpRequest.Header.Set("X-Request-Id", trace.RequestID)
	}
	if trace.TraceParent != "" {
		httpRequest.Header.Set("traceparent", trace.TraceParent)
	}
	if trace.TraceState != "" {
		httpRequest.Header.Set("tracestate", trace.TraceState)
	}

	response, err := c.httpClient.Do(httpRequest)
	if err != nil {
		return aggregation.BaselineResponse{}, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return aggregation.BaselineResponse{}, fmt.Errorf("%w: status %d", ErrUnavailable, response.StatusCode)
	}

	var payload aggregation.BaselineResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return aggregation.BaselineResponse{}, fmt.Errorf("%w: %v", ErrInvalidResponse, err)
	}
	if payload.SchemaVersion == 0 || payload.AlgorithmVersion == "" {
		return aggregation.BaselineResponse{}, ErrInvalidResponse
	}
	return payload, nil
}
