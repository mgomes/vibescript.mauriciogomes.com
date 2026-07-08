package runner

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/mgomes/vibescript.mauriciogomes.com/internal/catalog"
	"github.com/mgomes/vibescript/vibes"
	"github.com/mgomes/vibescript/vibes/value"
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
		StepQuota:              250_000,
		MemoryQuotaBytes:       256 << 10,
		RecursionLimit:         32,
		StrictEffects:          true,
		DefaultTaskConcurrency: 4,
		MaxTaskConcurrency:     8,
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
	result, err := script.Call(ctx, example.RunFunction, nil, vibes.CallOptions{})
	if err != nil {
		return Result{}, err
	}

	return Result{
		Kind:       result.Kind().String(),
		Display:    result.String(),
		Value:      exportValue(result),
		DurationUS: time.Since(started).Microseconds(),
	}, nil
}

func exportValue(v value.Value) any {
	switch v.Kind() {
	case value.KindNil:
		return nil
	case value.KindBool:
		return v.Bool()
	case value.KindInt:
		return v.Int()
	case value.KindFloat:
		return v.Float()
	case value.KindString, value.KindSymbol, value.KindMoney, value.KindDuration, value.KindTime, value.KindRange, value.KindEnum, value.KindEnumValue, value.KindClass, value.KindInstance:
		return v.String()
	case value.KindArray:
		items := v.Array()
		exported := make([]any, len(items))
		for i, item := range items {
			exported[i] = exportValue(item)
		}
		return exported
	case value.KindHash, value.KindObject:
		hash := v.Hash()
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
		return v.String()
	}
}
