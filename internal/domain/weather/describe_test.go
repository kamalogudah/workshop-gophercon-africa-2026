package weather_test

import (
	"testing"
	"workshop/internal/domain/weather"

	"github.com/stretchr/testify/assert"
)

func TestWind_Compass(t *testing.T) {
	t.Parallel()

	tt := map[string]struct {
		input int
		want  string
	}{
		"North":           {0, "N"},
		"Full circle":     {360, "N"},
		"Just west of N":  {350, "N"},
		"North east":      {45, "NE"},
		"Rounds up":       {12, "NNE"},
		"South":           {180, "S"},
		"West south west": {247, "WSW"},
		"Negative":        {-90, "W"},
	}

	for name, test := range tt {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, test.want, weather.Wind{Direction: test.input}.Compass())
		})
	}
}

func TestWind_Beaufort(t *testing.T) {
	t.Parallel()

	tt := map[string]struct {
		input     float64
		wantForce int
		wantName  string
	}{
		"Calm":      {0, 0, "Calm"},
		"Breeze":    {15, 3, "Gentle breeze"},
		"Boundary":  {20, 4, "Moderate breeze"},
		"Gale":      {70, 8, "Gale"},
		"Hurricane": {150, 12, "Hurricane force"},
	}

	for name, test := range tt {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			force, label := weather.Wind{Speed: test.input}.Beaufort()
			assert.Equal(t, test.wantForce, force)
			assert.Equal(t, test.wantName, label)
		})
	}
}
