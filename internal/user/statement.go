package user

const getByIDQuery = `
	SELECT
		u.id,
		u.org_id,
		u.name,
		u.email,
		u.is_system,
		u.is_admin,
		u.is_active,
		u.created_at,
		u.created_by,
		u.updated_at,
		u.updated_by,
		u.version
	FROM users u
	WHERE u.id = $1
`

const getAllQuery = `
	SELECT
		u.id,
		u.org_id,
		u.name,
		u.email,
		u.is_system,
		u.is_admin,
		u.is_active,
		u.created_at,
		u.created_by,
		u.updated_at,
		u.updated_by,
		u.version
	FROM users u
	ORDER BY u.email ASC, u.created_at DESC
`

const getAllByOrgIDQuery = `
	SELECT
		u.id,
		u.org_id,
		u.name,
		u.email,
		u.is_system,
		u.is_admin,
		u.is_active,
		u.created_at,
		u.created_by,
		u.updated_at,
		u.updated_by,
		u.version
	FROM users u
	WHERE u.org_id = $1
	ORDER BY u.email ASC, u.created_at DESC
`

const createQuery = `
	INSERT INTO users (
		id,
		org_id,
		name,
		email,
		is_admin,
		is_active,
		created_at,
		created_by,
		updated_at,
		updated_by,
		version
	) VALUES (
		:id,
		:org_id,
		:name,
		:email,
		:is_admin,
		:is_active,
		:created_at,
		:created_by,
		:updated_at,
		:updated_by,
		:version
	)
`

const updateQuery = `
	UPDATE users SET
		name = :name,
		is_admin = :is_admin,
		is_active = :is_active,
		updated_at = :updated_at,
		updated_by = :updated_by,
		version = 1 + :version
	WHERE id = :id
	AND version = :version
`

const deleteQuery = `
	DELETE FROM users
	WHERE id = :id
	AND version = :version
`
