package pages

import (
	"net/url"
	"workshop/internal/domain/sighting"
	"workshop/internal/domain/sighting/country"
	"workshop/internal/domain/species"
)

type (
	// Filter is what a page was asked to narrow its sightings to.
	Filter struct {
		Animal  species.Animal
		Country country.Code
	}
	// AnimalTally is an animal and how often it has been seen.
	AnimalTally struct {
		Animal species.Animal
		Count  int
		Class  string
		Status species.IUCNCategory
		Photo  string
	}
	// CountryTally is a country and how much has been seen there.
	CountryTally struct {
		Code      country.Code
		Name      string
		Continent string
		Count     int
		Species   int
	}
	// Point is a sighting as the map and globe plot it. The fields after
	// Group fill the globe's hologram when a point is clicked.
	Point struct {
		ID          string  `json:"id"`
		Lat         float64 `json:"lat"`
		Lng         float64 `json:"lng"`
		Name        string  `json:"name"`
		Date        string  `json:"date"`
		Place       string  `json:"place"`
		Group       string  `json:"group"`
		Scientific  string  `json:"scientific,omitzero"`
		Family      string  `json:"family,omitzero"`
		Status      string  `json:"status,omitzero"`
		StatusLabel string  `json:"statusLabel,omitzero"`
		Photo       string  `json:"photo,omitzero"`
		Flag        string  `json:"flag,omitzero"`
		Weather     string  `json:"weather,omitzero"`
		Temperature string  `json:"temperature,omitzero"`
	}
	// Photo is a photo in the gallery, with the sighting it shows.
	Photo struct {
		Media    sighting.Media
		Sighting sighting.Sighting
	}
	// StatsProps is everything on the statistics page.
	StatsProps struct {
		Summary         sighting.Summary
		Threatened      int
		MeanTemperature float64
		Charts          []Chart
	}
	// Chart is a bar chart. Columns stand the bars up, for a series
	// such as months that reads left to right.
	Chart struct {
		Title   string
		Note    string
		Bars    []Bar
		Columns bool
	}
	// Bar is one bar of a chart. Tone, if set, picks its colour, such
	// as an IUCN category's code.
	Bar struct {
		Label string
		Count int
		Href  string
		Tone  string
	}
)

// Query is the filter as a query string, without the "?".
func (f Filter) Query() string {
	q := url.Values{}
	if f.Animal != "" {
		q.Set("animal", f.Animal.String())
	}
	if f.Country != "" {
		q.Set("country", f.Country.String())
	}
	return q.Encode()
}

// IsZero reports whether the filter matches everything.
func (f Filter) IsZero() bool {
	return f == Filter{}
}

// Max is the longest bar in the chart, at least one so the bars can be
// drawn as a share of it.
func (c Chart) Max() int {
	m := 1
	for _, b := range c.Bars {
		m = max(m, b.Count)
	}
	return m
}
