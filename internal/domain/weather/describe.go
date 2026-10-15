package weather

// compassPoints are the sixteen points of the compass, clockwise from
// north, each covering 22.5 degrees.
var compassPoints = [...]string{
	"N", "NNE", "NE", "ENE", "E", "ESE", "SE", "SSE",
	"S", "SSW", "SW", "WSW", "W", "WNW", "NW", "NNW",
}

// Compass returns the point of the compass the wind blew from, such as
// "NE" or "WSW".
func (w Wind) Compass() string {
	degrees := ((w.Direction % 360) + 360) % 360
	return compassPoints[int((float64(degrees)+11.25)/22.5)%len(compassPoints)]
}

// beaufort holds the upper bound in km/h of each force on the Beaufort
// scale below 12, and its name.
var beaufort = [...]struct {
	below float64
	name  string
}{
	{1, "Calm"},
	{6, "Light air"},
	{12, "Light breeze"},
	{20, "Gentle breeze"},
	{29, "Moderate breeze"},
	{39, "Fresh breeze"},
	{50, "Strong breeze"},
	{62, "Near gale"},
	{75, "Gale"},
	{89, "Strong gale"},
	{103, "Storm"},
	{118, "Violent storm"},
}

// Beaufort returns the wind's force on the Beaufort scale, from 0 for
// calm to 12 for a hurricane, and its name.
func (w Wind) Beaufort() (force int, name string) {
	for i, f := range beaufort {
		if w.Speed < f.below {
			return i, f.name
		}
	}
	return len(beaufort), "Hurricane force"
}
