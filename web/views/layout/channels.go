package layout

import "fmt"

// Channel is one page of the site, tuned in from the nav or the knobs on the set.
type Channel struct {
	Href  string
	Label string
}

var channels = []Channel{
	{"/", "Sightings"},
	{"/map", "Map"},
	{"/gallery", "Gallery"},
	{"/animals", "Animals"},
	{"/countries", "Countries"},
	{"/stats", "Stats"},
}

// channelIndex returns the position of the active page in channels, or -1.
func channelIndex(active string) int {
	for i, c := range channels {
		if c.Label == active {
			return i
		}
	}
	return -1
}

// channelName returns the on-screen channel label, such as "CH 02".
func channelName(active string) string {
	i := channelIndex(active)
	if i < 0 {
		return "AV"
	}
	return fmt.Sprintf("CH %02d", i+1)
}

// tune returns the channel step places away from the active one, wrapping round.
func tune(active string, step int) Channel {
	i := channelIndex(active)
	n := len(channels)
	if i < 0 {
		if step > 0 {
			return channels[0]
		}
		return channels[n-1]
	}
	return channels[((i+step)%n+n)%n]
}

// tvClass returns the classes that turn the knob to the active channel.
func tvClass(active string) string {
	return fmt.Sprintf("tv tv--ch-%d", max(channelIndex(active), 0))
}
