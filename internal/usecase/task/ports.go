package task

import (
	"context"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type Repository interface {
	Create(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error)
	GetByID(ctx context.Context, id int64) (*taskdomain.Task, error)
	Update(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context) ([]taskdomain.Task, error)

	UpsertPeriodicity(ctx context.Context, p *taskdomain.Periodicity) (*taskdomain.Periodicity, error)
	DeletePeriodicity(ctx context.Context, taskID int64) error
}

type Usecase interface {
	Create(ctx context.Context, input CreateInput) (*taskdomain.Task, error)
	GetByID(ctx context.Context, id int64) (*taskdomain.Task, error)
	Update(ctx context.Context, id int64, input UpdateInput) (*taskdomain.Task, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context) ([]taskdomain.Task, error)
}

type PeriodicityInput struct {
	Daily   *int
	Monthly *int
	Dates   []time.Time
	IsEven  *bool
}

type CreateInput struct {
	Title        string
	Description  string
	Status       taskdomain.Status
	IsPeriodical bool
	Periodicity  *PeriodicityInput
}

type UpdateInput struct {
	Title        string
	Description  string
	Status       taskdomain.Status
	IsPeriodical bool
	Periodicity  *PeriodicityInput
}
