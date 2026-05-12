package fetcher

import "context"

// FetchResult is what every fetcher returns.
type FetchResult struct {
	Body         string
	StatusCode   int
	Headers      map[string]string
	SourceType   string
	NetworkCalls []NetworkCall
}

// NetworkCall captures intercepted HTTP during SPA render.
type NetworkCall struct {
	URL      string
	Method   string
	Body     string
	Response string
}

// Fetcher fetches a URL and returns content.
type Fetcher interface {
	Fetch(ctx context.Context, url string) (*FetchResult, error)
}

// New returns a fetcher for sourceType: "html", "spa", or "api".
// For API mode with custom endpoint/headers, use NewWithOptions.
func New(sourceType string) Fetcher {
	return NewWithOptions(Options{SourceType: sourceType})
}

// Options configures fetcher construction (API credentials, etc.).
type Options struct {
	SourceType string
	API        *APIFetcher
}

// NewWithOptions returns the appropriate fetcher; API fields override when SourceType is "api".
func NewWithOptions(opts Options) Fetcher {
	switch opts.SourceType {
	case "spa":
		return &SPAFetcher{}
	case "api":
		if opts.API != nil {
			return opts.API
		}
		return &APIFetcher{}
	default:
		return &HTMLFetcher{}
	}
}
