package site

import (
	"bytes"
	"encoding/json"
	"errors"
	"html/template"
	"io/fs"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/mgomes/vibescript.mauriciogomes.com/internal/catalog"
	"github.com/mgomes/vibescript.mauriciogomes.com/internal/runner"
)

type App struct {
	store     *catalog.Store
	runner    *runner.Service
	templates *template.Template
	static    http.Handler
}

type page struct {
	Title       string
	Description string
	Section     string
}

type viewData struct {
	ContentTemplate  string
	Content          template.HTML
	Page             page
	ShowcaseExamples int
	TotalExamples    int
	RunnableExamples int
	Featured         []catalog.Example
	Examples         []catalog.Example
	Example          catalog.Example
	UpstreamVersion  string
	UpstreamRepoURL  string
	Year             int
}

func New(store *catalog.Store, runService *runner.Service) (http.Handler, error) {
	templates, err := template.ParseFS(assets, "templates/*.html")
	if err != nil {
		return nil, err
	}

	staticFS, err := fs.Sub(assets, "static")
	if err != nil {
		return nil, err
	}

	app := &App{
		store:     store,
		runner:    runService,
		templates: templates,
		static:    http.FileServer(http.FS(staticFS)),
	}

	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Compress(5))
	router.Use(middleware.Timeout(30 * time.Second))

	router.Get("/", app.home)
	router.Get("/healthz", app.healthz)
	router.Handle("/static/*", http.StripPrefix("/static/", app.static))
	router.Post("/api/examples/{slug}/run", app.runExample)

	router.Route("/examples", func(r chi.Router) {
		r.Get("/", app.examplesIndex)
		r.Get("/{slug}", app.exampleDetail)
	})

	router.NotFound(app.notFound)

	return router, nil
}

func (a *App) home(w http.ResponseWriter, r *http.Request) {
	a.render(w, http.StatusOK, viewData{
		ContentTemplate: "home",
		Page: page{
			Title:       "Vibescript",
			Description: "An embeddable Ruby-like language for building safe, AI-friendly applications in Go. Explore examples and run them in the browser.",
			Section:     "home",
		},
		ShowcaseExamples: a.store.TaggedCount("showcase"),
		Featured:         a.store.Featured(4),
		Examples:         a.store.All(),
		TotalExamples:    a.store.Count(),
		RunnableExamples: a.store.RunnableCount(),
		UpstreamVersion:  catalog.UpstreamVersion,
		UpstreamRepoURL:  catalog.UpstreamRepoURL,
		Year:             time.Now().Year(),
	})
}

func (a *App) examplesIndex(w http.ResponseWriter, r *http.Request) {
	a.render(w, http.StatusOK, viewData{
		ContentTemplate: "examples",
		Page: page{
			Title:       "Examples",
			Description: "Browse Vibescript examples — from idiomatic showcases to upstream references — all compiled against the real interpreter.",
			Section:     "examples",
		},
		ShowcaseExamples: a.store.TaggedCount("showcase"),
		Examples:         a.store.All(),
		TotalExamples:    a.store.Count(),
		RunnableExamples: a.store.RunnableCount(),
		UpstreamVersion:  catalog.UpstreamVersion,
		UpstreamRepoURL:  catalog.UpstreamRepoURL,
		Year:             time.Now().Year(),
	})
}

func (a *App) exampleDetail(w http.ResponseWriter, r *http.Request) {
	identifier := chi.URLParam(r, "slug")
	example, ok := a.store.BySlug(identifier)
	if !ok {
		a.notFound(w, r)
		return
	}

	a.render(w, http.StatusOK, viewData{
		ContentTemplate: "example",
		Page: page{
			Title:       example.Title,
			Description: example.Summary,
			Section:     "examples",
		},
		ShowcaseExamples: a.store.TaggedCount("showcase"),
		Example:          example,
		TotalExamples:    a.store.Count(),
		RunnableExamples: a.store.RunnableCount(),
		UpstreamVersion:  catalog.UpstreamVersion,
		UpstreamRepoURL:  catalog.UpstreamRepoURL,
		Year:             time.Now().Year(),
	})
}

func (a *App) healthz(w http.ResponseWriter, r *http.Request) {
	a.writeJSON(w, http.StatusOK, map[string]any{
		"status":            "ok",
		"examples":          a.store.Count(),
		"runnable_examples": a.store.RunnableCount(),
	})
}

func (a *App) runExample(w http.ResponseWriter, r *http.Request) {
	result, err := a.runner.Run(r.Context(), chi.URLParam(r, "slug"))
	if err != nil {
		status := http.StatusInternalServerError
		switch {
		case errors.Is(err, runner.ErrExampleNotFound):
			status = http.StatusNotFound
		case errors.Is(err, runner.ErrExampleNotRunnable):
			status = http.StatusConflict
		}
		a.writeJSON(w, status, map[string]string{"error": err.Error()})
		return
	}

	a.writeJSON(w, http.StatusOK, map[string]any{"result": result})
}

func (a *App) notFound(w http.ResponseWriter, r *http.Request) {
	a.render(w, http.StatusNotFound, viewData{
		ContentTemplate: "not-found",
		Page: page{
			Title:       "Not Found",
			Description: "The requested page does not exist.",
			Section:     "",
		},
		ShowcaseExamples: a.store.TaggedCount("showcase"),
		TotalExamples:    a.store.Count(),
		RunnableExamples: a.store.RunnableCount(),
		UpstreamVersion:  catalog.UpstreamVersion,
		UpstreamRepoURL:  catalog.UpstreamRepoURL,
		Year:             time.Now().Year(),
	})
}

func (a *App) writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func (a *App) render(w http.ResponseWriter, status int, data viewData) {
	var body bytes.Buffer
	if err := a.templates.ExecuteTemplate(&body, data.ContentTemplate, data); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	data.Content = template.HTML(body.String())

	var page bytes.Buffer
	if err := a.templates.ExecuteTemplate(&page, "layout", data); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = page.WriteTo(w)
}
