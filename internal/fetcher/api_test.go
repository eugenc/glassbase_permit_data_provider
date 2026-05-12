package fetcher

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestAPIFetcher_usesPaginatorURLWhenEndpointAlsoSet(t *testing.T) {
	var gotRaw string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotRaw = r.URL.RawQuery
		_ = json.NewEncoder(w).Encode([]map[string]string{{"ok": "1"}})
	}))
	t.Cleanup(ts.Close)

	f := &APIFetcher{
		Endpoint: ts.URL + "/resource/demo.json",
	}
	q := url.Values{}
	q.Set("$offset", "99000")
	pageURL := ts.URL + "/resource/demo.json?" + q.Encode()

	_, err := f.Fetch(context.Background(), pageURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	if gotRaw == "" {
		t.Fatal("expected paginated query string on server, got empty")
	}
	decoded, err := url.QueryUnescape(gotRaw)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(decoded, "offset") || !strings.Contains(decoded, "99000") {
		t.Fatalf("server RawQuery=%q (decoded=%q) should carry $offset", gotRaw, decoded)
	}
}
