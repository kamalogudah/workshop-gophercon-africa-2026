package species_test

import (
	"testing"
	"workshop/internal/domain/species"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseIUCNCategory(t *testing.T) {
	t.Parallel()

	tt := map[string]struct {
		input   string
		want    species.IUCNCategory
		wantErr bool
	}{
		"Code":       {input: "VU", want: species.IUCNVulnerable},
		"Lower case": {input: " en ", want: species.IUCNEndangered},
		"Spelt out":  {input: "CRITICALLY_ENDANGERED", want: species.IUCNCriticallyEndangered},
		"Empty":      {input: "", want: species.IUCNUnknown},
		"Unknown":    {input: "ZZ", want: species.IUCNUnknown, wantErr: true},
	}

	for name, test := range tt {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := species.ParseIUCNCategory(test.input)
			if test.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, test.want, got)
		})
	}
}

func TestIUCNCategory(t *testing.T) {
	t.Parallel()

	t.Run("Label", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "Near threatened", species.IUCNNearThreatened.Label())
		assert.Equal(t, "Not assessed", species.IUCNUnknown.Label())
	})

	t.Run("Rank orders most at risk first", func(t *testing.T) {
		t.Parallel()
		assert.Less(t, species.IUCNCriticallyEndangered.Rank(), species.IUCNVulnerable.Rank())
		assert.Less(t, species.IUCNLeastConcern.Rank(), species.IUCNUnknown.Rank())
	})

	t.Run("Threatened", func(t *testing.T) {
		t.Parallel()
		assert.True(t, species.IUCNEndangered.Threatened())
		assert.False(t, species.IUCNNearThreatened.Threatened())
		assert.False(t, species.IUCNUnknown.Threatened())
	})
}
