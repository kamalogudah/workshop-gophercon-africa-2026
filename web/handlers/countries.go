package handlers

import (
	"cmp"
	"log/slog"
	"net/http"
	"slices"
	"workshop/internal/domain/sighting"
	"workshop/internal/domain/sighting/country"
	"workshop/web/views/pages"
)

// Countries renders every country with a sighting, most seen first.
func Countries(logger *slog.Logger, sightings SightingsLister) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		all, ok := listOrFail(w, r, logger, sightings, sighting.ListFilter{})
		if !ok {
			return
		}

		byCode := make(map[country.Code]*pages.CountryTally)
		species := make(map[country.Code]map[string]struct{})
		for _, s := range all {
			code := s.Location.CountryCode
			if code == "" {
				continue
			}
			t := byCode[code]
			if t == nil {
				t = &pages.CountryTally{Code: code, Name: s.Location.Country, Continent: s.Location.Continent}
				byCode[code] = t
				species[code] = make(map[string]struct{})
			}
			t.Count++
			species[code][s.Species.ScientificName] = struct{}{}
			if t.Continent == "" {
				t.Continent = s.Location.Continent
			}
		}

		tallies := make([]pages.CountryTally, 0, len(byCode))
		for code, t := range byCode {
			t.Species = len(species[code])
			tallies = append(tallies, *t)
		}
		slices.SortFunc(tallies, func(a, b pages.CountryTally) int {
			if c := cmp.Compare(b.Count, a.Count); c != 0 {
				return c
			}
			return cmp.Compare(a.Name, b.Name)
		})

		render(w, r, logger, pages.Countries(tallies))
	}
}

// Country renders the sightings in the country in the path.
func Country(logger *slog.Logger, sightings SightingsLister) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code, err := country.Parse(r.PathValue("code"))
		if err != nil || code == "" {
			renderError(r.Context(), w, logger, http.StatusNotFound, "That isn't a country code we recognise.")
			return
		}

		items, ok := listOrFail(w, r, logger, sightings, sighting.ListFilter{Country: code})
		if !ok {
			return
		}

		render(w, r, logger, pages.Country(code, items))
	}
}
