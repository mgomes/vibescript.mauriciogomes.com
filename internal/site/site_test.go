package site

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mgomes/vibescript.mauriciogomes.com/internal/catalog"
)

func TestHomePageRendersFeaturedExamples(t *testing.T) {
	app := newTestApp(t)

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	recorder := httptest.NewRecorder()

	app.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}

	body := recorder.Body.String()
	if !strings.Contains(body, "The home for practical Vibescript examples") {
		t.Fatalf("expected home headline, got %q", body)
	}

	if !strings.Contains(body, "Hello, Vibescript") {
		t.Fatalf("expected featured example title, got %q", body)
	}
}

func TestExamplesPageRendersCatalog(t *testing.T) {
	app := newTestApp(t)

	request := httptest.NewRequest(http.MethodGet, "/examples", nil)
	recorder := httptest.NewRecorder()

	app.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}

	body := recorder.Body.String()
	if !strings.Contains(body, "Build your Vibescript mental model") {
		t.Fatalf("expected catalog intro, got %q", body)
	}

	if !strings.Contains(body, "Prompt To Structured Output") {
		t.Fatalf("expected example listing, got %q", body)
	}
}

func TestExamplePageRendersDetail(t *testing.T) {
	app := newTestApp(t)

	request := httptest.NewRequest(http.MethodGet, "/examples/agentic-file-loop", nil)
	recorder := httptest.NewRecorder()

	app.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}

	body := recorder.Body.String()
	if !strings.Contains(body, "Agentic File Loop") {
		t.Fatalf("expected detail title, got %q", body)
	}

	if !strings.Contains(body, "Runnable sandbox landing soon") {
		t.Fatalf("expected runnable placeholder copy, got %q", body)
	}
}

func TestMissingExampleReturnsNotFound(t *testing.T) {
	app := newTestApp(t)

	request := httptest.NewRequest(http.MethodGet, "/examples/does-not-exist", nil)
	recorder := httptest.NewRecorder()

	app.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", recorder.Code)
	}
}

func TestHealthz(t *testing.T) {
	app := newTestApp(t)

	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	recorder := httptest.NewRecorder()

	app.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}

	if strings.TrimSpace(recorder.Body.String()) != `{"status":"ok"}` {
		t.Fatalf("expected health payload, got %q", recorder.Body.String())
	}
}

func newTestApp(t *testing.T) http.Handler {
	t.Helper()

	app, err := New(catalog.NewStore(catalog.SeedExamples()))
	if err != nil {
		t.Fatalf("new app: %v", err)
	}

	return app
}
