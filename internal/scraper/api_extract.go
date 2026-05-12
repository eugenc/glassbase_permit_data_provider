package scraper

import (
	"fmt"

	"github.com/echayko/glassbase_permit_data_provider/internal/generator"
	"github.com/tidwall/gjson"
)

// ExtractFromAPI parses a JSON response using JSON paths from the connector config.
func ExtractFromAPI(jsonBody string, config *generator.ConnectorConfig) ([]PermitRecord, error) {
	path := config.Extraction.RecordsPath
	if path == "" {
		return nil, fmt.Errorf("records_path is required for api extraction")
	}

	result := gjson.Get(jsonBody, path)
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
