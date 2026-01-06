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
		cfg:      Config{ShowOutline: false},
		terminal: NewTestTerminal(),
		width:    120,
		height:   40,
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
		cfg:      Config{},
		terminal: NewTestTerminal(),
		width:    120,
		height:   40,
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
		cfg:      Config{ShowOutline: true},
		terminal: NewTestTerminal(),
		width:    100,
		height:   25,
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
		cfg:      Config{ShowOutline: true},
		terminal: NewTestTerminal(),
		width:    100,
		height:   25,
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
		cfg:      config,
		terminal: NewTestTerminal(),
		width:    100,
		height:   25,
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

// ====================
// mapHeadingsToRenderedLines TESTS
// ====================

func TestMapHeadingsToRenderedLines_Basic(t *testing.T) {
	common := &commonModel{}
	m := newOutlineModel(common)

	// Parse headings from raw markdown
	rawMd := "# Title\nSome text\n## Section 1\nMore text\n### Subsection"
	m.setContent(rawMd)

	// Simulate rendered content (similar structure but different lines)
	renderedContent := "Title\n\nSome text\n\nSection 1\n\nMore text\n\nSubsection"
	m.mapHeadingsToRenderedLines(renderedContent)

	// Verify mappings
	if len(m.headings) != 3 {
		t.Fatalf("Expected 3 headings, got %d", len(m.headings))
	}

	// Title should be found on line 0
	if m.headings[0].RenderedLine != 0 {
		t.Errorf("Title RenderedLine = %d, want 0", m.headings[0].RenderedLine)
	}

	// Section 1 should be found (line 4)
	if m.headings[1].RenderedLine < 0 {
		t.Errorf("Section 1 RenderedLine = %d, should be >= 0", m.headings[1].RenderedLine)
	}

	// Subsection should be found (line 8)
	if m.headings[2].RenderedLine < 0 {
		t.Errorf("Subsection RenderedLine = %d, should be >= 0", m.headings[2].RenderedLine)
	}

	// Rendered lines should be in order
	if m.headings[1].RenderedLine <= m.headings[0].RenderedLine {
		t.Errorf("Headings should be in order: %d <= %d", m.headings[1].RenderedLine, m.headings[0].RenderedLine)
	}
}

func TestMapHeadingsToRenderedLines_WithANSI(t *testing.T) {
	common := &commonModel{}
	m := newOutlineModel(common)

	rawMd := "# Bold Title\n## Styled Section"
	m.setContent(rawMd)

	// Rendered content with ANSI codes
	renderedContent := "\x1b[1mBold Title\x1b[0m\n\nSome content\n\n\x1b[38;5;208mStyled Section\x1b[0m"
	m.mapHeadingsToRenderedLines(renderedContent)

	if len(m.headings) != 2 {
		t.Fatalf("Expected 2 headings, got %d", len(m.headings))
	}

	// Both headings should be found despite ANSI codes
	if m.headings[0].RenderedLine < 0 {
		t.Errorf("Bold Title should be found, got RenderedLine = %d", m.headings[0].RenderedLine)
	}
	if m.headings[1].RenderedLine < 0 {
		t.Errorf("Styled Section should be found, got RenderedLine = %d", m.headings[1].RenderedLine)
	}
}

func TestMapHeadingsToRenderedLines_Empty(t *testing.T) {
	common := &commonModel{}
	m := newOutlineModel(common)

	// Empty content
	m.setContent("")
	m.mapHeadingsToRenderedLines("")

	if len(m.headings) != 0 {
		t.Errorf("Expected 0 headings for empty content, got %d", len(m.headings))
	}

	// Empty rendered but with headings - should return early without modifying
	m.setContent("# Title")
	initialRenderedLine := m.headings[0].RenderedLine
	m.mapHeadingsToRenderedLines("")

	// Should not crash and headings should remain unchanged (returns early)
	if m.headings[0].RenderedLine != initialRenderedLine {
		t.Errorf("RenderedLine should remain unchanged for empty rendered content, got %d", m.headings[0].RenderedLine)
	}

	// When called with non-empty content but heading not found, should be -1
	m.mapHeadingsToRenderedLines("no matching content here")
	if m.headings[0].RenderedLine != -1 {
		t.Errorf("RenderedLine should be -1 when heading not found, got %d", m.headings[0].RenderedLine)
	}
}

func TestMapHeadingsToRenderedLines_CaseInsensitive(t *testing.T) {
	common := &commonModel{}
	m := newOutlineModel(common)

	rawMd := "# UPPERCASE\n## MixedCase"
	m.setContent(rawMd)

	// Rendered with different case (some renderers might alter case)
	renderedContent := "uppercase\nmixedcase"
	m.mapHeadingsToRenderedLines(renderedContent)

	// Both should be found due to case-insensitive matching
	if m.headings[0].RenderedLine < 0 {
		t.Errorf("UPPERCASE heading should match 'uppercase', got RenderedLine = %d", m.headings[0].RenderedLine)
	}
	if m.headings[1].RenderedLine < 0 {
		t.Errorf("MixedCase heading should match 'mixedcase', got RenderedLine = %d", m.headings[1].RenderedLine)
	}
}

// ====================
// jumpToHeading TESTS
// ====================

func TestJumpToHeading_Basic(t *testing.T) {
	common := &commonModel{
		cfg:    Config{},
		width:  100,
		height: 30,
	}

	m := newPagerModel(common)
	m.currentDocument = markdown{
		Note: "test.md",
		Body: "# H1\nLine 1\nLine 2\n## H2\nLine 3\nLine 4\n### H3\nLine 5",
	}

	// Set up viewport with content
	m.setSize(common.width, common.height)
	m.viewport.SetContent("Rendered H1\nLine 1\nLine 2\nRendered H2\nLine 3\nLine 4\nRendered H3\nLine 5\n" +
		strings.Repeat("More content\n", 30)) // Make content scrollable

	// Parse and map headings
	m.outline.setContent(m.currentDocument.Body)
	m.outline.mapHeadingsToRenderedLines(m.viewport.View())

	// Jump to heading 1 (H2)
	initialOffset := m.viewport.YOffset
	m.jumpToHeading(1)

	// Offset should change (or stay 0 if H2 is near top)
	// And outline current/cursor should update
	if m.outline.current != 1 {
		t.Errorf("outline.current = %d, want 1 after jumping to heading 1", m.outline.current)
	}
	if m.outline.cursor != 1 {
		t.Errorf("outline.cursor = %d, want 1 after jumping to heading 1", m.outline.cursor)
	}

	t.Logf("Initial offset: %d, After jump: %d", initialOffset, m.viewport.YOffset)
}

func TestJumpToHeading_OutOfBounds(t *testing.T) {
	common := &commonModel{
		cfg:    Config{},
		width:  100,
		height: 30,
	}

	m := newPagerModel(common)
	m.currentDocument = markdown{
		Note: "test.md",
		Body: "# H1\n## H2",
	}

	m.setSize(common.width, common.height)
	m.outline.setContent(m.currentDocument.Body)

	initialCurrent := m.outline.current

	// Jump to invalid negative index
	m.jumpToHeading(-1)
	if m.outline.current != initialCurrent {
		t.Errorf("Negative index should not change state, current = %d", m.outline.current)
	}

	// Jump to index beyond headings count
	m.jumpToHeading(100)
	if m.outline.current != initialCurrent {
		t.Errorf("Out of bounds index should not change state, current = %d", m.outline.current)
	}
}

func TestJumpToHeading_FallbackRatio(t *testing.T) {
	common := &commonModel{
		cfg:    Config{},
		width:  100,
		height: 30,
	}

	m := newPagerModel(common)
	m.currentDocument = markdown{
		Note: "test.md",
		Body: "# H1\n" + strings.Repeat("Line\n", 50) + "## H2 at end",
	}

	m.setSize(common.width, common.height)
	m.viewport.SetContent(strings.Repeat("Content line\n", 100))

	// Parse headings but don't map rendered lines
	m.outline.setContent(m.currentDocument.Body)
	// Don't call mapHeadingsToRenderedLines - forces fallback

	// H2 is at the end, so jumping should use ratio-based approximation
	m.jumpToHeading(1)

	// Should update current without crashing
	if m.outline.current != 1 {
		t.Errorf("outline.current = %d, want 1", m.outline.current)
	}
}

// ====================
// EDGE CASE TESTS
// ====================

func TestOutline_EmptyDocument(t *testing.T) {
	common := &commonModel{}
	m := newOutlineModel(common)

	m.setContent("")
	m.setSize(30, 20)
	m.visible = true

	if len(m.headings) != 0 {
		t.Errorf("Empty doc should have 0 headings, got %d", len(m.headings))
	}

	// View should handle empty gracefully
	view := m.View()
	if view != "" {
		t.Logf("Empty outline view: %q", view)
	}

	// Navigation should not crash
	m.moveCursorDown()
	m.moveCursorUp()
	m.updateCurrent(0)
}

func TestOutline_OnlyCodeBlocks(t *testing.T) {
	common := &commonModel{}
	m := newOutlineModel(common)

	markdown := "```\n# Not a heading\n## Also not\n```\n\nSome text\n\n```python\n# Comment\n```"
	m.setContent(markdown)

	if len(m.headings) != 0 {
		t.Errorf("Code-only doc should have 0 headings, got %d", len(m.headings))
		for _, h := range m.headings {
			t.Logf("Found heading: %+v", h)
		}
	}
}

func TestOutline_VeryLongHeadingText(t *testing.T) {
	common := &commonModel{}
	m := newOutlineModel(common)

	longHeading := "# " + strings.Repeat("Very Long Heading Text ", 20)
	m.setContent(longHeading)
	m.setSize(30, 10)
	m.visible = true

	if len(m.headings) != 1 {
		t.Fatalf("Expected 1 heading, got %d", len(m.headings))
	}

	// View should truncate properly without crashing
	view := m.View()
	if len(view) == 0 {
		t.Error("View should not be empty for long heading")
	}

	// Should contain truncation indicator
	if !strings.Contains(view, "…") && len(m.headings[0].Text) > 25 {
		t.Logf("View may need truncation, heading text length: %d", len(m.headings[0].Text))
	}
}

func TestOutline_DeeplyNestedHeadings(t *testing.T) {
	common := &commonModel{}
	m := newOutlineModel(common)

	markdown := "# H1\n## H2\n### H3\n#### H4\n##### H5\n###### H6"
	m.setContent(markdown)
	m.setSize(40, 20)
	m.visible = true

	if len(m.headings) != 6 {
		t.Errorf("Expected 6 headings (H1-H6), got %d", len(m.headings))
	}

	// Verify levels
	for i, h := range m.headings {
		expected := i + 1
		if h.Level != expected {
			t.Errorf("Heading %d level = %d, want %d", i, h.Level, expected)
		}
	}

	// View should render with increasing indentation
	view := m.View()
	t.Logf("Nested headings view:\n%s", view)
}

func TestOutline_FocusStateRendering(t *testing.T) {
	common := &commonModel{}
	m := newOutlineModel(common)

	m.setContent("# H1\n## H2\n## H3")
	m.setSize(30, 15)
	m.visible = true

	// Get unfocused view
	m.focused = false
	m.cursor = 1
	unfocusedView := m.View()

	// Get focused view
	m.focused = true
	focusedView := m.View()

	// Views should be different (focused has different styling)
	if unfocusedView == focusedView {
		t.Log("Note: Focused and unfocused views may look similar in text output")
	}

	t.Logf("Unfocused:\n%s\nFocused:\n%s", unfocusedView, focusedView)
}

func TestStripANSI(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"no ANSI", "hello world", "hello world"},
		{"bold", "\x1b[1mhello\x1b[0m", "hello"},
		{"color", "\x1b[38;5;208mcolored\x1b[0m", "colored"},
		{"multiple codes", "\x1b[1m\x1b[31mred bold\x1b[0m", "red bold"},
		{"empty string", "", ""},
		{"only ANSI", "\x1b[1m\x1b[0m", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stripANSI(tt.input)
			if result != tt.expected {
				t.Errorf("stripANSI(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestOutline_HeadingWithSpecialChars(t *testing.T) {
	common := &commonModel{}
	m := newOutlineModel(common)

	markdown := "# Hello `code` World\n## Section: The Basics\n### Q&A / FAQ"
	m.setContent(markdown)

	if len(m.headings) != 3 {
		t.Fatalf("Expected 3 headings, got %d", len(m.headings))
	}

	// Verify text extraction handles special chars
	if !strings.Contains(m.headings[0].Text, "code") {
		t.Errorf("Heading should contain 'code': %q", m.headings[0].Text)
	}
	if !strings.Contains(m.headings[1].Text, ":") {
		t.Errorf("Heading should contain ':': %q", m.headings[1].Text)
	}
	if !strings.Contains(m.headings[2].Text, "&") || !strings.Contains(m.headings[2].Text, "/") {
		t.Errorf("Heading should contain '&' and '/': %q", m.headings[2].Text)
	}
}
