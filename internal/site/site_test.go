package site

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mgomes/vibescript.mauriciogomes.com/internal/catalog"
	"github.com/mgomes/vibescript.mauriciogomes.com/internal/runner"
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
	if !strings.Contains(body, "Imported Vibescript examples") {
		t.Fatalf("expected home headline, got %q", body)
	}

	if !strings.Contains(body, "Case When") {
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
	if !strings.Contains(body, "Browse the upstream Vibescript corpus") {
		t.Fatalf("expected catalog intro, got %q", body)
	}

	if !strings.Contains(body, "Strings Operations") {
		t.Fatalf("expected example listing, got %q", body)
	}
}

func TestExamplePageRendersDetail(t *testing.T) {
	app := newTestApp(t)

	request := httptest.NewRequest(http.MethodGet, "/examples/strings-operations", nil)
	recorder := httptest.NewRecorder()

	app.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}

	body := recorder.Body.String()
	if !strings.Contains(body, "Strings Operations") {
		t.Fatalf("expected detail title, got %q", body)
	}

	if !strings.Contains(body, "Run this example") {
		t.Fatalf("expected runner copy, got %q", body)
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

	if !strings.Contains(recorder.Body.String(), `"runnable_examples":7`) {
		t.Fatalf("expected health payload, got %q", recorder.Body.String())
	}
}

func TestRunExample(t *testing.T) {
	app := newTestApp(t)

	request := httptest.NewRequest(http.MethodPost, "/api/examples/control-flow-case-when/run", nil)
	recorder := httptest.NewRecorder()

	app.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}

	body := recorder.Body.String()
	if !strings.Contains(body, `"kind":"array"`) {
		t.Fatalf("expected array result, got %q", body)
	}

	if !strings.Contains(body, `"perfect"`) {
		t.Fatalf("expected runnable output, got %q", body)
	}
}

func TestRunNonRunnableExample(t *testing.T) {
	app := newTestApp(t)

	request := httptest.NewRequest(http.MethodPost, "/api/examples/basics-functions-and-calls/run", nil)
	recorder := httptest.NewRecorder()

	app.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusConflict {
		t.Fatalf("expected status 409, got %d", recorder.Code)
	}
}

func newTestApp(t *testing.T) http.Handler {
	t.Helper()

	store, err := catalog.Load()
	if err != nil {
		t.Fatalf("load catalog: %v", err)
	}

	runService, err := runner.New(store)
	if err != nil {
		t.Fatalf("new runner: %v", err)
	}

	app, err := New(store, runService)
	if err != nil {
		t.Fatalf("new app: %v", err)
	}

	return app
}
