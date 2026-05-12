package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStaticHandlerRootReturnsHTML(t *testing.T) {
	h := StaticHandler()
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("GET / want status %d got %d Location=%q", http.StatusOK, rr.Code, rr.Header().Get("Location"))
	}
}
