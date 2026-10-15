package sighting

import (
	"context"
	"errors"
	"log/slog"
	"regexp"
	"strings"
	"time"
	"uuid"
	"workshop/internal/clients/gbif"
	"workshop/internal/clients/openmeteo"
	"workshop/internal/domain/species"
	"workshop/internal/domain/weather"
)

// Store defines the operations used to persist and query
// sighting data.
type Store interface {
	Find(ctx context.Context, id uuid.UUID) (Sighting, error)
	FindByGbifKey(ctx context.Context, gbifKey int) (Sighting, error)
	List(ctx context.Context, filter ListFilter) ([]Sighting, error)
	Create(ctx context.Context, id uuid.UUID, in CreateParams) (Sighting, error)
	ListUnenriched(ctx context.Context, limit int) ([]Unenriched, error)
	Enrich(ctx context.Context, id uuid.UUID, details Details) error
}

// Service is the entry point for working with sighting data.
// Handlers and other domains should use this type instead
// of interacting with the store directly.
type Service struct {
	logger    *slog.Logger
	store     Store
	species   SpeciesReadWriter
	weather   WeatherWriter
	gbif      *gbif.Client
	openmeteo *openmeteo.Client
}

// NewService returns a Service backed by the given store.
func NewService(
	logger *slog.Logger,
	store Store,
	speciesSvc SpeciesReadWriter,
	weatherSvc WeatherWriter,
	gbifClient *gbif.Client,
	openMeteoClient *openmeteo.Client,
) *Service {
	return &Service{
		logger:    logger,
		store:     store,
		species:   speciesSvc,
		weather:   weatherSvc,
		gbif:      gbifClient,
		openmeteo: openMeteoClient,
	}
}

// SpeciesReadWriter is required by ingest to find species and
// create the ones that don't exist yet.
type SpeciesReadWriter interface {
	FindByGbifKey(ctx context.Context, gbifKey int) (species.Species, error)
	Create(ctx context.Context, params species.CreateParams) (species.Species, error)
	Enrich(ctx context.Context, id uuid.UUID, details species.Details) error
}

// WeatherWriter is required by ingest to create weather records.
type WeatherWriter interface {
	Create(ctx context.Context, params weather.CreateParams) (weather.Weather, error)
}

var (
	// ErrNotFound represents an error where a sighting is
	// not found in the data store.
	ErrNotFound = errors.New("sighting not found")

	// ErrAlreadyExists represents an error where a sighting with
	// the same GBIF key is already in the data store.
	ErrAlreadyExists = errors.New("sighting already exists")

	// ErrInvalid represents an error where a sighting breaks
	// one of the rules in Validate.
	ErrInvalid = errors.New("invalid sighting")
)

// Find returns the sighting with the given ID, or ErrNotFound.
func (s Service) Find(ctx context.Context, id uuid.UUID) (Sighting, error) {
	sight, err := s.store.Find(ctx, id)
	if err != nil {
		return Sighting{}, err
	}

	sight.RecordedBy = observers(sight.RecordedBy)

	return sight, nil
}

// List returns the sightings matching the filter, with each one's
// observers tidied for reading.
func (s Service) List(ctx context.Context, filter ListFilter) ([]Sighting, error) {
	sightings, err := s.store.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	for i := range sightings {
		sightings[i].RecordedBy = observers(sightings[i].RecordedBy)
	}

	return sightings, nil
}

// Create validates the params and assigns the sighting its identity
// before handing it to the store.
func (s Service) Create(ctx context.Context, params CreateParams) (Sighting, error) {
	if err := params.Validate(time.Now()); err != nil {
		return Sighting{}, err
	}

	return s.store.Create(ctx, uuid.New(), params)
}

// observers turns GBIF's pipe-separated observer list into a readable,
// comma-separated one, stripping any bracketed collector ID GBIF tags a
// name with, e.g. "J. Smith [12345]".
func observers(raw string) string {
	if raw == "" {
		return ""
	}

	tag := regexp.MustCompile(`\s*\[[^\]]*\]`)

	names := strings.Split(raw, "|")
	for i, n := range names {
		names[i] = strings.TrimSpace(tag.ReplaceAllString(n, ""))
	}

	return strings.Join(names, ", ")
}
