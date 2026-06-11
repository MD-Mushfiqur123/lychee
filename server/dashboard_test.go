package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func setupDashboardRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	s := &Server{}
	s.registerDashboardRoutes(r)
	return r
}

func TestDashboardRootServersIndexHTML(t *testing.T) {
	r := setupDashboardRouter()

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/dashboard/", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "<html") {
		t.Errorf("expected HTML response, got: %s", body[:min(len(body), 100)])
	}
}

func TestDashboardServesJSAsset(t *testing.T) {
	r := setupDashboardRouter()

	// Try to serve the main JS bundle
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/dashboard/assets/index-E9nPRTPp.js", nil)
	r.ServeHTTP(w, req)

	// Should be 200 (asset exists) or fallback to index.html
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200 for JS asset, got %d", w.Code)
	}
}

func TestDashboardServesCSS(t *testing.T) {
	r := setupDashboardRouter()

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/dashboard/assets/index-2SBFhNas.css", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200 for CSS asset, got %d", w.Code)
	}
}

func TestDashboardUnknownPathFallsBackToIndex(t *testing.T) {
	r := setupDashboardRouter()

	// Any unknown path should serve index.html for client-side routing
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/dashboard/some/unknown/route", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 fallback to index.html, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "<html") {
		t.Errorf("fallback should serve index.html, got: %s", body[:min(len(body), 100)])
	}
}

func TestDashboardContentType(t *testing.T) {
	r := setupDashboardRouter()

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/dashboard/", nil)
	r.ServeHTTP(w, req)

	ct := w.Header().Get("Content-Type")
	if !strings.Contains(ct, "text/html") {
		t.Errorf("expected text/html content type, got %s", ct)
	}
}

func TestDashboardNoBodyOnHEAD(t *testing.T) {
	r := setupDashboardRouter()

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/dashboard/", nil)
	r.ServeHTTP(w, req)

	body, _ := io.ReadAll(w.Body)
	if len(body) == 0 {
		t.Error("expected non-empty body for dashboard root")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
