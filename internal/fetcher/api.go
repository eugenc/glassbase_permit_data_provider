package fetcher

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
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

func expandAPIBodyTemplate(body string, fc *FetchContext) string {
	if !strings.Contains(body, "{{PAGE}}") && !strings.Contains(body, "{{PAGE_SIZE}}") {
		return body
	}
	page := 1
	size := 25
	if fc != nil {
		if fc.PageNum > 0 {
			page = fc.PageNum
		}
		if fc.PageSize > 0 {
			size = fc.PageSize
		}
	}
	body = strings.ReplaceAll(body, "{{PAGE}}", strconv.Itoa(page))
	body = strings.ReplaceAll(body, "{{PAGE_SIZE}}", strconv.Itoa(size))
	return body
}

func (f *APIFetcher) Fetch(ctx context.Context, url string, fc *FetchContext) (*FetchResult, error) {
	client := &http.Client{Timeout: 30 * time.Second}

	// Prefer the URL built by the paginator (offsets, pages, tokens). Older connectors also set
	// api.endpoint; when both are present, overwriting with Endpoint discards paging and repeats
	// the same page forever (classic Socrata symptom: ~rows_affected=1000 and flat COUNT(*)).
	target := strings.TrimSpace(url)
	if target == "" {
		target = strings.TrimSpace(f.Endpoint)
	}
	if target == "" {
		return nil, fmt.Errorf("api fetch: empty URL")
	}

	log.Printf("fetcher/api: %s %s", f.methodOrDefault(), target)

	var bodyReader io.Reader
	if f.Body != "" {
		payload := expandAPIBodyTemplate(f.Body, fc)
		bodyReader = strings.NewReader(payload)
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
	// Full response body is buffered here; callers parse and persist only after this returns.

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("api fetch %s: HTTP %d", target, resp.StatusCode)
	}

	s := strings.TrimSpace(string(body))
	if len(s) == 0 {
		return nil, fmt.Errorf("api fetch %s: empty body", target)
	}

	log.Printf("fetcher/api: OK status=%d bytes=%d", resp.StatusCode, len(body))
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
