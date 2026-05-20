package decision

import (
	"encoding/json"
	"fmt"
	"strings"
)

type parsedKESMIResult struct {
	Recommendation string
	Message        string
	DecisionCode   string
	RiskClass      string
	Patterns       []string
}

type ParsedResultView struct {
	DecisionCode string
	RiskClass    string
	Patterns     []string
}

type kesmiResponseEnvelope struct {
	RequiredExploredParameters []kesmiParameter `json:"requiredExploredParameters"`
}

type kesmiParameter struct {
	ID    string `json:"id"`
	Value any    `json:"value"`
}

func parseKESMIResult(raw []byte) (parsedKESMIResult, error) {
	var envelope kesmiResponseEnvelope
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return parsedKESMIResult{}, fmt.Errorf("decode kesmi response: %w", err)
	}

	values := make(map[string]string, len(envelope.RequiredExploredParameters))
	for _, item := range envelope.RequiredExploredParameters {
		value, ok := item.Value.(string)
		if !ok {
			continue
		}
		values[item.ID] = strings.TrimSpace(value)
	}

	decisionCode := values["p32"]
	if decisionCode == "" {
		return parsedKESMIResult{}, fmt.Errorf("missing final_decision p32")
	}
	summary := values["p33"]
	if summary == "" {
		return parsedKESMIResult{}, fmt.Errorf("missing final_summary p33")
	}

	recommendation, err := normalizeRecommendation(decisionCode)
	if err != nil {
		return parsedKESMIResult{}, err
	}

	riskClass, summaryDecision, patterns, err := parseKESMISummary(summary)
	if err != nil {
		return parsedKESMIResult{}, err
	}
	if summaryDecision != "" && summaryDecision != decisionCode {
		return parsedKESMIResult{}, fmt.Errorf("decision mismatch between p32=%q and p33=%q", decisionCode, summaryDecision)
	}

	return parsedKESMIResult{
		Recommendation: recommendation,
		Message:        summary,
		DecisionCode:   decisionCode,
		RiskClass:      riskClass,
		Patterns:       patterns,
	}, nil
}

func ParseResultForView(raw []byte) (ParsedResultView, error) {
	parsed, err := parseKESMIResult(raw)
	if err != nil {
		return ParsedResultView{}, err
	}
	return ParsedResultView{
		DecisionCode: parsed.DecisionCode,
		RiskClass:    parsed.RiskClass,
		Patterns:     append([]string(nil), parsed.Patterns...),
	}, nil
}

func normalizeRecommendation(decisionCode string) (string, error) {
	switch strings.TrimSpace(decisionCode) {
	case DecisionCodeAllow:
		return DecisionRecommendationAllowed, nil
	case DecisionCodeMonitoring, DecisionCodeExtendedCheck:
		return DecisionRecommendationRisk, nil
	case DecisionCodeNoAccess:
		return DecisionRecommendationDenied, nil
	default:
		return "", fmt.Errorf("unknown final_decision %q", decisionCode)
	}
}

func parseKESMISummary(summary string) (riskClass string, decisionCode string, patterns []string, err error) {
	summary = strings.TrimSpace(summary)
	if summary == "" {
		return "", "", nil, fmt.Errorf("empty final_summary")
	}

	const riskPrefix = "risk="
	const decisionPrefix = "; decision="
	const patternsPrefix = "; patterns="

	riskIndex := strings.Index(summary, riskPrefix)
	decisionIndex := strings.Index(summary, decisionPrefix)
	patternsIndex := strings.Index(summary, patternsPrefix)
	if riskIndex != 0 || decisionIndex == -1 || patternsIndex == -1 || decisionIndex <= len(riskPrefix) || patternsIndex <= decisionIndex+len(decisionPrefix) {
		return "", "", nil, fmt.Errorf("unexpected final_summary format")
	}

	riskClass = strings.TrimSpace(summary[len(riskPrefix):decisionIndex])
	decisionCode = strings.TrimSpace(summary[decisionIndex+len(decisionPrefix) : patternsIndex])
	patternsRaw := strings.TrimSpace(summary[patternsIndex+len(patternsPrefix):])
	patternsRaw = strings.TrimSuffix(patternsRaw, ";")
	if patternsRaw == "" || patternsRaw == "none" {
		return riskClass, decisionCode, nil, nil
	}

	parts := strings.Split(patternsRaw, ";")
	patterns = make([]string, 0, len(parts))
	for _, part := range parts {
		value := strings.TrimSpace(part)
		if value == "" || value == "none" {
			continue
		}
		patterns = append(patterns, value)
	}
	return riskClass, decisionCode, patterns, nil
}
