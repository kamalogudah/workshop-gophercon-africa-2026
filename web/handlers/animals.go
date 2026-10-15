package handlers

import (
	"cmp"
	"log/slog"
	"net/http"
	"slices"
	"workshop/internal/domain/sighting"
	"workshop/internal/domain/species"
	"workshop/web/views/format"
	"workshop/web/views/pages"
)

// Animals renders every animal the site knows, with how often each has
// been seen.
func Animals(logger *slog.Logger, sightings SightingsLister) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		all, ok := listOrFail(w, r, logger, sightings, sighting.ListFilter{})
		if !ok {
			return
		}

		byAnimal := make(map[species.Animal]*pages.AnimalTally)
		for _, s := range all {
			animal, ok := s.Species.Animal()
			if !ok {
				continue
			}
			t := byAnimal[animal]
			if t == nil {
				t = &pages.AnimalTally{Animal: animal, Class: s.Species.Class, Status: s.Species.IUCNCategory}
				byAnimal[animal] = t
			}
			t.Count++
			// Sightings come newest first, so this keeps the latest photo.
			if t.Photo == "" {
				t.Photo = format.FirstPhoto(s)
			}
		}

		tallies := make([]pages.AnimalTally, 0, len(species.Animals()))
		for _, animal := range species.Animals() {
			if t := byAnimal[animal]; t != nil {
				tallies = append(tallies, *t)
			} else {
				tallies = append(tallies, pages.AnimalTally{Animal: animal})
			}
		}
		slices.SortStableFunc(tallies, func(a, b pages.AnimalTally) int {
			return cmp.Compare(b.Count, a.Count)
		})

		render(w, r, logger, pages.Animals(tallies))
	}
}

// Animal renders the sightings of the animal in the path.
func Animal(logger *slog.Logger, sightings SightingsLister) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		animal, err := species.ParseAnimal(r.PathValue("animal"))
		if err != nil {
			renderError(r.Context(), w, logger, http.StatusNotFound, "We don't track that animal yet.")
			return
		}

		filter := pages.Filter{Animal: animal}
		items, ok := listOrFail(w, r, logger, sightings, listFilter(filter))
		if !ok {
			return
		}

		render(w, r, logger, pages.Animal(animal, items))
	}
}
