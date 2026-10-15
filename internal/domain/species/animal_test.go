package species

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAnimal_TaxonKeys(t *testing.T) {
	t.Parallel()

	tt := map[string]struct {
		input Animal
		want  []int
	}{
		"One species":                 {input: AnimalHumpbackWhale, want: []int{5220086}},
		"Several species":             {input: AnimalZebra, want: []int{2440892, 2440888, 2440894}},
		"A genus key":                 {input: AnimalGiraffe, want: []int{7716694}},
		"A group and its own species": {input: AnimalAlligator, want: []int{2441370, 2441368}},
		"Unknown animal":              {input: Animal("pangolin")},
		"Empty animal":                {input: Animal("")},
	}

	for name, test := range tt {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, test.want, test.input.TaxonKeys())
		})
	}
}

func TestAnimal_TaxonKeysAreCopied(t *testing.T) {
	t.Parallel()

	keys := AnimalZebra.TaxonKeys()
	keys[0] = 0

	assert.Equal(t, []int{2440892, 2440888, 2440894}, AnimalZebra.TaxonKeys())
}

func TestAnimal_Known(t *testing.T) {
	t.Parallel()

	tt := map[string]struct {
		input Animal
		want  bool
	}{
		"Supported":   {input: AnimalLion, want: true},
		"Unsupported": {input: Animal("pangolin")},
		"Empty":       {input: Animal("")},
	}

	for name, test := range tt {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, test.want, test.input.Known())
		})
	}
}

func TestAnimals(t *testing.T) {
	t.Parallel()

	all := Animals()

	assert.NotEmpty(t, all)
	assert.IsIncreasing(t, all, "listed in name order so the CLI is stable")
	assert.Contains(t, all, AnimalHumpbackWhale)

	for _, a := range all {
		assert.True(t, a.Known(), "every listed animal resolves to at least one taxon")
	}
}

func TestParseAnimal(t *testing.T) {
	t.Parallel()

	tt := map[string]struct {
		input   string
		want    Animal
		wantErr bool
	}{
		"Exact name":    {input: "humpback_whale", want: AnimalHumpbackWhale},
		"Title case":    {input: "Humpback Whale", want: AnimalHumpbackWhale},
		"Hyphenated":    {input: "humpback-whale", want: AnimalHumpbackWhale},
		"Padded":        {input: "  lion  ", want: AnimalLion},
		"Unsupported":   {input: "pangolin", wantErr: true},
		"Empty":         {input: "", wantErr: true},
		"Only spacing":  {input: "   ", wantErr: true},
		"Mangled input": {input: "humpback  whale", wantErr: true},
	}

	for name, test := range tt {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := ParseAnimal(test.input)
			assert.Equal(t, test.wantErr, err != nil)
			assert.Equal(t, test.want, got)
		})
	}
}

func TestAnimal_String(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "humpback_whale", AnimalHumpbackWhale.String())
}

func TestAnimalOf(t *testing.T) {
	t.Parallel()

	t.Run("Found", func(t *testing.T) {
		t.Parallel()
		got, ok := AnimalOf(2435349) // Loxodonta cyclotis
		assert.True(t, ok)
		assert.Equal(t, AnimalElephant, got)
	})

	t.Run("Prefers the narrower animal", func(t *testing.T) {
		t.Parallel()
		got, ok := AnimalOf(2441370) // Alligator mississippiensis
		assert.True(t, ok)
		assert.Equal(t, AnimalAmericanAlligator, got)
	})

	t.Run("Not found", func(t *testing.T) {
		t.Parallel()
		_, ok := AnimalOf(1)
		assert.False(t, ok)
	})
}

func TestAnimal_Name(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "Spotted hyena", AnimalSpottedHyena.Name())
	assert.Equal(t, "Lion", AnimalLion.Name())
}
