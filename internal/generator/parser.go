package generator

import (
	"encoding/json"
	"fmt"
	"strings"
)

func parseConnectorConfig(raw string) (*ConnectorConfig, error) {
	cleaned := strings.TrimSpace(raw)
	cleaned = strings.TrimPrefix(cleaned, "```json")
	cleaned = strings.TrimPrefix(cleaned, "```")
	cleaned = strings.TrimSuffix(cleaned, "```")
	cleaned = strings.TrimSpace(cleaned)

	var config ConnectorConfig
	if err := json.Unmarshal([]byte(cleaned), &config); err != nil {
		return nil, fmt.Errorf("parse connector config: %w\nRaw response: %s", err, raw)
	}

	if len(config.Extraction.Fields) == 0 {
		return nil, fmt.Errorf("connector config has no fields — AI analysis failed")
	}
	if config.Dedup.UniqueField == "" {
		return nil, fmt.Errorf("connector config missing dedup.unique_field")
	}

	return &config, nil
}
