package mermaid

import (
	"errors"
	"strings"
	"testing"
)

// MockRenderer is a mock implementation of Renderer for testing.
type MockRenderer struct {
	// RenderFunc allows customizing the render behavior in tests.
	RenderFunc func(source string) (string, error)
	// Calls records all calls to Render for verification.
	Calls []string
}

// Render implements Renderer interface.
func (m *MockRenderer) Render(source string) (string, error) {
	m.Calls = append(m.Calls, source)
	if m.RenderFunc != nil {
		return m.RenderFunc(source)
	}
	// Default: return a simple box representation
	return "+---+\n| " + strings.Split(source, "\n")[0][:min(3, len(source))] + " |\n+---+", nil
}

func TestExtractDiagramSource(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "backtick fence",
			input:    "```mermaid\ngraph LR\n  A --> B\n```",
			expected: "graph LR\n  A --> B",
		},
		{
			name:     "tilde fence",
			input:    "~~~mermaid\nsequenceDiagram\n  A->>B: Hello\n~~~",
			expected: "sequenceDiagram\n  A->>B: Hello",
		},
		{
			name:     "with extra whitespace",
			input:    "```mermaid  \n\n  graph TD\n  A --> B\n\n```",
			expected: "graph TD\n  A --> B",
		},
		{
			name:     "empty block",
			input:    "```mermaid\n```",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractDiagramSource(tt.input)
			if result != tt.expected {
				t.Errorf("extractDiagramSource(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestPreprocessor_Process(t *testing.T) {
	tests := []struct {
		name           string
		markdown       string
		renderFunc     func(string) (string, error)
		wantContains   []string
		wantNotContain []string
		wantCalls      int
	}{
		{
			name: "single mermaid block",
			markdown: `# Title

Some text.

` + "```mermaid\ngraph LR\n  A --> B\n```" + `

More text.`,
			renderFunc: func(s string) (string, error) {
				return "RENDERED", nil
			},
			wantContains:   []string{"# Title", "Some text.", "RENDERED", "More text."},
			wantNotContain: []string{"```mermaid", "graph LR"},
			wantCalls:      1,
		},
		{
			name: "multiple mermaid blocks",
			markdown: "```mermaid\ngraph LR\n```\n\n```mermaid\nsequenceDiagram\n```",
			renderFunc: func(s string) (string, error) {
				if strings.Contains(s, "graph") {
					return "GRAPH", nil
				}
				return "SEQUENCE", nil
			},
			wantContains: []string{"GRAPH", "SEQUENCE"},
			wantCalls:    2,
		},
		{
			name:     "no mermaid blocks",
			markdown: "# Just a title\n\n```go\nfunc main() {}\n```",
			renderFunc: func(s string) (string, error) {
				return "RENDERED", nil
			},
			wantContains:   []string{"# Just a title", "```go"},
			wantNotContain: []string{"RENDERED"},
			wantCalls:      0,
		},
		{
			name:     "render error preserves original",
			markdown: "```mermaid\ninvalid\n```",
			renderFunc: func(s string) (string, error) {
				return "", errors.New("parse error")
			},
			wantContains: []string{"```mermaid", "invalid", "mermaid rendering error"},
			wantCalls:    1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockRenderer{RenderFunc: tt.renderFunc}
			p := NewPreprocessor(mock)

			result := p.Process(tt.markdown)

			// Check expected content is present
			for _, want := range tt.wantContains {
				if !strings.Contains(result, want) {
					t.Errorf("Process() result should contain %q, got:\n%s", want, result)
				}
			}

			// Check unwanted content is absent
			for _, notWant := range tt.wantNotContain {
				if strings.Contains(result, notWant) {
					t.Errorf("Process() result should not contain %q, got:\n%s", notWant, result)
				}
			}

			// Verify number of render calls
			if len(mock.Calls) != tt.wantCalls {
				t.Errorf("Render() called %d times, want %d", len(mock.Calls), tt.wantCalls)
			}
		})
	}
}

func TestDefaultRenderer_Render(t *testing.T) {
	r := NewRenderer()

	tests := []struct {
		name    string
		source  string
		wantErr bool
	}{
		{
			name:    "simple flowchart",
			source:  "graph LR\n  A --> B",
			wantErr: false,
		},
		{
			name:    "invalid syntax",
			source:  "not a valid mermaid diagram",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := r.Render(tt.source)
			if (err != nil) != tt.wantErr {
				t.Errorf("Render() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result == "" {
				t.Error("Render() returned empty result for valid input")
			}
		})
	}
}

func TestCodeBlockRegex(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		matches bool
	}{
		{
			name:    "backtick mermaid",
			input:   "```mermaid\ngraph LR\n```",
			matches: true,
		},
		{
			name:    "tilde mermaid",
			input:   "~~~mermaid\ngraph LR\n~~~",
			matches: true,
		},
		{
			name:    "other language",
			input:   "```go\nfunc main() {}\n```",
			matches: false,
		},
		{
			name:    "mermaid with extra space",
			input:   "```mermaid \ngraph LR\n```",
			matches: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := codeBlockRegex.MatchString(tt.input)
			if result != tt.matches {
				t.Errorf("codeBlockRegex.MatchString(%q) = %v, want %v", tt.input, result, tt.matches)
			}
		})
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
