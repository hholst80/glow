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
func (m *MockRenderer) Render(source string, maxWidth int) (string, error) {
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
			p := NewPreprocessor(mock, 0) // 0 = no width limit

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
			result, err := r.Render(tt.source, 0) // 0 = no width limit
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

// Integration tests for supported diagram types.
// These tests use the real mermaid-ascii renderer to verify actual output.

func TestSupportedDiagramTypes(t *testing.T) {
	r := NewRenderer()

	t.Run("Flowchart_LeftToRight", func(t *testing.T) {
		source := `graph LR
    A[Start] --> B{Decision}
    B -->|Yes| C[OK]
    B -->|No| D[Cancel]`

		result, err := r.Render(source, 0)
		if err != nil {
			t.Fatalf("Flowchart LR rendering failed: %v", err)
		}

		// Verify key elements are present in the output
		expectations := []string{
			"Start",    // Node A label
			"Decision", // Node B label
			"OK",       // Node C label
			"Cancel",   // Node D label
		}

		for _, exp := range expectations {
			if !strings.Contains(result, exp) {
				t.Errorf("Flowchart LR output missing expected element %q\nGot:\n%s", exp, result)
			}
		}

		// Verify it contains box-drawing characters (Unicode rendering)
		if !strings.ContainsAny(result, "┌┐└┘├┤┬┴─│►") {
			t.Errorf("Flowchart LR output should contain box-drawing characters\nGot:\n%s", result)
		}
	})

	t.Run("Flowchart_TopDown", func(t *testing.T) {
		source := `graph TD
    A[Top] --> B[Middle]
    B --> C[Bottom]`

		result, err := r.Render(source, 0)
		if err != nil {
			t.Fatalf("Flowchart TD rendering failed: %v", err)
		}

		expectations := []string{
			"Top",
			"Middle",
			"Bottom",
		}

		for _, exp := range expectations {
			if !strings.Contains(result, exp) {
				t.Errorf("Flowchart TD output missing expected element %q\nGot:\n%s", exp, result)
			}
		}
	})

	t.Run("SequenceDiagram", func(t *testing.T) {
		source := `sequenceDiagram
    Alice->>Bob: Hello Bob
    Bob-->>Alice: Hi Alice
    Alice->>Bob: How are you?
    Bob-->>Alice: I'm good!`

		result, err := r.Render(source, 0)
		if err != nil {
			t.Fatalf("Sequence diagram rendering failed: %v", err)
		}

		// Verify participants are present
		participants := []string{
			"Alice",
			"Bob",
		}

		for _, p := range participants {
			if !strings.Contains(result, p) {
				t.Errorf("Sequence diagram output missing participant %q\nGot:\n%s", p, result)
			}
		}

		// Verify messages are present
		messages := []string{
			"Hello Bob",
			"Hi Alice",
		}

		for _, m := range messages {
			if !strings.Contains(result, m) {
				t.Errorf("Sequence diagram output missing message %q\nGot:\n%s", m, result)
			}
		}

		// Verify arrow indicators are present
		if !strings.ContainsAny(result, "►◄→←") && !strings.Contains(result, "->") {
			t.Errorf("Sequence diagram should contain arrow indicators\nGot:\n%s", result)
		}
	})
}

func TestFlowchartWithLabels(t *testing.T) {
	r := NewRenderer()

	source := `graph LR
    A --> |label1| B
    B --> |label2| C`

	result, err := r.Render(source, 0)
	if err != nil {
		t.Fatalf("Flowchart with labels rendering failed: %v", err)
	}

	// Edge labels should be present
	if !strings.Contains(result, "label1") {
		t.Errorf("Flowchart output missing edge label 'label1'\nGot:\n%s", result)
	}
	if !strings.Contains(result, "label2") {
		t.Errorf("Flowchart output missing edge label 'label2'\nGot:\n%s", result)
	}
}

func TestProcessMarkdownIntegration(t *testing.T) {
	markdown := `# My Document

Here's a flowchart:

` + "```mermaid\ngraph LR\n    A[Hello] --> B[World]\n```" + `

And here's a sequence diagram:

` + "```mermaid\nsequenceDiagram\n    Client->>Server: Request\n    Server-->>Client: Response\n```" + `

The end.`

	result := ProcessMarkdown(markdown, 0) // 0 = no width limit

	// The mermaid blocks should be replaced with rendered output
	if strings.Contains(result, "```mermaid") {
		t.Error("ProcessMarkdown should replace mermaid code blocks")
	}

	// Key content should be present
	expectations := []string{
		"# My Document",
		"Hello",
		"World",
		"Client",
		"Server",
		"Request",
		"Response",
		"The end.",
	}

	for _, exp := range expectations {
		if !strings.Contains(result, exp) {
			t.Errorf("ProcessMarkdown output missing expected content %q", exp)
		}
	}
}

func TestIsTooComplex(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		expected bool
	}{
		{
			name:     "simple flowchart",
			source:   "graph LR\n    A --> B --> C",
			expected: false,
		},
		{
			name:     "sequence diagram (not a flowchart)",
			source:   "sequenceDiagram\n    A->>B: Hello",
			expected: false,
		},
		{
			name:     "single subgraph with few edges",
			source:   "flowchart LR\n    subgraph SG[Group]\n        A --> B\n    end\n    C --> A",
			expected: false,
		},
		{
			name: "multiple subgraphs with many edges",
			source: `flowchart LR
    subgraph SENSE[Sense]
        cam[Camera]
        unity[Unity]
    end
    subgraph PERCEIVE[Perceive]
        perc[Perception]
        backend[Backend]
    end
    cam --> perc
    cam --> backend
    unity --> perc
    unity --> backend
    perc --> backend
    backend --> cam
    perc --> unity`,
			expected: true,
		},
		{
			name: "too many edges overall",
			source: `graph LR
    A --> B --> C --> D --> E --> F --> G --> H
    H --> I --> J --> K --> L --> M --> N --> O --> P --> Q`,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isTooComplex(tt.source)
			if result != tt.expected {
				t.Errorf("isTooComplex() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestComplexDiagramFromEXAMPLE(t *testing.T) {
	// This is the actual complex diagram from EXAMPLE.md that should be detected as too complex
	source := `flowchart LR
    subgraph SENSE[Sense & Simulate]
        cam["ros-basler<br/>Basler camera node"]
        unity["ros-tcp-endpoint<br/>Unity / DTS"]
        tests["test-scripts<br/>Scenario publishers"]
    end
    subgraph PERCEIVE[Perceive & Decide]
        perc["perception<br/>YOLO + tracker"]
        backend["backend<br/>FastAPI + ROS"]
        commander["commander<br/>Control arbiter"]
    end
    subgraph CONTROL[Control & Actuate]
        gimbal["gimbal-control<br/>Radians → servo"]
        servo["servo-c (ref) / servo<br/>EtherCAT bridge"]
        mech[(Pan/Tilt + weapon HW)]
    end
    subgraph EXPERIENCE[Experience & Ops]
        frontend["frontend<br/>Svelte / Android UI"]
        manager["manager<br/>Docker/DCN orchestration"]
        av["audiovisual-control<br/>Speaker + LEDs"]
        servoDriver["servo-driver<br/>Waypoint UI"]
    end

    cam -->|/sensor/visual_light_full/image_raw| perc
    cam -->|/sensor/visual_light/image_raw/compressed| backend
    perc -->|/gimbal/cmd/absolute/dts<br/>+ detections| commander
    unity -->|/gimbal/cmd/*/dts<br/>/drone positions| commander
    tests -->|/gimbal/cmd/*/user| commander
    backend -->|/command_mode<br/>/gimbal cmds - user| commander
    commander -->|/gimbal/cmd/absolute| gimbal
    commander -->|/gimbal/cmd/relative| gimbal
    commander -->|/gimbal/cmd/velocity| gimbal
    gimbal -->|/servo/cmd/trajectory<br/>JointTrajectory| servo
    servo -->|EtherCAT CSP| mech
    mech -->|Encoders| servo
    servo -->|/servo/position<br/>ros2_interfaces/ServoState| gimbal
    gimbal -->|/gimbal/state| backend
    commander -->|/turret/command_mode<br/>/turret/drone_detected| backend
    backend -->|HTTP MJPEG + SSE| frontend
    backend -->|mDNS info| frontend
    backend -->|/command_mode| av
    commander -->|/fire/cmd| mech
    commander -->|/fire/cmd| av
    manager --> backend
    manager --> commander
    manager --> gimbal
    manager --> servo
    servoDriver -->|/servo/cmd/trajectory<br/>JointTrajectory| servo`

	if !isTooComplex(source) {
		t.Error("EXAMPLE.md flowchart should be detected as too complex")
	}

	// Also verify that the renderer returns ErrTooComplex
	r := NewRenderer()
	_, err := r.Render(source, 0)
	if !errors.Is(err, ErrTooComplex) {
		t.Errorf("Render() should return ErrTooComplex for complex diagram, got: %v", err)
	}
}

func TestWidthLimitEnforcement(t *testing.T) {
	r := NewRenderer()

	// A simple diagram that renders to ~50 chars wide
	source := `graph LR
    A[Start] --> B[End]`

	// Should succeed with no width limit
	result, err := r.Render(source, 0)
	if err != nil {
		t.Fatalf("Render with no limit should succeed: %v", err)
	}

	// Get actual width
	actualWidth := 0
	for _, line := range strings.Split(result, "\n") {
		if len([]rune(line)) > actualWidth {
			actualWidth = len([]rune(line))
		}
	}

	// Should succeed with generous width limit
	_, err = r.Render(source, actualWidth+10)
	if err != nil {
		t.Errorf("Render with generous limit should succeed: %v", err)
	}

	// Should fail with tight width limit
	_, err = r.Render(source, 10)
	if !errors.Is(err, ErrTooComplex) {
		t.Errorf("Render with tight limit should return ErrTooComplex, got: %v", err)
	}
}

func TestTooComplexShowsOriginalWithNote(t *testing.T) {
	// Create a mock renderer that returns ErrTooComplex
	mock := &MockRenderer{
		RenderFunc: func(source string) (string, error) {
			return "", ErrTooComplex
		},
	}
	p := NewPreprocessor(mock, 0)

	markdown := "# Title\n\n```mermaid\ngraph LR\n    A --> B\n```\n\nText"
	result := p.Process(markdown)

	// Should contain the visual cue
	if !strings.Contains(result, "⚠") {
		t.Error("Result should contain warning symbol for too complex diagram")
	}
	if !strings.Contains(result, "too complex") {
		t.Error("Result should mention 'too complex'")
	}

	// Should preserve the original mermaid block
	if !strings.Contains(result, "```mermaid") {
		t.Error("Result should preserve original mermaid code block")
	}
	if !strings.Contains(result, "graph LR") {
		t.Error("Result should preserve original diagram source")
	}
}
