package catalog

import (
	"fmt"
	"io/fs"
	"path"
	"regexp"
	"sort"
	"strings"
)

const (
	UpstreamRepoURL = "https://github.com/mgomes/vibescript"
	UpstreamVersion = "v0.21.1"
)

var featuredExamples = map[string]int{
	"control_flow/case_when.vibe": 0,
	"strings/operations.vibe":     1,
	"stdlib/core_utilities.vibe":  2,
	"enums/operations.vibe":       3,
}

var runEntryPointPattern = regexp.MustCompile(`(?m)^def run\b`)

type Example struct {
	Slug        string
	Title       string
	Summary     string
	Description string
	Category    string
	Difficulty  string
	Stage       string
	Featured    bool
	Runnable    bool
	Tags        []string
	Source      string
	SourcePath  string
	SourceURL   string
	RunFunction string
}

type Store struct {
	examples      []Example
	featured      []Example
	bySlug        map[string]Example
	runnableCount int
}

func Load() (*Store, error) {
	examples := make([]Example, 0, 34)

	err := fs.WalkDir(content, "content/upstream", func(filePath string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || path.Ext(filePath) != ".vibe" {
			return nil
		}

		source, err := fs.ReadFile(content, filePath)
		if err != nil {
			return fmt.Errorf("read %s: %w", filePath, err)
		}

		relativePath := strings.TrimPrefix(filePath, "content/upstream/")
		categoryKey := path.Dir(relativePath)
		titleKey := strings.TrimSuffix(path.Base(relativePath), ".vibe")
		runnable := runEntryPointPattern.Match(source)

		stage := "Imported"
		summary := fmt.Sprintf(
			"Imported from the upstream Vibescript examples at %s and ready for browser discovery.",
			relativePath,
		)
		description := fmt.Sprintf(
			"This example is synced from the upstream Vibescript repository and serves as part of the site's growing examples corpus.",
		)
		runFunction := ""
		if runnable {
			stage = "Runnable"
			summary = fmt.Sprintf(
				"Imported from the upstream Vibescript examples at %s and runnable in the browser today.",
				relativePath,
			)
			description = "This example defines a top-level run function, so the site can compile and execute it directly through the browser runner."
			runFunction = "run"
		}

		tags := []string{"upstream", slugPart(categoryKey)}
		if runnable {
			tags = append(tags, "browser-runner")
		}

		examples = append(examples, Example{
			Slug:        slugPart(strings.TrimSuffix(relativePath, ".vibe")),
			Title:       titleize(titleKey),
			Summary:     summary,
			Description: description,
			Category:    titleize(categoryKey),
			Difficulty:  "Reference",
			Stage:       stage,
			Featured:    isFeatured(relativePath),
			Runnable:    runnable,
			Tags:        tags,
			Source:      string(source),
			SourcePath:  relativePath,
			SourceURL:   fmt.Sprintf("%s/blob/%s/examples/%s", UpstreamRepoURL, UpstreamVersion, relativePath),
			RunFunction: runFunction,
		})

		return nil
	})
	if err != nil {
		return nil, err
	}

	titleCounts := make(map[string]int, len(examples))
	for _, example := range examples {
		titleCounts[example.Title]++
	}
	for i := range examples {
		if titleCounts[examples[i].Title] > 1 {
			examples[i].Title = examples[i].Category + " " + examples[i].Title
		}
	}

	sort.Slice(examples, func(i, j int) bool {
		left := examples[i]
		right := examples[j]
		if left.Category != right.Category {
			return left.Category < right.Category
		}
		return left.Title < right.Title
	})

	store := &Store{
		examples: make([]Example, len(examples)),
		bySlug:   make(map[string]Example, len(examples)),
	}
	copy(store.examples, examples)

	for _, example := range store.examples {
		store.bySlug[example.Slug] = example
		if example.Runnable {
			store.runnableCount++
		}
		if example.Featured {
			store.featured = append(store.featured, example)
		}
	}

	sort.Slice(store.featured, func(i, j int) bool {
		return featuredExamples[store.featured[i].SourcePath] < featuredExamples[store.featured[j].SourcePath]
	})

	return store, nil
}

func (s *Store) All() []Example {
	ordered := make([]Example, len(s.examples))
	copy(ordered, s.examples)
	return ordered
}

func (s *Store) Featured(limit int) []Example {
	if limit > len(s.featured) {
		limit = len(s.featured)
	}
	featured := make([]Example, limit)
	copy(featured, s.featured[:limit])
	return featured
}

func (s *Store) BySlug(slug string) (Example, bool) {
	example, ok := s.bySlug[slug]
	return example, ok
}

func (s *Store) Count() int {
	return len(s.examples)
}

func (s *Store) RunnableCount() int {
	return s.runnableCount
}

func isFeatured(relativePath string) bool {
	_, ok := featuredExamples[relativePath]
	return ok
}

func slugPart(value string) string {
	replacer := strings.NewReplacer("/", "-", "_", "-", ".", "-")
	return replacer.Replace(strings.ToLower(value))
}

func titleize(value string) string {
	value = strings.ReplaceAll(value, "/", " ")
	value = strings.ReplaceAll(value, "_", " ")
	parts := strings.Fields(value)
	for i, part := range parts {
		parts[i] = strings.ToUpper(part[:1]) + part[1:]
	}
	return strings.Join(parts, " ")
}
