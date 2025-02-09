package utils

import "time"

type Layout struct {
	Pattern   string
	Timezoned bool
}

func timezoned(layout string) Layout {
	return Layout{
		Pattern:   layout,
		Timezoned: true,
	}
}

func timezoneless(layout string) Layout {
	return Layout{
		Pattern:   layout,
		Timezoned: false,
	}
}

var AcceptedLayouts = []Layout{
	// Sorted from most probably to less probably
	timezoneless("DD/MM/YYYY"),
	timezoneless("2006-1-2"),
	timezoneless("1/2/2006"),
	timezoned(time.RFC3339),
}

func ParseDateLikeInput(element string) (out time.Time, parsedFrom Layout, ok bool) {
	for _, layout := range AcceptedLayouts {
		parsed, err := time.Parse(layout.Pattern, element)
		if err == nil {
			if !layout.Timezoned {
				return parsed.Local(), layout, true
			}

			return parsed, layout, true
		}
	}

	return
}
