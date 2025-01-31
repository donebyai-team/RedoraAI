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
	timezoneless("2006-1-2"),
	timezoneless("1/2/2006"),
	timezoned(time.RFC3339),
	timezoneless("01/2006"),
	timezoneless("2006"),
	timezoneless("2006-1"), // Add this as saw there are few cases in PEP Active where the endDate is parsed in this format
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
