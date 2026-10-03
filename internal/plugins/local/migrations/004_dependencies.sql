-- Task dependency edges: task X is blocked by task Y when a row
-- (task_id = X, depends_on_id = Y) exists. Blocked/ready is computed at
-- read time from these edges plus task statuses — nothing is stored, so
-- it can never go stale across the processes sharing this database.
--
-- Both foreign keys CASCADE: deleting a task removes its edges in both
-- directions (contrast tasks.parent_id, which SETs NULL — a subtask
-- survives its parent, but a dependency is meaningless without its
-- blocker). The CHECK makes self-dependencies unrepresentable.

CREATE TABLE task_dependencies (
  task_id       TEXT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
  depends_on_id TEXT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
  created_at    TEXT NOT NULL,
  PRIMARY KEY (task_id, depends_on_id),
  CHECK (task_id <> depends_on_id)
);

CREATE INDEX idx_task_dependencies_depends_on ON task_dependencies(depends_on_id);
