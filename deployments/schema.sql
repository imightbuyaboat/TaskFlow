CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT now()
);

CREATE TABLE tasks (
    id UUID PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    payload JSONB NOT NULL,
    status TEXT NOT NULL DEFAULT 'queued',
    retries INTEGER DEFAULT 0,
    max_retries INTEGER DEFAULT 3,
    run_at TIMESTAMP DEFAULT now(),
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);

CREATE TABLE task_logs (
    id SERIAL PRIMARY KEY,
    task_id UUID REFERENCES tasks(id),
    message TEXT,
    created_at TIMESTAMP DEFAULT now()
);

CREATE OR REPLACE FUNCTION log_tasks()
RETURNS TRIGGER AS $$
BEGIN
    CASE
        WHEN TG_OP = 'INSERT' THEN
            INSERT INTO task_logs (task_id, message)
            VALUES (NEW.id, 'task ' || NEW.id || ' created');

        WHEN TG_OP = 'UPDATE' THEN
            INSERT INTO task_logs (task_id, message)
            VALUES (
                NEW.id,
                'updated status for task ' || OLD.id ||
                ' from "' || OLD.status || '" to "' || NEW.status || '"'
            );
    END CASE;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE TRIGGER tr_log_tasks
AFTER INSERT OR UPDATE
ON tasks
FOR EACH ROW
EXECUTE FUNCTION log_tasks();