package sighting

import (
	"fmt"
	"time"
	"uuid"
	"workshop/internal/domain/sighting/country"
	"workshop/internal/domain/species"
	"workshop/internal/domain/weather"
)

type (
	// Sighting represents an individual sighting of a species.
	// It's a record of a specific observation at a particular location and time.
	Sighting struct {
		ID              uuid.UUID       `json:"id"`
		GBIFKey         int64           `json:"gbifKey"`
		OccurrenceID    string          `json:"occurrenceID,omitzero"`
		HappenedAt      time.Time       `json:"happenedAt"`
		BasisOfRecord   string          `json:"basisOfRecord,omitzero"`
		RecordedBy      string          `json:"recordedBy,omitzero"`
		IndividualCount *int            `json:"individualCount,omitzero"`
		Remarks         string          `json:"remarks,omitzero"`
		ReferenceURL    string          `json:"referenceURL,omitzero"`
		Location        Location        `json:"location"`
		Species         species.Species `json:"species"`
		Weather         weather.Weather `json:"weather"`
		Media           []Media         `json:"media,omitzero"`
		Organism        Organism        `json:"organism,omitzero"`
		IdentifiedBy    string          `json:"identifiedBy,omitzero"`
		DatasetName     string          `json:"datasetName,omitzero"`
		CreatedAt       time.Time       `json:"createdAt"`
		UpdatedAt       time.Time       `json:"updatedAt"`
	}
	// Location represents the geographical coordinates and other
	// location-related information associated with a sighting.
	Location struct {
		Locality      string       `json:"locality,omitzero"`
		StateProvince string       `json:"stateProvince,omitzero"`
		Country       string       `json:"country"`
		CountryCode   country.Code `json:"countryCode"`
		// Continent is GBIF's name for it, such as "NORTH_AMERICA".
		Continent string `json:"continent,omitzero"`
		// Coordinates is embedded, so a Location still reads as though it
		// held a latitude and a longitude of its own.
		Coordinates
		CoordinateUncertaintyInMetres *float64 `json:"coordinateUncertaintyInMetres,omitzero"`
		ElevationInMetres             *float64 `json:"elevationInMetres,omitzero"`
	}
	// Coordinates is a position on the surface of the earth. Build one with
	// ParseCoordinates rather than by hand: the whole point of the type is
	// that holding one is proof the numbers in it are in range.
	Coordinates struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	}
	// Organism describes the animal itself, as far as the observer could
	// tell. Each field is free text from GBIF, such as "Female" or
	// "Juvenile", and most records leave them blank.
	Organism struct {
		Sex       string `json:"sex,omitzero"`
		LifeStage string `json:"lifeStage,omitzero"`
		Behaviour string `json:"behaviour,omitzero"`
	}
	// Details are the parts of an occurrence that were added to a
	// sighting after it was first released, and so can be filled in for
	// a sighting saved before then. See Service.Enrich.
	Details struct {
		Continent    string
		Organism     Organism
		IdentifiedBy string
		DatasetName  string
	}
	// Unenriched points at a sighting whose details are still empty.
	Unenriched struct {
		ID        uuid.UUID
		GBIFKey   int64
		SpeciesID uuid.UUID
	}
	// CreateParams represents the parameters required to
	// create a new sighting.
	CreateParams struct {
		GBIFKey         int64     `json:"gbifKey"`
		OccurrenceID    string    `json:"occurrenceID,omitzero"`
		HappenedAt      time.Time `json:"happenedAt"`
		BasisOfRecord   string    `json:"basisOfRecord,omitzero"`
		RecordedBy      string    `json:"recordedBy,omitzero"`
		IndividualCount *int      `json:"individualCount,omitzero"`
		Remarks         string    `json:"remarks,omitzero"`
		ReferenceURL    string    `json:"referenceURL,omitzero"`
		Location        Location  `json:"location"`
		SpeciesID       uuid.UUID `json:"speciesID,omitzero"`
		WeatherID       uuid.UUID `json:"weatherID,omitzero"`
		Media           []Media   `json:"media,omitzero"`
		Organism        Organism  `json:"organism,omitzero"`
		IdentifiedBy    string    `json:"identifiedBy,omitzero"`
		DatasetName     string    `json:"datasetName,omitzero"`
	}
	// Media represents a media item associated with a sighting.
	// This could be abstracted into its own domain type, but
	// for now, it's simple and stored alongside the sighting.
	Media struct {
		ID           uuid.UUID `json:"id"`
		Type         string    `json:"type"`
		Format       string    `json:"format"`
		URL          string    `json:"url"`
		Title        string    `json:"title,omitzero"`
		Description  string    `json:"description,omitzero"`
		Created      time.Time `json:"created,omitzero"`
		Creator      string    `json:"creator,omitzero"`
		License      string    `json:"license,omitzero"`
		RightsHolder string    `json:"rightsHolder,omitzero"`
	}
)

// CreateParams returns the sighting fields without its ID or timestamps.
func (s Sighting) CreateParams() CreateParams {
	return CreateParams{
		GBIFKey:         s.GBIFKey,
		OccurrenceID:    s.OccurrenceID,
		HappenedAt:      s.HappenedAt,
		BasisOfRecord:   s.BasisOfRecord,
		RecordedBy:      s.RecordedBy,
		IndividualCount: s.IndividualCount,
		Remarks:         s.Remarks,
		ReferenceURL:    s.ReferenceURL,
		Location:        s.Location,
		SpeciesID:       s.Species.ID,
		WeatherID:       s.Weather.ID,
		Media:           s.Media,
		Organism:        s.Organism,
		IdentifiedBy:    s.IdentifiedBy,
		DatasetName:     s.DatasetName,
	}
}

// IsZero reports whether the observer recorded nothing about the animal.
func (o Organism) IsZero() bool {
	return o == Organism{}
}

// ParseCoordinates returns the coordinates for a latitude and longitude, and
// is how numbers that came from outside become a position. Values out of
// range are an error rather than a point nobody can plot, so every
// Coordinates that exists is one the rest of the application can trust.
func ParseCoordinates(latitude, longitude float64) (Coordinates, error) {
	if latitude < -90 || latitude > 90 {
		return Coordinates{}, fmt.Errorf("%w: latitude must be between -90 and 90, got %v", ErrInvalid, latitude)
	}
	if longitude < -180 || longitude > 180 {
		return Coordinates{}, fmt.Errorf("%w: longitude must be between -180 and 180, got %v", ErrInvalid, longitude)
	}

	return Coordinates{Latitude: latitude, Longitude: longitude}, nil
}
