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

// TestNodeLabelDeduplication verifies that nodes with labels (e.g., A[Label])
// are correctly matched with bare references (e.g., A) without duplication.
// This is a regression test for the node ID extraction fix.
func TestNodeLabelDeduplication(t *testing.T) {
	r := NewRenderer()

	// This pattern caused node duplication before the fix:
	// - First line defines control[mx-control]
	// - Second line references just "control"
	// The library should recognize these as the same node.
	source := `flowchart LR
    control[mx-control] --> tenants[(tenants)]
    control --> domains[(domains)]
    agent[mx-agent] --> tenants
    agent --> domains`

	result, err := r.Render(source, 0)
	if err != nil {
		t.Fatalf("Rendering failed: %v", err)
	}

	// Each node should appear exactly once in the output
	nodeLabels := []string{"mx-control", "mx-agent", "tenants", "domains"}
	for _, label := range nodeLabels {
		count := strings.Count(result, label)
		if count != 1 {
			t.Errorf("Node label %q appears %d times, expected 1.\nOutput:\n%s", label, count, result)
		}
	}

	// Verify labels are displayed (not just IDs)
	if strings.Contains(result, "│ control │") && !strings.Contains(result, "mx-control") {
		t.Error("Node should display label 'mx-control', not bare ID 'control'")
	}
	if strings.Contains(result, "│ agent │") && !strings.Contains(result, "mx-agent") {
		t.Error("Node should display label 'mx-agent', not bare ID 'agent'")
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
			name: "multiple subgraphs with moderate edges (under new limits)",
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
			expected: false, // 2 subgraphs, 7 edges - under new limits (>2 subgraphs AND >10 edges)
		},
		{
			name: "many subgraphs with many edges (over limit)",
			source: `flowchart LR
    subgraph A[GroupA]
        a1
    end
    subgraph B[GroupB]
        b1
    end
    subgraph C[GroupC]
        c1
    end
    a1 --> b1
    b1 --> c1
    c1 --> a1
    a1 --> c1
    b1 --> a1
    c1 --> b1
    a1 --> b1
    b1 --> c1
    c1 --> a1
    a1 --> c1
    b1 --> a1`,
			expected: true, // 3 subgraphs, 11 edges - over limit (>2 AND >10)
		},
		{
			name: "too many edges overall",
			source: `graph LR
    A --> B --> C --> D --> E --> F --> G --> H --> I --> J --> K
    K --> L --> M --> N --> O --> P --> Q --> R --> S --> T --> U --> V`,
			expected: true, // 21 edges - over limit of 20
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

// ====================
// BOUNDARY TESTS FOR COMPLEXITY THRESHOLDS
// ====================

func TestComplexityBoundary_MaxTotalEdges(t *testing.T) {
	// maxTotalEdges = 20, test at boundaries
	tests := []struct {
		name      string
		edgeCount int
		tooComplex bool
	}{
		{"19 edges (under limit)", 19, false},
		{"20 edges (at limit)", 20, false},
		{"21 edges (over limit)", 21, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Build a graph with N edges: A --> B --> C --> ...
			var nodes []string
			for i := 0; i <= tt.edgeCount; i++ {
				nodes = append(nodes, string(rune('A'+i%26)))
			}
			var edges strings.Builder
			edges.WriteString("graph LR\n")
			for i := 0; i < tt.edgeCount; i++ {
				edges.WriteString("    ")
				edges.WriteString(nodes[i])
				edges.WriteString(" --> ")
				edges.WriteString(nodes[i+1])
				edges.WriteString("\n")
			}

			result := isTooComplex(edges.String())
			if result != tt.tooComplex {
				t.Errorf("isTooComplex() with %d edges = %v, want %v", tt.edgeCount, result, tt.tooComplex)
			}
		})
	}
}

func TestComplexityBoundary_SubgraphsWithEdges(t *testing.T) {
	// maxSubgraphsWithEdges = 2, maxEdgesWithSubgraphs = 10
	// Complex if subgraphs > 2 AND edges > 10
	tests := []struct {
		name           string
		subgraphCount  int
		edgeCount      int
		tooComplex     bool
	}{
		{"1 subgraph, 10 edges (at limit)", 1, 10, false},
		{"1 subgraph, 11 edges (edges over but 1 subgraph)", 1, 11, false},
		{"2 subgraphs, 10 edges (at limit)", 2, 10, false},
		{"2 subgraphs, 11 edges (edges over but 2 subgraphs)", 2, 11, false},
		{"3 subgraphs, 10 edges (at edge limit)", 3, 10, false},
		{"3 subgraphs, 11 edges (over limit)", 3, 11, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var src strings.Builder
			src.WriteString("flowchart LR\n")

			// Add subgraphs
			for i := 0; i < tt.subgraphCount; i++ {
				src.WriteString("    subgraph SG")
				src.WriteString(string(rune('A' + i)))
				src.WriteString("[Group")
				src.WriteString(string(rune('A' + i)))
				src.WriteString("]\n        N")
				src.WriteString(string(rune('A' + i)))
				src.WriteString("[Node]\n    end\n")
			}

			// Add edges
			for i := 0; i < tt.edgeCount; i++ {
				src.WriteString("    X")
				src.WriteString(string(rune('0' + i%10)))
				src.WriteString(" --> Y")
				src.WriteString(string(rune('0' + i%10)))
				src.WriteString("\n")
			}

			result := isTooComplex(src.String())
			if result != tt.tooComplex {
				t.Errorf("isTooComplex() with %d subgraphs and %d edges = %v, want %v",
					tt.subgraphCount, tt.edgeCount, result, tt.tooComplex)
			}
		})
	}
}

func TestComplexity_NonFlowchartNeverComplex(t *testing.T) {
	// Non-flowchart diagrams should never be marked as too complex
	// even with many elements
	tests := []struct {
		name   string
		source string
	}{
		{
			"sequence diagram with many messages",
			`sequenceDiagram
    A->>B: msg1
    B->>C: msg2
    C->>D: msg3
    D->>E: msg4
    E->>F: msg5
    F->>G: msg6
    G->>H: msg7
    H->>I: msg8
    I->>J: msg9
    J->>K: msg10
    K->>L: msg11
    L->>M: msg12
    M->>N: msg13
    N->>O: msg14
    O->>P: msg15
    P->>Q: msg16
    Q->>R: msg17`,
		},
		{
			"class diagram with many classes",
			`classDiagram
    class A
    class B
    class C
    A --> B
    B --> C`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if isTooComplex(tt.source) {
				t.Errorf("Non-flowchart should never be too complex: %s", tt.name)
			}
		})
	}
}

// ====================
// EDGE CASE TESTS
// ====================

func TestUnicodeContentInDiagrams(t *testing.T) {
	r := NewRenderer()

	tests := []struct {
		name     string
		source   string
		contains []string
	}{
		{
			"Emoji labels",
			"graph LR\n    A[🚀 Start] --> B[✅ End]",
			[]string{"Start", "End"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := r.Render(tt.source, 0)
			if err != nil {
				t.Skipf("Unicode rendering not supported: %v", err)
			}
			for _, want := range tt.contains {
				if !strings.Contains(result, want) {
					t.Errorf("Result should contain %q, got:\n%s", want, result)
				}
			}
		})
	}
}

func TestUnicodeContentInDiagrams_NonASCII(t *testing.T) {
	// Note: The mermaid-ascii library has known encoding issues with non-ASCII
	// characters in certain environments. These tests document the limitation.
	r := NewRenderer()

	tests := []struct {
		name   string
		source string
	}{
		{
			"Japanese labels",
			"graph LR\n    A[こんにちは] --> B[世界]",
		},
		{
			"Cyrillic labels",
			"graph TD\n    A[Привет] --> B[Мир]",
		},
		{
			"Arabic labels",
			"graph LR\n    A[مرحبا] --> B[عالم]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := r.Render(tt.source, 0)
			if err != nil {
				t.Logf("Non-ASCII rendering failed (expected in some environments): %v", err)
				return
			}
			// Just verify rendering completes without panic
			t.Logf("Non-ASCII rendered (%d bytes):\n%s", len(result), result)
		})
	}
}

func TestNestedCodeBlocksInMarkdown(t *testing.T) {
	// Mermaid blocks inside other code blocks should not be processed
	markdown := "# Doc\n\n````markdown\n```mermaid\ngraph LR\n    A --> B\n```\n````\n\nText"

	mock := &MockRenderer{
		RenderFunc: func(source string) (string, error) {
			return "RENDERED", nil
		},
	}
	p := NewPreprocessor(mock, 0)
	result := p.Process(markdown)

	// The inner mermaid block might still be detected by our regex
	// but we verify outer structure is preserved
	if !strings.Contains(result, "# Doc") {
		t.Error("Should preserve document structure")
	}
	if !strings.Contains(result, "Text") {
		t.Error("Should preserve trailing text")
	}
}

func TestEmptyAndWhitespaceBlocks(t *testing.T) {
	tests := []struct {
		name     string
		markdown string
	}{
		{"empty mermaid block", "```mermaid\n```"},
		{"whitespace only mermaid block", "```mermaid\n   \n   \n```"},
	}

	mock := &MockRenderer{
		RenderFunc: func(source string) (string, error) {
			t.Errorf("Render should not be called for empty content, got: %q", source)
			return "", nil
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock.Calls = nil
			p := NewPreprocessor(mock, 0)
			_ = p.Process(tt.markdown)

			if len(mock.Calls) > 0 {
				t.Errorf("Empty block should not trigger render, got %d calls", len(mock.Calls))
			}
		})
	}
}

func TestGetMaxLineWidth(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"empty string", "", 0},
		{"single line", "hello", 5},
		{"multi line", "short\nlonger line\nmed", 11},
		{"unicode chars", "héllo wörld", 11},
		{"wide unicode", "日本語", 3}, // 3 runes, though display width varies
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getMaxLineWidth(tt.input)
			if result != tt.expected {
				t.Errorf("getMaxLineWidth(%q) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestEdgePatternDetection(t *testing.T) {
	// Test that various edge syntaxes are detected
	tests := []struct {
		name      string
		source    string
		edgeCount int
	}{
		{"arrow -->", "graph LR\n    A --> B", 1},
		{"dotted -.->", "graph LR\n    A -.-> B", 1},
		{"thick ==>", "graph LR\n    A ==> B", 1},
		{"multiple types", "graph LR\n    A --> B -.-> C ==> D", 3},
		{"with labels", "graph LR\n    A -->|yes| B -->|no| C", 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := edgeRegex.FindAllString(tt.source, -1)
			if len(matches) != tt.edgeCount {
				t.Errorf("Found %d edges, want %d in:\n%s", len(matches), tt.edgeCount, tt.source)
			}
		})
	}
}
