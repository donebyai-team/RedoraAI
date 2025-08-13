package browser_automation

import (
	"encoding/json"
	"fmt"
	"github.com/playwright-community/playwright-go"
)

type CookieFromJSON struct {
	Name           string   `json:"name"`
	Value          string   `json:"value"`
	Domain         *string  `json:"domain"`
	Path           *string  `json:"path"`
	URL            *string  `json:"url"`
	ExpirationDate *float64 `json:"expirationDate"`
	HttpOnly       *bool    `json:"httpOnly"`
	Secure         *bool    `json:"secure"`
	SameSite       *string  `json:"sameSite"`
}

func ParseCookiesFromJSON(jsonStr string, isUserProvidedCookies bool) ([]playwright.OptionalCookie, error) {
	var rawCookies []playwright.OptionalCookie
	if err := json.Unmarshal([]byte(jsonStr), &rawCookies); err != nil {
		return nil, fmt.Errorf("failed to parse cookie JSON: %w", err)
	}

	if len(rawCookies) == 0 {
		return nil, fmt.Errorf("no cookies found in JSON")
	}

	if !isUserProvidedCookies {
		return rawCookies, nil
	}

	var cookies []playwright.OptionalCookie
	for _, rc := range rawCookies {
		cookie := playwright.OptionalCookie{
			Name:     rc.Name,
			Value:    rc.Value,
			Domain:   rc.Domain,
			Path:     rc.Path,
			URL:      rc.URL,
			HttpOnly: rc.HttpOnly,
			Secure:   rc.Secure,
		}

		if rc.Expires != nil {
			cookie.Expires = rc.Expires
		}

		if rc.SameSite != nil {
			switch *rc.SameSite {
			case "strict":
				s := playwright.SameSiteAttributeStrict
				cookie.SameSite = s
			case "lax":
				s := playwright.SameSiteAttributeLax
				cookie.SameSite = s
			case "no_restriction", "none":
				s := playwright.SameSiteAttributeNone
				cookie.SameSite = s
			}
		}

		cookies = append(cookies, cookie)
	}

	return cookies, nil
}
