package runner

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/mgomes/vibescript.mauriciogomes.com/internal/catalog"
	"github.com/mgomes/vibescript/vibes"
)

var (
	ErrExampleNotFound    = errors.New("example not found")
	ErrExampleNotRunnable = errors.New("example is not runnable")
)

type Result struct {
	Kind       string `json:"kind"`
	Display    string `json:"display"`
	Value      any    `json:"value"`
	DurationUS int64  `json:"duration_us"`
}

type Service struct {
	store    *catalog.Store
	compiled map[string]*vibes.Script
}

func New(store *catalog.Store) (*Service, error) {
	engine, err := vibes.NewEngine(vibes.Config{
		StepQuota:        20_000,
		MemoryQuotaBytes: 256 << 10,
		RecursionLimit:   32,
		StrictEffects:    true,
	})
	if err != nil {
		return nil, fmt.Errorf("new vibes engine: %w", err)
	}

	compiled := make(map[string]*vibes.Script, store.Count())
	for _, example := range store.All() {
		if !example.Runnable {
			continue
		}

		script, err := engine.Compile(example.Source)
		if err != nil {
			return nil, fmt.Errorf("compile %s: %w", example.SourcePath, err)
		}
		compiled[example.Slug] = script
	}

	return &Service{
		store:    store,
		compiled: compiled,
	}, nil
}

func (s *Service) Run(ctx context.Context, slug string) (Result, error) {
	example, ok := s.store.BySlug(slug)
	if !ok {
		return Result{}, ErrExampleNotFound
	}
	if !example.Runnable {
		return Result{}, ErrExampleNotRunnable
	}

	script, ok := s.compiled[slug]
	if !ok {
		return Result{}, fmt.Errorf("compiled script missing for %s", slug)
	}

	started := time.Now()
	value, err := script.Call(ctx, example.RunFunction, nil, vibes.CallOptions{})
	if err != nil {
		return Result{}, err
	}

	return Result{
		Kind:       value.Kind().String(),
		Display:    value.String(),
		Value:      exportValue(value),
		DurationUS: time.Since(started).Microseconds(),
	}, nil
}

func exportValue(value vibes.Value) any {
	switch value.Kind() {
	case vibes.KindNil:
		return nil
	case vibes.KindBool:
		return value.Bool()
	case vibes.KindInt:
		return value.Int()
	case vibes.KindFloat:
		return value.Float()
	case vibes.KindString, vibes.KindSymbol, vibes.KindMoney, vibes.KindDuration, vibes.KindTime, vibes.KindRange, vibes.KindEnum, vibes.KindEnumValue, vibes.KindClass, vibes.KindInstance:
		return value.String()
	case vibes.KindArray:
		items := value.Array()
		exported := make([]any, len(items))
		for i, item := range items {
			exported[i] = exportValue(item)
		}
		return exported
	case vibes.KindHash, vibes.KindObject:
		hash := value.Hash()
		keys := make([]string, 0, len(hash))
		for key := range hash {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		exported := make(map[string]any, len(hash))
		for _, key := range keys {
			exported[key] = exportValue(hash[key])
		}
		return exported
	default:
		return value.String()
	}
}
