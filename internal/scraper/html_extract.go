package scraper

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/echayko/glassbase_permit_data_provider/internal/generator"
)

type PermitRecord map[string]string

// ExtractFromHTML parses HTML using the connector config's CSS selectors.
func ExtractFromHTML(html string, config *generator.ConnectorConfig) ([]PermitRecord, error) {
	if config.Extraction.RecordSelector == "" {
		return nil, fmt.Errorf("record_selector is required for html/spa extraction")
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("parse html: %w", err)
	}

	var records []PermitRecord

	doc.Find(config.Extraction.RecordSelector).Each(func(i int, s *goquery.Selection) {
		record := make(PermitRecord)

		for _, field := range config.Extraction.Fields {
			var val string
			el := s.Find(field.Selector)
			if field.Attr != "" {
				val, _ = el.Attr(field.Attr)
			} else {
				val = strings.TrimSpace(el.Text())
			}
			record[field.Name] = val
		}

		if record[config.Dedup.UniqueField] != "" {
			records = append(records, record)
		}
	})

	return records, nil
}
