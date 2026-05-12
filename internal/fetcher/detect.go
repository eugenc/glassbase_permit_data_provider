package fetcher

import (
	"context"
	"log"
	"strings"
)

// DetectSourceType fetches the URL with the HTML fetcher first,
// then checks signals to determine the true source type.
func DetectSourceType(ctx context.Context, url string) (string, error) {
	log.Printf("fetcher/detect: probe GET (html fetcher) url=%s", url)
	html := &HTMLFetcher{}
	result, err := html.Fetch(ctx, url, nil)
	if err != nil {
		return "", err
	}

	body := result.Body
	trimmed := strings.TrimSpace(body)
	log.Printf("fetcher/detect: initial response bytes=%d trimmed_runes=%d", len(body), len([]rune(trimmed)))

	if len(trimmed) < 1000 {
		log.Printf("fetcher/detect: chose spa (short body < 1000)")
		return "spa", nil
	}

	spaMarkers := []string{
		"ng-app", "data-reactroot", "__NEXT_DATA__", "nuxt", "vue",
		"window.__INITIAL_STATE__", "application/json",
	}
	for _, marker := range spaMarkers {
		if strings.Contains(body, marker) {
			log.Printf("fetcher/detect: chose spa (marker %q)", marker)
			return "spa", nil
		}
	}

	if strings.Contains(url, "/api/") || strings.HasSuffix(url, ".json") {
		log.Printf("fetcher/detect: chose api (url pattern)")
		return "api", nil
	}

	log.Printf("fetcher/detect: chose html (default)")
	return "html", nil
}
