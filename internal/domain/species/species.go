package species

import (
	"context"
	"errors"
	"log/slog"
	"uuid"
)

// Store defines the operations used to persist and query
// species data.
type Store interface {
	Find(ctx context.Context, id uuid.UUID) (Species, error)
	FindByGbifKey(ctx context.Context, gbifKey int) (Species, error)
	List(ctx context.Context) ([]Species, error)
	Create(ctx context.Context, id uuid.UUID, in CreateParams) (Species, error)
	Enrich(ctx context.Context, id uuid.UUID, details Details) error
}

// Service is the entry point for working with species data.
// Handlers and other domains should use this type instead
// of interacting with the store directly.
type Service struct {
	logger *slog.Logger
	store  Store
}

// NewService returns a Service backed by the given store.
func NewService(
	logger *slog.Logger,
	store Store,
) *Service {
	return &Service{
		logger: logger,
		store:  store,
	}
}

var (
	// ErrNotFound represents an error where a species is
	// not found in the data store.
	ErrNotFound = errors.New("species not found")

	// ErrAlreadyExists represents an error where a species with
	// the same GBIF key is already in the data store.
	ErrAlreadyExists = errors.New("species already exists")
)

// Find returns the species with the given ID, or ErrNotFound.
func (s Service) Find(ctx context.Context, id uuid.UUID) (Species, error) {
	return s.store.Find(ctx, id)
}

// FindByGbifKey returns the species with the given GBIF key, or ErrNotFound.
func (s Service) FindByGbifKey(ctx context.Context, gbifKey int) (Species, error) {
	return s.store.FindByGbifKey(ctx, gbifKey)
}

// List returns every species in the data store.
func (s Service) List(ctx context.Context) ([]Species, error) {
	return s.store.List(ctx)
}

// Create assigns the species its identity before handing it to the
// store.
func (s Service) Create(ctx context.Context, params CreateParams) (Species, error) {
	return s.store.Create(ctx, uuid.New(), params)
}

// Enrich saves the details of the species with id.
func (s Service) Enrich(ctx context.Context, id uuid.UUID, details Details) error {
	return s.store.Enrich(ctx, id, details)
}
