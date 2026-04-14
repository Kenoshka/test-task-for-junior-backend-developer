package task

import (
	"context"
	"fmt"
	"strings"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type Service struct {
	repo Repository
	now  func() time.Time
}

func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
		now:  func() time.Time { return time.Now().UTC() },
	}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (*taskdomain.Task, error) {
	normalized, err := validateCreateInput(input)
	if err != nil {
		return nil, err
	}

	model := &taskdomain.Task{
		Title:        normalized.Title,
		Description:  normalized.Description,
		Status:       normalized.Status,
		IsPeriodical: normalized.IsPeriodical,
	}
	now := s.now()
	model.CreatedAt = now
	model.UpdatedAt = now

	created, err := s.repo.Create(ctx, model)
	if err != nil {
		return nil, err
	}

	if normalized.IsPeriodical {
		periodicity := buildPeriodicity(created.ID, normalized.Periodicity)
		if _, err := s.repo.UpsertPeriodicity(ctx, periodicity); err != nil {
			return nil, err
		}
	}

	return created, nil
}

func (s *Service) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	return s.repo.GetByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, id int64, input UpdateInput) (*taskdomain.Task, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	normalized, err := validateUpdateInput(input)
	if err != nil {
		return nil, err
	}

	model := &taskdomain.Task{
		ID:           id,
		Title:        normalized.Title,
		Description:  normalized.Description,
		Status:       normalized.Status,
		IsPeriodical: normalized.IsPeriodical,
		UpdatedAt:    s.now(),
	}

	updated, err := s.repo.Update(ctx, model)
	if err != nil {
		return nil, err
	}

	if normalized.IsPeriodical {
		periodicity := buildPeriodicity(updated.ID, normalized.Periodicity)
		if _, err := s.repo.UpsertPeriodicity(ctx, periodicity); err != nil {
			return nil, err
		}
	} else {
		if err := s.repo.DeletePeriodicity(ctx, updated.ID); err != nil {
			return nil, err
		}
	}

	return updated, nil
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	return s.repo.Delete(ctx, id)
}

func (s *Service) List(ctx context.Context) ([]taskdomain.Task, error) {
	return s.repo.List(ctx)
}

func validateCreateInput(input CreateInput) (CreateInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	if input.Title == "" {
		return CreateInput{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	if input.Status == "" {
		input.Status = taskdomain.StatusNew
	}

	if !input.Status.Valid() {
		return CreateInput{}, fmt.Errorf("%w: invalid status", ErrInvalidInput)
	}

	if err := validatePeriodicityInput(input.IsPeriodical, input.Periodicity); err != nil {
		return CreateInput{}, err
	}

	return input, nil
}

func validateUpdateInput(input UpdateInput) (UpdateInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	if input.Title == "" {
		return UpdateInput{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	if !input.Status.Valid() {
		return UpdateInput{}, fmt.Errorf("%w: invalid status", ErrInvalidInput)
	}

	if err := validatePeriodicityInput(input.IsPeriodical, input.Periodicity); err != nil {
		return UpdateInput{}, err
	}

	return input, nil
}

func validatePeriodicityInput(isPeriodical bool, p *PeriodicityInput) error {
	if !isPeriodical {
		return nil
	}
	if p == nil {
		return fmt.Errorf("%w: periodicity is required when is_periodical is true", ErrInvalidInput)
	}

	domainP := &taskdomain.Periodicity{
		Daily:   p.Daily,
		Monthly: p.Monthly,
		Dates:   p.Dates,
		IsEven:  p.IsEven,
	}
	if err := domainP.Validate(); err != nil {
		return fmt.Errorf("%w: %s", ErrInvalidInput, err.Error())
	}
	return nil
}

func buildPeriodicity(taskID int64, p *PeriodicityInput) *taskdomain.Periodicity {
	return &taskdomain.Periodicity{
		TaskID:  taskID,
		Daily:   p.Daily,
		Monthly: p.Monthly,
		Dates:   p.Dates,
		IsEven:  p.IsEven,
	}
}
