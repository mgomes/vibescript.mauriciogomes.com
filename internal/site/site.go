package site

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"strings"
	"time"

	"github.com/mgomes/ohm"
	"github.com/mgomes/vibescript-lang.org/internal/catalog"
	"github.com/mgomes/vibescript-lang.org/internal/runner"
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
	CacheBust        string
	SiteBaseURL      string
	CanonicalURL     string
}

const siteBaseURL = "https://vibescript-lang.org"

var cacheBust = fmt.Sprintf("%d", time.Now().UnixMilli())

func New(store *catalog.Store, runService *runner.Service) (*ohm.App, error) {
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

	application := ohm.New()
	application.Use(ohm.Recoverer(nil), realIP, requestTimeout(30*time.Second), gzipResponse, redirectLegacyHosts)
	application.Get("/", app.home)
	application.Get("/healthz", app.healthz)
	application.GetHTTP("/static/*", http.StripPrefix("/static/", app.static))
	application.Post("/api/examples/{slug}/run", app.runExample)
	application.Get("/examples", app.examplesIndex)
	application.Get("/examples/", app.examplesIndex)
	application.Get("/examples/{slug}", app.exampleDetail)
	application.NotFound(app.notFound)

	return application, nil
}

func redirectLegacyHosts(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Host {
		case "vibescript.mauriciogomes.com", "www.vibescript-lang.org":
			http.Redirect(w, r, "https://vibescript-lang.org"+r.URL.RequestURI(), http.StatusMovedPermanently)
		default:
			next.ServeHTTP(w, r)
		}
	})
}

func realIP(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
			host, _, _ := strings.Cut(forwarded, ",")
			r.RemoteAddr = strings.TrimSpace(host)
		}
		if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
			r.RemoteAddr = realIP
		}
		next.ServeHTTP(w, r)
	})
}

func requestTimeout(timeout time.Duration) ohm.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func (a *App) home(req *ohm.Request) error {
	return a.render(req, http.StatusOK, viewData{
		ContentTemplate: "home",
		Page: page{
			Title:       "Vibescript",
			Description: "An embeddable Ruby-like language for Go — safe by default and easy for AI to write. Explore examples and run them in the browser.",
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

func (a *App) examplesIndex(req *ohm.Request) error {
	return a.render(req, http.StatusOK, viewData{
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

func (a *App) exampleDetail(req *ohm.Request) error {
	identifier := req.Param("slug")
	example, ok := a.store.BySlug(identifier)
	if !ok {
		return a.renderNotFound(req.ResponseWriter(), req.HTTPRequest())
	}

	return a.render(req, http.StatusOK, viewData{
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

func (a *App) healthz(req *ohm.Request) error {
	req.JSON(http.StatusOK, map[string]any{
		"status":            "ok",
		"examples":          a.store.Count(),
		"runnable_examples": a.store.RunnableCount(),
	})
	return nil
}

func (a *App) runExample(req *ohm.Request) error {
	result, err := a.runner.Run(req.Context(), req.Param("slug"))
	if err != nil {
		status := http.StatusInternalServerError
		switch {
		case errors.Is(err, runner.ErrExampleNotFound):
			status = http.StatusNotFound
		case errors.Is(err, runner.ErrExampleNotRunnable):
			status = http.StatusConflict
		}
		req.JSON(status, map[string]string{"error": err.Error()})
		return nil
	}

	req.JSON(http.StatusOK, map[string]any{"result": result})
	return nil
}

func (a *App) notFound(w http.ResponseWriter, r *http.Request) {
	if err := a.renderNotFound(w, r); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func (a *App) renderNotFound(w http.ResponseWriter, r *http.Request) error {
	return a.renderHTTP(w, r, http.StatusNotFound, viewData{
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

func (a *App) render(req *ohm.Request, status int, data viewData) error {
	return a.renderHTTP(req.ResponseWriter(), req.HTTPRequest(), status, data)
}

func (a *App) renderHTTP(w http.ResponseWriter, r *http.Request, status int, data viewData) error {
	data.CacheBust = cacheBust
	data.SiteBaseURL = siteBaseURL
	data.CanonicalURL = siteBaseURL + r.URL.Path
	return ohm.RenderHTML(w, r, status, ohm.HTMLFunc(func(ctx context.Context, w io.Writer) error {
		if err := ctx.Err(); err != nil {
			return err
		}

		var body bytes.Buffer
		if err := a.templates.ExecuteTemplate(&body, data.ContentTemplate, data); err != nil {
			return fmt.Errorf("execute content template %q: %w", data.ContentTemplate, err)
		}

		data.Content = template.HTML(body.String())
		if err := a.templates.ExecuteTemplate(w, "layout", data); err != nil {
			return fmt.Errorf("execute layout template: %w", err)
		}
		return nil
	}))
}
