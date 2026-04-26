package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthCheckReturnsUnavailableWhenDBIsNil(t *testing.T) {
	server := &Server{}
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	server.healthCheck(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %d", http.StatusServiceUnavailable, rec.Code)
	}
	if body := rec.Body.String(); body != `{"status":"unhealthy","database":"unavailable"}` {
		t.Fatalf("unexpected body: %s", body)
	}
}
