package handlers

import (
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
	taskusecase "example.com/taskservice/internal/usecase/task"
)

type periodicityDTO struct {
	Daily   *int        `json:"daily,omitempty"`
	Monthly *int        `json:"monthly,omitempty"`
	Dates   []time.Time `json:"dates,omitempty"`
	IsEven  *bool       `json:"is_even,omitempty"`
}

type taskMutationDTO struct {
	Title        string            `json:"title"`
	Description  string            `json:"description"`
	Status       taskdomain.Status `json:"status"`
	IsPeriodical bool              `json:"is_periodical"`
	Periodicity  *periodicityDTO   `json:"periodicity,omitempty"`
}

type taskDTO struct {
	ID           int64             `json:"id"`
	Title        string            `json:"title"`
	Description  string            `json:"description"`
	Status       taskdomain.Status `json:"status"`
	IsPeriodical bool              `json:"is_periodical"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
}

func toPeriodicityInput(p *periodicityDTO) *taskusecase.PeriodicityInput {
	if p == nil {
		return nil
	}
	return &taskusecase.PeriodicityInput{
		Daily:   p.Daily,
		Monthly: p.Monthly,
		Dates:   p.Dates,
		IsEven:  p.IsEven,
	}
}

func newTaskDTO(task *taskdomain.Task) taskDTO {
	return taskDTO{
		ID:           task.ID,
		Title:        task.Title,
		Description:  task.Description,
		Status:       task.Status,
		IsPeriodical: task.IsPeriodical,
		CreatedAt:    task.CreatedAt,
		UpdatedAt:    task.UpdatedAt,
	}
}
