package decision

import "testing"

func TestBlocksBaselineRecommendationAllowsRiskDuringBootstrap(t *testing.T) {
	if blocksBaselineRecommendation(DecisionRecommendationRisk) {
		t.Fatal("expected risk recommendation not to block baseline bootstrap")
	}
}

func TestBlocksBaselineRecommendationStillBlocksDenied(t *testing.T) {
	if !blocksBaselineRecommendation(DecisionRecommendationDenied) {
		t.Fatal("expected denied recommendation to block baseline updates")
	}
}
