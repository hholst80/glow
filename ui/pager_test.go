package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

// newTestPagerModel creates a pagerModel configured for testing.
// It uses a TestMarkdownRenderer and TestTerminal and sets up minimal state.
func newTestPagerModel() pagerModel {
	renderer := &TestMarkdownRenderer{
		RenderFunc: func(markdown string, width int, style string, filename string, preserveNewLines bool) (string, error) {
			// Return markdown as-is for predictable test behavior
			return markdown, nil
		},
	}

	terminal := &TestTerminal{
		DarkBackground: true,
	}

	common := &commonModel{
		cfg: Config{
			GlamourEnabled:   true,
			GlamourStyle:     "dark",
			GlamourMaxWidth:  80,
			ShowLineNumbers:  false,
			PreserveNewLines: false,
		},
		terminal: terminal,
		renderer: renderer,
		width:    80,
		height:   24,
	}

	vp := viewport.New(80, 20)
	vp.YPosition = 0

	return pagerModel{
		common:   common,
		state:    pagerStateBrowse,
		viewport: vp,
		outline:  newOutlineModel(common),
		currentDocument: markdown{
			Note: "test.md",
			Body: "# Test\n\nContent",
		},
	}
}

// TestPagerUpdate_NavigationKeys tests viewport navigation keys.
func TestPagerUpdate_NavigationKeys(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		setup    func(*pagerModel)
		validate func(*testing.T, pagerModel)
	}{
		{
			name: "home goes to top",
			key:  "home",
			setup: func(m *pagerModel) {
				m.viewport.SetContent(strings.Repeat("line\n", 100))
				m.viewport.YOffset = 50
			},
			validate: func(t *testing.T, m pagerModel) {
				if m.viewport.YOffset != 0 {
					t.Errorf("expected YOffset=0, got %d", m.viewport.YOffset)
				}
			},
		},
		{
			name: "g goes to top",
			key:  "g",
			setup: func(m *pagerModel) {
				m.viewport.SetContent(strings.Repeat("line\n", 100))
				m.viewport.YOffset = 50
			},
			validate: func(t *testing.T, m pagerModel) {
				if m.viewport.YOffset != 0 {
					t.Errorf("expected YOffset=0, got %d", m.viewport.YOffset)
				}
			},
		},
		{
			name: "end goes to bottom",
			key:  "end",
			setup: func(m *pagerModel) {
				m.viewport.SetContent(strings.Repeat("line\n", 100))
				m.viewport.YOffset = 0
			},
			validate: func(t *testing.T, m pagerModel) {
				// Should be at or near bottom
				maxOffset := m.viewport.TotalLineCount() - m.viewport.Height
				if maxOffset < 0 {
					maxOffset = 0
				}
				if m.viewport.YOffset != maxOffset {
					t.Errorf("expected YOffset=%d, got %d", maxOffset, m.viewport.YOffset)
				}
			},
		},
		{
			name: "G goes to bottom",
			key:  "G",
			setup: func(m *pagerModel) {
				m.viewport.SetContent(strings.Repeat("line\n", 100))
				m.viewport.YOffset = 0
			},
			validate: func(t *testing.T, m pagerModel) {
				maxOffset := m.viewport.TotalLineCount() - m.viewport.Height
				if maxOffset < 0 {
					maxOffset = 0
				}
				if m.viewport.YOffset != maxOffset {
					t.Errorf("expected YOffset=%d, got %d", maxOffset, m.viewport.YOffset)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newTestPagerModel()
			if tt.setup != nil {
				tt.setup(&m)
			}

			msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)}
			if tt.key == "home" {
				msg = tea.KeyMsg{Type: tea.KeyHome}
			} else if tt.key == "end" {
				msg = tea.KeyMsg{Type: tea.KeyEnd}
			}

			newM, _ := m.update(msg)

			if tt.validate != nil {
				tt.validate(t, newM)
			}
		})
	}
}

// TestPagerUpdate_OutlineToggle tests the 'o' key for toggling outline.
func TestPagerUpdate_OutlineToggle(t *testing.T) {
	m := newTestPagerModel()
	m.currentDocument.Note = "test.md" // Ensure it's a markdown file
	m.showOutline = false

	// Toggle outline on
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("o")}
	newM, _ := m.update(msg)

	if !newM.showOutline {
		t.Error("expected showOutline=true after pressing 'o'")
	}

	// Toggle outline off
	newM, _ = newM.update(msg)
	if newM.showOutline {
		t.Error("expected showOutline=false after pressing 'o' again")
	}
}

// TestPagerUpdate_OutlineToggle_NonMarkdown tests that outline toggle is ignored for non-markdown files.
func TestPagerUpdate_OutlineToggle_NonMarkdown(t *testing.T) {
	m := newTestPagerModel()
	m.currentDocument.Note = "test.go" // Non-markdown file
	m.showOutline = false

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("o")}
	newM, _ := m.update(msg)

	if newM.showOutline {
		t.Error("expected outline toggle to be ignored for non-markdown files")
	}
}

// TestPagerUpdate_OutlineFocusSwitch tests the Tab key for switching focus.
func TestPagerUpdate_OutlineFocusSwitch(t *testing.T) {
	m := newTestPagerModel()
	m.showOutline = true
	m.outline.visible = true
	m.outlineFocused = false

	// Press Tab to focus outline
	msg := tea.KeyMsg{Type: tea.KeyTab}
	newM, _ := m.update(msg)

	if !newM.outlineFocused {
		t.Error("expected outlineFocused=true after pressing Tab")
	}
	if !newM.outline.focused {
		t.Error("expected outline.focused=true after pressing Tab")
	}

	// Press Tab again to unfocus
	newM, _ = newM.update(msg)
	if newM.outlineFocused {
		t.Error("expected outlineFocused=false after pressing Tab again")
	}
}

// TestPagerUpdate_HeadingNavigation tests bracket keys for heading navigation.
func TestPagerUpdate_HeadingNavigation(t *testing.T) {
	m := newTestPagerModel()
	m.showOutline = true
	m.outline.headings = []Heading{
		{Level: 1, Text: "First", Line: 0, RenderedLine: 0},
		{Level: 2, Text: "Second", Line: 5, RenderedLine: 5},
		{Level: 2, Text: "Third", Line: 10, RenderedLine: 10},
	}
	m.outline.current = 0
	m.viewport.SetContent(strings.Repeat("line\n", 100))

	// Press ] to go to next heading
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("]")}
	newM, _ := m.update(msg)

	if newM.outline.current != 1 {
		t.Errorf("expected outline.current=1, got %d", newM.outline.current)
	}

	// Press [ to go to previous heading
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("[")}
	newM, _ = newM.update(msg)

	if newM.outline.current != 0 {
		t.Errorf("expected outline.current=0, got %d", newM.outline.current)
	}
}

// TestPagerUpdate_ContentRenderedMsg tests content rendering message handling.
func TestPagerUpdate_ContentRenderedMsg(t *testing.T) {
	m := newTestPagerModel()
	m.currentDocument.Body = "# Test Heading\n\nSome content."

	content := "rendered content\nline 2\nline 3"
	msg := contentRenderedMsg(content)
	newM, _ := m.update(msg)

	// Viewport should have the rendered content
	if !strings.Contains(newM.viewport.View(), "rendered content") {
		t.Error("expected viewport to contain rendered content")
	}
}

// TestPagerUpdate_WindowSizeMsg tests window resize handling.
func TestPagerUpdate_WindowSizeMsg(t *testing.T) {
	m := newTestPagerModel()
	m.currentDocument.Body = "# Test\n\nContent"

	msg := tea.WindowSizeMsg{Width: 120, Height: 40}
	_, cmd := m.update(msg)

	// Should return a command to re-render content
	if cmd == nil {
		t.Error("expected a command to be returned for window resize")
	}
}

// TestPagerUpdate_QuitFromBrowse tests q key in browse state.
func TestPagerUpdate_QuitFromBrowse(t *testing.T) {
	m := newTestPagerModel()
	m.state = pagerStateBrowse

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")}
	newM, _ := m.update(msg)

	// In pager, 'q' should just return to browse state (quit is handled at model level)
	if newM.state != pagerStateBrowse {
		t.Errorf("expected state to remain pagerStateBrowse, got %v", newM.state)
	}
}

// TestPagerUpdate_EscFromStatusMessage tests Esc key from status message state.
func TestPagerUpdate_EscFromStatusMessage(t *testing.T) {
	m := newTestPagerModel()
	m.state = pagerStateStatusMessage

	msg := tea.KeyMsg{Type: tea.KeyEsc}
	newM, _ := m.update(msg)

	if newM.state != pagerStateBrowse {
		t.Errorf("expected state=pagerStateBrowse, got %v", newM.state)
	}
}

// TestPagerUpdate_HalfPageNavigation tests d and u keys for half-page scrolling.
func TestPagerUpdate_HalfPageNavigation(t *testing.T) {
	m := newTestPagerModel()
	m.viewport.SetContent(strings.Repeat("line\n", 100))
	m.viewport.YOffset = 20

	// Press d to scroll down half page
	initialOffset := m.viewport.YOffset
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")}
	newM, _ := m.update(msg)

	if newM.viewport.YOffset <= initialOffset {
		t.Error("expected YOffset to increase after pressing 'd'")
	}

	// Press u to scroll up half page
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("u")}
	newM, _ = newM.update(msg)

	// Should be back near initial position
	if newM.viewport.YOffset >= newM.viewport.TotalLineCount()-newM.viewport.Height {
		t.Error("expected YOffset to decrease after pressing 'u'")
	}
}

// TestPagerUpdate_HelpToggle tests ? key for help toggle.
func TestPagerUpdate_HelpToggle(t *testing.T) {
	m := newTestPagerModel()
	m.showHelp = false

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")}
	newM, _ := m.update(msg)

	if !newM.showHelp {
		t.Error("expected showHelp=true after pressing '?'")
	}

	// Toggle off
	newM, _ = newM.update(msg)
	if newM.showHelp {
		t.Error("expected showHelp=false after pressing '?' again")
	}
}

// TestPagerUpdate_EnterOnFocusedOutline tests Enter key when outline is focused.
func TestPagerUpdate_EnterOnFocusedOutline(t *testing.T) {
	m := newTestPagerModel()
	m.showOutline = true
	m.outline.visible = true
	m.outlineFocused = true
	m.outline.headings = []Heading{
		{Level: 1, Text: "First", Line: 0, RenderedLine: 0},
		{Level: 2, Text: "Second", Line: 10, RenderedLine: 10},
	}
	m.outline.cursor = 1
	m.viewport.SetContent(strings.Repeat("line\n", 100))
	m.viewport.YOffset = 0

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	newM, _ := m.update(msg)

	// Should jump to the cursor position
	if newM.outline.current != 1 {
		t.Errorf("expected outline.current=1, got %d", newM.outline.current)
	}
}

// TestPagerUpdate_CursorMovement tests j/k keys when outline is focused.
func TestPagerUpdate_CursorMovement(t *testing.T) {
	m := newTestPagerModel()
	m.showOutline = true
	m.outline.visible = true
	m.outlineFocused = true
	m.outline.headings = []Heading{
		{Level: 1, Text: "First", Line: 0},
		{Level: 2, Text: "Second", Line: 5},
		{Level: 2, Text: "Third", Line: 10},
	}
	m.outline.cursor = 0

	// Press j to move down
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")}
	newM, _ := m.update(msg)

	if newM.outline.cursor != 1 {
		t.Errorf("expected cursor=1, got %d", newM.outline.cursor)
	}

	// Press k to move up
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")}
	newM, _ = newM.update(msg)

	if newM.outline.cursor != 0 {
		t.Errorf("expected cursor=0, got %d", newM.outline.cursor)
	}
}

// TestGlamourRender_GlamourDisabled tests glamourRender when glamour is disabled.
func TestGlamourRender_GlamourDisabled(t *testing.T) {
	m := newTestPagerModel()
	m.common.cfg.GlamourEnabled = false

	input := "# Test\n\nSome content"
	out, err := glamourRender(m, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// When glamour is disabled, output should equal input
	if out != input {
		t.Errorf("expected output=%q, got %q", input, out)
	}
}

// TestGlamourRender_WithLineNumbers tests glamourRender with line numbers enabled.
func TestGlamourRender_WithLineNumbers(t *testing.T) {
	m := newTestPagerModel()
	m.common.cfg.GlamourEnabled = true
	m.common.cfg.ShowLineNumbers = true
	m.viewport.Width = 80

	input := "# Test\n\nLine 1\nLine 2"
	out, err := glamourRender(m, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Output should contain line numbers
	lines := strings.Split(out, "\n")
	if len(lines) == 0 {
		t.Fatal("expected non-empty output")
	}

	// First line should start with a number (line number)
	firstLine := lines[0]
	// Line numbers are right-padded to lineNumberWidth (4)
	if len(firstLine) < 4 {
		t.Errorf("expected line to be at least 4 chars, got %d", len(firstLine))
	}
}

// TestGlamourRender_CodeFile tests glamourRender with a code file.
func TestGlamourRender_CodeFile(t *testing.T) {
	m := newTestPagerModel()
	m.common.cfg.GlamourEnabled = true
	m.currentDocument.Note = "test.go"
	m.viewport.Width = 80

	input := "package main\n\nfunc main() {}"
	out, err := glamourRender(m, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Code files should always have line numbers
	if out == "" {
		t.Error("expected non-empty output for code file")
	}
}

// TestGlamourRender_UsesInjectedRenderer verifies the renderer is called.
func TestGlamourRender_UsesInjectedRenderer(t *testing.T) {
	renderer := &TestMarkdownRenderer{}
	terminal := &TestTerminal{DarkBackground: true}
	common := &commonModel{
		cfg: Config{
			GlamourEnabled:  true,
			GlamourStyle:    "dark",
			GlamourMaxWidth: 80,
		},
		terminal: terminal,
		renderer: renderer,
		width:    80,
		height:   24,
	}

	vp := viewport.New(80, 20)
	m := pagerModel{
		common:   common,
		viewport: vp,
		currentDocument: markdown{
			Note: "test.md",
		},
	}

	input := "# Test\n\nContent"
	_, _ = glamourRender(m, input)

	// Verify the renderer was called
	if len(renderer.RenderCalls) != 1 {
		t.Errorf("expected 1 render call, got %d", len(renderer.RenderCalls))
	}

	call := renderer.RenderCalls[0]
	if call.Markdown != input {
		t.Errorf("expected markdown=%q, got %q", input, call.Markdown)
	}
	if call.Style != "dark" {
		t.Errorf("expected style='dark', got %q", call.Style)
	}
	if call.Filename != "test.md" {
		t.Errorf("expected filename='test.md', got %q", call.Filename)
	}
}
