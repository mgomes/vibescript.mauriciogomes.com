// Package reference renders the language reference page from its embedded
// markdown source and derives the sidebar structure from its headings.
package reference

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

//go:embed content/reference.md
var content embed.FS

const sourcePath = "content/reference.md"

// Item is one linkable subsection (h3) in the sidebar.
type Item struct {
	Title string
	ID    string
}

// Section is a sidebar group: an h2 heading and its h3 subsections.
type Section struct {
	Title string
	ID    string
	Items []Item
}

// Reference is the rendered language reference page.
type Reference struct {
	Content  template.HTML
	Sections []Section
}

// Load renders the embedded markdown and collects the h2/h3 heading tree.
// Every h2 and h3 must carry an explicit {#id} attribute so anchor URLs stay
// stable when headings are reworded.
func Load() (*Reference, error) {
	source, err := content.ReadFile(sourcePath)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", sourcePath, err)
	}

	markdown := goldmark.New(
		goldmark.WithExtensions(extension.Table),
		goldmark.WithParserOptions(parser.WithAttribute()),
	)

	doc := markdown.Parser().Parse(text.NewReader(source))
	sections, err := collectSections(doc, source)
	if err != nil {
		return nil, err
	}

	var body bytes.Buffer
	if err := markdown.Renderer().Render(&body, source, doc); err != nil {
		return nil, fmt.Errorf("render %s: %w", sourcePath, err)
	}

	return &Reference{
		Content:  template.HTML(body.String()),
		Sections: sections,
	}, nil
}

func collectSections(doc ast.Node, source []byte) ([]Section, error) {
	var sections []Section

	for node := doc.FirstChild(); node != nil; node = node.NextSibling() {
		heading, ok := node.(*ast.Heading)
		if !ok || (heading.Level != 2 && heading.Level != 3) {
			continue
		}

		title := headingText(heading, source)
		id, ok := headingID(heading)
		if !ok {
			return nil, fmt.Errorf("heading %q is missing an explicit {#id}", title)
		}

		switch heading.Level {
		case 2:
			sections = append(sections, Section{Title: title, ID: id})
		case 3:
			if len(sections) == 0 {
				return nil, fmt.Errorf("subsection %q appears before any section", title)
			}
			last := &sections[len(sections)-1]
			last.Items = append(last.Items, Item{Title: title, ID: id})
		}
	}

	if len(sections) == 0 {
		return nil, fmt.Errorf("no sections found in %s", sourcePath)
	}
	return sections, nil
}

func headingText(heading *ast.Heading, source []byte) string {
	var sb strings.Builder
	for child := heading.FirstChild(); child != nil; child = child.NextSibling() {
		if textNode, ok := child.(*ast.Text); ok {
			sb.Write(textNode.Segment.Value(source))
		}
	}
	return sb.String()
}

func headingID(heading *ast.Heading) (string, bool) {
	value, ok := heading.AttributeString("id")
	if !ok {
		return "", false
	}
	switch id := value.(type) {
	case []byte:
		return string(id), true
	case string:
		return id, true
	default:
		return "", false
	}
}
