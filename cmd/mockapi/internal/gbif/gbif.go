package gbif

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"workshop/internal/common/httputil"
)

// Handler serves the mocked endpoints under /gbif/v1.
func (s *Store) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /gbif/v1/occurrence/search", s.handleSearch)
	mux.HandleFunc("GET /gbif/v1/occurrence/{key}", s.handleOccurrence)
	mux.HandleFunc("GET /gbif/v1/species/{key}", s.handleSpecies)
	mux.HandleFunc("GET /gbif/v1/species/{key}/media", s.handleMedia)
	mux.HandleFunc("/gbif/", func(w http.ResponseWriter, r *http.Request) {
		writeError(w, http.StatusNotFound, "Entity not found for uri: "+r.URL.Path)
	})
	return mux
}

// handleSearch mocks GET /occurrence/search.
func (s *Store) handleSearch(w http.ResponseWriter, r *http.Request) {
	q, err := parseSearch(r.URL.Query())
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	ctx := r.Context()

	var count int
	if err = s.db.QueryRowContext(ctx, searchCountQuery, q.args()...).Scan(&count); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	rows, err := s.db.QueryContext(ctx, searchQuery, q.args()...)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close() //nolint:errcheck

	results := make([]json.RawMessage, 0, q.limit)
	for rows.Next() {
		var raw string
		if err = rows.Scan(&raw); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		results = append(results, json.RawMessage(raw))
	}
	if err = rows.Err(); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Field order matches GBIF's response.
	httputil.WriteJSON(w, http.StatusOK, struct {
		Offset       int               `json:"offset"`
		Limit        int               `json:"limit"`
		EndOfRecords bool              `json:"endOfRecords"`
		Count        int               `json:"count"`
		Results      []json.RawMessage `json:"results"`
		Facets       []any             `json:"facets"`
	}{
		Offset:       q.offset,
		Limit:        q.limit,
		EndOfRecords: q.offset+q.limit >= count,
		Count:        count,
		Results:      results,
		Facets:       []any{},
	})
}

// handleOccurrence mocks GET /occurrence/{key}.
func (s *Store) handleOccurrence(w http.ResponseWriter, r *http.Request) {
	raw, err := s.raw(r, "SELECT raw FROM occurrences WHERE key = ?")
	switch {
	case errors.Is(err, sql.ErrNoRows):
		writeError(w, http.StatusNotFound, "Entity not found for uri: /occurrence/"+r.PathValue("key"))
	case err != nil:
		writeError(w, http.StatusInternalServerError, err.Error())
	default:
		httputil.WriteRawJSON(w, raw)
	}
}

// handleSpecies mocks GET /species/{key}.
func (s *Store) handleSpecies(w http.ResponseWriter, r *http.Request) {
	raw, err := s.raw(r, "SELECT raw FROM species WHERE key = ?")
	switch {
	case errors.Is(err, sql.ErrNoRows):
		writeError(w, http.StatusNotFound, "Entity not found for uri: /species/"+r.PathValue("key"))
	case err != nil:
		writeError(w, http.StatusInternalServerError, err.Error())
	default:
		httputil.WriteRawJSON(w, raw)
	}
}

// handleMedia mocks GET /species/{key}/media. GBIF answers an unknown
// species with an empty page rather than a 404.
func (s *Store) handleMedia(w http.ResponseWriter, r *http.Request) {
	raw, err := s.raw(r, "SELECT raw FROM media WHERE key = ?")
	switch {
	case errors.Is(err, sql.ErrNoRows):
		httputil.WriteRawJSON(w, `{"offset":0,"limit":20,"endOfRecords":true,"results":[]}`)
	case err != nil:
		writeError(w, http.StatusInternalServerError, err.Error())
	default:
		httputil.WriteRawJSON(w, raw)
	}
}

// raw runs query for the {key} in the request path.
func (s *Store) raw(r *http.Request, query string) (string, error) {
	key, err := strconv.Atoi(r.PathValue("key"))
	if err != nil {
		return "", sql.ErrNoRows
	}
	var raw string
	err = s.db.QueryRowContext(r.Context(), query, key).Scan(&raw)
	return raw, err
}
