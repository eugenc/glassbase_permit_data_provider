package scraper

import (
	"testing"

	"github.com/echayko/glassbase_permit_data_provider/internal/generator"
)

func TestExtractFromAPI_rootJSONArray_recordsPathDollar(t *testing.T) {
	body := `[{"permit_no":"BP-1","addr":"Main"},{"permit_no":"BP-2","addr":"Oak"}]`

	cfg := &generator.ConnectorConfig{
		Extraction: generator.Extraction{
			RecordsPath: "$",
			Fields: []generator.FieldMapping{
				{Name: "permit_number", JSONPath: "permit_no", Type: "string"},
				{Name: "permit_address", JSONPath: "addr", Type: "string"},
			},
		},
		Dedup: generator.Dedup{UniqueField: "permit_number"},
	}

	recs, err := ExtractFromAPI(body, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(recs) != 2 || recs[0]["permit_number"] != "BP-1" {
		t.Fatalf("got %+v", recs)
	}
}
