package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"uuid"
	"workshop/internal/common/logs"
	"workshop/internal/domain/sighting"
	"workshop/internal/domain/species"
	"workshop/internal/domain/weather"
	"workshop/web/handlers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var lionSighting = sighting.Sighting{
	ID:         uuid.MustParse("6f1c6a58-9b1e-4c39-9d43-6a3b1f0b2c11"),
	GBIFKey:    1,
	HappenedAt: time.Date(2024, 3, 2, 0, 0, 0, 0, time.UTC),
	Location: sighting.Location{
		Country:     "Kenya",
		CountryCode: "KE",
		Continent:   "AFRICA",
		Coordinates: sighting.Coordinates{Latitude: -1.5, Longitude: 35.1},
	},
	Species: species.Species{
		GBIFKey:        5219404,
		Class:          "Mammalia",
		VernacularName: "lion",
		ScientificName: "Panthera leo",
		IUCNCategory:   species.IUCNVulnerable,
	},
	Weather: weather.Weather{
		Temperature: weather.Temperature{Actual: 24},
		Condition:   weather.Condition{Code: 1, Description: "Mainly clear"},
	},
	Organism: sighting.Organism{Sex: "Female"},
}

type fakeFinder struct {
	item sighting.Sighting
	err  error
}

func (f fakeFinder) Find(context.Context, uuid.UUID) (sighting.Sighting, error) {
	return f.item, f.err
}

func serve(t *testing.T, h func(*slog.Logger) http.Handler, path string) *httptest.ResponseRecorder {
	t.Helper()

	mux := http.NewServeMux()
	logger := logs.NewProduction(io.Discard, slog.LevelError)
	mux.Handle("GET /sightings/{id}", h(logger))
	mux.Handle("GET /animals/{animal}", h(logger))
	mux.Handle("GET /countries/{code}", h(logger))
	mux.Handle("/", h(logger))

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, path, nil))
	return rec
}

func TestSighting(t *testing.T) {
	t.Parallel()

	handler := func(f fakeFinder) func(*slog.Logger) http.Handler {
		return func(l *slog.Logger) http.Handler { return handlers.Sighting(l, f) }
	}

	t.Run("OK", func(t *testing.T) {
		t.Parallel()
		got := serve(t, handler(fakeFinder{item: lionSighting}), "/sightings/"+lionSighting.ID.String())
		assert.Equal(t, http.StatusOK, got.Code)
		assert.Contains(t, got.Body.String(), "Lion")
		assert.Contains(t, got.Body.String(), "Vulnerable")
	})

	t.Run("Bad ID", func(t *testing.T) {
		t.Parallel()
		got := serve(t, handler(fakeFinder{}), "/sightings/not-a-uuid")
		assert.Equal(t, http.StatusNotFound, got.Code)
	})

	t.Run("Not found", func(t *testing.T) {
		t.Parallel()
		got := serve(t, handler(fakeFinder{err: sighting.ErrNotFound}), "/sightings/"+lionSighting.ID.String())
		assert.Equal(t, http.StatusNotFound, got.Code)
	})

	t.Run("Store error", func(t *testing.T) {
		t.Parallel()
		got := serve(t, handler(fakeFinder{err: errors.New("boom")}), "/sightings/"+lionSighting.ID.String())
		assert.Equal(t, http.StatusInternalServerError, got.Code)
	})
}

func TestListingPages(t *testing.T) {
	t.Parallel()

	ok := fakeLister{items: []sighting.Sighting{lionSighting}}
	failing := fakeLister{err: errors.New("store unavailable")}

	tt := map[string]struct {
		handler func(handlers.SightingsLister) func(*slog.Logger) http.Handler
		path    string
		lister  fakeLister
		want    int
	}{
		"Animals":          {animalsH, "/", ok, http.StatusOK},
		"Animal":           {animalH, "/animals/lion", ok, http.StatusOK},
		"Unknown animal":   {animalH, "/animals/dragon", ok, http.StatusNotFound},
		"Countries":        {countriesH, "/", ok, http.StatusOK},
		"Country":          {countryH, "/countries/KE", ok, http.StatusOK},
		"Bad country":      {countryH, "/countries/KEN", ok, http.StatusNotFound},
		"Gallery":          {galleryH, "/", ok, http.StatusOK},
		"Gallery filtered": {galleryH, "/?animal=lion&country=KE", ok, http.StatusOK},
		"Gallery bad":      {galleryH, "/?animal=dragon", ok, http.StatusBadRequest},
		"Stats":            {statsH, "/", ok, http.StatusOK},
		"Stats empty":      {statsH, "/", fakeLister{}, http.StatusOK},
		"Stats error":      {statsH, "/", failing, http.StatusInternalServerError},
		"Animals error":    {animalsH, "/", failing, http.StatusInternalServerError},
	}

	for name, test := range tt {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got := serve(t, test.handler(test.lister), test.path)
			assert.Equal(t, test.want, got.Code)
			assert.Equal(t, "text/html; charset=utf-8", got.Header().Get("Content-Type"))
		})
	}
}

func TestPoints(t *testing.T) {
	t.Parallel()

	got := serve(t, pointsH(fakeLister{items: []sighting.Sighting{lionSighting}}), "/?animal=lion")
	require.Equal(t, http.StatusOK, got.Code)

	var points []map[string]any
	require.NoError(t, json.Unmarshal(got.Body.Bytes(), &points))
	require.Len(t, points, 1)
	assert.Equal(t, "Lion", points[0]["name"])
	assert.Equal(t, "Mammals", points[0]["group"])
	assert.InDelta(t, -1.5, points[0]["lat"], 0.0001)
	assert.Equal(t, "Panthera leo", points[0]["scientific"])
	assert.Equal(t, "VU", points[0]["status"])
	assert.Equal(t, "Vulnerable", points[0]["statusLabel"])
	assert.Equal(t, "🇰🇪", points[0]["flag"])
	assert.Equal(t, "🌤️ Mainly clear", points[0]["weather"])
	assert.Equal(t, "24°C", points[0]["temperature"])
	assert.NotContains(t, points[0], "photo")
}

func TestNotFound(t *testing.T) {
	t.Parallel()

	got := serve(t, func(l *slog.Logger) http.Handler { return handlers.NotFound(l) }, "/nowhere")
	assert.Equal(t, http.StatusNotFound, got.Code)
}

func animalsH(s handlers.SightingsLister) func(*slog.Logger) http.Handler {
	return func(l *slog.Logger) http.Handler { return handlers.Animals(l, s) }
}

func animalH(s handlers.SightingsLister) func(*slog.Logger) http.Handler {
	return func(l *slog.Logger) http.Handler { return handlers.Animal(l, s) }
}

func countriesH(s handlers.SightingsLister) func(*slog.Logger) http.Handler {
	return func(l *slog.Logger) http.Handler { return handlers.Countries(l, s) }
}

func countryH(s handlers.SightingsLister) func(*slog.Logger) http.Handler {
	return func(l *slog.Logger) http.Handler { return handlers.Country(l, s) }
}

func galleryH(s handlers.SightingsLister) func(*slog.Logger) http.Handler {
	return func(l *slog.Logger) http.Handler { return handlers.Gallery(l, s) }
}

func statsH(s handlers.SightingsLister) func(*slog.Logger) http.Handler {
	return func(l *slog.Logger) http.Handler { return handlers.Stats(l, s) }
}

func pointsH(s handlers.SightingsLister) func(*slog.Logger) http.Handler {
	return func(l *slog.Logger) http.Handler { return handlers.Points(l, s) }
}
