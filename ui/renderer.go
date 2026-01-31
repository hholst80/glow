package ui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/hholst80/glow/mermaid"
	"github.com/hholst80/glow/utils"
)

// MarkdownRenderer abstracts markdown-to-terminal rendering for testability.
// This interface allows injecting test doubles that don't require
// the glamour library or terminal output.
type MarkdownRenderer interface {
	// Render converts markdown to styled terminal output.
	// Parameters:
	//   - markdown: the raw markdown content
	//   - width: maximum width for word wrapping (0 = no limit)
	//   - style: glamour style name (e.g., "dark", "light", "dracula")
	//   - filename: used to detect code files by extension
	//   - preserveNewLines: whether to preserve newlines in output
	// Returns the rendered output or an error.
	Render(markdown string, width int, style string, filename string, preserveNewLines bool) (string, error)
}

// RealMarkdownRenderer implements MarkdownRenderer using glamour and mermaid.
type RealMarkdownRenderer struct{}

// NewMarkdownRenderer creates a new RealMarkdownRenderer.
func NewMarkdownRenderer() *RealMarkdownRenderer {
	return &RealMarkdownRenderer{}
}

// Render converts markdown to styled terminal output using glamour.
func (r *RealMarkdownRenderer) Render(markdown string, width int, style string, filename string, preserveNewLines bool) (string, error) {
	isCode := !utils.IsMarkdownFile(filename)

	// For code files, don't apply width limit
	renderWidth := width
	if isCode {
		renderWidth = 0
	}

	options := []glamour.TermRendererOption{
		utils.GlamourStyle(style, isCode),
		glamour.WithWordWrap(renderWidth),
	}

	if preserveNewLines {
		options = append(options, glamour.WithPreservedNewLines())
	}

	renderer, err := glamour.NewTermRenderer(options...)
	if err != nil {
		return "", fmt.Errorf("error creating glamour renderer: %w", err)
	}

	// For code files, wrap in a code block
	content := markdown
	if isCode {
		content = utils.WrapCodeBlock(markdown, filepath.Ext(filename))
	}

	// Preprocess mermaid diagrams before rendering
	content = mermaid.ProcessMarkdown(content, renderWidth)

	out, err := renderer.Render(content)
	if err != nil {
		return "", fmt.Errorf("error rendering markdown: %w", err)
	}

	if isCode {
		out = strings.TrimSpace(out)
	}

	return out, nil
}

// Ensure RealMarkdownRenderer implements MarkdownRenderer.
var _ MarkdownRenderer = (*RealMarkdownRenderer)(nil)
