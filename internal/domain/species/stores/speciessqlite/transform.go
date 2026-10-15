package speciessqlite

import (
	"uuid"
	"workshop/internal/domain/species"
	"workshop/internal/infra/db/sqlc"
)

func transform(r db.Species) species.Species {
	return species.Species{
		ID:             r.ID,
		GBIFKey:        int(r.GbifKey),
		Kingdom:        r.Kingdom,
		Phylum:         r.Phylum,
		Class:          r.Class,
		Order:          r.Order,
		Family:         r.Family,
		Genus:          r.Genus,
		Species:        r.Species,
		Rank:           r.Rank,
		ScientificName: r.ScientificName,
		CanonicalName:  r.CanonicalName,
		VernacularName: r.VernacularName,
		GenusKey:       int(r.GenusKey),
		IUCNCategory:   species.IUCNCategory(r.IucnCategory),
		CreatedAt:      r.CreatedAt,
		UpdatedAt:      r.UpdatedAt,
	}
}

func toSpeciesCreateParams(id uuid.UUID, in species.CreateParams) db.SpeciesCreateParams {
	return db.SpeciesCreateParams{
		ID:             id,
		GbifKey:        int64(in.GBIFKey),
		Kingdom:        in.Kingdom,
		Phylum:         in.Phylum,
		Class:          in.Class,
		Order:          in.Order,
		Family:         in.Family,
		Genus:          in.Genus,
		Species:        in.Species,
		Rank:           in.Rank,
		ScientificName: in.ScientificName,
		CanonicalName:  in.CanonicalName,
		VernacularName: in.VernacularName,
		GenusKey:       int64(in.GenusKey),
		IucnCategory:   in.IUCNCategory.String(),
	}
}
