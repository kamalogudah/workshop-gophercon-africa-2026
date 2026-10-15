-- +goose Up
-- +goose StatementBegin
-- The site lists sightings newest first, by country and by species, so
-- each of those gets an index rather than a scan of every row.
CREATE INDEX sightings_happened_at ON sightings (happened_at DESC, id DESC);
CREATE INDEX sightings_country_code ON sightings (country_code, happened_at DESC);
CREATE INDEX sightings_species_id ON sightings (species_id);
CREATE INDEX species_genus_key ON species (genus_key);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS species_genus_key;
DROP INDEX IF EXISTS sightings_species_id;
DROP INDEX IF EXISTS sightings_country_code;
DROP INDEX IF EXISTS sightings_happened_at;
-- +goose StatementEnd
