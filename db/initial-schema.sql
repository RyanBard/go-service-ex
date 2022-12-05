CREATE TABLE orgs(
	org_id TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	description TEXT,
	is_archived BOOLEAN,
	created_at TIMESTAMP NOT NULL,
	updated_at TIMESTAMP NOT NULL,
	version BIGINT NOT NULL
);
