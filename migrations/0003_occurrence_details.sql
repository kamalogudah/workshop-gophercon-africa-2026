-- +goose Up
-- +goose StatementBegin
-- More of what GBIF says about each occurrence: where in the world it
-- was, who the animal was, and who vouched for the record. Rows saved
-- before this migration read as empty until `workshop sightings enrich`
-- fetches them again.
ALTER TABLE sightings ADD COLUMN continent     TEXT NOT NULL DEFAULT '';
ALTER TABLE sightings ADD COLUMN sex           TEXT NOT NULL DEFAULT '';
ALTER TABLE sightings ADD COLUMN life_stage    TEXT NOT NULL DEFAULT '';
ALTER TABLE sightings ADD COLUMN behaviour     TEXT NOT NULL DEFAULT '';
ALTER TABLE sightings ADD COLUMN identified_by TEXT NOT NULL DEFAULT '';
ALTER TABLE sightings ADD COLUMN dataset_name  TEXT NOT NULL DEFAULT '';

-- An animal such as "giraffe" names a genus, while a species is stored
-- under its own key, so the genus is kept to match the two up.
ALTER TABLE species ADD COLUMN genus_key INTEGER NOT NULL DEFAULT 0;

-- The IUCN Red List category is a property of the species, though GBIF
-- reports it on every occurrence.
ALTER TABLE species ADD COLUMN iucn_category TEXT NOT NULL DEFAULT ''
	CHECK (iucn_category IN ('', 'EX', 'EW', 'CR', 'EN', 'VU', 'NT', 'LC', 'DD', 'NE'));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE species DROP COLUMN iucn_category;
ALTER TABLE species DROP COLUMN genus_key;
ALTER TABLE sightings DROP COLUMN dataset_name;
ALTER TABLE sightings DROP COLUMN identified_by;
ALTER TABLE sightings DROP COLUMN behaviour;
ALTER TABLE sightings DROP COLUMN life_stage;
ALTER TABLE sightings DROP COLUMN sex;
ALTER TABLE sightings DROP COLUMN continent;
-- +goose StatementEnd
