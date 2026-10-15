package species

import (
	"fmt"
	"strings"
)

// IUCNCategory is a species' status on the IUCN Red List, as the two
// letter code GBIF reports it. The empty value means it hasn't been
// assessed, or GBIF didn't say.
type IUCNCategory string

// The IUCN Red List categories, from gone to doing fine.
//
// See: https://www.iucnredlist.org/about/faqs
const (
	IUCNExtinct              IUCNCategory = "EX"
	IUCNExtinctInTheWild     IUCNCategory = "EW"
	IUCNCriticallyEndangered IUCNCategory = "CR"
	IUCNEndangered           IUCNCategory = "EN"
	IUCNVulnerable           IUCNCategory = "VU"
	IUCNNearThreatened       IUCNCategory = "NT"
	IUCNLeastConcern         IUCNCategory = "LC"
	IUCNDataDeficient        IUCNCategory = "DD"
	IUCNNotEvaluated         IUCNCategory = "NE"
	IUCNUnknown              IUCNCategory = ""
)

// iucnCategories holds each category's name and its place on the scale,
// most at risk first.
var iucnCategories = map[IUCNCategory]struct {
	label string
	rank  int
}{
	IUCNExtinct:              {"Extinct", 1},
	IUCNExtinctInTheWild:     {"Extinct in the wild", 2},
	IUCNCriticallyEndangered: {"Critically endangered", 3},
	IUCNEndangered:           {"Endangered", 4},
	IUCNVulnerable:           {"Vulnerable", 5},
	IUCNNearThreatened:       {"Near threatened", 6},
	IUCNLeastConcern:         {"Least concern", 7},
	IUCNDataDeficient:        {"Data deficient", 8},
	IUCNNotEvaluated:         {"Not evaluated", 9},
}

// ParseIUCNCategory converts a code such as "VU" or "vu" into a category.
// GBIF also spells some out in full, such as "VULNERABLE", so those are
// accepted too. An empty value parses to IUCNUnknown.
func ParseIUCNCategory(value string) (IUCNCategory, error) {
	v := strings.ToUpper(strings.TrimSpace(value))
	if v == "" {
		return IUCNUnknown, nil
	}

	if _, ok := iucnCategories[IUCNCategory(v)]; ok {
		return IUCNCategory(v), nil
	}
	for code, c := range iucnCategories {
		if strings.EqualFold(strings.ReplaceAll(c.label, " ", "_"), v) {
			return code, nil
		}
	}

	return IUCNUnknown, fmt.Errorf("unknown IUCN category %q", value)
}

// String implements fmt.Stringer on IUCNCategory.
func (c IUCNCategory) String() string {
	return string(c)
}

// Label is the category in words, such as "Vulnerable".
func (c IUCNCategory) Label() string {
	if v, ok := iucnCategories[c]; ok {
		return v.label
	}
	return "Not assessed"
}

// Rank orders the categories most at risk first, with an unknown
// category last.
func (c IUCNCategory) Rank() int {
	if v, ok := iucnCategories[c]; ok {
		return v.rank
	}
	return len(iucnCategories) + 1
}

// Threatened reports whether the Red List counts the species as
// threatened: critically endangered, endangered or vulnerable.
func (c IUCNCategory) Threatened() bool {
	switch c {
	case IUCNCriticallyEndangered, IUCNEndangered, IUCNVulnerable:
		return true
	default:
		return false
	}
}
