package generator

// ConnectorConfig is stored as JSONB in county_connectors.connector_config.
type ConnectorConfig struct {
	SourceType string            `json:"source_type"`
	Extraction Extraction        `json:"extraction"`
	Pagination Pagination        `json:"pagination"`
	Dedup      Dedup             `json:"dedup"`
	RateLimit  RateLimit         `json:"rate_limit"`
	API        *APIRequestConfig `json:"api,omitempty"`
}

// APIRequestConfig optional HTTP details when source_type is "api" or "csv" (GET bulk download).
// Body may contain "{{PAGE}}" and "{{PAGE_SIZE}}" placeholders; the API fetcher substitutes
// the current page index (1-based) and pagination.page_size each request.
type APIRequestConfig struct {
	Endpoint string            `json:"endpoint,omitempty"`
	Method   string            `json:"method,omitempty"`
	Headers  map[string]string `json:"headers,omitempty"`
	Body     string            `json:"body,omitempty"`
}

// Extraction tells the scraper where to find permit records and their fields.
type Extraction struct {
	RecordSelector string         `json:"record_selector,omitempty"`
	RecordsPath    string         `json:"records_path,omitempty"`
	Fields         []FieldMapping `json:"fields"`
}

type FieldMapping struct {
	Name      string `json:"name"`
	Selector  string `json:"selector,omitempty"`
	JSONPath  string `json:"json_path,omitempty"`
	CsvColumn string `json:"csv_column,omitempty"`
	Attr      string `json:"attr,omitempty"`
	Type      string `json:"type"`
}

// Pagination tells the scraper how to get the next page.
type Pagination struct {
	Type           string `json:"type"`
	PageParam      string `json:"page_param,omitempty"`
	PageSize       int    `json:"page_size,omitempty"`
	OffsetParam    string `json:"offset_param,omitempty"`
	CursorParam    string `json:"cursor_param,omitempty"`
	CursorJSONPath string `json:"cursor_json_path,omitempty"`
	NextSelector   string `json:"next_selector,omitempty"`
	MaxPages       int    `json:"max_pages,omitempty"`
}

// Dedup tells the scraper how to identify a permit uniquely.
type Dedup struct {
	UniqueField string `json:"unique_field"`
}

// RateLimit is per-county politeness config.
type RateLimit struct {
	DelayBetweenRequestsMs int `json:"delay_between_requests_ms"`
	MaxConcurrent          int `json:"max_concurrent"`
	// UpsertBatchMaxRows caps multi-row INSERT batch size for this county (overrides global default when > 0).
	UpsertBatchMaxRows int `json:"upsert_batch_max_rows,omitempty"`
}
