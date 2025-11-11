CREATE TABLE IF NOT EXISTS users (
	id   TEXT PRIMARY KEY,
	role TEXT NOT NULL,
	pswd TEXT NOT NULL,
	user_name TEXT NOT NULL UNIQUE
);

CREATE INDEX IF NOT EXISTS idx_id ON users(id);
CREATE INDEX IF NOT EXISTS idx_user_name ON users(user_name);
