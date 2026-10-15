package sighting_test

import (
	"context"
	"encoding/json/v2"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"slices"
	"strconv"
	"testing"
	"time"
	"uuid"
	"workshop/internal/clients/gbif"
	"workshop/internal/clients/openmeteo"
	"workshop/internal/domain/sighting"
	"workshop/internal/domain/sighting/country"
	"workshop/internal/domain/species"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// apis stands in for GBIF and Open-Meteo on one test server, so ingest
// runs through the real clients without leaving the process.
type apis struct {
	t         *testing.T
	gbif      *gbif.Client
	openmeteo *openmeteo.Client

	// occurrences is everything the search matches, served a page at a time.
	occurrences []gbif.Occurrence
	// pageSize caps each page, as GBIF caps it at 300. Zero means no cap.
	pageSize int
	// reading is the hour of weather every archive request gets back.
	reading openmeteo.Reading
	// searchStatus, speciesStatus and weatherStatus fail that endpoint
	// with the given HTTP status when set.
	searchStatus, speciesStatus, weatherStatus int

	// searches and lookups record what ingest asked for.
	searches []url.Values
	lookups  int
}

// newAPIs serves both APIs from one server, under the same paths the
// bundled mock uses.
func newAPIs(t *testing.T) *apis {
	t.Helper()

	a := &apis{t: t, reading: settledReading}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /gbif/v1/occurrence/search", a.search)
	mux.HandleFunc("GET /gbif/v1/occurrence/{key}", a.occurrence)
	mux.HandleFunc("GET /gbif/v1/species/{key}", a.species)
	mux.HandleFunc("GET /weather/v1/archive", a.archive)

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	a.gbif = gbif.New(srv.URL + "/gbif/v1")
	a.openmeteo = openmeteo.New(srv.URL + "/weather/v1")

	return a
}

func (a *apis) search(w http.ResponseWriter, r *http.Request) {
	a.searches = append(a.searches, r.URL.Query())
	if a.searchStatus != 0 {
		w.WriteHeader(a.searchStatus)
		return
	}

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	start := min(offset, len(a.occurrences))
	end := len(a.occurrences)
	if limit > 0 {
		end = min(start+limit, end)
	}
	if a.pageSize > 0 {
		end = min(start+a.pageSize, end)
	}

	a.write(w, gbif.PaginatedResponse[gbif.Occurrence]{
		Offset:       offset,
		Limit:        limit,
		EndOfRecords: end == len(a.occurrences),
		Count:        int64(len(a.occurrences)),
		Results:      a.occurrences[start:end],
	})
}

// occurrence returns the occurrence with the key asked for, or a 404.
func (a *apis) occurrence(w http.ResponseWriter, r *http.Request) {
	key, _ := strconv.ParseInt(r.PathValue("key"), 10, 64)
	for _, o := range a.occurrences {
		if o.Key == key {
			a.write(w, o)
			return
		}
	}
	w.WriteHeader(http.StatusNotFound)
}

// species returns a lion under whichever key is asked for, so each new
// key saves a species of its own.
func (a *apis) species(w http.ResponseWriter, r *http.Request) {
	a.lookups++
	if a.speciesStatus != 0 {
		w.WriteHeader(a.speciesStatus)
		return
	}

	key, _ := strconv.Atoi(r.PathValue("key"))
	lion := lionParams()
	a.write(w, gbif.Species{
		Key:            key,
		Kingdom:        lion.Kingdom,
		Phylum:         lion.Phylum,
		Class:          lion.Class,
		Order:          lion.Order,
		Family:         lion.Family,
		Genus:          lion.Genus,
		Species:        lion.Species,
		Rank:           lion.Rank,
		ScientificName: lion.ScientificName,
		CanonicalName:  lion.CanonicalName,
		VernacularName: lion.VernacularName,
	})
}

// archive answers with the same hour of weather, placed where and when
// it was asked for, in the archive's one-entry-per-hour arrays.
func (a *apis) archive(w http.ResponseWriter, r *http.Request) {
	if a.weatherStatus != 0 {
		w.WriteHeader(a.weatherStatus)
		return
	}

	q := r.URL.Query()
	latitude, _ := strconv.ParseFloat(q.Get("latitude"), 64)
	longitude, _ := strconv.ParseFloat(q.Get("longitude"), 64)

	a.write(w, map[string]any{
		"latitude":  latitude,
		"longitude": longitude,
		"hourly": map[string]any{
			"time":                 []string{q.Get("start_hour")},
			"temperature_2m":       []float64{a.reading.Temperature},
			"apparent_temperature": []float64{a.reading.ApparentTemperature},
			"relative_humidity_2m": []int{a.reading.Humidity},
			"precipitation":        []float64{a.reading.Precipitation},
			"rain":                 []float64{a.reading.Rain},
			"snowfall":             []float64{a.reading.Snowfall},
			"cloud_cover":          []int{a.reading.CloudCover},
			"weather_code":         []int{int(a.reading.Code)},
			"wind_speed_10m":       []float64{a.reading.WindSpeed},
			"wind_gusts_10m":       []float64{a.reading.WindGusts},
			"wind_direction_10m":   []int{a.reading.WindDirection},
		},
	})
}

func (a *apis) write(w http.ResponseWriter, v any) {
	require.NoError(a.t, json.MarshalWrite(w, v))
}

// settledReading is the weather the archive reports by default.
var settledReading = openmeteo.Reading{
	Temperature:         26.3,
	ApparentTemperature: 28.1,
	Humidity:            74,
	CloudCover:          22,
	WindSpeed:           14.2,
	WindGusts:           27.4,
	WindDirection:       118,
	Code:                openmeteo.WeatherCode(1),
}

// newOccurrence returns a fully populated lion occurrence in the Maasai Mara.
func newOccurrence(key int64) gbif.Occurrence {
	return gbif.Occurrence{
		Key:                           key,
		OccurrenceID:                  fmt.Sprintf("occurrence-%d", key),
		SpeciesKey:                    lionKey,
		EventDate:                     happenedAt,
		BasisOfRecord:                 "HUMAN_OBSERVATION",
		RecordedBy:                    "Ada|Grace",
		IndividualCount:               new(2),
		OccurrenceRemarks:             "Resting in the shade",
		References:                    "https://www.inaturalist.org/observations/1",
		Country:                       "Kenya",
		CountryCode:                   "KE",
		StateProvince:                 "Narok",
		Locality:                      "Maasai Mara",
		DecimalLatitude:               -1.5,
		DecimalLongitude:              35.1,
		CoordinateUncertaintyInMeters: new(10.0),
		Elevation:                     new(1600.0),
	}
}

// The one branch left uncovered is saving a species GBIF has just
// described, which needs a store that reads but won't write.
func TestService_Ingest(t *testing.T) {
	type Test struct {
		ctx     context.Context
		svc     *sighting.Service
		species *species.Service
		apis    *apis
	}

	setupIngest := func(t *testing.T, fail bool) Test {
		t.Helper()

		test := Test{apis: newAPIs(t)}
		test.ctx, test.svc, test.species = setup(t, fail, test.apis.gbif, test.apis.openmeteo)

		return test
	}

	t.Run("Creates a sighting per occurrence", func(t *testing.T) {
		test := setupIngest(t, false)
		test.apis.occurrences = []gbif.Occurrence{newOccurrence(1), newOccurrence(2)}

		got, err := test.svc.Ingest(test.ctx, sighting.IngestRequest{Limit: 10})
		require.NoError(t, err)
		assert.Len(t, got.Created, 2)
		assert.Equal(t, 2, got.Fetched)
		assert.Zero(t, got.Skipped+got.Duplicates+got.Failed)
	})

	t.Run("Maps the occurrence onto the sighting", func(t *testing.T) {
		test := setupIngest(t, false)
		occ := newOccurrence(1)
		test.apis.occurrences = []gbif.Occurrence{occ}

		res, err := test.svc.Ingest(test.ctx, sighting.IngestRequest{Limit: 10})
		require.NoError(t, err)
		require.Len(t, res.Created, 1)

		got := res.Created[0]
		assert.Equal(t, sighting.CreateParams{
			GBIFKey:         occ.Key,
			OccurrenceID:    occ.OccurrenceID,
			HappenedAt:      occ.EventDate,
			BasisOfRecord:   occ.BasisOfRecord,
			RecordedBy:      occ.RecordedBy,
			IndividualCount: occ.IndividualCount,
			Remarks:         occ.OccurrenceRemarks,
			ReferenceURL:    occ.References,
			Location: sighting.Location{
				Country:                       occ.Country,
				CountryCode:                   country.Kenya,
				StateProvince:                 occ.StateProvince,
				Locality:                      occ.Locality,
				Coordinates:                   sighting.Coordinates{Latitude: -1.5, Longitude: 35.1},
				CoordinateUncertaintyInMetres: occ.CoordinateUncertaintyInMeters,
				ElevationInMetres:             occ.Elevation,
			},
			SpeciesID: got.Species.ID,
			WeatherID: got.Weather.ID,
			Media:     []sighting.Media{},
		}, got.CreateParams())
	})

	t.Run("Copies the media", func(t *testing.T) {
		test := setupIngest(t, false)
		occ := newOccurrence(1)
		// GBIF's media is an anonymous struct, so grow the slice rather
		// than spell its type out.
		occ.Media = slices.Grow(occ.Media, 1)[:1]
		occ.Media[0].Type = "StillImage"
		occ.Media[0].Format = "image/jpeg"
		occ.Media[0].Identifier = "https://example.com/lion.jpg"
		occ.Media[0].Creator = "Ada"
		occ.Media[0].License = "CC BY 4.0"
		test.apis.occurrences = []gbif.Occurrence{occ}

		res, err := test.svc.Ingest(test.ctx, sighting.IngestRequest{Limit: 10})
		require.NoError(t, err)
		require.Len(t, res.Created, 1)
		require.Len(t, res.Created[0].Media, 1)

		got := res.Created[0].Media[0]
		assert.NotEqual(t, uuid.Nil(), got.ID)
		// The ID is random, so compare everything else.
		got.ID = uuid.Nil()
		assert.Equal(t, sighting.Media{
			Type:    "StillImage",
			Format:  "image/jpeg",
			URL:     "https://example.com/lion.jpg",
			Creator: "Ada",
			License: "CC BY 4.0",
		}, got)
	})

	t.Run("Fetches a new species from GBIF once", func(t *testing.T) {
		test := setupIngest(t, false)
		test.apis.occurrences = []gbif.Occurrence{newOccurrence(1), newOccurrence(2)}

		res, err := test.svc.Ingest(test.ctx, sighting.IngestRequest{Limit: 10})
		require.NoError(t, err)
		assert.Equal(t, 1, test.apis.lookups)

		got, err := test.svc.Find(test.ctx, res.Created[0].ID)
		require.NoError(t, err)
		assert.Equal(t, lionKey, got.Species.GBIFKey)
	})

	t.Run("Reuses a species already stored", func(t *testing.T) {
		test := setupIngest(t, false)
		lion := seedLion(t, test.ctx, test.species)
		test.apis.occurrences = []gbif.Occurrence{newOccurrence(1)}

		res, err := test.svc.Ingest(test.ctx, sighting.IngestRequest{Limit: 10})
		require.NoError(t, err)
		require.Len(t, res.Created, 1)
		assert.Zero(t, test.apis.lookups)
		assert.Equal(t, lion.ID, res.Created[0].Species.ID)
	})

	t.Run("Records the weather", func(t *testing.T) {
		test := setupIngest(t, false)
		test.apis.occurrences = []gbif.Occurrence{newOccurrence(1)}

		res, err := test.svc.Ingest(test.ctx, sighting.IngestRequest{Limit: 10})
		require.NoError(t, err)
		require.Len(t, res.Created, 1)

		got, err := test.svc.Find(test.ctx, res.Created[0].ID)
		require.NoError(t, err)
		assert.Equal(t, happenedAt.Truncate(time.Hour), got.Weather.ObservedAt)
		assert.Equal(t, settledReading.Temperature, got.Weather.Temperature.Actual)
		assert.Equal(t, settledReading.WindDirection, got.Weather.Wind.Direction)
		assert.Equal(t, settledReading.Code.String(), got.Weather.Condition.Description)
	})

	t.Run("Builds the search from the request", func(t *testing.T) {
		test := setupIngest(t, false)
		test.apis.occurrences = []gbif.Occurrence{newOccurrence(1)}
		from := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2025, time.December, 31, 0, 0, 0, 0, time.UTC)

		_, err := test.svc.Ingest(test.ctx, sighting.IngestRequest{
			Limit:       10,
			Offset:      20,
			Country:     country.Kenya,
			Animals:     []species.Animal{species.AnimalLion, species.AnimalCheetah},
			Latitude:    -1.5,
			Longitude:   35.1,
			RadiusMiles: 25,
			DateFrom:    from,
			DateTo:      to,
		})
		require.NoError(t, err)
		require.Len(t, test.apis.searches, 1)
		assert.Equal(t, url.Values{
			"limit":        {"10"},
			"offset":       {"20"},
			"country":      {"KE"},
			"geo_distance": {"25mi,-1.5,35.1"},
			"eventDate":    {"2025-01-01,2025-12-31"},
			"taxonKey":     {"5219404", "2435270"},
			"checklistKey": {"d7dddbf4-2cf0-4f39-9b2a-bb099caae36c"},
		}, test.apis.searches[0])
	})

	t.Run("Pages until the limit", func(t *testing.T) {
		test := setupIngest(t, false)
		for key := range int64(6) {
			test.apis.occurrences = append(test.apis.occurrences, newOccurrence(key+1))
		}
		test.apis.pageSize = 2

		var pages []int
		res, err := test.svc.Ingest(test.ctx, sighting.IngestRequest{
			Limit:  5,
			OnPage: func(fetched int, _ int64) { pages = append(pages, fetched) },
		})
		require.NoError(t, err)
		assert.Equal(t, 5, res.Fetched)
		assert.Len(t, res.Created, 5)
		assert.Equal(t, []int{2, 4, 5}, pages)
		assert.Len(t, test.apis.searches, 3)
	})

	t.Run("Stops at the end of records", func(t *testing.T) {
		test := setupIngest(t, false)
		test.apis.occurrences = []gbif.Occurrence{newOccurrence(1), newOccurrence(2), newOccurrence(3)}
		test.apis.pageSize = 2

		res, err := test.svc.Ingest(test.ctx, sighting.IngestRequest{Limit: 10})
		require.NoError(t, err)
		assert.Equal(t, 3, res.Fetched)
		assert.Len(t, test.apis.searches, 2)
	})

	t.Run("Skips what a skipper rejects", func(t *testing.T) {
		test := setupIngest(t, false)
		test.apis.occurrences = []gbif.Occurrence{newOccurrence(1), newOccurrence(2)}

		res, err := test.svc.Ingest(test.ctx, sighting.IngestRequest{
			Limit: 10,
			Skippers: []sighting.IngestShouldSkipFn{
				func(o gbif.Occurrence) bool { return o.Key == 2 },
			},
		})
		require.NoError(t, err)
		assert.Equal(t, 1, res.Skipped)
		require.Len(t, res.Created, 1)
		assert.Equal(t, int64(1), res.Created[0].GBIFKey)
	})

	t.Run("Counts duplicates on re-ingest", func(t *testing.T) {
		test := setupIngest(t, false)
		test.apis.occurrences = []gbif.Occurrence{newOccurrence(1), newOccurrence(2)}

		_, err := test.svc.Ingest(test.ctx, sighting.IngestRequest{Limit: 10})
		require.NoError(t, err)

		res, err := test.svc.Ingest(test.ctx, sighting.IngestRequest{Limit: 10})
		require.NoError(t, err)
		assert.Equal(t, 2, res.Duplicates)
		assert.Empty(t, res.Created)
	})

	t.Run("Counts a failed species lookup", func(t *testing.T) {
		test := setupIngest(t, false)
		test.apis.occurrences = []gbif.Occurrence{newOccurrence(1)}
		test.apis.speciesStatus = http.StatusInternalServerError

		res, err := test.svc.Ingest(test.ctx, sighting.IngestRequest{Limit: 10})
		require.NoError(t, err)
		assert.Equal(t, 1, res.Failed)
		assert.Empty(t, res.Created)
	})

	t.Run("Counts a failed weather reading", func(t *testing.T) {
		test := setupIngest(t, false)
		test.apis.occurrences = []gbif.Occurrence{newOccurrence(1)}
		test.apis.weatherStatus = http.StatusServiceUnavailable

		res, err := test.svc.Ingest(test.ctx, sighting.IngestRequest{Limit: 10})
		require.NoError(t, err)
		assert.Equal(t, 1, res.Failed)
		assert.Empty(t, res.Created)
	})

	t.Run("Counts coordinates out of range", func(t *testing.T) {
		test := setupIngest(t, false)
		occ := newOccurrence(1)
		occ.DecimalLatitude = 100
		test.apis.occurrences = []gbif.Occurrence{occ}

		res, err := test.svc.Ingest(test.ctx, sighting.IngestRequest{Limit: 10})
		require.NoError(t, err)
		assert.Equal(t, 1, res.Failed)
		assert.Empty(t, res.Created)
	})

	t.Run("Counts an unreadable country code", func(t *testing.T) {
		test := setupIngest(t, false)
		occ := newOccurrence(1)
		occ.CountryCode = "KEN"
		test.apis.occurrences = []gbif.Occurrence{occ}

		res, err := test.svc.Ingest(test.ctx, sighting.IngestRequest{Limit: 10})
		require.NoError(t, err)
		assert.Equal(t, 1, res.Failed)
		assert.Empty(t, res.Created)
	})

	t.Run("Counts weather that breaks the rules", func(t *testing.T) {
		test := setupIngest(t, false)
		test.apis.occurrences = []gbif.Occurrence{newOccurrence(1)}
		test.apis.reading.Humidity = 101

		res, err := test.svc.Ingest(test.ctx, sighting.IngestRequest{Limit: 10})
		require.NoError(t, err)
		assert.Equal(t, 1, res.Failed)
		assert.Empty(t, res.Created)
	})

	t.Run("Counts a sighting that breaks the rules", func(t *testing.T) {
		test := setupIngest(t, false)
		occ := newOccurrence(1)
		occ.EventDate = time.Now().AddDate(0, 0, 1)
		test.apis.occurrences = []gbif.Occurrence{occ}

		res, err := test.svc.Ingest(test.ctx, sighting.IngestRequest{Limit: 10})
		require.NoError(t, err)
		assert.Equal(t, 1, res.Failed)
		assert.Empty(t, res.Created)
	})

	t.Run("Counts a store failure", func(t *testing.T) {
		test := setupIngest(t, true)
		test.apis.occurrences = []gbif.Occurrence{newOccurrence(1)}

		res, err := test.svc.Ingest(test.ctx, sighting.IngestRequest{Limit: 10})
		require.NoError(t, err)
		assert.Equal(t, 1, res.Failed)
		assert.Zero(t, test.apis.lookups)
	})

	t.Run("No results", func(t *testing.T) {
		test := setupIngest(t, false)

		_, err := test.svc.Ingest(test.ctx, sighting.IngestRequest{Limit: 10})
		assert.ErrorIs(t, err, sighting.ErrIngestNoResults)
	})

	t.Run("GBIF error", func(t *testing.T) {
		test := setupIngest(t, false)
		test.apis.searchStatus = http.StatusInternalServerError

		_, err := test.svc.Ingest(test.ctx, sighting.IngestRequest{Limit: 10})
		require.Error(t, err)
	})
}
