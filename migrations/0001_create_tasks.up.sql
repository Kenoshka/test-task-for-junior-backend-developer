CREATE TABLE IF NOT EXISTS tasks (
	id BIGSERIAL PRIMARY KEY,
	title TEXT NOT NULL,
	description TEXT NOT NULL DEFAULT '',
	status TEXT NOT NULL,
	is_periodical BOOLEAN NOT NULL DEFAULT FALSE,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks (status);

CREATE TABLE IF NOT EXISTS periodicities (
	id BIGSERIAL PRIMARY KEY,
	task_id BIGINT NOT NULL UNIQUE REFERENCES tasks (id) ON DELETE CASCADE,
	daily INTEGER NULL,
	monthly INTEGER NULL,
	dates DATE[] NULL,
	is_even BOOLEAN NULL,
	last_usage DATE NOT NULL DEFAULT CURRENT_DATE,
	CONSTRAINT periodicities_exactly_one_kind CHECK (
		num_nonnulls(daily, monthly, dates, is_even) = 1
	),
	CONSTRAINT periodicities_daily_positive CHECK (daily IS NULL OR daily >= 1),
	CONSTRAINT periodicities_monthly_range CHECK (monthly IS NULL OR (monthly BETWEEN 1 AND 31))
);
