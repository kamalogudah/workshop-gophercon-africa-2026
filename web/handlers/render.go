package handlers

import (
	"context"
	"log/slog"
	"net/http"
	"workshop/internal/domain/sighting"
	"workshop/internal/domain/sighting/country"
	"workshop/internal/domain/species"
	"workshop/web/views/pages"

	"github.com/a-h/templ"
)

// SightingsLister lists the sightings matching a filter.
type SightingsLister interface {
	List(ctx context.Context, filter sighting.ListFilter) ([]sighting.Sighting, error)
}

// render writes component as the response, falling back to the error
// page if it can't be rendered.
func render(w http.ResponseWriter, r *http.Request, logger *slog.Logger, component templ.Component) {
	ctx := r.Context()

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := component.Render(ctx, w); err != nil {
		logger.ErrorContext(ctx, "handlers: rendering page: "+err.Error())
		renderError(ctx, w, logger, http.StatusInternalServerError, "Unable to render page.")
	}
}

// listOrFail lists sightings, rendering the error page and reporting
// false if the store fails.
func listOrFail(w http.ResponseWriter, r *http.Request, logger *slog.Logger, lister SightingsLister, filter sighting.ListFilter) ([]sighting.Sighting, bool) {
	ctx := r.Context()

	items, err := lister.List(ctx, filter)
	if err != nil {
		logger.ErrorContext(ctx, "handlers: listing sightings: "+err.Error())
		renderError(ctx, w, logger, http.StatusInternalServerError, "Unable to load sightings.")
		return nil, false
	}

	return items, true
}

// filterFromQuery reads the ?animal= and ?country= a page was asked
// for, rendering a 400 and reporting false if either is unknown.
func filterFromQuery(w http.ResponseWriter, r *http.Request, logger *slog.Logger) (pages.Filter, bool) {
	q := r.URL.Query()

	var f pages.Filter
	if v := q.Get("animal"); v != "" {
		animal, err := species.ParseAnimal(v)
		if err != nil {
			renderError(r.Context(), w, logger, http.StatusBadRequest, "We don't know the animal "+v+".")
			return f, false
		}
		f.Animal = animal
	}

	code, err := country.Parse(q.Get("country"))
	if err != nil {
		renderError(r.Context(), w, logger, http.StatusBadRequest, "That isn't a country code we recognise.")
		return f, false
	}
	f.Country = code

	return f, true
}

// listFilter turns the page's filter into the domain's.
func listFilter(f pages.Filter) sighting.ListFilter {
	out := sighting.ListFilter{Country: f.Country}
	if f.Animal != "" {
		out.Animals = []species.Animal{f.Animal}
	}
	return out
}

// NotFound renders the 404 page for any path no other route matches.
func NotFound(logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		renderError(r.Context(), w, logger, http.StatusNotFound, "This trail has gone cold. The page you were tracking isn't here.")
	}
}
