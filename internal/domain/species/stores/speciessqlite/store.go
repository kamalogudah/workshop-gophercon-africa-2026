package speciessqlite

import (
	"context"
	"database/sql"
	"errors"
	"uuid"
	"workshop/internal/domain/species"
	"workshop/internal/infra/db/sqlc"
	"workshop/internal/infra/db/sqlite"
)

// Store persists species using sqlc-generated SQLite queries.
type Store struct {
	queries db.Querier
}

var _ species.Store = (*Store)(nil)

// New constructs a species SQLite store.
func New(queries db.Querier) *Store {
	return &Store{queries: queries}
}

// Find returns the species for id, or species.ErrNotFound if it
// doesn't exist.
func (s Store) Find(ctx context.Context, id uuid.UUID) (species.Species, error) {
	row, err := s.queries.SpeciesFind(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return species.Species{}, species.ErrNotFound
	} else if err != nil {
		return species.Species{}, err
	}

	return transform(row), nil
}

// FindByGbifKey returns the species for gbifKey, or
// species.ErrNotFound if it doesn't exist.
func (s Store) FindByGbifKey(ctx context.Context, gbifKey int) (species.Species, error) {
	row, err := s.queries.SpeciesFindByGbifKey(ctx, int64(gbifKey))
	if errors.Is(err, sql.ErrNoRows) {
		return species.Species{}, species.ErrNotFound
	} else if err != nil {
		return species.Species{}, err
	}

	return transform(row), nil
}

// List returns every species.
func (s Store) List(ctx context.Context) ([]species.Species, error) {
	rows, err := s.queries.SpeciesList(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]species.Species, len(rows))
	for i, row := range rows {
		result[i] = transform(row)
	}

	return result, nil
}

// Create persists a new species under id, or returns
// species.ErrAlreadyExists if one is already there.
func (s Store) Create(ctx context.Context, id uuid.UUID, in species.CreateParams) (species.Species, error) {
	row, err := s.queries.SpeciesCreate(ctx, toSpeciesCreateParams(id, in))
	if sqlite.IsUniqueViolation(err) {
		return species.Species{}, species.ErrAlreadyExists
	} else if err != nil {
		return species.Species{}, err
	}

	return transform(row), nil
}

// Enrich saves the details of the species with id, leaving any that
// are zero as they were.
func (s Store) Enrich(ctx context.Context, id uuid.UUID, details species.Details) error {
	return s.queries.SpeciesEnrich(ctx, db.SpeciesEnrichParams{
		GenusKey:     int64(details.GenusKey),
		IucnCategory: details.IUCNCategory.String(),
		ID:           id,
	})
}
