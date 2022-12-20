INSERT INTO orgs (
	id,
	name,
	description,
	is_system,
	created_at,
	created_by,
	updated_at,
	updated_by,
	version
)
VALUES(
	'a517c24e-9b5f-4e5a-b840-e4f70a74725f',
	'System Org',
	'Root org for the whole system',
	TRUE,
	CURRENT_TIMESTAMP,
	'init-script',
	CURRENT_TIMESTAMP,
	'init-script',
	1
);

INSERT INTO users (
	id,
	org_id,
	name,
	email,
	is_system,
	is_admin,
	is_active,
	created_at,
	created_by,
	updated_at,
	updated_by,
	version
)
VALUES(
	'fc83cf36-bba0-41f0-8125-2ebc03087140',
	'a517c24e-9b5f-4e5a-b840-e4f70a74725f',
	'System Admin',
	'john.ryan.bard@gmail.com',
	TRUE,
	TRUE,
	TRUE,
	CURRENT_TIMESTAMP,
	'init-script',
	CURRENT_TIMESTAMP,
	'init-script',
	1
);
