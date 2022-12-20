CREATE TABLE orgs(
	id TEXT NOT NULL,
	name TEXT NOT NULL,
	description TEXT,
	is_system BOOLEAN NOT NULL DEFAULT FALSE,
	created_at TIMESTAMP NOT NULL,
	created_by TEXT NOT NULL,
	updated_at TIMESTAMP NOT NULL,
	updated_by TEXT NOT NULL,
	version BIGINT NOT NULL,
	CONSTRAINT orgs_pk PRIMARY KEY(id),
	CONSTRAINT orgs_name_uk UNIQUE (name)
);

CREATE TABLE users(
	id TEXT NOT NULL,
	-- TODO - should this be nullable (allow pending users to not be associated with an org)
	org_id TEXT NOT NULL,
	name TEXT NOT NULL,
	email TEXT NOT NULL,
	is_system BOOLEAN NOT NULL DEFAULT FALSE,
	is_admin BOOLEAN NOT NULL DEFAULT FALSE,
	is_active BOOLEAN NOT NULL DEFAULT FALSE,
	created_at TIMESTAMP NOT NULL,
	created_by TEXT NOT NULL,
	updated_at TIMESTAMP NOT NULL,
	updated_by TEXT NOT NULL,
	version BIGINT NOT NULL,
	CONSTRAINT users_pk PRIMARY KEY(id),
	CONSTRAINT users_org_fk FOREIGN KEY (org_id) REFERENCES orgs(id),
	-- TODO - This should probably be tweaked to allow a user to be on multiple orgs
	CONSTRAINT users_email_uk UNIQUE (email)
);
