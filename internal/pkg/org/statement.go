package org

const getByIDQuery = `
	SELECT
		o.org_id,
		o.name,
		o.description,
		o.is_archived,
		o.created_at,
		o.updated_at,
		o.version
	FROM orgs o
	WHERE o.org_id = $1
`

const getAllQuery = `
	SELECT
		o.org_id,
		o.name,
		o.description,
		o.is_archived,
		o.created_at,
		o.updated_at,
		o.version
	FROM orgs o
	ORDER BY o.name ASC, o.created_at DESC
`

const searchByNameQuery = `
	SELECT
		o.org_id,
		o.name,
		o.description,
		o.is_archived,
		o.created_at,
		o.updated_at,
		o.version
	FROM orgs o
	WHERE LOWER(o.name) LIKE $1
	ORDER BY o.name ASC, o.created_at DESC
`

const createQuery = `
	INSERT INTO orgs (
		org_id,
		name,
		description,
		is_archived,
		created_at,
		updated_at,
		version
	) VALUES (
		:org_id,
		:name,
		:description,
		:is_archived,
		:created_at,
		:updated_at,
		:version
	)
`

const updateQuery = `
	UPDATE orgs SET
		name = :name,
		description = :description,
		is_archived = :is_archived,
		updated_at = :updated_at,
		version = 1 + :version
	WHERE org_id = :org_id
	AND version = :version
`

const deleteQuery = `
	DELETE FROM orgs
	WHERE org_id = :org_id
	AND version = :version
`
