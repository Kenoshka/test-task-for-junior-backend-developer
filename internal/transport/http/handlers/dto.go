package handlers

import (
	"fmt"
	"strings"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
	taskusecase "example.com/taskservice/internal/usecase/task"
)

const dateLayout = "2006-01-02"

type jsonDate struct {
	time.Time
}

func (d *jsonDate) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), `"`)
	if s == "" || s == "null" {
		return nil
	}
	t, err := time.Parse(dateLayout, s)
	if err != nil {
		return fmt.Errorf("invalid date %q, expected YYYY-MM-DD", s)
	}
	d.Time = t
	return nil
}

func (d jsonDate) MarshalJSON() ([]byte, error) {
	return []byte(`"` + d.Time.Format(dateLayout) + `"`), nil
}

type periodicityDTO struct {
	Daily   *int       `json:"daily,omitempty"`
	Monthly *int       `json:"monthly,omitempty"`
	Dates   []jsonDate `json:"dates,omitempty"`
	IsEven  *bool      `json:"is_even,omitempty"`
}

type taskMutationDTO struct {
	Title        string            `json:"title"`
	Description  string            `json:"description"`
	Status       taskdomain.Status `json:"status"`
	IsPeriodical bool              `json:"is_periodical"`
	Periodicity  *periodicityDTO   `json:"periodicity,omitempty"`
}

type periodicityResponseDTO struct {
	ID        int64      `json:"id"`
	TaskID    int64      `json:"task_id"`
	Daily     *int       `json:"daily,omitempty"`
	Monthly   *int       `json:"monthly,omitempty"`
	Dates     []jsonDate `json:"dates,omitempty"`
	IsEven    *bool      `json:"is_even,omitempty"`
	LastUsage jsonDate   `json:"last_usage"`
}

type taskDTO struct {
	ID           int64                   `json:"id"`
	Title        string                  `json:"title"`
	Description  string                  `json:"description"`
	Status       taskdomain.Status       `json:"status"`
	IsPeriodical bool                    `json:"is_periodical"`
	CreatedAt    time.Time               `json:"created_at"`
	UpdatedAt    time.Time               `json:"updated_at"`
	Periodicity  *periodicityResponseDTO `json:"periodicity,omitempty"`
}

func toPeriodicityInput(p *periodicityDTO) *taskusecase.PeriodicityInput {
	if p == nil {
		return nil
	}
	var dates []time.Time
	if p.Dates != nil {
		dates = make([]time.Time, len(p.Dates))
		for i, d := range p.Dates {
			dates[i] = d.Time
		}
	}
	return &taskusecase.PeriodicityInput{
		Daily:   p.Daily,
		Monthly: p.Monthly,
		Dates:   dates,
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
		Periodicity:  newPeriodicityResponseDTO(task.Periodicity),
	}
}

func newPeriodicityResponseDTO(p *taskdomain.Periodicity) *periodicityResponseDTO {
	if p == nil {
		return nil
	}
	var dates []jsonDate
	if p.Dates != nil {
		dates = make([]jsonDate, len(p.Dates))
		for i, d := range p.Dates {
			dates[i] = jsonDate{Time: d}
		}
	}
	return &periodicityResponseDTO{
		ID:        p.ID,
		TaskID:    p.TaskID,
		Daily:     p.Daily,
		Monthly:   p.Monthly,
		Dates:     dates,
		IsEven:    p.IsEven,
		LastUsage: jsonDate{Time: p.LastUsage},
	}
}
