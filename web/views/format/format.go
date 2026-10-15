// Package format turns domain values into the words and symbols the
// pages show. It decides nothing about the domain, only how it reads.
package format

import (
	"cmp"
	"fmt"
	"slices"
	"strings"
	"time"
	"unicode"
	"workshop/internal/domain/sighting"
	"workshop/internal/domain/sighting/country"
	"workshop/internal/domain/species"
)

// DisplayName prefers the name a reader would recognise, falling back
// through the taxonomy as GBIF records thin out.
func DisplayName(sp species.Species) string {
	name := []rune(cmp.Or(sp.VernacularName, sp.CanonicalName, sp.ScientificName, "Unknown species"))
	name[0] = unicode.ToUpper(name[0])
	return string(name)
}

// Place joins whichever parts of the location GBIF supplied.
func Place(l sighting.Location) string {
	parts := slices.DeleteFunc(
		[]string{l.Locality, l.StateProvince, l.Country},
		func(s string) bool { return s == "" },
	)
	return cmp.Or(strings.Join(parts, ", "), "Location unknown")
}

// Individuals renders how many animals were seen in a single sighting.
func Individuals(n int) string {
	if n == 1 {
		return "1 individual"
	}
	return fmt.Sprintf("%d individuals", n)
}

// Date is a day as the site writes it, such as "2 March 2024".
func Date(t time.Time) string {
	return t.Format("2 January 2006")
}

// When is the date and, if GBIF recorded one, the time of day. GBIF
// fills a missing time with midnight, so midnight is left off.
func When(t time.Time) string {
	if t.Hour() == 0 && t.Minute() == 0 {
		return Date(t)
	}
	return t.Format("2 January 2006, 15:04 MST")
}

// Group names the broad kind of animal, such as "Mammals" or "Birds",
// from its class.
func Group(sp species.Species) string {
	switch sp.Class {
	case "Mammalia":
		return "Mammals"
	case "Aves":
		return "Birds"
	case "Crocodylia", "Testudines", "Squamata", "Reptilia":
		return "Reptiles"
	case "Elasmobranchii", "Chondrichthyes":
		return "Sharks & rays"
	case "Amphibia":
		return "Amphibians"
	case "":
		return "Other"
	default:
		return sp.Class
	}
}

// Humanise turns GBIF's upper snake case, such as "NORTH_AMERICA" or
// "HUMAN_OBSERVATION", into words: "North America", "Human observation".
func Humanise(s string) string {
	if s == "" {
		return ""
	}
	s = strings.ToLower(strings.ReplaceAll(s, "_", " "))
	return strings.ToUpper(s[:1]) + s[1:]
}

// Continent is the continent's name, such as "North America".
func Continent(s string) string {
	words := strings.Fields(strings.ToLower(strings.ReplaceAll(s, "_", " ")))
	for i, w := range words {
		words[i] = strings.ToUpper(w[:1]) + w[1:]
	}
	return strings.Join(words, " ")
}

// Flag is the emoji flag of a country, built from the regional
// indicator symbols its ISO code spells.
func Flag(code country.Code) string {
	c := strings.ToUpper(code.String())
	if len(c) != 2 || c[0] < 'A' || c[0] > 'Z' || c[1] < 'A' || c[1] > 'Z' {
		return "🏳️"
	}
	const base = 0x1F1E6
	return string([]rune{rune(base + int(c[0]-'A')), rune(base + int(c[1]-'A'))})
}

// WeatherIcon is an emoji for a WMO weather code.
func WeatherIcon(code int) string {
	switch {
	case code == 0:
		return "☀️"
	case code == 1:
		return "🌤️"
	case code == 2:
		return "⛅"
	case code == 3:
		return "☁️"
	case code == 45 || code == 48:
		return "🌫️"
	case code >= 51 && code <= 57:
		return "🌦️"
	case code >= 61 && code <= 67:
		return "🌧️"
	case code >= 71 && code <= 77:
		return "❄️"
	case code >= 80 && code <= 82:
		return "🌦️"
	case code == 85 || code == 86:
		return "🌨️"
	case code >= 95:
		return "⛈️"
	default:
		return "🌡️"
	}
}

// Temperature is a temperature in degrees Celsius, to the nearest degree.
func Temperature(c float64) string {
	return fmt.Sprintf("%.0f°C", c)
}

// FirstPhoto returns the URL of the sighting's first photo, if it has one.
func FirstPhoto(s sighting.Sighting) string {
	for _, m := range s.Media {
		if m.Type == "StillImage" && m.URL != "" {
			return m.URL
		}
	}
	return ""
}

// PhotoCount is how many photos the sighting has.
func PhotoCount(s sighting.Sighting) int {
	var n int
	for _, m := range s.Media {
		if m.Type == "StillImage" && m.URL != "" {
			n++
		}
	}
	return n
}

// AltText describes a media item for a reader who can't see it.
func AltText(m sighting.Media) string {
	return cmp.Or(m.Title, m.Description, "Sighting media")
}

// SightingURL is the address of a sighting's page.
func SightingURL(s sighting.Sighting) string {
	return "/sightings/" + s.ID.String()
}

// Plural writes n and the noun, adding an s unless n is one.
func Plural(n int, noun string) string {
	if n == 1 {
		return "1 " + noun
	}
	return fmt.Sprintf("%d %ss", n, noun)
}

// Number writes n with commas between the thousands.
func Number(n int) string {
	s := fmt.Sprint(n)
	if n < 0 {
		return "-" + Number(-n)
	}
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "," + s[i:]
	}
	return s
}
