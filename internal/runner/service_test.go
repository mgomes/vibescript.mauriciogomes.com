package runner

import (
	"context"
	"errors"
	"testing"

	"github.com/mgomes/vibescript.mauriciogomes.com/internal/catalog"
)

func TestRunRunnableExample(t *testing.T) {
	store, err := catalog.Load()
	if err != nil {
		t.Fatalf("load catalog: %v", err)
	}

	service, err := New(store)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	result, err := service.Run(context.Background(), "control-flow-case-when")
	if err != nil {
		t.Fatalf("run example: %v", err)
	}

	if result.Kind != "array" {
		t.Fatalf("expected array result, got %q", result.Kind)
	}

	values, ok := result.Value.([]any)
	if !ok {
		t.Fatalf("expected array export, got %T", result.Value)
	}

	if len(values) != 3 {
		t.Fatalf("expected 3 values, got %d", len(values))
	}

	if values[0] != "perfect" || values[1] != "great" || values[2] != "ok" {
		t.Fatalf("unexpected output: %#v", values)
	}
}

func TestRunNonRunnableExample(t *testing.T) {
	store, err := catalog.Load()
	if err != nil {
		t.Fatalf("load catalog: %v", err)
	}

	service, err := New(store)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	_, err = service.Run(context.Background(), "basics-functions-and-calls")
	if !errors.Is(err, ErrExampleNotRunnable) {
		t.Fatalf("expected ErrExampleNotRunnable, got %v", err)
	}
}

func TestRunAllRunnableExamples(t *testing.T) {
	store, err := catalog.Load()
	if err != nil {
		t.Fatalf("load catalog: %v", err)
	}

	service, err := New(store)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	for _, example := range store.All() {
		if !example.Runnable {
			continue
		}

		example := example
		t.Run(example.Slug, func(t *testing.T) {
			if _, err := service.Run(context.Background(), example.Slug); err != nil {
				t.Fatalf("run %s: %v", example.Slug, err)
			}
		})
	}
}
