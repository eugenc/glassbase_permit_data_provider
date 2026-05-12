package scraper

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/echayko/glassbase_permit_data_provider/internal/generator"
	"github.com/tidwall/gjson"
)

// ExtractFromAPI parses a JSON response using JSON paths from the connector config.
func ExtractFromAPI(jsonBody string, config *generator.ConnectorConfig) ([]PermitRecord, error) {
	path := config.Extraction.RecordsPath
	if path == "" {
		return nil, fmt.Errorf("records_path is required for api extraction")
	}

	raw := strings.TrimSpace(jsonBody)
	if raw == "" {
		return nil, fmt.Errorf("empty API response body")
	}
	if !json.Valid([]byte(raw)) {
		return nil, fmt.Errorf("API response is not valid JSON (truncated or non-JSON body)")
	}

	// Tyler EnerGov wraps errors in HTTP 200 with Success=false and Result=null (e.g. Elasticsearch
	// max_result_window exceeded when page * pageSize > 10000). Surface that instead of "path not found".
	if succ := gjson.Get(raw, "Success"); succ.Exists() && succ.Type == gjson.False {
		msg := strings.TrimSpace(gjson.Get(raw, "ErrorMessage").String())
		if msg == "" {
			msg = "Success=false (no ErrorMessage)"
		}
		return nil, fmt.Errorf("api reported failure: %s", msg)
	}

	// tidwall/gjson has no JSONPath-style "$" root. Top-level arrays (common for Socrata
	// /resource/....json feeds) must use the @this modifier.
	if path == "$" {
		path = "@this"
	}

	result := gjson.Get(raw, path)
	if !result.Exists() {
		return nil, fmt.Errorf("records_path %q not found in response", path)
	}
	if !result.IsArray() {
		return nil, fmt.Errorf("records_path %q is not an array", path)
	}

	var records []PermitRecord

	result.ForEach(func(_, item gjson.Result) bool {
		record := make(PermitRecord)
		for _, field := range config.Extraction.Fields {
			jp := field.JSONPath
			if jp == "" {
				jp = field.Name
			}
			val := item.Get(jp)
			record[field.Name] = val.String()
		}
		if record[config.Dedup.UniqueField] != "" {
			records = append(records, record)
		}
		return true
	})

	return records, nil
}
