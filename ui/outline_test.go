package ui

import (
	"strings"
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

func TestPagerOutlineIntegration(t *testing.T) {
	// Test that outline works correctly when toggled
	common := &commonModel{
		cfg:    Config{ShowOutline: false},
		width:  120,
		height: 40,
	}

	m := newPagerModel(common)
	m.currentDocument = markdown{
		Note: "test.md",
		Body: "# Title\n## Section 1\nSome content\n## Section 2\nMore content",
	}

	// Initially outline should be off
	if m.showOutline {
		t.Error("showOutline should be false initially")
	}

	// Toggle outline on
	m.showOutline = true
	m.setSize(common.width, common.height)

	// Outline should now be visible (width > 80)
	if !m.outline.visible {
		t.Errorf("outline.visible should be true after enabling, got false (width=%d)", common.width)
	}

	// Parse headings
	m.outline.setContent(m.currentDocument.Body)

	// Should have 3 headings
	if len(m.outline.headings) != 3 {
		t.Errorf("expected 3 headings, got %d", len(m.outline.headings))
	}

	// Test with narrow terminal (< 80)
	common.width = 60
	m.setSize(common.width, common.height)

	// Outline should not be visible on narrow terminal
	if m.outline.visible {
		t.Error("outline.visible should be false when terminal width < 80")
	}
}

func TestPagerIsMarkdownFile(t *testing.T) {
	common := &commonModel{
		cfg:    Config{},
		width:  120,
		height: 40,
	}

	tests := []struct {
		note     string
		expected bool
	}{
		{"README.md", true},
		{"doc.markdown", true},
		{"file.txt", false},
		{"code.go", false},
		{"", true}, // Empty extension defaults to markdown
	}

	for _, tt := range tests {
		m := newPagerModel(common)
		m.currentDocument = markdown{Note: tt.note}
		got := m.isMarkdownFile()
		if got != tt.expected {
			t.Errorf("isMarkdownFile() with Note=%q = %v, want %v", tt.note, got, tt.expected)
		}
	}
}

func TestOutlineViewRendering(t *testing.T) {
	common := &commonModel{}
	m := newOutlineModel(common)

	// Set up outline with test content
	testMd := "# Title\n## Section 1\nContent here\n## Section 2\nMore content\n### Subsection"
	m.setContent(testMd)
	m.setSize(30, 20)
	m.visible = true

	// Verify headings were parsed
	if len(m.headings) != 4 {
		t.Errorf("Expected 4 headings, got %d", len(m.headings))
	}

	// Get the view output
	view := m.View()

	// View should not be empty
	if view == "" {
		t.Error("outline.View() returned empty string")
	}

	// View should contain "OUTLINE" title
	if !strings.Contains(view, "OUTLINE") {
		t.Errorf("outline.View() should contain 'OUTLINE', got: %q", view)
	}

	// View should contain heading text
	if !strings.Contains(view, "Title") {
		t.Errorf("outline.View() should contain 'Title', got: %q", view)
	}

	t.Logf("Outline view output:\n%s", view)
	t.Logf("View length: %d chars, %d lines", len(view), strings.Count(view, "\n")+1)
}

func TestPagerViewWithOutline(t *testing.T) {
	common := &commonModel{
		cfg:    Config{ShowOutline: true},
		width:  100,
		height: 25,
	}

	m := newPagerModel(common)
	m.currentDocument = markdown{
		Note: "test.md",
		Body: "# Title\n## Section 1\nContent here\n## Section 2\nMore content",
	}

	// Enable outline
	m.showOutline = true
	m.setSize(common.width, common.height)

	// Parse headings
	m.outline.setContent(m.currentDocument.Body)

	// Set some viewport content
	m.viewport.SetContent("Line 1\nLine 2\nLine 3\nLine 4\nLine 5")

	t.Logf("Pager state:")
	t.Logf("  showOutline: %v", m.showOutline)
	t.Logf("  outline.visible: %v", m.outline.visible)
	t.Logf("  outline.headings: %d", len(m.outline.headings))
	t.Logf("  outline.width: %d", m.outline.width)
	t.Logf("  outline.height: %d", m.outline.height)
	t.Logf("  viewport.Width: %d", m.viewport.Width)
	t.Logf("  viewport.Height: %d", m.viewport.Height)

	// Get pager view
	view := m.View()

	t.Logf("Pager view output (%d chars):\n%s", len(view), view)

	// View should contain outline
	if !strings.Contains(view, "OUTLINE") {
		t.Error("Pager view should contain 'OUTLINE'")
	}
}

func TestPagerViewWithANSIContent(t *testing.T) {
	common := &commonModel{
		cfg:    Config{ShowOutline: true},
		width:  100,
		height: 25,
	}

	m := newPagerModel(common)
	m.currentDocument = markdown{
		Note: "test.md",
		Body: "# Title\n## Section 1\nContent here",
	}

	// Enable outline
	m.showOutline = true
	m.setSize(common.width, common.height)

	// Parse headings
	m.outline.setContent(m.currentDocument.Body)

	// Set viewport content with ANSI codes (simulating glamour output)
	// This includes bold, colors, etc.
	ansiContent := "\x1b[1m# Title\x1b[0m\n\x1b[38;5;208mSection 1\x1b[0m\nSome \x1b[4munderlined\x1b[0m text\nLine 4\nLine 5"
	m.viewport.SetContent(ansiContent)

	t.Logf("Pager state:")
	t.Logf("  viewport.Width: %d", m.viewport.Width)
	t.Logf("  outline.width: %d", m.outline.width)

	// Get pager view
	view := m.View()

	t.Logf("Pager view (showing first 1000 chars):\n%s", view[:min(len(view), 1000)])

	// View should contain outline
	if !strings.Contains(view, "OUTLINE") {
		t.Error("Pager view should contain 'OUTLINE' with ANSI content")
	}
}

func TestPagerViewWithHighPerfRendering(t *testing.T) {
	// This test simulates the actual runtime environment
	// where HighPerformanceRendering may be enabled
	config = Config{
		HighPerformancePager: true,
		ShowOutline:          true,
	}

	common := &commonModel{
		cfg:    config,
		width:  100,
		height: 25,
	}

	m := newPagerModel(common)

	t.Logf("HighPerformanceRendering: %v", m.viewport.HighPerformanceRendering)

	m.currentDocument = markdown{
		Note: "test.md",
		Body: "# Title\n## Section 1\nContent here",
	}

	// Enable outline
	m.showOutline = true
	m.setSize(common.width, common.height)

	// Parse headings
	m.outline.setContent(m.currentDocument.Body)

	// Set viewport content
	m.viewport.SetContent("Line 1\nLine 2\nLine 3\nLine 4\nLine 5")

	// Get pager view
	view := m.View()

	t.Logf("View length: %d", len(view))
	t.Logf("Pager view:\n%s", view)

	// View should contain outline
	if !strings.Contains(view, "OUTLINE") {
		t.Error("Pager view should contain 'OUTLINE' with HighPerformanceRendering")
	}
}
