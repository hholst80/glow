package ui

import (
	"testing"
)

func TestParseHeadings(t *testing.T) {
	tests := []struct {
		name     string
		markdown string
		expected []Heading
	}{
		{
			name:     "single heading",
			markdown: "# Hello World",
			expected: []Heading{
				{Level: 1, Text: "Hello World", Line: 0},
			},
		},
		{
			name: "multiple headings",
			markdown: `# Title
Some text
## Section 1
More text
### Subsection
#### Deep`,
			expected: []Heading{
				{Level: 1, Text: "Title", Line: 0},
				{Level: 2, Text: "Section 1", Line: 2},
				{Level: 3, Text: "Subsection", Line: 4},
				{Level: 4, Text: "Deep", Line: 5},
			},
		},
		{
			name: "headings in code blocks should be ignored",
			markdown: `# Real Heading
` + "```" + `
# This is not a heading
` + "```" + `
## Another Real Heading`,
			expected: []Heading{
				{Level: 1, Text: "Real Heading", Line: 0},
				{Level: 2, Text: "Another Real Heading", Line: 4},
			},
		},
		{
			name: "headings in tilde code blocks should be ignored",
			markdown: `# Real Heading
~~~
# This is not a heading
~~~
## Another Real Heading`,
			expected: []Heading{
				{Level: 1, Text: "Real Heading", Line: 0},
				{Level: 2, Text: "Another Real Heading", Line: 4},
			},
		},
		{
			name: "heading with trailing hashes",
			markdown: "## Hello World ##",
			expected: []Heading{
				{Level: 2, Text: "Hello World", Line: 0},
			},
		},
		{
			name:     "all heading levels",
			markdown: "# H1\n## H2\n### H3\n#### H4\n##### H5\n###### H6",
			expected: []Heading{
				{Level: 1, Text: "H1", Line: 0},
				{Level: 2, Text: "H2", Line: 1},
				{Level: 3, Text: "H3", Line: 2},
				{Level: 4, Text: "H4", Line: 3},
				{Level: 5, Text: "H5", Line: 4},
				{Level: 6, Text: "H6", Line: 5},
			},
		},
		{
			name:     "empty markdown",
			markdown: "",
			expected: nil,
		},
		{
			name:     "no headings",
			markdown: "Just some text\nwith no headings",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseHeadings(tt.markdown)
			if len(got) != len(tt.expected) {
				t.Errorf("parseHeadings() returned %d headings, want %d", len(got), len(tt.expected))
				return
			}
			for i, h := range got {
				if h.Level != tt.expected[i].Level {
					t.Errorf("heading[%d].Level = %d, want %d", i, h.Level, tt.expected[i].Level)
				}
				if h.Text != tt.expected[i].Text {
					t.Errorf("heading[%d].Text = %q, want %q", i, h.Text, tt.expected[i].Text)
				}
				if h.Line != tt.expected[i].Line {
					t.Errorf("heading[%d].Line = %d, want %d", i, h.Line, tt.expected[i].Line)
				}
			}
		})
	}
}

func TestCalculateOutlineWidth(t *testing.T) {
	tests := []struct {
		termWidth int
		expected  int
	}{
		{60, 0},  // Too narrow
		{80, 20}, // Minimum width
		{100, 25},
		{120, 30},
		{200, 40}, // Max width
	}

	for _, tt := range tests {
		got := calculateOutlineWidth(tt.termWidth)
		if got != tt.expected {
			t.Errorf("calculateOutlineWidth(%d) = %d, want %d", tt.termWidth, got, tt.expected)
		}
	}
}

func TestOutlineModelNavigation(t *testing.T) {
	common := &commonModel{}
	m := newOutlineModel(common)
	m.setContent("# H1\n## H2\n### H3")
	m.setSize(30, 10)

	// Test initial state
	if m.cursor != 0 {
		t.Errorf("initial cursor = %d, want 0", m.cursor)
	}

	// Test moveCursorDown
	m.moveCursorDown()
	if m.cursor != 1 {
		t.Errorf("after moveCursorDown, cursor = %d, want 1", m.cursor)
	}

	// Test moveCursorDown at end
	m.moveCursorDown()
	m.moveCursorDown() // Should not go past the end
	if m.cursor != 2 {
		t.Errorf("cursor should not exceed headings count, got %d, want 2", m.cursor)
	}

	// Test moveCursorUp
	m.moveCursorUp()
	if m.cursor != 1 {
		t.Errorf("after moveCursorUp, cursor = %d, want 1", m.cursor)
	}

	// Test moveCursorUp at start
	m.moveCursorUp()
	m.moveCursorUp() // Should not go below 0
	if m.cursor != 0 {
		t.Errorf("cursor should not go below 0, got %d", m.cursor)
	}
}

func TestOutlineModelNextPrevHeading(t *testing.T) {
	common := &commonModel{}
	m := newOutlineModel(common)
	m.setContent("# H1\n## H2\n### H3")

	// Test nextHeadingIndex from start
	m.current = 0
	if idx := m.nextHeadingIndex(); idx != 1 {
		t.Errorf("nextHeadingIndex() = %d, want 1", idx)
	}

	// Test nextHeadingIndex from end
	m.current = 2
	if idx := m.nextHeadingIndex(); idx != -1 {
		t.Errorf("nextHeadingIndex() at end = %d, want -1", idx)
	}

	// Test prevHeadingIndex from middle
	m.current = 1
	if idx := m.prevHeadingIndex(); idx != 0 {
		t.Errorf("prevHeadingIndex() = %d, want 0", idx)
	}

	// Test prevHeadingIndex from start
	m.current = 0
	if idx := m.prevHeadingIndex(); idx != -1 {
		t.Errorf("prevHeadingIndex() at start = %d, want -1", idx)
	}
}

func TestOutlineModelUpdateCurrent(t *testing.T) {
	common := &commonModel{}
	m := newOutlineModel(common)
	m.setContent("# H1\ntext\n## H2\nmore text\n### H3")
	m.setSize(30, 10)

	// Line 0 should select H1
	m.updateCurrent(0)
	if m.current != 0 {
		t.Errorf("updateCurrent(0) set current = %d, want 0", m.current)
	}

	// Line 2 should select H2
	m.updateCurrent(2)
	if m.current != 1 {
		t.Errorf("updateCurrent(2) set current = %d, want 1", m.current)
	}

	// Line 4 should select H3
	m.updateCurrent(4)
	if m.current != 2 {
		t.Errorf("updateCurrent(4) set current = %d, want 2", m.current)
	}
}
