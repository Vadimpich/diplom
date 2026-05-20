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

const (
	outputFinalDecisionID = "p32"
	outputFinalSummaryID  = "p33"
)

var parameterIDs = map[string]string{
	"text_negativity_score":        "i1",
	"text_anxiety_score":           "i2",
	"text_confidence_score":        "i3",
	"text_coherence_score":         "i4",
	"text_evasion_score":           "i5",
	"acoustic_stress_score":        "i6",
	"voice_stability_score":        "i7",
	"intensity_variability_score":  "i8",
	"hesitation_score":             "i9",
	"speech_disorganization_score": "i10",
	"baseline_available":           "i11",
	"baseline_deviation_index":     "i12",
	"overall_deviation_z":          "i13",
	"text_risk_z":                  "i14",
	"acoustic_stress_z":            "i15",
	"paralinguistic_behavior_z":    "i16",
	"speech_stability_z":           "i17",
	"data_reliability":             "i18",
	"semantic_stress_index":        "i19",
	"acoustic_activation_index":    "i20",
	"speech_disorganization_index": "i21",
	"baseline_shift_index":         "i22",
}

type Client struct {
	baseURL    string
	modelID    string
	httpClient *http.Client
}

type parameterValue struct {
	ID    string  `json:"id"`
	Value float64 `json:"value"`
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
	parameters := buildIncomingParameters(input)
	body, err := json.Marshal(map[string]any{
		"modelID":             c.modelID,
		"incommingParameters": parameters,
		"outputParameters":    []string{outputFinalDecisionID, outputFinalSummaryID},
		"service": map[string]any{
			"outputFields": []string{
				"requiredExploredParameters",
			},
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

func buildIncomingParameters(input decision.DecisionInput) []parameterValue {
	values := map[string]float64{
		"data_reliability":             input.DataReliability,
		"text_negativity_score":        input.Channels["text"].Scores["text_negativity_score"],
		"text_anxiety_score":           input.Channels["text"].Scores["text_anxiety_score"],
		"text_confidence_score":        input.Channels["text"].Scores["text_confidence_score"],
		"text_coherence_score":         input.Channels["text"].Scores["text_coherence_score"],
		"text_evasion_score":           input.Channels["text"].Scores["text_evasion_score"],
		"acoustic_stress_score":        input.Channels["acoustic"].Scores["acoustic_stress_score"],
		"voice_stability_score":        input.Channels["acoustic"].Scores["voice_stability_score"],
		"intensity_variability_score":  input.Channels["acoustic"].Scores["intensity_variability_score"],
		"hesitation_score":             input.Channels["paralinguistic"].Scores["hesitation_score"],
		"speech_disorganization_score": input.Channels["paralinguistic"].Scores["speech_disorganization_score"],
		"baseline_available":           boolToFloat(input.Baseline.Available),
		"baseline_deviation_index":     input.Baseline.BaselineDeviationIndex,
		"overall_deviation_z":          input.Baseline.ZScores["overall_deviation_index"],
		"text_risk_z":                  input.Baseline.ZScores["text_risk_signal"],
		"acoustic_stress_z":            input.Baseline.ZScores["acoustic_stress_signal"],
		"paralinguistic_behavior_z":    input.Baseline.ZScores["paralinguistic_behavior_signal"],
		"speech_stability_z":           input.Baseline.ZScores["speech_stability_score"],
		"semantic_stress_index":        input.DerivedIndicators.SemanticStressIndex,
		"acoustic_activation_index":    input.DerivedIndicators.AcousticActivationIndex,
		"speech_disorganization_index": input.DerivedIndicators.SpeechDisorganizationIndex,
		"baseline_shift_index":         input.DerivedIndicators.BaselineShiftIndex,
	}

	keys := []string{
		"text_negativity_score",
		"text_anxiety_score",
		"text_confidence_score",
		"text_coherence_score",
		"text_evasion_score",
		"acoustic_stress_score",
		"voice_stability_score",
		"intensity_variability_score",
		"hesitation_score",
		"speech_disorganization_score",
		"baseline_available",
		"baseline_deviation_index",
		"overall_deviation_z",
		"text_risk_z",
		"acoustic_stress_z",
		"paralinguistic_behavior_z",
		"speech_stability_z",
		"data_reliability",
		"semantic_stress_index",
		"acoustic_activation_index",
		"speech_disorganization_index",
		"baseline_shift_index",
	}

	result := make([]parameterValue, 0, len(keys))
	for _, key := range keys {
		result = append(result, parameterValue{
			ID:    parameterIDs[key],
			Value: values[key],
		})
	}
	return result
}

func boolToFloat(value bool) float64 {
	if value {
		return 1
	}
	return 0
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
