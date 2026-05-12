package fetcher

import (
	"context"
	"strings"
)

// DetectSourceType fetches the URL with the HTML fetcher first,
// then checks signals to determine the true source type.
func DetectSourceType(ctx context.Context, url string) (string, error) {
	html := &HTMLFetcher{}
	result, err := html.Fetch(ctx, url)
	if err != nil {
		return "", err
	}

	body := result.Body

	if len(strings.TrimSpace(body)) < 1000 {
		return "spa", nil
	}

	spaMarkers := []string{
		"ng-app", "data-reactroot", "__NEXT_DATA__", "nuxt", "vue",
		"window.__INITIAL_STATE__", "application/json",
	}
	for _, marker := range spaMarkers {
		if strings.Contains(body, marker) {
			return "spa", nil
		}
	}

	if strings.Contains(url, "/api/") || strings.HasSuffix(url, ".json") {
		return "api", nil
	}

	return "html", nil
}
