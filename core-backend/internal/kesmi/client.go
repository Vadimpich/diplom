package kesmi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"diplom/internal/decision"
	"diplom/internal/observability"
)

type Client struct {
	baseURL    string
	modelID    string
	httpClient *http.Client
}

func NewClient(baseURL, modelID string, timeout time.Duration) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		modelID: modelID,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *Client) Execute(ctx context.Context, input decision.DecisionInput, correlationID string) (decision.ExecutionResponse, error) {
	body, err := json.Marshal(map[string]any{
		"modelID":             c.modelID,
		"incommingParameters": []any{},
		"outputParameters":    []any{},
		"service": map[string]any{
			"outputFields":   []string{"timing"},
			"correlationID":  correlationID,
			"payloadVersion": input.PayloadVersion,
		},
	})
	if err != nil {
		return decision.ExecutionResponse{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/ModelCalc", bytes.NewReader(body))
	if err != nil {
		return decision.ExecutionResponse{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	trace := observability.TraceFromContext(ctx)
	if trace.RequestID != "" {
		req.Header.Set("X-Request-Id", trace.RequestID)
	}
	if trace.TraceParent != "" {
		req.Header.Set("traceparent", trace.TraceParent)
	}
	if trace.TraceState != "" {
		req.Header.Set("tracestate", trace.TraceState)
	}
	if correlationID != "" {
		req.Header.Set("X-Correlation-Id", correlationID)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		classification := ClassifyFailure(err, 0, "")
		return decision.ExecutionResponse{
			State:        classification.State,
			ErrorClass:   classification.ErrorClass,
			ErrorCode:    "transport_error",
			ErrorMessage: err.Error(),
			Retryable:    classification.Retryable,
		}, nil
	}
	defer resp.Body.Close()

	raw := new(bytes.Buffer)
	if _, err := raw.ReadFrom(resp.Body); err != nil {
		return decision.ExecutionResponse{}, err
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return decision.ExecutionResponse{
			HTTPStatus:  resp.StatusCode,
			RawResponse: raw.Bytes(),
		}, nil
	}

	errorCode, errorMessage := extractError(raw.Bytes())
	classification := ClassifyFailure(nil, resp.StatusCode, errorCode)
	return decision.ExecutionResponse{
		State:        classification.State,
		ErrorClass:   classification.ErrorClass,
		ErrorCode:    errorCode,
		ErrorMessage: errorMessage,
		HTTPStatus:   resp.StatusCode,
		Retryable:    classification.Retryable,
		RawResponse:  raw.Bytes(),
	}, nil
}

func (c *Client) Models(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/Models", nil)
	if err != nil {
		return err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("kesmi models status=%d", resp.StatusCode)
	}
	return nil
}

func extractError(raw []byte) (string, string) {
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return "", ""
	}
	return firstString(payload, "error_code", "code", "constraint"), firstString(payload, "error_message", "message", "description")
}

func firstString(payload map[string]any, keys ...string) string {
	for _, key := range keys {
		if value, ok := payload[key].(string); ok && value != "" {
			return value
		}
	}
	return ""
}
