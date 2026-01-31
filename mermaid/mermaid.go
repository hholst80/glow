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
	"errors"
	"regexp"
	"strings"

	"github.com/charmbracelet/glow/v2/mermaid/ascii"
)

// ErrTooComplex is returned when a diagram is too complex for ASCII rendering.
var ErrTooComplex = errors.New("diagram too complex for ASCII rendering")

// Renderer defines the interface for rendering Mermaid diagrams.
// This interface enables dependency injection for testing.
type Renderer interface {
	// Render converts a Mermaid diagram source to ASCII art.
	// Returns the rendered output or an error if rendering fails.
	// maxWidth specifies the maximum allowed output width (0 = no limit).
	Render(source string, maxWidth int) (string, error)
}

// DefaultRenderer implements Renderer using the mermaid-ascii library.
type DefaultRenderer struct{}

// NewRenderer creates a new DefaultRenderer.
func NewRenderer() *DefaultRenderer {
	return &DefaultRenderer{}
}

// Render converts a Mermaid diagram source to ASCII art using mermaid-ascii.
// Returns ErrTooComplex if the diagram is too complex or the output exceeds maxWidth.
func (r *DefaultRenderer) Render(source string, maxWidth int) (string, error) {
	if isTooComplex(source) {
		return "", ErrTooComplex
	}
	// Pass nil config to use defaults (Unicode box-drawing characters)
	result, err := ascii.RenderDiagram(source, nil)
	if err != nil {
		return "", err
	}

	// Check if output exceeds max width
	if maxWidth > 0 && getMaxLineWidth(result) > maxWidth {
		return "", ErrTooComplex
	}

	return result, nil
}

// getMaxLineWidth returns the maximum line width in the given text.
func getMaxLineWidth(text string) int {
	maxWidth := 0
	for _, line := range strings.Split(text, "\n") {
		// Count runes for proper Unicode handling
		width := len([]rune(line))
		if width > maxWidth {
			maxWidth = width
		}
	}
	return maxWidth
}

// Complexity thresholds for flowcharts with subgraphs.
// The mermaid-ascii library has layout issues with subgraphs that have
// external connections, causing node duplication and garbled output.
const (
	maxSubgraphsWithEdges = 1  // Max subgraphs when diagram has cross-subgraph edges
	maxEdgesWithSubgraphs = 6  // Max edges when subgraphs are present
	maxTotalEdges         = 15 // Max edges for any flowchart
)

// Regex patterns for complexity detection
var (
	flowchartRegex = regexp.MustCompile(`(?i)^\s*(graph|flowchart)\s+(LR|RL|TD|TB|BT)`)
	subgraphRegex  = regexp.MustCompile(`(?i)\bsubgraph\b`)
	edgeRegex      = regexp.MustCompile(`-->|--[^>]|-.->|-\.-|==>|~~~|&`)
)

// isTooComplex checks if a diagram is too complex for the ASCII renderer.
// Flowcharts with multiple subgraphs and cross-subgraph edges render poorly
// due to node duplication bugs in the mermaid-ascii library.
func isTooComplex(source string) bool {
	// Only apply complexity checks to flowcharts
	if !flowchartRegex.MatchString(source) {
		return false
	}

	subgraphCount := len(subgraphRegex.FindAllString(source, -1))
	edgeCount := len(edgeRegex.FindAllString(source, -1))

	// Too many edges overall
	if edgeCount > maxTotalEdges {
		return true
	}

	// Subgraphs with significant edges cause layout issues
	if subgraphCount > maxSubgraphsWithEdges && edgeCount > maxEdgesWithSubgraphs {
		return true
	}

	return false
}

// codeBlockRegex matches fenced code blocks with mermaid language identifier.
// Matches: ```mermaid ... ``` or ~~~mermaid ... ~~~
var codeBlockRegex = regexp.MustCompile("(?s)```mermaid\\s*\n(.*?)```|~~~mermaid\\s*\n(.*?)~~~")

// Preprocessor handles preprocessing of markdown to render Mermaid diagrams.
type Preprocessor struct {
	renderer Renderer
	maxWidth int
}

// NewPreprocessor creates a new Preprocessor with the given renderer.
// maxWidth specifies the maximum allowed output width (0 = no limit).
func NewPreprocessor(renderer Renderer, maxWidth int) *Preprocessor {
	return &Preprocessor{renderer: renderer, maxWidth: maxWidth}
}

// tooComplexNote is shown when a diagram cannot be rendered in ASCII.
const tooComplexNote = "  ⚠ [Diagram too complex for terminal - view in markdown renderer]"

// Process finds and renders all Mermaid code blocks in the given markdown.
// Returns the markdown with Mermaid blocks replaced by their ASCII rendering.
// If rendering fails for a block, it is left unchanged with an error comment.
// If a diagram is too complex, it shows the original source with a note.
func (p *Preprocessor) Process(markdown string) string {
	return codeBlockRegex.ReplaceAllStringFunc(markdown, func(match string) string {
		// Extract the diagram source from the code block
		source := extractDiagramSource(match)
		if source == "" {
			return match
		}

		// Render the diagram
		rendered, err := p.renderer.Render(source, p.maxWidth)
		if err != nil {
			if errors.Is(err, ErrTooComplex) {
				// Show original source with a visual cue
				return tooComplexNote + "\n" + match
			}
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
// maxWidth specifies the maximum allowed output width (0 = no limit).
func ProcessMarkdown(markdown string, maxWidth int) string {
	p := NewPreprocessor(NewRenderer(), maxWidth)
	return p.Process(markdown)
}
