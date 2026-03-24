package processing

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"

	"diplom/internal/audit"
	"diplom/internal/examinations"
)

var ErrProcessingStatusUnavailable = errors.New("processing: status unavailable")

type FinishRepository interface {
	FinishLaunch(context.Context, int64) (examinations.Examination, []ProcessingCommandEnvelope, error)
	GetProcessingStatus(context.Context, int64) (ProcessingStatusResponse, error)
}

type Service struct {
	repo    FinishRepository
	auditor *audit.Service
}

func NewService(repo FinishRepository, auditors ...*audit.Service) *Service {
	var auditor *audit.Service
	if len(auditors) > 0 {
		auditor = auditors[0]
	}
	return &Service{repo: repo, auditor: auditor}
}

func (s *Service) Finish(ctx context.Context, examinationID int64) (examinations.Examination, error) {
	exam, commands, err := s.repo.FinishLaunch(ctx, examinationID)
	if err != nil {
		return examinations.Examination{}, err
	}
	s.appendAudit(ctx, audit.Event{
		Type:          audit.EventTypeProcessingLaunch,
		Key:           "processing-launch:examination:" + strconv.FormatInt(exam.ID, 10),
		Outcome:       audit.OutcomeSucceeded,
		CorrelationID: firstCorrelation(commands),
		Resource: audit.ResourceRef{
			Kind: "examination",
			ID:   exam.ID,
		},
		DomainRefs: audit.DomainRefs{
			ExaminationID: &exam.ID,
			SpecialistID:  &exam.SpecialistID,
		},
		Payload: processingLaunchPayload(commands),
	})
	return exam, nil
}

func (s *Service) GetStatus(ctx context.Context, examinationID int64) (ProcessingStatusResponse, error) {
	return s.repo.GetProcessingStatus(ctx, examinationID)
}

func (s *Service) appendAudit(ctx context.Context, event audit.Event) {
	if s.auditor == nil {
		return
	}
	_ = s.auditor.AppendFromContext(ctx, event)
}

func firstCorrelation(commands []ProcessingCommandEnvelope) string {
	if len(commands) == 0 {
		return ""
	}
	return commands[0].CorrelationID
}

func processingLaunchPayload(commands []ProcessingCommandEnvelope) json.RawMessage {
	channels := make([]string, 0, len(commands))
	for _, command := range commands {
		channels = append(channels, command.Channel)
	}
	data, _ := json.Marshal(map[string]any{"channels": channels})
	return data
}
