package results

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"sort"
	"time"

	"diplom/internal/decision"
	"diplom/internal/processing"
	"diplom/internal/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var scoreLabelByKey = map[string]string{
	"text_negativity_score":        "Негативная окраска текста",
	"text_anxiety_score":           "Тревожность текста",
	"text_confidence_score":        "Уверенность ответа",
	"text_coherence_score":         "Связность ответа",
	"text_evasion_score":           "Уклончивость ответа",
	"acoustic_stress_score":        "Акустическое напряжение",
	"voice_stability_score":        "Стабильность голоса",
	"intensity_variability_score":  "Вариативность интенсивности",
	"hesitation_score":             "Выраженность пауз и колебаний",
	"speech_disorganization_score": "Речевая дезорганизация",
}

var channelOrder = map[string]int{
	"text":           0,
	"acoustic":       1,
	"paralinguistic": 2,
}

type SQLRepository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *SQLRepository {
	return &SQLRepository{pool: pool}
}

func (r *SQLRepository) GetExaminationResult(ctx context.Context, examinationID int64) (ExaminationResultResponse, error) {
	const profileQuery = `
SELECT
	p.examination_id,
	p.specialist_id,
	p.schema_version,
	p.aggregation_version,
	e.status,
	p.generated_at,
	p.overall_score,
	p.overall_band,
	p.primary_metric_key,
	p.neutral_recommendation_placeholder,
	b.algorithm_version,
	b.refreshed_at,
	b.general_delta,
	b.general_band,
	b.general_baseline_available,
	b.general_baseline_source,
	b.general_reference_population_version,
	b.personal_delta,
	b.personal_band,
	b.personal_baseline_available,
	b.personal_baseline_source,
	b.baseline_exam_count,
	b.data_reliability,
	b.update_eligible,
	ds.status,
	ds.recommendation,
	ds.message,
	ds.correlation_id,
	ds.attempt_count,
	ds.max_attempts,
	ds.last_attempt_at,
	ds.diagnostics_json,
	ds.raw_response_json,
	(ds.raw_response_json IS NOT NULL) AS raw_response_available
FROM aggregated_examination_profiles p
JOIN examinations e
	ON e.id = p.examination_id
LEFT JOIN examination_baseline_snapshots b
	ON b.profile_id = p.id
LEFT JOIN decision_snapshots ds
	ON ds.examination_id = p.examination_id
WHERE p.examination_id = $1
  AND e.status IN ('decision_pending', 'completed')`

	var result ExaminationResultResponse
	var generatedAt time.Time
	var refreshedAt *time.Time
	var generalRef *string
	var generalDelta, personalDelta *float64
	var generalBand, personalBand *string
	var generalAvailable, personalAvailable *bool
	var generalSource, personalSource *string
	var baselineExamCount *int
	var dataReliability *float64
	var updateEligible *bool
	var decisionState, decisionRecommendation, decisionMessage, decisionCorrelation *string
	var decisionAttemptCount, decisionMaxAttempts *int32
	var lastAttemptAt *time.Time
	var diagnosticsJSON []byte
	var rawResponseJSON []byte
	var rawResponseAvailable bool
	err := r.pool.QueryRow(ctx, profileQuery, examinationID).Scan(
		&result.ExaminationID,
		&result.SpecialistID,
		&result.SchemaVersion,
		&result.AggregationVersion,
		&result.Status,
		&generatedAt,
		&result.Summary.OverallScore,
		&result.Summary.OverallBand,
		&result.Summary.PrimaryMetricKey,
		&result.Summary.NeutralRecommendationPlaceholder,
		&result.BaselineSnapshot.AlgorithmVersion,
		&refreshedAt,
		&generalDelta,
		&generalBand,
		&generalAvailable,
		&generalSource,
		&generalRef,
		&personalDelta,
		&personalBand,
		&personalAvailable,
		&personalSource,
		&baselineExamCount,
		&dataReliability,
		&updateEligible,
		&decisionState,
		&decisionRecommendation,
		&decisionMessage,
		&decisionCorrelation,
		&decisionAttemptCount,
		&decisionMaxAttempts,
		&lastAttemptAt,
		&diagnosticsJSON,
		&rawResponseJSON,
		&rawResponseAvailable,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ExaminationResultResponse{}, repository.ErrNotFound
		}
		return ExaminationResultResponse{}, err
	}

	result.GeneratedAt = generatedAt.UTC().Format(time.RFC3339)

	if refreshedAt != nil {
		result.BaselineSnapshot.RefreshedAt = refreshedAt.UTC().Format(time.RFC3339)
	}
	if generalDelta != nil {
		result.BaselineSnapshot.General.Delta = *generalDelta
	}
	if generalBand != nil {
		result.BaselineSnapshot.General.Band = *generalBand
	}
	if generalAvailable != nil {
		result.BaselineSnapshot.General.BaselineAvailable = *generalAvailable
	}
	if generalSource != nil {
		result.BaselineSnapshot.General.BaselineSource = *generalSource
	}
	if generalRef != nil {
		result.BaselineSnapshot.General.ReferencePopulationVersion = *generalRef
	}
	if personalDelta != nil {
		result.BaselineSnapshot.Personal.Delta = *personalDelta
	}
	if personalBand != nil {
		result.BaselineSnapshot.Personal.Band = *personalBand
	}
	if personalAvailable != nil {
		result.BaselineSnapshot.Personal.BaselineAvailable = *personalAvailable
	}
	if personalSource != nil {
		result.BaselineSnapshot.Personal.BaselineSource = *personalSource
	}
	if baselineExamCount != nil {
		result.BaselineSnapshot.Personal.BaselineExamCount = *baselineExamCount
	}
	if dataReliability != nil {
		result.BaselineSnapshot.Personal.DataReliability = *dataReliability
	}
	if updateEligible != nil {
		result.BaselineSnapshot.Personal.UpdateEligible = *updateEligible
	}

	if decisionState != nil {
		result.Decision.State = *decisionState
	}
	if decisionRecommendation != nil {
		result.Decision.Recommendation = *decisionRecommendation
	}
	if decisionMessage != nil {
		result.Decision.Message = *decisionMessage
	}
	if decisionCorrelation != nil {
		result.Decision.CorrelationID = *decisionCorrelation
	}
	if decisionAttemptCount != nil {
		result.Decision.AttemptCount = *decisionAttemptCount
	}
	if decisionMaxAttempts != nil {
		result.Decision.MaxAttempts = *decisionMaxAttempts
	}
	if lastAttemptAt != nil {
		value := lastAttemptAt.UTC().Format(time.RFC3339)
		result.Decision.LastAttemptAt = &value
	}
	result.Decision.RawResponseAvailable = rawResponseAvailable
	if len(diagnosticsJSON) > 0 {
		if err := json.Unmarshal(diagnosticsJSON, &result.Decision.Diagnostics); err != nil {
			return ExaminationResultResponse{}, err
		}
	}
	if result.Decision.State == decision.DecisionStateSucceeded && len(rawResponseJSON) > 0 {
		if parsed, err := decision.ParseResultForView(rawResponseJSON); err == nil {
			result.Decision.DecisionCode = parsed.DecisionCode
			result.Decision.RiskClass = parsed.RiskClass
			result.Decision.Patterns = append([]string(nil), parsed.Patterns...)
		}
	}

	metrics, err := r.loadMetrics(ctx, examinationID)
	if err != nil {
		return ExaminationResultResponse{}, err
	}
	result.Metrics = metrics
	contributions, err := r.loadContributions(ctx, examinationID)
	if err != nil {
		return ExaminationResultResponse{}, err
	}
	result.ChannelContributions = contributions
	channelReports, err := r.loadChannelReports(ctx, examinationID)
	if err != nil {
		return ExaminationResultResponse{}, err
	}
	result.ChannelReports = channelReports
	explanations, err := r.loadExplanations(ctx, examinationID)
	if err != nil {
		return ExaminationResultResponse{}, err
	}
	result.Explanations = explanations

	return result, nil
}

func (r *SQLRepository) loadChannelReports(ctx context.Context, examinationID int64) ([]ChannelReport, error) {
	const query = `
SELECT channel, payload
FROM channel_results
WHERE examination_id = $1
  AND status = 'succeeded'
ORDER BY completed_at DESC`

	rows, err := r.pool.Query(ctx, query, examinationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	reportByChannel := map[string]ChannelReport{}
	for rows.Next() {
		var channel string
		var raw []byte
		if err := rows.Scan(&channel, &raw); err != nil {
			return nil, err
		}
		if _, exists := reportByChannel[channel]; exists {
			continue
		}
		var payload processing.CanonicalChannelPayload
		if err := json.Unmarshal(raw, &payload); err != nil {
			return nil, fmt.Errorf("decode channel report payload %s: %w", channel, err)
		}
		reportByChannel[channel] = buildChannelReport(payload)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	items := make([]ChannelReport, 0, len(reportByChannel))
	for _, item := range reportByChannel {
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool {
		left, leftKnown := channelOrder[items[i].Channel]
		right, rightKnown := channelOrder[items[j].Channel]
		switch {
		case leftKnown && rightKnown:
			return left < right
		case leftKnown:
			return true
		case rightKnown:
			return false
		default:
			return items[i].Channel < items[j].Channel
		}
	})
	return items, nil
}

func buildChannelReport(payload processing.CanonicalChannelPayload) ChannelReport {
	scoreKeys := make([]string, 0, len(payload.Scores))
	for key := range payload.Scores {
		scoreKeys = append(scoreKeys, key)
	}
	sort.Slice(scoreKeys, func(i, j int) bool {
		return scoreSortOrder(scoreKeys[i]) < scoreSortOrder(scoreKeys[j])
	})

	scores := make([]ChannelReportScore, 0, len(scoreKeys))
	for _, key := range scoreKeys {
		value, ok := payload.Scores[key].(float64)
		if !ok {
			continue
		}
		scores = append(scores, ChannelReportScore{
			Key:   key,
			Label: scoreLabel(key),
			Value: value,
		})
	}

	return ChannelReport{
		Channel:      payload.Channel,
		ModelVersion: payload.ModelVersion,
		QualityFlags: append([]string(nil), payload.QualityFlags...),
		Evidence:     append([]string(nil), payload.Evidence...),
		Scores:       scores,
	}
}

func scoreLabel(key string) string {
	if label, ok := scoreLabelByKey[key]; ok {
		return label
	}
	return key
}

func scoreSortOrder(key string) int {
	order := []string{
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
	}
	index := slices.Index(order, key)
	if index >= 0 {
		return index
	}
	if key == "" {
		return len(order)
	}
	return len(order) + int(key[0])
}

func (r *SQLRepository) GetSpecialistHistory(ctx context.Context, specialistID int64) (SpecialistHistoryResponse, error) {
	const specialistQuery = `SELECT 1 FROM specialists WHERE id = $1`
	if err := r.pool.QueryRow(ctx, specialistQuery, specialistID).Scan(new(int)); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return SpecialistHistoryResponse{}, repository.ErrNotFound
		}
		return SpecialistHistoryResponse{}, err
	}

	const query = `
SELECT
	p.examination_id,
	p.generated_at,
	p.status,
	p.overall_score,
	p.overall_band,
	COALESCE(b.algorithm_version, ''),
	b.refreshed_at,
	COALESCE(b.general_delta, 0),
	COALESCE(b.personal_delta, 0),
	COALESCE(b.baseline_exam_count, 0)
FROM aggregated_examination_profiles p
LEFT JOIN examination_baseline_snapshots b
	ON b.profile_id = p.id
WHERE p.specialist_id = $1
  AND p.status = 'aggregated'
ORDER BY p.generated_at DESC, p.id DESC`
	rows, err := r.pool.Query(ctx, query, specialistID)
	if err != nil {
		return SpecialistHistoryResponse{}, err
	}
	defer rows.Close()

	items := make([]SpecialistHistoryItem, 0)
	previousMetrics := map[string]float64{}
	for rows.Next() {
		var item SpecialistHistoryItem
		var generatedAt time.Time
		var refreshedAt *time.Time
		if err := rows.Scan(
			&item.ExaminationID,
			&generatedAt,
			&item.Status,
			&item.Summary.OverallScore,
			&item.Summary.OverallBand,
			&item.BaselineSnapshot.AlgorithmVersion,
			&refreshedAt,
			&item.BaselineSnapshot.GeneralDelta,
			&item.BaselineSnapshot.PersonalDelta,
			&item.BaselineSnapshot.BaselineExamCount,
		); err != nil {
			return SpecialistHistoryResponse{}, err
		}
		item.GeneratedAt = generatedAt.UTC().Format(time.RFC3339)
		if refreshedAt != nil {
			item.BaselineSnapshot.RefreshedAt = refreshedAt.UTC().Format(time.RFC3339)
		}

		metrics, err := r.loadMetrics(ctx, item.ExaminationID)
		if err != nil {
			return SpecialistHistoryResponse{}, err
		}
		item.KeyMetrics = make([]HistoryMetric, 0, len(metrics))
		for _, metric := range metrics {
			record := HistoryMetric{Key: metric.Key, Label: metric.Label, Value: metric.Value}
			if previous, ok := previousMetrics[metric.Key]; ok {
				prev := previous
				delta := metric.Value - previous
				record.PreviousValue = &prev
				record.DeltaFromPrevious = &delta
			}
			previousMetrics[metric.Key] = metric.Value
			item.KeyMetrics = append(item.KeyMetrics, record)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return SpecialistHistoryResponse{}, err
	}

	return SpecialistHistoryResponse{SpecialistID: specialistID, Items: items}, nil
}

func (r *SQLRepository) loadMetrics(ctx context.Context, examinationID int64) ([]ExaminationMetric, error) {
	const query = `
SELECT m.metric_key, m.label, m.value, m.scale, m.direction
FROM aggregated_profile_metrics m
JOIN aggregated_examination_profiles p
	ON p.id = m.profile_id
WHERE p.examination_id = $1
ORDER BY m.metric_key ASC`
	rows, err := r.pool.Query(ctx, query, examinationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]ExaminationMetric, 0)
	for rows.Next() {
		var item ExaminationMetric
		if err := rows.Scan(&item.Key, &item.Label, &item.Value, &item.Scale, &item.Direction); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *SQLRepository) loadContributions(ctx context.Context, examinationID int64) ([]ChannelContribution, error) {
	const query = `
SELECT c.channel, c.metric_key, c.weight, c.contribution, c.evidence_keys
FROM aggregated_profile_channel_contributions c
JOIN aggregated_examination_profiles p
	ON p.id = c.profile_id
WHERE p.examination_id = $1
ORDER BY c.contribution DESC, c.channel ASC`
	rows, err := r.pool.Query(ctx, query, examinationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]ChannelContribution, 0)
	for rows.Next() {
		var item ChannelContribution
		var evidence []byte
		if err := rows.Scan(&item.Channel, &item.MetricKey, &item.Weight, &item.Contribution, &evidence); err != nil {
			return nil, err
		}
		if len(evidence) > 0 {
			if err := json.Unmarshal(evidence, &item.EvidenceKeys); err != nil {
				return nil, err
			}
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *SQLRepository) loadExplanations(ctx context.Context, examinationID int64) ([]Explanation, error) {
	const query = `
SELECT e.position, e.kind, e.text
FROM aggregated_profile_explanations e
JOIN aggregated_examination_profiles p
	ON p.id = e.profile_id
WHERE p.examination_id = $1
ORDER BY e.position ASC`
	rows, err := r.pool.Query(ctx, query, examinationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]Explanation, 0)
	for rows.Next() {
		var item Explanation
		if err := rows.Scan(&item.Position, &item.Kind, &item.Text); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
