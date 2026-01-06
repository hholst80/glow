package ui

import (
	"strings"
	"testing"
)

// TestMarkdownRenderer is a test double for MarkdownRenderer.
// It allows tests to control rendering behavior and track calls.
type TestMarkdownRenderer struct {
	// RenderFunc allows tests to control the render output
	RenderFunc func(markdown string, width int, style string, filename string, preserveNewLines bool) (string, error)

	// RenderCalls tracks all calls to Render for verification
	RenderCalls []TestRenderCall
}

// TestRenderCall records the arguments of a Render call.
type TestRenderCall struct {
	Markdown        string
	Width           int
	Style           string
	Filename        string
	PreserveNewLines bool
}

// Render implements MarkdownRenderer.
func (r *TestMarkdownRenderer) Render(markdown string, width int, style string, filename string, preserveNewLines bool) (string, error) {
	r.RenderCalls = append(r.RenderCalls, TestRenderCall{
		Markdown:        markdown,
		Width:           width,
		Style:           style,
		Filename:        filename,
		PreserveNewLines: preserveNewLines,
	})

	if r.RenderFunc != nil {
		return r.RenderFunc(markdown, width, style, filename, preserveNewLines)
	}

	// Default: return markdown as-is
	return markdown, nil
}

// Ensure TestMarkdownRenderer implements MarkdownRenderer.
var _ MarkdownRenderer = (*TestMarkdownRenderer)(nil)

// TestRealMarkdownRenderer_MarkdownFile tests rendering a markdown file.
func TestRealMarkdownRenderer_MarkdownFile(t *testing.T) {
	r := NewMarkdownRenderer()

	input := "# Hello World\n\nThis is a **test**."
	out, err := r.Render(input, 80, "dark", "test.md", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Glamour adds ANSI styling, so output should differ from input
	if out == input {
		t.Error("expected glamour to transform markdown")
	}

	// Should contain heading text (ANSI codes may be interspersed)
	if !strings.Contains(out, "Hello") || !strings.Contains(out, "World") {
		t.Error("expected output to contain heading text")
	}
}

// TestRealMarkdownRenderer_CodeFile tests rendering a code file.
func TestRealMarkdownRenderer_CodeFile(t *testing.T) {
	r := NewMarkdownRenderer()

	input := "package main\n\nfunc main() {}\n"
	out, err := r.Render(input, 80, "dark", "test.go", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Code files get wrapped in code blocks and syntax highlighted
	// The output should be trimmed (no leading/trailing whitespace)
	if strings.HasPrefix(out, "\n") || strings.HasSuffix(out, "\n\n") {
		t.Error("expected code file output to be trimmed")
	}
}

// TestRealMarkdownRenderer_PreserveNewLines tests the preserveNewLines option.
func TestRealMarkdownRenderer_PreserveNewLines(t *testing.T) {
	r := NewMarkdownRenderer()

	input := "Line 1\n\n\n\nLine 2"
	_, err := r.Render(input, 80, "dark", "test.md", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Just verify it doesn't error - the exact newline behavior is glamour's concern
}

// TestRealMarkdownRenderer_Styles tests different style options.
func TestRealMarkdownRenderer_Styles(t *testing.T) {
	r := NewMarkdownRenderer()

	styles := []string{"dark", "light", "dracula", "notty"}
	input := "# Test\n\nSome text."

	for _, style := range styles {
		t.Run(style, func(t *testing.T) {
			_, err := r.Render(input, 80, style, "test.md", false)
			if err != nil {
				t.Errorf("style %q: unexpected error: %v", style, err)
			}
		})
	}
}

// TestRealMarkdownRenderer_ZeroWidth tests rendering with zero width.
func TestRealMarkdownRenderer_ZeroWidth(t *testing.T) {
	r := NewMarkdownRenderer()

	input := "# Test\n\nSome text that is longer than usual."
	_, err := r.Render(input, 0, "dark", "test.md", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Just verify it doesn't error with zero width
}

// TestRealMarkdownRenderer_MermaidDiagram tests mermaid diagram processing.
func TestRealMarkdownRenderer_MermaidDiagram(t *testing.T) {
	r := NewMarkdownRenderer()

	input := "# Diagram\n\n```mermaid\ngraph LR\n    A --> B\n```\n"
	out, err := r.Render(input, 80, "dark", "test.md", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Mermaid should be processed - output should contain something
	if out == "" {
		t.Error("expected non-empty output for mermaid diagram")
	}
}

// TestTestMarkdownRenderer_TracksCalls verifies the test double tracks calls.
func TestTestMarkdownRenderer_TracksCalls(t *testing.T) {
	r := &TestMarkdownRenderer{}

	_, _ = r.Render("test content", 80, "dark", "test.md", true)
	_, _ = r.Render("more content", 40, "light", "other.txt", false)

	if len(r.RenderCalls) != 2 {
		t.Fatalf("expected 2 calls, got %d", len(r.RenderCalls))
	}

	call1 := r.RenderCalls[0]
	if call1.Markdown != "test content" {
		t.Errorf("call 1: expected markdown 'test content', got %q", call1.Markdown)
	}
	if call1.Width != 80 {
		t.Errorf("call 1: expected width 80, got %d", call1.Width)
	}
	if call1.Style != "dark" {
		t.Errorf("call 1: expected style 'dark', got %q", call1.Style)
	}
	if call1.Filename != "test.md" {
		t.Errorf("call 1: expected filename 'test.md', got %q", call1.Filename)
	}
	if !call1.PreserveNewLines {
		t.Error("call 1: expected preserveNewLines to be true")
	}

	call2 := r.RenderCalls[1]
	if call2.Markdown != "more content" {
		t.Errorf("call 2: expected markdown 'more content', got %q", call2.Markdown)
	}
	if call2.PreserveNewLines {
		t.Error("call 2: expected preserveNewLines to be false")
	}
}

// TestTestMarkdownRenderer_CustomRenderFunc tests injecting custom render behavior.
func TestTestMarkdownRenderer_CustomRenderFunc(t *testing.T) {
	r := &TestMarkdownRenderer{
		RenderFunc: func(markdown string, width int, style string, filename string, preserveNewLines bool) (string, error) {
			return "CUSTOM: " + markdown, nil
		},
	}

	out, err := r.Render("input", 80, "dark", "test.md", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if out != "CUSTOM: input" {
		t.Errorf("expected 'CUSTOM: input', got %q", out)
	}
}
