package specialists

import (
	"context"
	"errors"
	"strings"
	"time"
)

var ErrInvalidInput = errors.New("specialists: invalid input")

type Specialist struct {
	ID              int64     `json:"id"`
	FullName        string    `json:"full_name"`
	PersonnelNumber *string   `json:"personnel_number,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type CreateInput struct {
	FullName        string
	PersonnelNumber *string
}

type UpdateInput struct {
	ID              int64
	FullName        string
	PersonnelNumber *string
}

type Repository interface {
	Create(context.Context, CreateInput) (Specialist, error)
	List(context.Context) ([]Specialist, error)
	GetByID(context.Context, int64) (Specialist, error)
	Update(context.Context, UpdateInput) (Specialist, error)
	Delete(context.Context, int64) error
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (Specialist, error) {
	input = normalizeCreate(input)
	if input.FullName == "" {
		return Specialist{}, ErrInvalidInput
	}

	return s.repo.Create(ctx, input)
}

func (s *Service) List(ctx context.Context) ([]Specialist, error) {
	return s.repo.List(ctx)
}

func (s *Service) GetByID(ctx context.Context, id int64) (Specialist, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, input UpdateInput) (Specialist, error) {
	input = normalizeUpdate(input)
	if input.FullName == "" {
		return Specialist{}, ErrInvalidInput
	}

	return s.repo.Update(ctx, input)
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

func normalizeCreate(input CreateInput) CreateInput {
	input.FullName = strings.TrimSpace(input.FullName)
	if input.PersonnelNumber != nil {
		value := strings.TrimSpace(*input.PersonnelNumber)
		if value == "" {
			input.PersonnelNumber = nil
		} else {
			input.PersonnelNumber = &value
		}
	}
	return input
}

func normalizeUpdate(input UpdateInput) UpdateInput {
	return UpdateInput{
		ID:              input.ID,
		FullName:        normalizeCreate(CreateInput{FullName: input.FullName, PersonnelNumber: input.PersonnelNumber}).FullName,
		PersonnelNumber: normalizeCreate(CreateInput{PersonnelNumber: input.PersonnelNumber}).PersonnelNumber,
	}
}
