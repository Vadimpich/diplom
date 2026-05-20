package decision

import "testing"

func TestParseKESMIResult(t *testing.T) {
	t.Run("allow", func(t *testing.T) {
		parsed, err := parseKESMIResult([]byte(`{"requiredExploredParameters":[{"id":"p32","value":"allow"},{"id":"p33","value":"risk=low; decision=allow; patterns=none"}]}`))
		if err != nil {
			t.Fatalf("parse kesmi result: %v", err)
		}
		if parsed.Recommendation != DecisionRecommendationAllowed {
			t.Fatalf("expected allowed, got %q", parsed.Recommendation)
		}
		if parsed.DecisionCode != DecisionCodeAllow || parsed.RiskClass != "low" || len(parsed.Patterns) != 0 {
			t.Fatalf("unexpected parsed result: %#v", parsed)
		}
	})

	t.Run("extended_check", func(t *testing.T) {
		parsed, err := parseKESMIResult([]byte(`{"requiredExploredParameters":[{"id":"p32","value":"extended_check"},{"id":"p33","value":"risk=high; decision=extended_check; patterns=open_stress;cognitive_overload;fatigue_state;"}]}`))
		if err != nil {
			t.Fatalf("parse kesmi result: %v", err)
		}
		if parsed.Recommendation != DecisionRecommendationRisk {
			t.Fatalf("expected coarse risk, got %q", parsed.Recommendation)
		}
		if parsed.DecisionCode != DecisionCodeExtendedCheck || parsed.RiskClass != "high" || len(parsed.Patterns) != 3 {
			t.Fatalf("unexpected parsed result: %#v", parsed)
		}
	})

	t.Run("no_access", func(t *testing.T) {
		parsed, err := parseKESMIResult([]byte(`{"requiredExploredParameters":[{"id":"p32","value":"no_access"},{"id":"p33","value":"risk=critical; decision=no_access; patterns=open_stress;cognitive_overload;fatigue_state;compensation_breakdown;"}]}`))
		if err != nil {
			t.Fatalf("parse kesmi result: %v", err)
		}
		if parsed.Recommendation != DecisionRecommendationDenied {
			t.Fatalf("expected denied, got %q", parsed.Recommendation)
		}
		if parsed.DecisionCode != DecisionCodeNoAccess || parsed.RiskClass != "critical" || len(parsed.Patterns) != 4 {
			t.Fatalf("unexpected parsed result: %#v", parsed)
		}
	})
}

func TestParseKESMIResultRejectsInvalidResponse(t *testing.T) {
	if _, err := parseKESMIResult([]byte(`{"requiredExploredParameters":[{"id":"p32","value":"allow"}]}`)); err == nil {
		t.Fatal("expected missing p33 to fail")
	}
	if _, err := parseKESMIResult([]byte(`{"requiredExploredParameters":[{"id":"p32","value":"weird"},{"id":"p33","value":"risk=low; decision=weird; patterns=none"}]}`)); err == nil {
		t.Fatal("expected unknown decision code to fail")
	}
}
