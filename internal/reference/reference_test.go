package reference

import (
	"regexp"
	"strings"
	"testing"

	"github.com/mgomes/vibescript-lang.org/internal/runner"
	"github.com/mgomes/vibescript/vibes"
)

func TestLoadBuildsSidebarFromHeadings(t *testing.T) {
	ref, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}

	wantSections := []string{"basics", "functions", "calls", "operators", "control-flow", "types", "runtime"}
	if len(ref.Sections) != len(wantSections) {
		t.Fatalf("got %d sections, want %d", len(ref.Sections), len(wantSections))
	}
	for i, want := range wantSections {
		if ref.Sections[i].ID != want {
			t.Errorf("section %d ID = %q, want %q", i, ref.Sections[i].ID, want)
		}
		if len(ref.Sections[i].Items) == 0 {
			t.Errorf("section %q has no subsections", ref.Sections[i].ID)
		}
	}
}

func TestLoadRendersAnchorsAndCode(t *testing.T) {
	ref, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}

	html := string(ref.Content)
	for _, want := range []string{`id="basics"`, `id="parameters"`, `id="sandbox"`, `class="language-vibe"`, "<table>"} {
		if !strings.Contains(html, want) {
			t.Errorf("rendered content missing %q", want)
		}
	}
}

var fencePattern = regexp.MustCompile("(?ms)^```vibe\n(.*?)^```$")

// TestCodeBlocksCompile keeps the reference honest: every fenced vibe example
// must compile against the same pinned interpreter that runs the site's
// examples. Fragment-style blocks are retried wrapped in a function body,
// mirroring how they would appear inside a script.
func TestCodeBlocksCompile(t *testing.T) {
	source, err := content.ReadFile(sourcePath)
	if err != nil {
		t.Fatalf("read %s: %v", sourcePath, err)
	}

	engine, err := vibes.NewEngine(runner.EngineConfig)
	if err != nil {
		t.Fatalf("new vibes engine: %v", err)
	}

	blocks := fencePattern.FindAllStringSubmatch(string(source), -1)
	if len(blocks) == 0 {
		t.Fatal("no fenced vibe blocks found in reference.md")
	}

	for i, block := range blocks {
		snippet := block[1]
		if _, err := engine.Compile(snippet); err == nil {
			continue
		}

		wrapped := "def __reference_check\n" + snippet + "\nend\n"
		if _, err := engine.Compile(wrapped); err != nil {
			t.Errorf("block %d does not compile: %v\n%s", i+1, err, snippet)
		}
	}
}
