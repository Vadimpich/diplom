package processing

import (
	"context"

	"dimplom/internal/examinations"
)

type FinishRepository interface {
	FinishLaunch(context.Context, int64) (examinations.Examination, []ProcessingCommandEnvelope, error)
}

type Service struct {
	repo FinishRepository
}

func NewService(repo FinishRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Finish(ctx context.Context, examinationID int64) (examinations.Examination, error) {
	exam, _, err := s.repo.FinishLaunch(ctx, examinationID)
	if err != nil {
		return examinations.Examination{}, err
	}
	return exam, nil
}
