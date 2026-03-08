package site

import (
	"bytes"
	"encoding/json"
	"html/template"
	"io/fs"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/mgomes/vibescript.mauriciogomes.com/internal/catalog"
)

type App struct {
	store     *catalog.Store
	templates *template.Template
	static    http.Handler
}

type page struct {
	Title       string
	Description string
	Section     string
}

type viewData struct {
	ContentTemplate string
	Content         template.HTML
	Page            page
	TotalExamples   int
	Featured        []catalog.Example
	Examples        []catalog.Example
	Example         catalog.Example
	Year            int
}

func New(store *catalog.Store) (http.Handler, error) {
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
			Description: "A site for learning Vibescript through examples that will become runnable in the browser.",
			Section:     "home",
		},
		Featured:      a.store.Featured(3),
		Examples:      a.store.All(),
		TotalExamples: a.store.Count(),
		Year:          time.Now().Year(),
	})
}

func (a *App) examplesIndex(w http.ResponseWriter, r *http.Request) {
	a.render(w, http.StatusOK, viewData{
		ContentTemplate: "examples",
		Page: page{
			Title:       "Examples",
			Description: "A growing catalog of Vibescript examples, designed to scale into a large runnable library.",
			Section:     "examples",
		},
		Examples:      a.store.All(),
		TotalExamples: a.store.Count(),
		Year:          time.Now().Year(),
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
		Example:       example,
		TotalExamples: a.store.Count(),
		Year:          time.Now().Year(),
	})
}

func (a *App) healthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (a *App) notFound(w http.ResponseWriter, r *http.Request) {
	a.render(w, http.StatusNotFound, viewData{
		ContentTemplate: "not-found",
		Page: page{
			Title:       "Not Found",
			Description: "The requested page does not exist.",
			Section:     "",
		},
		TotalExamples: a.store.Count(),
		Year:          time.Now().Year(),
	})
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
