package catalog

import "testing"

func TestLoad(t *testing.T) {
	store, err := Load()
	if err != nil {
		t.Fatalf("load catalog: %v", err)
	}

	if store.Count() != 34 {
		t.Fatalf("expected 34 examples, got %d", store.Count())
	}

	if store.RunnableCount() != 7 {
		t.Fatalf("expected 7 runnable examples, got %d", store.RunnableCount())
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
}
