package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type Repository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
	const query = `
		INSERT INTO tasks (title, description, status, is_periodical, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, title, description, status, is_periodical, created_at, updated_at
	`

	row := r.pool.QueryRow(ctx, query, task.Title, task.Description, task.Status, task.IsPeriodical, task.CreatedAt, task.UpdatedAt)
	created, err := scanTask(row)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (r *Repository) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	const query = `
		SELECT id, title, description, status, is_periodical, created_at, updated_at
		FROM tasks
		WHERE id = $1
	`

	row := r.pool.QueryRow(ctx, query, id)
	found, err := scanTask(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, taskdomain.ErrNotFound
		}

		return nil, err
	}

	if found.IsPeriodical {
		periodicity, err := r.getPeriodicityByTaskID(ctx, found.ID)
		if err != nil {
			return nil, err
		}
		found.Periodicity = periodicity
	}

	return found, nil
}

func (r *Repository) getPeriodicityByTaskID(ctx context.Context, taskID int64) (*taskdomain.Periodicity, error) {
	const query = `
		SELECT id, task_id, daily, monthly, dates, is_even, last_usage
		FROM periodicities
		WHERE task_id = $1
	`

	row := r.pool.QueryRow(ctx, query, taskID)

	var p taskdomain.Periodicity
	if err := row.Scan(&p.ID, &p.TaskID, &p.Daily, &p.Monthly, &p.Dates, &p.IsEven, &p.LastUsage); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &p, nil
}

func (r *Repository) Update(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
	const query = `
		UPDATE tasks
		SET title = $1,
			description = $2,
			status = $3,
			is_periodical = $4,
			updated_at = $5
		WHERE id = $6
		RETURNING id, title, description, status, is_periodical, created_at, updated_at
	`

	row := r.pool.QueryRow(ctx, query, task.Title, task.Description, task.Status, task.IsPeriodical, task.UpdatedAt, task.ID)
	updated, err := scanTask(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, taskdomain.ErrNotFound
		}

		return nil, err
	}

	return updated, nil
}

func (r *Repository) Delete(ctx context.Context, id int64) error {
	const query = `DELETE FROM tasks WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return taskdomain.ErrNotFound
	}

	return nil
}

func (r *Repository) List(ctx context.Context) ([]taskdomain.Task, error) {
	const query = `
		SELECT id, title, description, status, is_periodical, created_at, updated_at
		FROM tasks
		ORDER BY id DESC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := make([]taskdomain.Task, 0)
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return nil, err
		}

		tasks = append(tasks, *task)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

func (r *Repository) UpsertPeriodicity(ctx context.Context, p *taskdomain.Periodicity) (*taskdomain.Periodicity, error) {
	const query = `
		INSERT INTO periodicities (task_id, daily, monthly, dates, is_even)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (task_id) DO UPDATE SET
			daily = EXCLUDED.daily,
			monthly = EXCLUDED.monthly,
			dates = EXCLUDED.dates,
			is_even = EXCLUDED.is_even
		RETURNING id, task_id, daily, monthly, dates, is_even, last_usage
	`

	row := r.pool.QueryRow(ctx, query, p.TaskID, p.Daily, p.Monthly, p.Dates, p.IsEven)

	var out taskdomain.Periodicity
	if err := row.Scan(&out.ID, &out.TaskID, &out.Daily, &out.Monthly, &out.Dates, &out.IsEven, &out.LastUsage); err != nil {
		return nil, err
	}

	return &out, nil
}

func (r *Repository) ListPeriodicities(ctx context.Context) ([]taskdomain.Periodicity, error) {
	const query = `
		SELECT id, task_id, daily, monthly, dates, is_even, last_usage
		FROM periodicities
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]taskdomain.Periodicity, 0)
	for rows.Next() {
		var p taskdomain.Periodicity
		if err := rows.Scan(&p.ID, &p.TaskID, &p.Daily, &p.Monthly, &p.Dates, &p.IsEven, &p.LastUsage); err != nil {
			return nil, err
		}
		out = append(out, p)
	}

	return out, rows.Err()
}

func (r *Repository) TouchPeriodicityLastUsage(ctx context.Context, id int64, lastUsage time.Time) error {
	const query = `UPDATE periodicities SET last_usage = $2 WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id, lastUsage)
	return err
}

func (r *Repository) SetTaskStatus(ctx context.Context, taskID int64, status taskdomain.Status, updatedAt time.Time) error {
	const query = `UPDATE tasks SET status = $2, updated_at = $3 WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, taskID, status, updatedAt)
	return err
}

func (r *Repository) DeletePeriodicity(ctx context.Context, taskID int64) error {
	const query = `DELETE FROM periodicities WHERE task_id = $1`

	if _, err := r.pool.Exec(ctx, query, taskID); err != nil {
		return err
	}

	return nil
}

type taskScanner interface {
	Scan(dest ...any) error
}

func scanTask(scanner taskScanner) (*taskdomain.Task, error) {
	var (
		task   taskdomain.Task
		status string
	)

	if err := scanner.Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&status,
		&task.IsPeriodical,
		&task.CreatedAt,
		&task.UpdatedAt,
	); err != nil {
		return nil, err
	}

	task.Status = taskdomain.Status(status)

	return &task, nil
}
