package catalog

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
}

type Store struct {
	examples []Example
	bySlug   map[string]Example
}

func NewStore(examples []Example) *Store {
	ordered := make([]Example, len(examples))
	copy(ordered, examples)

	bySlug := make(map[string]Example, len(ordered))
	for _, example := range ordered {
		bySlug[example.Slug] = example
	}

	return &Store{
		examples: ordered,
		bySlug:   bySlug,
	}
}

func (s *Store) All() []Example {
	ordered := make([]Example, len(s.examples))
	copy(ordered, s.examples)
	return ordered
}

func (s *Store) Featured(limit int) []Example {
	featured := make([]Example, 0, limit)
	for _, example := range s.examples {
		if !example.Featured {
			continue
		}
		featured = append(featured, example)
		if len(featured) == limit {
			break
		}
	}

	return featured
}

func (s *Store) BySlug(slug string) (Example, bool) {
	example, ok := s.bySlug[slug]
	return example, ok
}

func (s *Store) Count() int {
	return len(s.examples)
}

func SeedExamples() []Example {
	return []Example{
		{
			Slug:        "hello-vibescript",
			Title:       "Hello, Vibescript",
			Summary:     "The smallest possible Vibescript example, focused on syntax, shape, and the first successful output.",
			Description: "Start with the smallest complete example so the site can introduce Vibescript without forcing readers through a long setup story.",
			Category:    "Foundations",
			Difficulty:  "Intro",
			Stage:       "Planned",
			Featured:    true,
			Tags:        []string{"syntax", "quickstart", "basics"},
		},
		{
			Slug:        "data-reshape-pipeline",
			Title:       "Data Reshape Pipeline",
			Summary:     "Walk through transforming messy structured data into a cleaner schema with predictable output.",
			Description: "This example is a strong fit for an early runnable sandbox because the input, transformation, and output are all easy to inspect.",
			Category:    "Data",
			Difficulty:  "Intermediate",
			Stage:       "Planned",
			Featured:    true,
			Tags:        []string{"json", "mapping", "transforms"},
		},
		{
			Slug:        "agentic-file-loop",
			Title:       "Agentic File Loop",
			Summary:     "Show how a Vibescript flow can iterate through files, classify them, and produce a useful report.",
			Description: "A file-processing example makes the language feel practical quickly, and it sets up later examples around tools and automation.",
			Category:    "Automation",
			Difficulty:  "Intermediate",
			Stage:       "Drafting",
			Featured:    true,
			Tags:        []string{"files", "loops", "reports"},
		},
		{
			Slug:        "api-client-workflow",
			Title:       "API Client Workflow",
			Summary:     "Chain a request, validation step, and formatted response into one compact script.",
			Description: "This will eventually become a good benchmark example because it covers I/O, control flow, and response shaping in one place.",
			Category:    "Integrations",
			Difficulty:  "Intermediate",
			Stage:       "Planned",
			Tags:        []string{"http", "validation", "responses"},
		},
		{
			Slug:        "component-state-machine",
			Title:       "Component State Machine",
			Summary:     "Model a UI flow with explicit states so readers can understand how Vibescript handles branching logic.",
			Description: "State-driven examples tend to surface language ergonomics quickly, which makes them useful both for docs and future playground demos.",
			Category:    "UI",
			Difficulty:  "Advanced",
			Stage:       "Drafting",
			Tags:        []string{"ui", "states", "branching"},
		},
		{
			Slug:        "prompt-to-structured-output",
			Title:       "Prompt To Structured Output",
			Summary:     "Take a loose prompt, constrain it with a schema, and return something downstream systems can trust.",
			Description: "This example connects Vibescript to a common production concern: turning flexible language tasks into stable machine-readable output.",
			Category:    "LLM",
			Difficulty:  "Intro",
			Stage:       "Planned",
			Tags:        []string{"schemas", "prompts", "validation"},
		},
	}
}
