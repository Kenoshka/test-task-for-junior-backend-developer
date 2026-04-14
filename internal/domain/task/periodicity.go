package task

import (
	"errors"
	"time"
)

var ErrInvalidPeriodicity = errors.New("invalid periodicity")

type Periodicity struct {
	ID      int64       `json:"id"`
	TaskID  int64       `json:"task_id"`
	Daily   *int        `json:"daily,omitempty"`
	Monthly *int        `json:"monthly,omitempty"`
	Dates   []time.Time `json:"dates,omitempty"`
	IsEven  *bool       `json:"is_even,omitempty"`
	LastUsage time.Time `json:"last_usage"`
}

func (p *Periodicity) Validate() error {
	set := 0
	if p.Daily != nil {
		set++
	}
	if p.Monthly != nil {
		set++
	}
	if p.Dates != nil {
		set++
	}
	if p.IsEven != nil {
		set++
	}
	if set != 1 {
		return errors.New("periodicity must have exactly one of: daily, monthly, dates, is_even")
	}

	if p.Daily != nil && *p.Daily < 1 {
		return errors.New("daily must be >= 1")
	}
	if p.Monthly != nil && (*p.Monthly < 1 || *p.Monthly > 31) {
		return errors.New("monthly must be between 1 and 31")
	}
	if p.Dates != nil && len(p.Dates) == 0 {
		return errors.New("dates must contain at least one date")
	}

	return nil
}
