package ui

import (
	"regexp"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/truncate"
)

const (
	outlineMinWidth     = 20
	outlineMaxWidth     = 40
	outlineWidthPercent = 25
	minTerminalWidth    = 80
)

// Heading represents a markdown heading extracted from the document.
type Heading struct {
	Level int    // 1-6 for # through ######
	Text  string // The heading text (without # prefix)
	Line  int    // Line number in raw markdown (0-indexed)
}

// outlineModel manages the outline sidebar state.
type outlineModel struct {
	common   *commonModel
	headings []Heading
	cursor   int            // Currently selected heading (when focused)
	current  int            // Current heading based on scroll position
	width    int            // Sidebar width
	height   int            // Available height
	visible  bool           // Whether outline is shown
	focused  bool           // Whether outline has keyboard focus
	viewport viewport.Model // For scrollable outline when many headings
}

// Regex patterns for heading extraction.
var (
	// Match ATX headings: # through ######
	headingRegex = regexp.MustCompile(`(?m)^(#{1,6})\s+(.+?)(?:\s+#+)?$`)
)

func newOutlineModel(common *commonModel) outlineModel {
	vp := viewport.New(0, 0)
	return outlineModel{
		common:   common,
		viewport: vp,
		visible:  false,
		focused:  false,
	}
}

// parseHeadings extracts headings from raw markdown content.
func parseHeadings(markdown string) []Heading {
	var headings []Heading
	lines := strings.Split(markdown, "\n")

	// Track which lines are in code blocks
	inCodeBlock := false
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
			inCodeBlock = !inCodeBlock
			continue
		}
		if inCodeBlock {
			continue
		}

		// Check for heading
		match := headingRegex.FindStringSubmatch(line)
		if match != nil {
			level := len(match[1])
			text := strings.TrimSpace(match[2])
			headings = append(headings, Heading{
				Level: level,
				Text:  text,
				Line:  i,
			})
		}
	}

	return headings
}

// setContent updates the outline with new markdown content.
func (m *outlineModel) setContent(markdown string) {
	m.headings = parseHeadings(markdown)
	m.cursor = 0
	m.current = 0
	m.updateViewport()
}

// setSize updates the outline dimensions.
func (m *outlineModel) setSize(width, height int) {
	m.width = width
	m.height = height
	m.viewport.Width = width
	m.viewport.Height = height
	m.updateViewport()
}

// calculateOutlineWidth returns the appropriate outline width for a given terminal width.
func calculateOutlineWidth(termWidth int) int {
	if termWidth < minTerminalWidth {
		return 0
	}
	width := termWidth * outlineWidthPercent / 100
	if width < outlineMinWidth {
		width = outlineMinWidth
	}
	if width > outlineMaxWidth {
		width = outlineMaxWidth
	}
	return width
}

// updateViewport refreshes the viewport content.
func (m *outlineModel) updateViewport() {
	if len(m.headings) == 0 {
		m.viewport.SetContent("")
		return
	}

	var b strings.Builder
	for i, h := range m.headings {
		line := m.renderHeadingLine(i, h)
		b.WriteString(line)
		if i < len(m.headings)-1 {
			b.WriteString("\n")
		}
	}
	m.viewport.SetContent(b.String())
}

// renderHeadingLine renders a single heading line with appropriate styling.
func (m *outlineModel) renderHeadingLine(index int, h Heading) string {
	// Indentation based on heading level
	indent := strings.Repeat("  ", h.Level-1)

	// Prefix indicator
	prefix := "  "
	if index == m.current {
		prefix = "> "
	}

	// Calculate available width for text
	availWidth := m.width - len(indent) - len(prefix) - 2 // -2 for padding
	if availWidth < 5 {
		availWidth = 5
	}

	// Truncate text if needed
	text := truncate.StringWithTail(h.Text, uint(availWidth), "…")

	// Full line content
	content := indent + prefix + text

	// Pad to full width
	if len(content) < m.width {
		content += strings.Repeat(" ", m.width-len(content))
	}

	// Apply styling
	if m.focused && index == m.cursor {
		return outlineCursorStyle.Width(m.width).Render(content)
	} else if index == m.current {
		return outlineCurrentStyle.Width(m.width).Render(content)
	}
	return outlineNormalStyle.Width(m.width).Render(content)
}

// moveCursorUp moves the cursor up in the heading list.
func (m *outlineModel) moveCursorUp() {
	if m.cursor > 0 {
		m.cursor--
		m.ensureCursorVisible()
	}
}

// moveCursorDown moves the cursor down in the heading list.
func (m *outlineModel) moveCursorDown() {
	if m.cursor < len(m.headings)-1 {
		m.cursor++
		m.ensureCursorVisible()
	}
}

// ensureCursorVisible scrolls the viewport to keep cursor in view.
func (m *outlineModel) ensureCursorVisible() {
	if m.cursor < m.viewport.YOffset {
		m.viewport.YOffset = m.cursor
	} else if m.cursor >= m.viewport.YOffset+m.height {
		m.viewport.YOffset = m.cursor - m.height + 1
	}
	m.updateViewport()
}

// ensureCurrentVisible scrolls the viewport to keep current heading in view.
func (m *outlineModel) ensureCurrentVisible() {
	if m.current < m.viewport.YOffset {
		m.viewport.YOffset = m.current
	} else if m.current >= m.viewport.YOffset+m.height {
		m.viewport.YOffset = m.current - m.height + 1
	}
}

// updateCurrent sets the current heading based on line number.
func (m *outlineModel) updateCurrent(lineNum int) {
	if len(m.headings) == 0 {
		return
	}

	// Find the heading that corresponds to the current position
	newCurrent := 0
	for i, h := range m.headings {
		if h.Line <= lineNum {
			newCurrent = i
		} else {
			break
		}
	}

	if newCurrent != m.current {
		m.current = newCurrent
		m.ensureCurrentVisible()
		m.updateViewport()
	}
}

// selectedHeading returns the currently selected heading, or nil if none.
func (m *outlineModel) selectedHeading() *Heading {
	if m.cursor >= 0 && m.cursor < len(m.headings) {
		return &m.headings[m.cursor]
	}
	return nil
}

// nextHeadingIndex returns the index of the next heading, or -1 if at end.
func (m *outlineModel) nextHeadingIndex() int {
	if m.current < len(m.headings)-1 {
		return m.current + 1
	}
	return -1
}

// prevHeadingIndex returns the index of the previous heading, or -1 if at start.
func (m *outlineModel) prevHeadingIndex() int {
	if m.current > 0 {
		return m.current - 1
	}
	return -1
}

// update handles messages for the outline model.
func (m outlineModel) update(msg tea.Msg) (outlineModel, tea.Cmd) {
	if !m.visible || !m.focused {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			m.moveCursorDown()
			return m, nil
		case "k", "up":
			m.moveCursorUp()
			return m, nil
		case "g", "home":
			m.cursor = 0
			m.ensureCursorVisible()
			return m, nil
		case "G", "end":
			if len(m.headings) > 0 {
				m.cursor = len(m.headings) - 1
				m.ensureCursorVisible()
			}
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

// View renders the outline sidebar.
func (m outlineModel) View() string {
	if !m.visible || len(m.headings) == 0 {
		return ""
	}

	// Title
	title := outlineTitleStyle.Width(m.width).Render("OUTLINE")

	// Build content
	var lines []string
	lines = append(lines, title)

	for i, h := range m.headings {
		if i >= m.height-1 { // -1 for title
			break
		}
		lines = append(lines, m.renderHeadingLine(i, h))
	}

	content := strings.Join(lines, "\n")

	// Apply panel styling with left border
	return outlinePanelStyle.Height(m.height).Render(content)
}

// Outline styles.
var (
	outlineBorderColor = lipgloss.AdaptiveColor{Light: "#DCDCDC", Dark: "#3C3C3C"}

	outlineTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(fuchsia).
				Padding(0, 1)

	outlineNormalStyle = lipgloss.NewStyle().
				Foreground(gray)

	outlineCurrentStyle = lipgloss.NewStyle().
				Foreground(yellowGreen).
				Bold(true)

	outlineCursorStyle = lipgloss.NewStyle().
				Background(darkGray).
				Foreground(cream)

	outlinePanelStyle = lipgloss.NewStyle().
				BorderLeft(true).
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(outlineBorderColor)
)
