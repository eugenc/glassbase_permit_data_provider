package fetcher

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// APIFetcher performs HTTP requests for JSON/XML API sources.
// Endpoint, Method, Headers, Body come from connector_config when set.
type APIFetcher struct {
	Endpoint string
	Method   string
	Headers  map[string]string
	Body     string
}

func (f *APIFetcher) Fetch(ctx context.Context, url string) (*FetchResult, error) {
	client := &http.Client{Timeout: 30 * time.Second}

	target := url
	if f.Endpoint != "" {
		target = f.Endpoint
	}

	var bodyReader io.Reader
	if f.Body != "" {
		bodyReader = strings.NewReader(f.Body)
	}

	req, err := http.NewRequestWithContext(ctx, f.methodOrDefault(), target, bodyReader)
	if err != nil {
		return nil, err
	}

	for k, v := range f.Headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("api fetch %s: %w", target, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("api fetch %s: HTTP %d", target, resp.StatusCode)
	}

	s := strings.TrimSpace(string(body))
	if len(s) == 0 {
		return nil, fmt.Errorf("api fetch %s: empty body", target)
	}

	return &FetchResult{
		Body:       string(body),
		StatusCode: resp.StatusCode,
		SourceType: "api",
	}, nil
}

func (f *APIFetcher) methodOrDefault() string {
	if f.Method != "" {
		return f.Method
	}
	return http.MethodGet
}
