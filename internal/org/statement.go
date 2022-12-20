package org

const getByIDQuery = `
	SELECT
		o.id,
		o.name,
		o.description,
		o.is_system,
		o.created_at,
		o.created_by,
		o.updated_at,
		o.updated_by,
		o.version
	FROM orgs o
	WHERE o.id = $1
`

const getAllQuery = `
	SELECT
		o.id,
		o.name,
		o.description,
		o.is_system,
		o.created_at,
		o.created_by,
		o.updated_at,
		o.updated_by,
		o.version
	FROM orgs o
	ORDER BY o.name ASC, o.created_at DESC
`

const searchByNameQuery = `
	SELECT
		o.id,
		o.name,
		o.description,
		o.is_system,
		o.created_at,
		o.created_by,
		o.updated_at,
		o.updated_by,
		o.version
	FROM orgs o
	WHERE LOWER(o.name) LIKE $1
	ORDER BY o.name ASC, o.created_at DESC
`

const createQuery = `
	INSERT INTO orgs (
		id,
		name,
		description,
		created_at,
		created_by,
		updated_at,
		updated_by,
		version
	) VALUES (
		:id,
		:name,
		:description,
		:created_at,
		:created_by,
		:updated_at,
		:updated_by,
		:version
	)
`

const updateQuery = `
	UPDATE orgs SET
		name = :name,
		description = :description,
		updated_at = :updated_at,
		updated_by = :updated_by,
		version = 1 + :version
	WHERE id = :id
	AND version = :version
`

const deleteQuery = `
	DELETE FROM orgs
	WHERE id = :id
	AND version = :version
`
