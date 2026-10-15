package handlers

import (
	"log/slog"
	"net/http"
	"workshop/internal/domain/sighting"
	"workshop/web/views/pages"
)

func Home(logger *slog.Logger, sightings SightingsLister) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		items, ok := listOrFail(w, r, logger, sightings, sighting.ListFilter{
			Limit: 48,
		})
		if !ok {
			return
		}

		render(w, r, logger, pages.Home(items))
	}
}
