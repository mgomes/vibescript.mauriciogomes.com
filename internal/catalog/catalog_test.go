package catalog

import (
	"testing"

	"github.com/mgomes/vibescript/vibes"
)

func TestLoad(t *testing.T) {
	store, err := Load()
	if err != nil {
		t.Fatalf("load catalog: %v", err)
	}

	if store.Count() < 49 {
		t.Fatalf("expected at least 49 examples, got %d", store.Count())
	}

	if store.RunnableCount() < 22 {
		t.Fatalf("expected at least 22 runnable examples, got %d", store.RunnableCount())
	}

	example, ok := store.BySlug("strings-operations")
	if !ok {
		t.Fatalf("expected strings-operations to be present")
	}

	if !example.Runnable {
		t.Fatalf("expected strings-operations to be runnable")
	}

	if example.RunFunction != "run" {
		t.Fatalf("expected run entrypoint, got %q", example.RunFunction)
	}

	rosettaExample, ok := store.BySlug("rosettacode-popular-fizzbuzz")
	if !ok {
		t.Fatalf("expected rosettacode-popular-fizzbuzz to be present")
	}

	if !rosettaExample.Runnable {
		t.Fatalf("expected rosetta example to be runnable")
	}

	showcaseExample, ok := store.BySlug("showcase-workflows-release-readiness")
	if !ok {
		t.Fatalf("expected showcase-workflows-release-readiness to be present")
	}

	if !showcaseExample.Runnable {
		t.Fatalf("expected showcase example to be runnable")
	}
}

func TestAllExamplesCompileAndPassStaticChecks(t *testing.T) {
	store, err := Load()
	if err != nil {
		t.Fatalf("load catalog: %v", err)
	}

	engine := vibes.MustNewEngine(vibes.Config{})
	for _, example := range store.All() {
		t.Run(example.SourcePath, func(t *testing.T) {
			script, err := engine.Compile(example.Source)
			if err != nil {
				t.Fatalf("Compile(%q) error = %v, want nil", example.SourcePath, err)
			}

			if !example.Runnable {
				return
			}

			if warnings := script.CheckWarningsForFunction(example.RunFunction); len(warnings) > 0 {
				t.Errorf(
					"CheckWarningsForFunction(%q, %q) = %#v, want none",
					example.SourcePath,
					example.RunFunction,
					warnings,
				)
			}
		})
	}
}

// The site promises every catalog example runs in the browser, and the runner
// silently skips any example without a top-level run entry point. Assert the
// invariant so an import can never quietly reintroduce a dead example.
func TestEveryExampleIsRunnable(t *testing.T) {
	store, err := Load()
	if err != nil {
		t.Fatalf("load catalog: %v", err)
	}

	for _, example := range store.All() {
		if !example.Runnable {
			t.Errorf("%s (%s) defines no top-level `def run`", example.Slug, example.SourcePath)
		}
	}

	if store.RunnableCount() != store.Count() {
		t.Fatalf("runnable count = %d, want all %d examples", store.RunnableCount(), store.Count())
	}
}
