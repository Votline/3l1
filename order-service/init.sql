CREATE TABLE IF NOT EXISTS orders(
	id TEXT PRIMARY KEY,
	user_id TEXT NOT NULL,
	user_role TEXT NOT NULL,
	status TEXT NOT NULL,
	service_url TEXT NOT NULL,
	target_url TEXT NOT NULL,
	order_type TEXT NOT NULL,
	quantity INTEGER NOT NULL,
	created_at TIMESTAMP DEFAULT NOW(),
	updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_user_id ON orders(user_id);
CREATE INDEX IF NOT EXISTS idx_user_role ON orders(user_role);
CREATE INDEX IF NOT EXISTS idx_id_user_id ON orders(id, user_role);
