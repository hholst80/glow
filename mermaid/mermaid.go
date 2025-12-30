// Package mermaid provides rendering of Mermaid diagrams to ASCII art.
//
// Supported diagram types:
//   - Flowcharts: graph LR (left-to-right) and graph TD (top-down)
//   - Sequence diagrams: sequenceDiagram with participants and messages
//
// The package uses the mermaid-ascii library for rendering, which produces
// Unicode box-drawing characters for clean terminal output.
package mermaid

import (
	"regexp"
	"strings"

	mermaidcmd "github.com/AlexanderGrooff/mermaid-ascii/cmd"
)

// Renderer defines the interface for rendering Mermaid diagrams.
// This interface enables dependency injection for testing.
type Renderer interface {
	// Render converts a Mermaid diagram source to ASCII art.
	// Returns the rendered output or an error if rendering fails.
	Render(source string) (string, error)
}

// DefaultRenderer implements Renderer using the mermaid-ascii library.
type DefaultRenderer struct{}

// NewRenderer creates a new DefaultRenderer.
func NewRenderer() *DefaultRenderer {
	return &DefaultRenderer{}
}

// Render converts a Mermaid diagram source to ASCII art using mermaid-ascii.
func (r *DefaultRenderer) Render(source string) (string, error) {
	// Pass nil config to use defaults (Unicode box-drawing characters)
	return mermaidcmd.RenderDiagram(source, nil)
}

// codeBlockRegex matches fenced code blocks with mermaid language identifier.
// Matches: ```mermaid ... ``` or ~~~mermaid ... ~~~
var codeBlockRegex = regexp.MustCompile("(?s)```mermaid\\s*\n(.*?)```|~~~mermaid\\s*\n(.*?)~~~")

// Preprocessor handles preprocessing of markdown to render Mermaid diagrams.
type Preprocessor struct {
	renderer Renderer
}

// NewPreprocessor creates a new Preprocessor with the given renderer.
func NewPreprocessor(renderer Renderer) *Preprocessor {
	return &Preprocessor{renderer: renderer}
}

// Process finds and renders all Mermaid code blocks in the given markdown.
// Returns the markdown with Mermaid blocks replaced by their ASCII rendering.
// If rendering fails for a block, it is left unchanged with an error comment.
func (p *Preprocessor) Process(markdown string) string {
	return codeBlockRegex.ReplaceAllStringFunc(markdown, func(match string) string {
		// Extract the diagram source from the code block
		source := extractDiagramSource(match)
		if source == "" {
			return match
		}

		// Render the diagram
		rendered, err := p.renderer.Render(source)
		if err != nil {
			// If rendering fails, keep the original code block with an error note
			return match + "\n<!-- mermaid rendering error: " + err.Error() + " -->"
		}

		// Return the rendered diagram as a preformatted block
		return "```\n" + strings.TrimSpace(rendered) + "\n```"
	})
}

// extractDiagramSource extracts the diagram content from a mermaid code block.
func extractDiagramSource(block string) string {
	matches := codeBlockRegex.FindStringSubmatch(block)
	if len(matches) < 2 {
		return ""
	}
	// Check both capture groups (for ``` and ~~~)
	if matches[1] != "" {
		return strings.TrimSpace(matches[1])
	}
	if len(matches) > 2 && matches[2] != "" {
		return strings.TrimSpace(matches[2])
	}
	return ""
}

// ProcessMarkdown is a convenience function that processes markdown with the default renderer.
func ProcessMarkdown(markdown string) string {
	p := NewPreprocessor(NewRenderer())
	return p.Process(markdown)
}
