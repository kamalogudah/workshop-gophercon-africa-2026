package handlers

import (
	"log/slog"
	"math"
	"net/http"
	"workshop/internal/common/httputil"
	"workshop/web/views/format"
	"workshop/web/views/pages"
)

// Map renders the map page. Its points load separately, from Points.
func Map(logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		filter, ok := filterFromQuery(w, r, logger)
		if !ok {
			return
		}

		render(w, r, logger, pages.Map(filter))
	}
}

// Points serves the sightings matching ?animal= and ?country= as JSON,
// for the map and the globe to plot.
func Points(logger *slog.Logger, sightings SightingsLister) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		filter, ok := filterFromQuery(w, r, logger)
		if !ok {
			return
		}

		items, err := sightings.List(ctx, listFilter(filter))
		if err != nil {
			logger.ErrorContext(ctx, "handlers: listing points: "+err.Error())
			httputil.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "Unable to load sightings."})
			return
		}

		points := make([]pages.Point, len(items))
		for i, s := range items {
			points[i] = pages.Point{
				ID:    s.ID.String(),
				Lat:   round(s.Location.Latitude, 4),
				Lng:   round(s.Location.Longitude, 4),
				Name:  format.DisplayName(s.Species),
				Date:  s.HappenedAt.Format("2 Jan 2006"),
				Place: format.Place(s.Location),
				Group: format.Group(s.Species),
			}
		}

		w.Header().Set("Cache-Control", "public, max-age=60")
		httputil.WriteJSON(w, http.StatusOK, points)
	}
}

// Gallery renders the photos taken of the sightings matching ?animal=
// and ?country=, newest first.
func Gallery(logger *slog.Logger, sightings SightingsLister) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		filter, ok := filterFromQuery(w, r, logger)
		if !ok {
			return
		}

		items, ok := listOrFail(w, r, logger, sightings, listFilter(filter))
		if !ok {
			return
		}

		const most = 120
		photos := make([]pages.Photo, 0, most)
	sightings:
		for _, s := range items {
			for _, m := range s.Media {
				if m.Type != "StillImage" || m.URL == "" {
					continue
				}
				photos = append(photos, pages.Photo{Media: m, Sighting: s})
				if len(photos) == most {
					break sightings
				}
			}
		}

		render(w, r, logger, pages.Gallery(filter, photos))
	}
}

// round keeps places decimal places of v, which for a coordinate is
// plenty to plot and keeps the payload small.
func round(v float64, places int) float64 {
	p := math.Pow10(places)
	return math.Round(v*p) / p
}
