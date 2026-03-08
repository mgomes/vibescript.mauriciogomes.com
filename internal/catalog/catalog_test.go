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

func TestAllImportedExamplesCompile(t *testing.T) {
	store, err := Load()
	if err != nil {
		t.Fatalf("load catalog: %v", err)
	}

	engine := vibes.MustNewEngine(vibes.Config{})
	for _, example := range store.All() {
		if _, err := engine.Compile(example.Source); err != nil {
			t.Fatalf("compile %s: %v", example.SourcePath, err)
		}
	}
}
