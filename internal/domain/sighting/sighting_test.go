package sighting_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"
	"uuid"
	"workshop/internal/clients/gbif"
	"workshop/internal/clients/openmeteo"
	"workshop/internal/common/logs"
	"workshop/internal/domain/sighting"
	"workshop/internal/domain/sighting/country"
	"workshop/internal/domain/sighting/stores/sightingsqlite"
	"workshop/internal/domain/species"
	"workshop/internal/domain/species/stores/speciessqlite"
	"workshop/internal/domain/weather"
	"workshop/internal/domain/weather/stores/weathersqlite"
	"workshop/internal/infra/db/dbtest"
	db "workshop/internal/infra/db/sqlc"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errInternal = errors.New("internal error")

var happenedAt = time.Date(2025, time.September, 22, 11, 12, 13, 0, time.UTC)

// testWeatherID is the weather every hand-built test sighting points at.
// Each test gets its own database, and setup seeds this row into all of
// them, so newCreateParams can name it without any plumbing.
var testWeatherID = uuid.New()

const lionKey = 5219404

func setup(
	t *testing.T,
	fail bool,
	gbifClient *gbif.Client,
	openMeteoClient *openmeteo.Client,
) (context.Context, *sighting.Service, *species.Service) {
	t.Helper()

	ctx, conn, teardown := dbtest.Setup(t)
	t.Cleanup(teardown)

	q := db.Querier(db.New(conn))
	if fail {
		q = dbtest.FailingQuerier(errInternal)
	}

	// Weather is mandatory on a sighting, so every database needs the row
	// that newCreateParams points at. Seeded through the real queries so
	// it exists even when the service is given a failing one.
	_, err := weathersqlite.New(db.New(conn)).Create(ctx, testWeatherID, weather.CreateParams{
		ObservedAt:    happenedAt,
		Temperature:   weather.Temperature{Actual: 26.8, Apparent: 29.5},
		Precipitation: weather.Precipitation{Total: 1.2, Rain: 1.2},
		Wind:          weather.Wind{Speed: 24.7, Gusts: 41.3, Direction: 160},
		Condition:     weather.Condition{Code: 61, Description: "Slight rain"},
		CloudCover:    68,
		Humidity:      57,
	})
	require.NoError(t, err)

	logger := logs.NewProduction(io.Discard, slog.LevelError)
	speciesSvc := species.NewService(logger, speciessqlite.New(q))

	return ctx, sighting.NewService(
		logger,
		sightingsqlite.New(q),
		speciesSvc,
		weather.NewService(logger, weathersqlite.New(q)),
		gbifClient,
		openMeteoClient,
	), speciesSvc
}

func lionParams() species.CreateParams {
	return species.CreateParams{
		GBIFKey:        lionKey,
		Kingdom:        "Animalia",
		Phylum:         "Chordata",
		Class:          "Mammalia",
		Order:          "Carnivora",
		Family:         "Felidae",
		Genus:          "Panthera",
		Species:        "Panthera leo",
		Rank:           "SPECIES",
		ScientificName: "Panthera leo (Linnaeus, 1758)",
		CanonicalName:  "Panthera leo",
		VernacularName: "lion",
	}
}

// seedLion saves the lion locally, so callers reuse it instead of fetching
// the taxonomy from GBIF.
func seedLion(t *testing.T, ctx context.Context, speciesSvc *species.Service) species.Species {
	t.Helper()

	lion, err := speciesSvc.Create(ctx, lionParams())
	require.NoError(t, err)

	return lion
}

// newCreateParams returns a fully populated sighting of the given species.
func newCreateParams(speciesID uuid.UUID, gbifKey int64) sighting.CreateParams {
	return sighting.CreateParams{
		GBIFKey:         gbifKey,
		OccurrenceID:    "occurrence-123",
		HappenedAt:      happenedAt,
		BasisOfRecord:   "HUMAN_OBSERVATION",
		RecordedBy:      "observer",
		IndividualCount: new(2),
		Remarks:         "remarks",
		ReferenceURL:    "https://example.com/sighting/123",
		Location: sighting.Location{
			Locality:                      "London",
			StateProvince:                 "England",
			Country:                       "United Kingdom",
			CountryCode:                   "GB",
			Coordinates:                   sighting.Coordinates{Latitude: 51.5, Longitude: -0.12},
			CoordinateUncertaintyInMetres: new(3.5),
			ElevationInMetres:             new(125.25),
		},
		SpeciesID: speciesID,
		WeatherID: testWeatherID,
		Media: []sighting.Media{
			{
				ID:     uuid.New(),
				Type:   "StillImage",
				Format: "image/jpeg",
				URL:    "https://example.com/lion.jpg",
			},
		},
	}
}

func create(t *testing.T, ctx context.Context, svc *sighting.Service, in sighting.CreateParams) sighting.Sighting {
	t.Helper()

	got, err := svc.Create(ctx, in)
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil(), got.ID)
	assert.Equal(t, in, got.CreateParams())

	return got
}

func TestService_Find(t *testing.T) {
	t.Run("Found", func(t *testing.T) {
		ctx, svc, speciesSvc := setup(t, false, nil, nil)
		lion := seedLion(t, ctx, speciesSvc)
		created := create(t, ctx, svc, newCreateParams(lion.ID, 123))

		got, err := svc.Find(ctx, created.ID)
		require.NoError(t, err)
		assert.Equal(t, created.ID, got.ID)
		assert.Equal(t, created.CreateParams(), got.CreateParams())
	})

	t.Run("Joins the species", func(t *testing.T) {
		ctx, svc, speciesSvc := setup(t, false, nil, nil)
		lion := seedLion(t, ctx, speciesSvc)
		created := create(t, ctx, svc, newCreateParams(lion.ID, 123))

		got, err := svc.Find(ctx, created.ID)
		require.NoError(t, err)
		assert.Equal(t, lion, got.Species)
	})

	t.Run("Not found", func(t *testing.T) {
		ctx, svc, _ := setup(t, false, nil, nil)

		_, err := svc.Find(ctx, uuid.New())
		assert.ErrorIs(t, err, sighting.ErrNotFound)
	})

	t.Run("Store error", func(t *testing.T) {
		ctx, svc, _ := setup(t, true, nil, nil)

		_, err := svc.Find(ctx, uuid.New())
		assert.ErrorIs(t, err, errInternal)
	})
}

func TestService_List(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		ctx, svc, _ := setup(t, false, nil, nil)

		got, err := svc.List(ctx, sighting.ListFilter{})
		require.NoError(t, err)
		assert.Empty(t, got)
	})

	t.Run("Lists newest first", func(t *testing.T) {
		ctx, svc, speciesSvc := setup(t, false, nil, nil)
		lion := seedLion(t, ctx, speciesSvc)

		older := newCreateParams(lion.ID, 123)
		newer := newCreateParams(lion.ID, 124)
		newer.HappenedAt = happenedAt.AddDate(0, 1, 0)

		create(t, ctx, svc, older)
		create(t, ctx, svc, newer)

		got, err := svc.List(ctx, sighting.ListFilter{})
		require.NoError(t, err)
		require.Len(t, got, 2)
		assert.Equal(t, int64(124), got[0].GBIFKey)
		assert.Equal(t, int64(123), got[1].GBIFKey)
	})

	t.Run("Passes the filter through", func(t *testing.T) {
		ctx, svc, speciesSvc := setup(t, false, nil, nil)
		lion := seedLion(t, ctx, speciesSvc)

		london := newCreateParams(lion.ID, 123)
		nairobi := newCreateParams(lion.ID, 124)
		nairobi.Location.Country = "Kenya"
		nairobi.Location.CountryCode = "KE"

		create(t, ctx, svc, london)
		create(t, ctx, svc, nairobi)

		got, err := svc.List(ctx, sighting.ListFilter{Country: country.Kenya})
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, int64(124), got[0].GBIFKey)
	})

	t.Run("Filters by an animal that names a genus", func(t *testing.T) {
		ctx, svc, speciesSvc := setup(t, false, nil, nil)

		// GBIF keeps Giraffa giraffa under its own key, so only its genus
		// matches the giraffe.
		params := lionParams()
		params.GBIFKey = 8959277
		params.GenusKey = species.AnimalGiraffe.TaxonKeys()[0]
		giraffe, err := speciesSvc.Create(ctx, params)
		require.NoError(t, err)
		create(t, ctx, svc, newCreateParams(giraffe.ID, 123))

		got, err := svc.List(ctx, sighting.ListFilter{Animals: []species.Animal{species.AnimalGiraffe}})
		require.NoError(t, err)
		assert.Len(t, got, 1)
	})

	t.Run("Filters by animal", func(t *testing.T) {
		ctx, svc, speciesSvc := setup(t, false, nil, nil)
		lion := seedLion(t, ctx, speciesSvc)
		create(t, ctx, svc, newCreateParams(lion.ID, 123))

		got, err := svc.List(ctx, sighting.ListFilter{Animals: []species.Animal{species.AnimalLion}})
		require.NoError(t, err)
		assert.Len(t, got, 1)

		got, err = svc.List(ctx, sighting.ListFilter{Animals: []species.Animal{species.AnimalCheetah}})
		require.NoError(t, err)
		assert.Empty(t, got)
	})

	t.Run("Tidies the observers", func(t *testing.T) {
		ctx, svc, speciesSvc := setup(t, false, nil, nil)
		lion := seedLion(t, ctx, speciesSvc)

		params := newCreateParams(lion.ID, 123)
		params.RecordedBy = "Ada [1]|Grace"
		create(t, ctx, svc, params)

		got, err := svc.List(ctx, sighting.ListFilter{})
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, "Ada, Grace", got[0].RecordedBy)
	})

	t.Run("Store error", func(t *testing.T) {
		ctx, svc, _ := setup(t, true, nil, nil)

		_, err := svc.List(ctx, sighting.ListFilter{})
		assert.ErrorIs(t, err, errInternal)
	})
}

func TestService_Create(t *testing.T) {
	t.Run("Creates sighting", func(t *testing.T) {
		ctx, svc, speciesSvc := setup(t, false, nil, nil)
		lion := seedLion(t, ctx, speciesSvc)
		in := newCreateParams(lion.ID, 200)

		got, err := svc.Create(ctx, in)
		require.NoError(t, err)
		assert.Equal(t, in, got.CreateParams())
		assert.WithinDuration(t, time.Now(), got.CreatedAt, 5*time.Second)
		assert.Equal(t, got.CreatedAt, got.UpdatedAt)
	})

	t.Run("Assigns an identity", func(t *testing.T) {
		ctx, svc, speciesSvc := setup(t, false, nil, nil)
		lion := seedLion(t, ctx, speciesSvc)

		first := create(t, ctx, svc, newCreateParams(lion.ID, 201))
		second := create(t, ctx, svc, newCreateParams(lion.ID, 202))

		assert.NotEqual(t, uuid.Nil(), first.ID)
		assert.NotEqual(t, first.ID, second.ID)
	})

	t.Run("Is retrievable once created", func(t *testing.T) {
		ctx, svc, speciesSvc := setup(t, false, nil, nil)
		lion := seedLion(t, ctx, speciesSvc)
		created := create(t, ctx, svc, newCreateParams(lion.ID, 203))

		got, err := svc.Find(ctx, created.ID)
		require.NoError(t, err)
		assert.Equal(t, created.CreateParams(), got.CreateParams())
	})

	t.Run("Already exists", func(t *testing.T) {
		ctx, svc, speciesSvc := setup(t, false, nil, nil)
		lion := seedLion(t, ctx, speciesSvc)

		in := newCreateParams(lion.ID, 204)
		create(t, ctx, svc, in)

		_, err := svc.Create(ctx, in)
		assert.ErrorIs(t, err, sighting.ErrAlreadyExists)
	})

	t.Run("Missing species", func(t *testing.T) {
		ctx, svc, _ := setup(t, false, nil, nil)

		in := newCreateParams(uuid.Nil(), 205)

		_, err := svc.Create(ctx, in)
		assert.ErrorIs(t, err, sighting.ErrInvalid)
	})

	t.Run("Future sighting", func(t *testing.T) {
		ctx, svc, speciesSvc := setup(t, false, nil, nil)
		lion := seedLion(t, ctx, speciesSvc)

		in := newCreateParams(lion.ID, 206)
		in.HappenedAt = time.Now().AddDate(0, 0, 1)

		_, err := svc.Create(ctx, in)
		assert.ErrorIs(t, err, sighting.ErrInvalid)
	})

	t.Run("Invalid params are not persisted", func(t *testing.T) {
		ctx, svc, _ := setup(t, true, nil, nil)
		in := newCreateParams(uuid.New(), 207)
		in.Location.Latitude = 100

		// The store would fail any call, so ErrInvalid proves it was never reached.
		_, err := svc.Create(ctx, in)
		assert.ErrorIs(t, err, sighting.ErrInvalid)
	})

	t.Run("Store error", func(t *testing.T) {
		ctx, svc, _ := setup(t, true, nil, nil)

		_, err := svc.Create(ctx, newCreateParams(uuid.New(), 208))
		assert.ErrorIs(t, err, errInternal)
	})
}
