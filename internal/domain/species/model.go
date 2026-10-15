package species

import (
	"time"
	"uuid"
)

type (
	// Species represents a taxonomy record as GBIF classifies it.
	// It can be an animal or a plant, at species rank or above.
	Species struct {
		ID             uuid.UUID    `json:"id"`
		GBIFKey        int          `json:"gbifKey"`
		Kingdom        string       `json:"kingdom,omitzero"`
		Phylum         string       `json:"phylum,omitzero"`
		Class          string       `json:"class,omitzero"`
		Order          string       `json:"order,omitzero"`
		Family         string       `json:"family,omitzero"`
		Genus          string       `json:"genus,omitzero"`
		Species        string       `json:"species,omitzero"`
		Rank           string       `json:"rank,omitzero"`
		ScientificName string       `json:"scientificName,omitzero"`
		CanonicalName  string       `json:"canonicalName,omitzero"`
		VernacularName string       `json:"vernacularName,omitzero"`
		GenusKey       int          `json:"genusKey,omitzero"`
		IUCNCategory   IUCNCategory `json:"iucnCategory,omitzero"`
		CreatedAt      time.Time    `json:"createdAt"`
		UpdatedAt      time.Time    `json:"updatedAt"`
	}
	// CreateParams represents the parameters required to
	// create a new species.
	CreateParams struct {
		GBIFKey        int          `json:"gbifKey"`
		Kingdom        string       `json:"kingdom,omitzero"`
		Phylum         string       `json:"phylum,omitzero"`
		Class          string       `json:"class,omitzero"`
		Order          string       `json:"order,omitzero"`
		Family         string       `json:"family,omitzero"`
		Genus          string       `json:"genus,omitzero"`
		Species        string       `json:"species,omitzero"`
		Rank           string       `json:"rank,omitzero"`
		ScientificName string       `json:"scientificName,omitzero"`
		CanonicalName  string       `json:"canonicalName,omitzero"`
		VernacularName string       `json:"vernacularName,omitzero"`
		GenusKey       int          `json:"genusKey,omitzero"`
		IUCNCategory   IUCNCategory `json:"iucnCategory,omitzero"`
	}
	// Details are the parts of a species that were added after it was
	// first released. A zero field leaves the stored value as it is.
	Details struct {
		GenusKey     int
		IUCNCategory IUCNCategory
	}
)

// Animal returns the animal the species belongs to, matching on the
// species first and then its genus.
func (s Species) Animal() (Animal, bool) {
	if a, ok := AnimalOf(s.GBIFKey); ok {
		return a, true
	}
	if s.GenusKey == 0 {
		return "", false
	}
	return AnimalOf(s.GenusKey)
}

// CreateParams returns the species fields without its ID or timestamps.
func (s Species) CreateParams() CreateParams {
	return CreateParams{
		GBIFKey:        s.GBIFKey,
		Kingdom:        s.Kingdom,
		Phylum:         s.Phylum,
		Class:          s.Class,
		Order:          s.Order,
		Family:         s.Family,
		Genus:          s.Genus,
		Species:        s.Species,
		Rank:           s.Rank,
		ScientificName: s.ScientificName,
		CanonicalName:  s.CanonicalName,
		VernacularName: s.VernacularName,
		GenusKey:       s.GenusKey,
		IUCNCategory:   s.IUCNCategory,
	}
}
