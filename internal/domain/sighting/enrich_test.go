package sighting_test

import (
	"testing"
	"workshop/internal/clients/gbif"
	"workshop/internal/domain/sighting"
	"workshop/internal/domain/species"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_Enrich(t *testing.T) {
	// ingestBare saves occurrences as they were before 0003, with none
	// of the details, then gives GBIF the details to hand back.
	ingestBare := func(t *testing.T, occurrences ...gbif.Occurrence) (*apis, *sighting.Service, *species.Service) {
		t.Helper()

		a := newAPIs(t)
		ctx, svc, speciesSvc := setup(t, false, a.gbif, a.openmeteo)

		for _, o := range occurrences {
			a.occurrences = append(a.occurrences, newOccurrence(o.Key))
		}
		_, err := svc.Ingest(ctx, sighting.IngestRequest{Limit: 10})
		require.NoError(t, err)

		a.occurrences = occurrences
		return a, svc, speciesSvc
	}

	detailed := func(key int64) gbif.Occurrence {
		o := newOccurrence(key)
		o.Continent = "AFRICA"
		o.Sex = "Female"
		o.LifeStage = "Unknown"
		o.Behavior = "Resting"
		o.IdentifiedBy = "Ada [123]"
		o.DatasetName = "iNaturalist research-grade observations"
		o.GenusKey = 2435194
		o.IucnRedListCategory = "VU"
		return o
	}

	t.Run("Fills in the details", func(t *testing.T) {
		_, svc, _ := ingestBare(t, detailed(1))
		ctx := t.Context()

		res, err := svc.Enrich(ctx, sighting.EnrichRequest{})
		require.NoError(t, err)
		assert.Equal(t, 1, res.Checked)
		assert.Equal(t, 1, res.Updated)
		assert.Equal(t, 1, res.Species)

		all, err := svc.List(ctx, sighting.ListFilter{})
		require.NoError(t, err)
		require.Len(t, all, 1)

		got := all[0]
		assert.Equal(t, "AFRICA", got.Location.Continent)
		// "Unknown" says no more than leaving it out.
		assert.Equal(t, sighting.Organism{Sex: "Female", Behaviour: "Resting"}, got.Organism)
		assert.Equal(t, "Ada", got.IdentifiedBy)
		assert.Equal(t, "iNaturalist research-grade observations", got.DatasetName)
		assert.Equal(t, species.IUCNVulnerable, got.Species.IUCNCategory)
		assert.Equal(t, 2435194, got.Species.GenusKey)
	})

	t.Run("Skips sightings already enriched", func(t *testing.T) {
		_, svc, _ := ingestBare(t, detailed(1))
		ctx := t.Context()

		_, err := svc.Enrich(ctx, sighting.EnrichRequest{})
		require.NoError(t, err)

		res, err := svc.Enrich(ctx, sighting.EnrichRequest{})
		require.NoError(t, err)
		assert.Zero(t, res.Checked)
	})

	t.Run("Counts occurrences GBIF no longer has", func(t *testing.T) {
		a, svc, _ := ingestBare(t, detailed(1))
		a.occurrences = nil

		res, err := svc.Enrich(t.Context(), sighting.EnrichRequest{})
		require.NoError(t, err)
		assert.Equal(t, 1, res.Failed)
		assert.Zero(t, res.Updated)
	})
}
