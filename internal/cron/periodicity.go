package cron

import (
	"context"
	"log/slog"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type Repository interface {
	ListPeriodicities(ctx context.Context) ([]taskdomain.Periodicity, error)
	TouchPeriodicityLastUsage(ctx context.Context, id int64, lastUsage time.Time) error
	SetTaskStatus(ctx context.Context, taskID int64, status taskdomain.Status, updatedAt time.Time) error
}

type PeriodicityCron struct {
	repo   Repository
	logger *slog.Logger
	now    func() time.Time
}

func New(repo Repository, logger *slog.Logger) *PeriodicityCron {
	return &PeriodicityCron{
		repo:   repo,
		logger: logger,
		now:    func() time.Time { return time.Now().UTC() },
	}
}

func (c *PeriodicityCron) Start(ctx context.Context) {
	go func() {
		c.Run(ctx)

		for {
			next := nextMidnight(c.now())
			timer := time.NewTimer(time.Until(next))

			select {
			case <-ctx.Done():
				timer.Stop()
				return
			case <-timer.C:
				c.Run(ctx)
			}
		}
	}()
}

func (c *PeriodicityCron) Run(ctx context.Context) {
	items, err := c.repo.ListPeriodicities(ctx)
	if err != nil {
		c.logger.Error("periodicity cron: list", "error", err)
		return
	}

	today := truncateToDate(c.now())

	for i := range items {
		p := items[i]
		if !shouldTrigger(&p, today) {
			continue
		}

		if err := c.repo.SetTaskStatus(ctx, p.TaskID, taskdomain.StatusNew, c.now()); err != nil {
			c.logger.Error("periodicity cron: set task status", "task_id", p.TaskID, "error", err)
			continue
		}

		if err := c.repo.TouchPeriodicityLastUsage(ctx, p.ID, today); err != nil {
			c.logger.Error("periodicity cron: touch last_usage", "periodicity_id", p.ID, "error", err)
			continue
		}

		c.logger.Info("periodicity cron: task moved to new", "task_id", p.TaskID, "periodicity_id", p.ID)
	}
}

func shouldTrigger(p *taskdomain.Periodicity, today time.Time) bool {
	next, ok := nextExpectedDate(p, truncateToDate(p.LastUsage))
	if !ok {
		return false
	}
	return !today.Before(next)
}

func nextExpectedDate(p *taskdomain.Periodicity, last time.Time) (time.Time, bool) {
	switch {
	case p.Daily != nil:
		return last.AddDate(0, 0, *p.Daily), true

	case p.Monthly != nil:
		year, month, day := last.Date()
		target := clampDay(*p.Monthly, year, month)
		candidate := time.Date(year, month, target, 0, 0, 0, 0, last.Location())
		if target <= day {
			nextMonth := time.Date(year, month+1, 1, 0, 0, 0, 0, last.Location())
			target = clampDay(*p.Monthly, nextMonth.Year(), nextMonth.Month())
			candidate = time.Date(nextMonth.Year(), nextMonth.Month(), target, 0, 0, 0, 0, last.Location())
		}
		return candidate, true

	case p.Dates != nil:
		var (
			best  time.Time
			found bool
		)
		for _, d := range p.Dates {
			d := truncateToDate(d)
			if !d.After(last) {
				continue
			}
			if !found || d.Before(best) {
				best = d
				found = true
			}
		}
		return best, found

	case p.IsEven != nil:
		candidate := last.AddDate(0, 0, 1)
		if (candidate.Day()%2 == 0) != *p.IsEven {
			candidate = candidate.AddDate(0, 0, 1)
		}
		return candidate, true
	}

	return time.Time{}, false
}

func daysInMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

func clampDay(day, year int, month time.Month) int {
	if max := daysInMonth(year, month); day > max {
		return max
	}
	return day
}

func truncateToDate(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

func nextMidnight(now time.Time) time.Time {
	y, m, d := now.Date()
	return time.Date(y, m, d+1, 0, 0, 0, 0, now.Location())
}
