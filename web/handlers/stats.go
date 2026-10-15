package handlers

import (
	"cmp"
	"fmt"
	"log/slog"
	"maps"
	"net/http"
	"slices"
	"time"
	"workshop/internal/domain/sighting"
	"workshop/internal/domain/species"
	"workshop/web/views/format"
	"workshop/web/views/pages"
)

// Stats renders the statistics page, worked out over every sighting.
func Stats(logger *slog.Logger, sightings SightingsLister) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		all, ok := listOrFail(w, r, logger, sightings, sighting.ListFilter{})
		if !ok {
			return
		}

		render(w, r, logger, pages.Stats(statsOf(all)))
	}
}

// statsOf works out the figures and charts on the statistics page.
func statsOf(all []sighting.Sighting) pages.StatsProps {
	props := pages.StatsProps{Summary: sighting.Summarise(all)}
	if len(all) == 0 {
		return props
	}

	threatened := make(map[string]struct{})
	var temperature float64
	for _, s := range all {
		if s.Species.IUCNCategory.Threatened() {
			threatened[s.Species.ScientificName] = struct{}{}
		}
		temperature += s.Weather.Temperature.Actual
	}
	props.Threatened = len(threatened)
	props.MeanTemperature = temperature / float64(len(all))

	props.Charts = []pages.Chart{
		topAnimals(all),
		topCountries(all),
		{
			Title: "By continent",
			Bars:  bars(sighting.Top(sighting.CountBy(all, func(s sighting.Sighting) string { return format.Continent(s.Location.Continent) }), 0), nil),
		},
		{
			Title: "By class",
			Note:  "Mammals, birds, reptiles and the rest.",
			Bars:  bars(sighting.Top(sighting.CountBy(all, func(s sighting.Sighting) string { return format.Group(s.Species) }), 0), nil),
		},
		conservation(all),
		byYear(all),
		byMonth(all),
		{
			Title: "Weather at the time",
			Bars:  bars(sighting.Top(sighting.CountBy(all, func(s sighting.Sighting) string { return s.Weather.Condition.Description }), 8), nil),
		},
		temperatures(all),
		winds(all),
		recorded("Sex", all, func(s sighting.Sighting) string { return s.Organism.Sex }),
		recorded("Life stage", all, func(s sighting.Sighting) string { return s.Organism.LifeStage }),
		{
			Title: "Most active recorders",
			Bars:  bars(sighting.Top(sighting.CountBy(all, func(s sighting.Sighting) string { return s.RecordedBy }), 8), nil),
		},
	}

	return props
}

// bars turns tallies into bars, linking each to href(name) when given.
func bars(tallies []sighting.Tally, href func(name string) string) []pages.Bar {
	out := make([]pages.Bar, len(tallies))
	for i, t := range tallies {
		out[i] = pages.Bar{Label: t.Name, Count: t.Count}
		if href != nil {
			out[i].Href = href(t.Name)
		}
	}
	return out
}

func topAnimals(all []sighting.Sighting) pages.Chart {
	links := make(map[string]string)
	counts := sighting.CountBy(all, func(s sighting.Sighting) string {
		name := format.DisplayName(s.Species)
		if a, ok := s.Species.Animal(); ok {
			links[name] = "/animals/" + a.String()
		}
		return name
	})
	return pages.Chart{
		Title: "Most seen species",
		Bars:  bars(sighting.Top(counts, 10), func(name string) string { return links[name] }),
	}
}

func topCountries(all []sighting.Sighting) pages.Chart {
	links := make(map[string]string)
	counts := sighting.CountBy(all, func(s sighting.Sighting) string {
		links[s.Location.Country] = "/countries/" + s.Location.CountryCode.String()
		return s.Location.Country
	})
	return pages.Chart{
		Title: "Top countries",
		Bars:  bars(sighting.Top(counts, 10), func(name string) string { return links[name] }),
	}
}

// conservation counts sightings by their species' Red List category,
// most at risk first.
func conservation(all []sighting.Sighting) pages.Chart {
	counts := make(map[species.IUCNCategory]int)
	for _, s := range all {
		counts[s.Species.IUCNCategory]++
	}

	categories := slices.SortedFunc(maps.Keys(counts), func(a, b species.IUCNCategory) int {
		return cmp.Compare(a.Rank(), b.Rank())
	})
	out := make([]pages.Bar, len(categories))
	for i, c := range categories {
		out[i] = pages.Bar{Label: c.Label(), Count: counts[c], Tone: string(c)}
	}

	return pages.Chart{
		Title: "Conservation status",
		Note:  "The IUCN Red List category of each sighting's species.",
		Bars:  out,
	}
}

func byYear(all []sighting.Sighting) pages.Chart {
	counts := sighting.CountByYear(all)
	years := slices.Sorted(maps.Keys(counts))
	out := make([]pages.Bar, len(years))
	for i, y := range years {
		out[i] = pages.Bar{Label: fmt.Sprint(y), Count: counts[y]}
	}
	return pages.Chart{Title: "By year", Bars: out, Columns: true}
}

func byMonth(all []sighting.Sighting) pages.Chart {
	var counts [12]int
	for _, s := range all {
		counts[s.HappenedAt.Month()-1]++
	}
	out := make([]pages.Bar, 12)
	for i, n := range counts {
		out[i] = pages.Bar{Label: time.Month(i + 1).String()[:3], Count: n}
	}
	return pages.Chart{Title: "By month", Note: "When in the year animals are seen.", Bars: out, Columns: true}
}

// temperatureBands are the upper bounds, in °C, of each band on the
// temperature chart, coldest first.
var temperatureBands = []struct {
	below float64
	label string
}{
	{-10, "Below −10°C"},
	{0, "−10 to 0°C"},
	{10, "0 to 10°C"},
	{20, "10 to 20°C"},
	{30, "20 to 30°C"},
}

func temperatures(all []sighting.Sighting) pages.Chart {
	counts := make([]int, len(temperatureBands)+1)
	for _, s := range all {
		band := len(temperatureBands)
		for i, b := range temperatureBands {
			if s.Weather.Temperature.Actual < b.below {
				band = i
				break
			}
		}
		counts[band]++
	}

	out := make([]pages.Bar, len(counts))
	for i, n := range counts {
		label := "30°C and above"
		if i < len(temperatureBands) {
			label = temperatureBands[i].label
		}
		out[i] = pages.Bar{Label: label, Count: n}
	}
	return pages.Chart{Title: "Temperature", Note: "The air temperature in the hour of the sighting.", Bars: out}
}

func winds(all []sighting.Sighting) pages.Chart {
	counts := make(map[int]int)
	names := make(map[int]string)
	for _, s := range all {
		force, name := s.Weather.Wind.Beaufort()
		counts[force]++
		names[force] = name
	}

	forces := slices.Sorted(maps.Keys(counts))
	out := make([]pages.Bar, len(forces))
	for i, f := range forces {
		out[i] = pages.Bar{Label: fmt.Sprintf("%d · %s", f, names[f]), Count: counts[f]}
	}
	return pages.Chart{Title: "Wind", Note: "On the Beaufort scale.", Bars: out}
}

// recorded charts a detail most observers leave blank, noting how many
// filled it in.
func recorded(title string, all []sighting.Sighting, key func(sighting.Sighting) string) pages.Chart {
	counts := sighting.CountBy(all, key)
	var total int
	for _, n := range counts {
		total += n
	}
	return pages.Chart{
		Title: title,
		Note:  fmt.Sprintf("Recorded for %d of %d sightings.", total, len(all)),
		Bars:  bars(sighting.Top(counts, 6), nil),
	}
}
