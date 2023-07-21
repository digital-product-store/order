CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS orders (
	id uuid DEFAULT uuid_generate_v4(),
	status VARCHAR NOT NULL,
	payment_id uuid,
	user_id uuid NOT NULL,
	total DECIMAL NOT NULL,
	items JSONB NOT NULL,
	PRIMARY KEY (id)
);

CREATE INDEX IF NOT EXISTS orders_user_id_idx ON orders (user_id);
