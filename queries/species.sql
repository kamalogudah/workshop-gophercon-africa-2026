-- name: SpeciesFind :one
SELECT * FROM species WHERE id = ?;

-- name: SpeciesFindByGbifKey :one
SELECT * FROM species WHERE gbif_key = ?;

-- name: SpeciesList :many
SELECT * FROM species;

-- name: SpeciesCreate :one
INSERT INTO species (id, "gbif_key", "kingdom", "phylum", "class", "order", "family", "genus", "species", "rank", "scientific_name", "canonical_name", "vernacular_name", "genus_key", "iucn_category")
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) RETURNING *;

-- name: SpeciesEnrich :exec
-- A blank value leaves what is already there.
UPDATE species
SET genus_key     = CASE WHEN CAST(sqlc.arg('genus_key') AS INTEGER) = 0 THEN genus_key ELSE sqlc.arg('genus_key') END,
	iucn_category = CASE WHEN CAST(sqlc.arg('iucn_category') AS TEXT) = '' THEN iucn_category ELSE sqlc.arg('iucn_category') END,
	updated_at    = CURRENT_TIMESTAMP
WHERE id = sqlc.arg('id');
