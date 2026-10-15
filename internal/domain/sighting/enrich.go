package sighting

import (
	"context"
	"log/slog"
	"strings"
	"uuid"
	"workshop/internal/clients/gbif"
	"workshop/internal/domain/species"
)

// EnrichRequest configures Enrich.
type EnrichRequest struct {
	// Limit is the most sightings to look at, zero for all of them.
	Limit int

	// OnProgress, if set, is called after each sighting with how many
	// have been looked at and how many there are.
	OnProgress func(done, total int)
}

// EnrichResult summarises what happened during an Enrich.
type EnrichResult struct {
	Checked int `json:"checked"` // The number of sightings looked at.
	Updated int `json:"updated"` // The number GBIF had details for.
	Species int `json:"species"` // The number of updates made to species.
	Failed  int `json:"failed"`  // The number that errored.
}

// Enrich fills in the details of sightings saved before they were
// recorded, fetching each one's occurrence from GBIF again. Its species
// gets its genus and conservation status along the way.
//
// A sighting GBIF has nothing more to say about stays as it was, and so
// is looked at again next time.
func (s Service) Enrich(ctx context.Context, req EnrichRequest) (EnrichResult, error) {
	limit := req.Limit
	if limit <= 0 {
		limit = -1
	}

	stale, err := s.store.ListUnenriched(ctx, limit)
	if err != nil {
		return EnrichResult{}, err
	}

	var out EnrichResult
	assessed := make(map[uuid.UUID]bool)

	for _, item := range stale {
		if err = ctx.Err(); err != nil {
			return out, err
		}

		out.Checked++
		if req.OnProgress != nil {
			req.OnProgress(out.Checked, len(stale))
		}

		occurrence, err := s.gbif.Occurrence(ctx, item.GBIFKey)
		if err != nil {
			s.logger.ErrorContext(ctx, "Fetching occurrence", "gbif_key", item.GBIFKey, "error", err)
			out.Failed++
			continue
		}

		details := detailsOf(occurrence)
		if details != (Details{}) {
			if err = s.store.Enrich(ctx, item.ID, details); err != nil {
				s.logger.ErrorContext(ctx, "Enriching sighting", "sighting_id", item.ID.String(), "error", err)
				out.Failed++
				continue
			}
			out.Updated++
		}

		// Not every occurrence carries a conservation status, so a species
		// is updated until one that does comes along.
		about := species.Details{GenusKey: occurrence.GenusKey, IUCNCategory: iucnCategory(occurrence)}
		if about != (species.Details{}) && !assessed[item.SpeciesID] {
			assessed[item.SpeciesID] = about.IUCNCategory != species.IUCNUnknown
			if err = s.species.Enrich(ctx, item.SpeciesID, about); err != nil {
				s.logger.ErrorContext(ctx, "Enriching species", "species_id", item.SpeciesID.String(), "error", err)
				continue
			}
			out.Species++
		}
	}

	s.logger.InfoContext(ctx, "Enriched sightings",
		slog.Int("checked", out.Checked),
		slog.Int("updated", out.Updated),
	)

	return out, nil
}

// detailsOf picks the details out of an occurrence.
func detailsOf(v gbif.Occurrence) Details {
	return Details{
		Continent:    v.Continent,
		Organism:     organism(v),
		IdentifiedBy: observers(v.IdentifiedBy),
		DatasetName:  v.DatasetName,
	}
}

// organism reads what the observer said about the animal. GBIF records
// "Unknown" where a field was asked for and left blank, which says no
// more than leaving it out.
func organism(v gbif.Occurrence) Organism {
	known := func(s string) string {
		s = strings.TrimSpace(s)
		if strings.EqualFold(s, "unknown") {
			return ""
		}
		return s
	}
	return Organism{
		Sex:       known(v.Sex),
		LifeStage: known(v.LifeStage),
		Behaviour: known(v.Behavior),
	}
}

// iucnCategory reads the occurrence's conservation status, treating one
// it can't parse as unknown rather than failing the whole sighting.
func iucnCategory(v gbif.Occurrence) species.IUCNCategory {
	category, err := species.ParseIUCNCategory(v.IucnRedListCategory)
	if err != nil {
		return species.IUCNUnknown
	}
	return category
}
