CREATE TABLE IF NOT EXISTS task_sagas (
  id BIGSERIAL PRIMARY KEY,
  task_id BIGINT NOT NULL,
  user_id BIGINT NOT NULL,
  saga_type VARCHAR(64) NOT NULL,
  status VARCHAR(32) NOT NULL,
  current_step VARCHAR(128) NOT NULL,
  compensation TEXT,
  last_error TEXT,
  next_run_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_task_sagas_task_id ON task_sagas (task_id);
CREATE INDEX IF NOT EXISTS idx_task_sagas_user_id ON task_sagas (user_id);
CREATE INDEX IF NOT EXISTS idx_task_sagas_type_status ON task_sagas (saga_type, status);
CREATE INDEX IF NOT EXISTS idx_task_sagas_next_run_at ON task_sagas (next_run_at);
