package catalog

import (
	"fmt"
	"io/fs"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const (
	UpstreamRepoURL = "https://github.com/mgomes/vibescript"
	UpstreamVersion = "v0.26.2"
)

var featuredExamples = map[string]int{
	"control_flow/case_when.vibe": 100,
	"strings/operations.vibe":     101,
	"stdlib/core_utilities.vibe":  102,
	"enums/operations.vibe":       103,
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
	FeatureRank int
}

type Store struct {
	examples      []Example
	featured      []Example
	bySlug        map[string]Example
	runnableCount int
}

func Load() (*Store, error) {
	examples := make([]Example, 0, 64)

	err := fs.WalkDir(content, "content", func(filePath string, entry fs.DirEntry, err error) error {
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

		example, ok, err := loadExample(filePath, source)
		if err != nil {
			return err
		}
		if !ok {
			return nil
		}

		examples = append(examples, example)

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
		left := store.featured[i]
		right := store.featured[j]
		if left.FeatureRank != right.FeatureRank {
			return left.FeatureRank < right.FeatureRank
		}
		if left.Category != right.Category {
			return left.Category < right.Category
		}
		return left.Title < right.Title
	})

	return store, nil
}

func loadExample(filePath string, source []byte) (Example, bool, error) {
	relativeToContent := strings.TrimPrefix(filePath, "content/")
	parts := strings.Split(relativeToContent, "/")
	if len(parts) < 2 {
		return Example{}, false, nil
	}

	switch parts[0] {
	case "showcase":
		return loadShowcaseExample(strings.Join(parts[1:], "/"), source), true, nil
	case "upstream":
		return loadUpstreamExample(strings.Join(parts[1:], "/"), source), true, nil
	case "rosettacode":
		return loadRosettaCodeExample(strings.Join(parts[1:], "/"), source), true, nil
	default:
		return Example{}, false, nil
	}
}

func loadUpstreamExample(relativePath string, source []byte) Example {
	categoryKey := path.Dir(relativePath)
	titleKey := strings.TrimSuffix(path.Base(relativePath), ".vibe")
	runnable := runEntryPointPattern.Match(source)

	stage := "Imported"
	summary := fmt.Sprintf(
		"Imported from the upstream Vibescript examples at %s and ready for browser discovery.",
		relativePath,
	)
	description := "This example is synced from the upstream Vibescript repository and serves as part of the site's growing examples corpus."
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

	return Example{
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
		FeatureRank: featuredExamples[relativePath],
	}
}

func loadRosettaCodeExample(relativePath string, source []byte) Example {
	metadata := parseMetadata(string(source))
	titleKey := strings.TrimSuffix(path.Base(relativePath), ".vibe")
	title := metadata["title"]
	if title == "" {
		title = titleize(titleKey)
	}

	category := metadata["category"]
	if category == "" {
		category = "Rosetta Code"
	}

	difficulty := metadata["difficulty"]
	if difficulty == "" {
		difficulty = "Reference"
	}

	runnable := runEntryPointPattern.Match(source)
	stage := metadata["stage"]
	if stage == "" {
		if runnable {
			stage = "Runnable"
		} else {
			stage = "Draft"
		}
	}

	summary := metadata["summary"]
	if summary == "" {
		if runnable {
			summary = fmt.Sprintf("A Vibescript implementation of the Rosetta Code task %q that runs in the browser.", title)
		} else {
			summary = fmt.Sprintf("A Vibescript implementation draft for the Rosetta Code task %q.", title)
		}
	}

	description := metadata["description"]
	if description == "" {
		description = "This example is part of the Rosetta Code task import for the Vibescript site."
	}

	sourceURL := metadata["source"]
	if sourceURL == "" {
		sourceURL = "https://rosettacode.org/wiki/" + strings.ReplaceAll(title, " ", "_")
	}

	tags := []string{"rosetta-code"}
	if extra := splitMetadataList(metadata["tags"]); len(extra) > 0 {
		tags = append(tags, extra...)
	}
	if runnable {
		tags = append(tags, "browser-runner")
	}

	runFunction := ""
	if runnable {
		runFunction = "run"
	}

	featured := metadata["featured"] == "true"
	featureRank := 0
	if featured {
		featureRank = parseFeatureRank(metadata["feature_rank"], 500)
	}

	return Example{
		Slug:        "rosettacode-" + slugPart(strings.TrimSuffix(relativePath, ".vibe")),
		Title:       title,
		Summary:     summary,
		Description: description,
		Category:    category,
		Difficulty:  difficulty,
		Stage:       stage,
		Featured:    featured,
		Runnable:    runnable,
		Tags:        dedupe(tags),
		Source:      string(source),
		SourcePath:  "rosettacode/" + relativePath,
		SourceURL:   sourceURL,
		RunFunction: runFunction,
		FeatureRank: featureRank,
	}
}

func loadShowcaseExample(relativePath string, source []byte) Example {
	metadata := parseMetadata(string(source))
	titleKey := strings.TrimSuffix(path.Base(relativePath), ".vibe")
	title := metadata["title"]
	if title == "" {
		title = titleize(titleKey)
	}

	category := metadata["category"]
	if category == "" {
		category = "Vibescript Showcase"
	}

	difficulty := metadata["difficulty"]
	if difficulty == "" {
		difficulty = "Showcase"
	}

	runnable := runEntryPointPattern.Match(source)
	stage := metadata["stage"]
	if stage == "" {
		if runnable {
			stage = "Showcase"
		} else {
			stage = "Draft"
		}
	}

	summary := metadata["summary"]
	if summary == "" {
		summary = fmt.Sprintf("An idiomatic Vibescript showcase example for %q.", title)
	}

	description := metadata["description"]
	if description == "" {
		description = "This example is written to demonstrate idiomatic Vibescript with semantic types, typed signatures, and structured outputs."
	}

	tags := []string{"showcase", "idiomatic-vibescript"}
	if extra := splitMetadataList(metadata["tags"]); len(extra) > 0 {
		tags = append(tags, extra...)
	}
	if runnable {
		tags = append(tags, "browser-runner")
	}

	runFunction := ""
	if runnable {
		runFunction = "run"
	}

	featured := metadata["featured"] == "true"
	featureRank := 0
	if featured {
		featureRank = parseFeatureRank(metadata["feature_rank"], 0)
	}

	return Example{
		Slug:        "showcase-" + slugPart(strings.TrimSuffix(relativePath, ".vibe")),
		Title:       title,
		Summary:     summary,
		Description: description,
		Category:    category,
		Difficulty:  difficulty,
		Stage:       stage,
		Featured:    featured,
		Runnable:    runnable,
		Tags:        dedupe(tags),
		Source:      string(source),
		SourcePath:  "showcase/" + relativePath,
		SourceURL:   metadata["source"],
		RunFunction: runFunction,
		FeatureRank: featureRank,
	}
}

func parseMetadata(source string) map[string]string {
	metadata := map[string]string{}
	for _, line := range strings.Split(source, "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "# ") || !strings.Contains(trimmed, ":") {
			if trimmed == "" {
				continue
			}
			break
		}

		parts := strings.SplitN(strings.TrimPrefix(trimmed, "# "), ":", 2)
		key := strings.ToLower(strings.TrimSpace(parts[0]))
		value := strings.TrimSpace(parts[1])
		metadata[key] = value
	}

	return metadata
}

func splitMetadataList(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}

	parts := strings.Split(value, ",")
	items := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			items = append(items, trimmed)
		}
	}

	return items
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

func (s *Store) TaggedCount(tag string) int {
	count := 0
	for _, example := range s.examples {
		for _, current := range example.Tags {
			if current == tag {
				count++
				break
			}
		}
	}
	return count
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

func dedupe(values []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func parseFeatureRank(value string, fallback int) int {
	if strings.TrimSpace(value) == "" {
		return fallback
	}

	rank, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return fallback
	}

	return rank
}
