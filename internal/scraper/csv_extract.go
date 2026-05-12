package scraper

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/echayko/glassbase_permit_data_provider/internal/generator"
)

// ExtractFromCSV parses a CSV response using csv_column (or name) headers from the connector config.
func ExtractFromCSV(csvBody string, config *generator.ConnectorConfig) ([]PermitRecord, error) {
	raw := strings.TrimSpace(csvBody)
	if raw == "" {
		return nil, fmt.Errorf("empty CSV body")
	}
	raw = strings.TrimPrefix(raw, "\ufeff")

	r := csv.NewReader(strings.NewReader(raw))
	r.ReuseRecord = true
	r.LazyQuotes = true

	headers, err := r.Read()
	if err != nil {
		return nil, fmt.Errorf("csv header: %w", err)
	}
	colIdx := make(map[string]int, len(headers))
	for i, h := range headers {
		h = strings.TrimSpace(h)
		colIdx[h] = i
	}

	if config.Dedup.UniqueField == "" {
		return nil, fmt.Errorf("dedup.unique_field is required for csv extraction")
	}

	uniqueCol, err := csvHeaderForFieldName(config, config.Dedup.UniqueField)
	if err != nil {
		return nil, err
	}
	if _, ok := colIdx[uniqueCol]; !ok {
		return nil, fmt.Errorf("csv missing dedup column %q (unique_field %q)", uniqueCol, config.Dedup.UniqueField)
	}

	var records []PermitRecord
	for {
		row, err := r.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("csv row: %w", err)
		}
		rec := make(PermitRecord)
		for _, field := range config.Extraction.Fields {
			hdr, ferr := csvHeaderForFieldName(config, field.Name)
			if ferr != nil {
				return nil, ferr
			}
			idx, ok := colIdx[hdr]
			if !ok {
				rec[field.Name] = ""
				continue
			}
			if idx >= len(row) {
				rec[field.Name] = ""
				continue
			}
			rec[field.Name] = strings.TrimSpace(row[idx])
		}
		if rec[config.Dedup.UniqueField] != "" {
			records = append(records, rec)
		}
	}

	return records, nil
}

func csvHeaderForFieldName(config *generator.ConnectorConfig, fieldName string) (string, error) {
	for _, f := range config.Extraction.Fields {
		if f.Name != fieldName {
			continue
		}
		if f.CsvColumn != "" {
			return f.CsvColumn, nil
		}
		if f.Name == "" {
			return "", fmt.Errorf("field has empty name")
		}
		// Fall back to using the logical field name as the CSV header.
		return f.Name, nil
	}
	return "", fmt.Errorf("no extraction field named %q for csv column lookup", fieldName)
}
