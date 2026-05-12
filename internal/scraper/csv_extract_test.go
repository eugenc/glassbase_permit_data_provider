package scraper

import (
	"strings"
	"testing"

	"github.com/echayko/glassbase_permit_data_provider/internal/generator"
)

func TestExtractFromCSV(t *testing.T) {
	csv := strings.Join([]string{
		"PermitNumber,ApplicationNumber,ScopeofWork",
		`P1,A1,"Roof"`,
		`,A2,Skip empty key`,
		`P2,A3,Wall`,
	}, "\n")

	cfg := &generator.ConnectorConfig{
		Pagination: generator.Pagination{Type: "none"},
		Dedup:      generator.Dedup{UniqueField: "permit_number"},
		Extraction: generator.Extraction{
			Fields: []generator.FieldMapping{
				{Name: "permit_number", CsvColumn: "PermitNumber", Type: "string"},
				{Name: "application_number", CsvColumn: "ApplicationNumber", Type: "string"},
				{Name: "scope_of_work", CsvColumn: "ScopeofWork", Type: "string"},
			},
		},
	}

	recs, err := ExtractFromCSV("\ufeff"+csv, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(recs) != 2 {
		t.Fatalf("records: want 2 got %d", len(recs))
	}
	if recs[0]["permit_number"] != "P1" || recs[0]["scope_of_work"] != "Roof" {
		t.Fatalf("row0: %#v", recs[0])
	}
	if recs[1]["permit_number"] != "P2" {
		t.Fatalf("row1: %#v", recs[1])
	}
}
