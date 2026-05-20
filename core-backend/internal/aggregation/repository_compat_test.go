package aggregation

import "testing"

func TestDecodeExistingBaselineMetricsCurrentSchema(t *testing.T) {
	raw := []byte(`{
		"text_risk_signal": {
			"baseline_mean": 0.12,
			"baseline_std": 0.04,
			"sample_count": 6,
			"method": "ewma"
		}
	}`)

	metrics, err := decodeExistingBaselineMetrics(raw)
	if err != nil {
		t.Fatalf("decode current metrics: %v", err)
	}
	item, ok := metrics["text_risk_signal"]
	if !ok {
		t.Fatal("expected current metric to be present")
	}
	if item.SampleCount != 6 {
		t.Fatalf("expected sample_count=6, got %d", item.SampleCount)
	}
}

func TestDecodeExistingBaselineMetricsLegacySchemaFallsBackToEmpty(t *testing.T) {
	raw := []byte(`{
		"centers": {"text_proxy_signal": 0.3},
		"scales": {"text_proxy_signal": 0.0},
		"exam_count": 1
	}`)

	metrics, err := decodeExistingBaselineMetrics(raw)
	if err != nil {
		t.Fatalf("decode legacy metrics: %v", err)
	}
	if len(metrics) != 0 {
		t.Fatalf("expected legacy metrics to be ignored, got %d entries", len(metrics))
	}
}
