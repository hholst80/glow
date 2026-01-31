package ui

import (
	"fmt"
	"math"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/hholst80/glow/utils"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/fsnotify/fsnotify"
	runewidth "github.com/mattn/go-runewidth"
	"github.com/muesli/reflow/ansi"
	"github.com/muesli/reflow/truncate"
)

const (
	statusBarHeight = 1
	lineNumberWidth = 4
)

var (
	pagerHelpHeight int

	mintGreen = lipgloss.AdaptiveColor{Light: "#89F0CB", Dark: "#89F0CB"}
	darkGreen = lipgloss.AdaptiveColor{Light: "#1C8760", Dark: "#1C8760"}

	lineNumberFg = lipgloss.AdaptiveColor{Light: "#656565", Dark: "#7D7D7D"}

	statusBarNoteFg = lipgloss.AdaptiveColor{Light: "#656565", Dark: "#7D7D7D"}
	statusBarBg     = lipgloss.AdaptiveColor{Light: "#E6E6E6", Dark: "#242424"}

	statusBarScrollPosStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#949494", Dark: "#5A5A5A"}).
				Background(statusBarBg).
				Render

	statusBarNoteStyle = lipgloss.NewStyle().
				Foreground(statusBarNoteFg).
				Background(statusBarBg).
				Render

	statusBarHelpStyle = lipgloss.NewStyle().
				Foreground(statusBarNoteFg).
				Background(lipgloss.AdaptiveColor{Light: "#DCDCDC", Dark: "#323232"}).
				Render

	statusBarMessageStyle = lipgloss.NewStyle().
				Foreground(mintGreen).
				Background(darkGreen).
				Render

	statusBarMessageScrollPosStyle = lipgloss.NewStyle().
					Foreground(mintGreen).
					Background(darkGreen).
					Render

	statusBarMessageHelpStyle = lipgloss.NewStyle().
					Foreground(lipgloss.Color("#B6FFE4")).
					Background(green).
					Render

	helpViewStyle = lipgloss.NewStyle().
			Foreground(statusBarNoteFg).
			Background(lipgloss.AdaptiveColor{Light: "#f2f2f2", Dark: "#1B1B1B"}).
			Render

	lineNumberStyle = lipgloss.NewStyle().
			Foreground(lineNumberFg).
			Render
)

type (
	contentRenderedMsg string
	reloadMsg          struct{}
)

type pagerState int

const (
	pagerStateBrowse pagerState = iota
	pagerStateStatusMessage
)

type pagerModel struct {
	common   *commonModel
	viewport viewport.Model
	state    pagerState
	showHelp bool

	statusMessage      string
	statusMessageTimer *time.Timer

	// Current document being rendered, sans-glamour rendering. We cache
	// it here so we can re-render it on resize.
	currentDocument markdown

	watcher *fsnotify.Watcher

	// Outline sidebar
	outline        outlineModel
	showOutline    bool
	outlineFocused bool
}

func newPagerModel(common *commonModel) pagerModel {
	// Init viewport
	vp := viewport.New(0, 0)
	vp.YPosition = 0
	vp.HighPerformanceRendering = config.HighPerformancePager

	m := pagerModel{
		common:      common,
		state:       pagerStateBrowse,
		viewport:    vp,
		outline:     newOutlineModel(common),
		showOutline: common.cfg.ShowOutline,
	}
	m.initWatcher()
	return m
}

func (m *pagerModel) setSize(w, h int) {
	contentWidth := w
	outlineWidth := 0

	// Calculate outline width if visible and viewing markdown
	if m.showOutline && m.isMarkdownFile() {
		outlineWidth = calculateOutlineWidth(w)
		if outlineWidth > 0 {
			contentWidth = w - outlineWidth
			m.outline.visible = true
			m.outline.setSize(outlineWidth, h-statusBarHeight)
		} else {
			m.outline.visible = false
		}
	} else {
		m.outline.visible = false
	}

	// Disable high performance rendering for markdown files because
	// outline toggle changes layout and scroll regions don't work with sidebars
	if m.isMarkdownFile() {
		m.viewport.HighPerformanceRendering = false
	} else {
		m.viewport.HighPerformanceRendering = config.HighPerformancePager
	}

	m.viewport.Width = contentWidth
	m.viewport.Height = h - statusBarHeight

	if m.showHelp {
		if pagerHelpHeight == 0 {
			pagerHelpHeight = strings.Count(m.helpView(), "\n")
		}
		m.viewport.Height -= (statusBarHeight + pagerHelpHeight)
	}
}

func (m *pagerModel) setContent(s string) {
	m.viewport.SetContent(s)
}

// isMarkdownFile returns true if the current document is a markdown file.
func (m *pagerModel) isMarkdownFile() bool {
	return utils.IsMarkdownFile(m.currentDocument.Note)
}

func (m *pagerModel) toggleHelp() {
	m.showHelp = !m.showHelp
	m.setSize(m.common.width, m.common.height)
	if m.viewport.PastBottom() {
		m.viewport.GotoBottom()
	}
}

type pagerStatusMessage struct {
	message string
	isError bool
}

// Perform stuff that needs to happen after a successful markdown stash. Note
// that the returned command should be sent back the through the pager
// update function.
func (m *pagerModel) showStatusMessage(msg pagerStatusMessage) tea.Cmd {
	// Show a success message to the user
	m.state = pagerStateStatusMessage
	m.statusMessage = msg.message
	if m.statusMessageTimer != nil {
		m.statusMessageTimer.Stop()
	}
	m.statusMessageTimer = time.NewTimer(statusMessageTimeout)

	return waitForStatusMessageTimeout(pagerContext, m.statusMessageTimer)
}

func (m *pagerModel) unload() {
	log.Debug("unload")
	if m.showHelp {
		m.toggleHelp()
	}
	if m.statusMessageTimer != nil {
		m.statusMessageTimer.Stop()
	}
	m.state = pagerStateBrowse
	m.viewport.SetContent("")
	m.viewport.YOffset = 0
	m.unwatchFile()
}

func (m pagerModel) update(msg tea.Msg) (pagerModel, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", keyEsc:
			if m.state != pagerStateBrowse {
				m.state = pagerStateBrowse
				return m, nil
			}
		case "home", "g":
			m.viewport.GotoTop()
			if m.viewport.HighPerformanceRendering {
				cmds = append(cmds, viewport.Sync(m.viewport))
			}
		case "end", "G":
			m.viewport.GotoBottom()
			if m.viewport.HighPerformanceRendering {
				cmds = append(cmds, viewport.Sync(m.viewport))
			}

		case "d":
			m.viewport.HalfViewDown()
			if m.viewport.HighPerformanceRendering {
				cmds = append(cmds, viewport.Sync(m.viewport))
			}

		case "u":
			m.viewport.HalfViewUp()
			if m.viewport.HighPerformanceRendering {
				cmds = append(cmds, viewport.Sync(m.viewport))
			}

		case "e":
			lineno := int(math.RoundToEven(float64(m.viewport.TotalLineCount()) * m.viewport.ScrollPercent()))
			if m.viewport.AtTop() {
				lineno = 0
			}
			log.Info(
				"opening editor",
				"file", m.currentDocument.localPath,
				"line", fmt.Sprintf("%d/%d", lineno, m.viewport.TotalLineCount()),
			)
			return m, openEditor(m.currentDocument.localPath, lineno)

		case "c":
			// Copy using OSC 52
			m.common.terminal.CopyOSC52(m.currentDocument.Body)
			// Copy using native system clipboard
			_ = m.common.terminal.CopyClipboard(m.currentDocument.Body)
			cmds = append(cmds, m.showStatusMessage(pagerStatusMessage{"Copied contents", false}))

		case "r":
			return m, loadLocalMarkdown(&m.currentDocument)

		case "?":
			m.toggleHelp()
			if m.viewport.HighPerformanceRendering {
				cmds = append(cmds, viewport.Sync(m.viewport))
			}

		case "o":
			// Toggle outline visibility (only for markdown files)
			if m.isMarkdownFile() {
				m.showOutline = !m.showOutline
				if !m.showOutline {
					m.outlineFocused = false
				} else {
					// Parse headings immediately when enabling outline
					m.outline.setContent(m.currentDocument.Body)
				}
				m.setSize(m.common.width, m.common.height)
				// Re-render content at new width
				return m, renderWithGlamour(m, m.currentDocument.Body)
			}

		case "tab":
			// Toggle focus between content and outline
			if m.showOutline && m.outline.visible {
				m.outlineFocused = !m.outlineFocused
				m.outline.focused = m.outlineFocused
				if m.outlineFocused {
					// Sync cursor to current heading when gaining focus
					m.outline.cursor = m.outline.current
				}
				m.outline.updateViewport()
			}

		case "]":
			// Jump to next heading
			if m.showOutline && len(m.outline.headings) > 0 {
				nextIdx := m.outline.nextHeadingIndex()
				if nextIdx >= 0 {
					m.jumpToHeading(nextIdx)
					if m.viewport.HighPerformanceRendering {
						cmds = append(cmds, viewport.Sync(m.viewport))
					}
				}
			}

		case "[":
			// Jump to previous heading
			if m.showOutline && len(m.outline.headings) > 0 {
				prevIdx := m.outline.prevHeadingIndex()
				if prevIdx >= 0 {
					m.jumpToHeading(prevIdx)
					if m.viewport.HighPerformanceRendering {
						cmds = append(cmds, viewport.Sync(m.viewport))
					}
				}
			}

		case "enter":
			// Jump to selected heading when outline is focused
			if m.outlineFocused && m.outline.visible {
				m.jumpToHeading(m.outline.cursor)
				if m.viewport.HighPerformanceRendering {
					cmds = append(cmds, viewport.Sync(m.viewport))
				}
			}

		case "j", "down":
			if m.outlineFocused && m.outline.visible {
				m.outline.moveCursorDown()
				return m, nil
			}

		case "k", "up":
			if m.outlineFocused && m.outline.visible {
				m.outline.moveCursorUp()
				return m, nil
			}
		}

	// Glow has rendered the content
	case contentRenderedMsg:
		log.Info("content rendered", "state", m.state)

		m.setSize(m.common.width, m.common.height)
		m.setContent(string(msg))

		if m.viewport.HighPerformanceRendering {
			cmds = append(cmds, viewport.Sync(m.viewport))
		}
		cmds = append(cmds, m.watchFile)

		// Always parse headings for markdown files (needed for navigation)
		// Then map them to rendered line positions
		if m.isMarkdownFile() {
			m.outline.setContent(m.currentDocument.Body)
			m.outline.mapHeadingsToRenderedLines(string(msg))
		}

	// The file was changed on disk and we're reloading it
	case reloadMsg:
		return m, loadLocalMarkdown(&m.currentDocument)

	// We've finished editing the document, potentially making changes. Let's
	// retrieve the latest version of the document so that we display
	// up-to-date contents.
	case editorFinishedMsg:
		return m, loadLocalMarkdown(&m.currentDocument)

	// We've received terminal dimensions, either for the first time or
	// after a resize
	case tea.WindowSizeMsg:
		return m, renderWithGlamour(m, m.currentDocument.Body)

	case statusMessageTimeoutMsg:
		m.state = pagerStateBrowse
	}

	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	// Update current heading based on scroll position
	if m.showOutline && m.outline.visible && !m.outlineFocused {
		m.updateCurrentHeading()
	}

	return m, tea.Batch(cmds...)
}

// scrollOff is the number of lines to keep visible above/below when jumping to headings.
// Similar to Vim's scrolloff setting.
const scrollOff = 5

// jumpToHeading scrolls the viewport to show the heading at the given index.
// The heading is positioned with scrollOff lines of context above it.
func (m *pagerModel) jumpToHeading(headingIndex int) {
	if headingIndex < 0 || headingIndex >= len(m.outline.headings) {
		return
	}

	heading := m.outline.headings[headingIndex]

	// Use pre-computed rendered line position
	targetLine := heading.RenderedLine
	if targetLine < 0 {
		// Fallback to ratio-based approximation if not mapped
		totalRawLines := strings.Count(m.currentDocument.Body, "\n") + 1
		if totalRawLines == 0 {
			return
		}
		ratio := float64(heading.Line) / float64(totalRawLines)
		targetLine = int(ratio * float64(m.viewport.TotalLineCount()))
	}

	// Apply scroll offset so heading appears scrollOff lines from top
	scrollTarget := targetLine - scrollOff
	if scrollTarget < 0 {
		scrollTarget = 0
	}

	// Clamp to valid range
	maxOffset := m.viewport.TotalLineCount() - m.viewport.Height
	if maxOffset < 0 {
		maxOffset = 0
	}
	if scrollTarget > maxOffset {
		scrollTarget = maxOffset
	}

	m.viewport.YOffset = scrollTarget
	m.outline.current = headingIndex
	m.outline.cursor = headingIndex
	m.outline.updateViewport()
}

// updateCurrentHeading updates the outline's current heading based on scroll position.
func (m *pagerModel) updateCurrentHeading() {
	if len(m.outline.headings) == 0 {
		return
	}

	// Calculate current line from scroll position
	currentLine := m.viewport.YOffset

	// Convert to approximate raw markdown line
	totalRenderedLines := m.viewport.TotalLineCount()
	if totalRenderedLines == 0 {
		return
	}

	totalRawLines := strings.Count(m.currentDocument.Body, "\n") + 1
	ratio := float64(currentLine) / float64(totalRenderedLines)
	rawLine := int(ratio * float64(totalRawLines))

	m.outline.updateCurrent(rawLine)
}

func (m pagerModel) View() string {
	var b strings.Builder

	// Main content
	content := m.viewport.View()

	// Add outline sidebar if visible
	if m.outline.visible && len(m.outline.headings) > 0 {
		content = m.joinContentAndOutline(content, m.outline.View())
	}

	fmt.Fprint(&b, content+"\n")

	// Footer
	m.statusBarView(&b)

	if m.showHelp {
		fmt.Fprint(&b, "\n"+m.helpView())
	}

	return b.String()
}

// joinContentAndOutline joins the main content and outline sidebar line by line.
// This is more reliable than lipgloss.JoinHorizontal for ANSI-styled content.
func (m pagerModel) joinContentAndOutline(content, outline string) string {
	contentLines := strings.Split(content, "\n")
	outlineLines := strings.Split(outline, "\n")

	// Ensure we have enough lines
	maxLines := len(contentLines)
	if len(outlineLines) > maxLines {
		maxLines = len(outlineLines)
	}

	var result strings.Builder
	for i := 0; i < maxLines; i++ {
		var contentLine, outlineLine string

		if i < len(contentLines) {
			contentLine = contentLines[i]
		}
		if i < len(outlineLines) {
			outlineLine = outlineLines[i]
		}

		// Pad content line to viewport width
		contentWidth := ansi.PrintableRuneWidth(contentLine)
		if contentWidth < m.viewport.Width {
			contentLine += strings.Repeat(" ", m.viewport.Width-contentWidth)
		}

		result.WriteString(contentLine)
		result.WriteString(outlineLine)

		if i < maxLines-1 {
			result.WriteString("\n")
		}
	}

	return result.String()
}

func (m pagerModel) statusBarView(b *strings.Builder) {
	const (
		minPercent               float64 = 0.0
		maxPercent               float64 = 1.0
		percentToStringMagnitude float64 = 100.0
	)

	showStatusMessage := m.state == pagerStateStatusMessage

	// Logo
	logo := glowLogoView()

	// Scroll percent
	percent := math.Max(minPercent, math.Min(maxPercent, m.viewport.ScrollPercent()))
	scrollPercent := fmt.Sprintf(" %3.f%% ", percent*percentToStringMagnitude)
	if showStatusMessage {
		scrollPercent = statusBarMessageScrollPosStyle(scrollPercent)
	} else {
		scrollPercent = statusBarScrollPosStyle(scrollPercent)
	}

	// "Help" note
	var helpNote string
	if showStatusMessage {
		helpNote = statusBarMessageHelpStyle(" ? Help ")
	} else {
		helpNote = statusBarHelpStyle(" ? Help ")
	}

	// Note
	var note string
	if showStatusMessage {
		note = m.statusMessage
	} else {
		note = m.currentDocument.Note
	}
	note = truncate.StringWithTail(" "+note+" ", uint(max(0, //nolint:gosec
		m.common.width-
			ansi.PrintableRuneWidth(logo)-
			ansi.PrintableRuneWidth(scrollPercent)-
			ansi.PrintableRuneWidth(helpNote),
	)), ellipsis)
	if showStatusMessage {
		note = statusBarMessageStyle(note)
	} else {
		note = statusBarNoteStyle(note)
	}

	// Empty space
	padding := max(0,
		m.common.width-
			ansi.PrintableRuneWidth(logo)-
			ansi.PrintableRuneWidth(note)-
			ansi.PrintableRuneWidth(scrollPercent)-
			ansi.PrintableRuneWidth(helpNote),
	)
	emptySpace := strings.Repeat(" ", padding)
	if showStatusMessage {
		emptySpace = statusBarMessageStyle(emptySpace)
	} else {
		emptySpace = statusBarNoteStyle(emptySpace)
	}

	fmt.Fprintf(b, "%s%s%s%s%s",
		logo,
		note,
		emptySpace,
		scrollPercent,
		helpNote,
	)
}

func (m pagerModel) helpView() (s string) {
	col1 := []string{
		"g/home  go to top",
		"G/end   go to bottom",
		"c       copy contents",
		"e       edit this document",
		"r       reload this document",
		"esc     back to files",
		"q       quit",
	}

	col2 := []string{
		"o       toggle outline",
		"tab     focus outline",
		"]/[     next/prev heading",
	}

	s += "\n"
	s += "k/↑      up                  " + col1[0] + "\n"
	s += "j/↓      down                " + col1[1] + "\n"
	s += "b/pgup   page up             " + col1[2] + "\n"
	s += "f/pgdn   page down           " + col1[3] + "\n"
	s += "u        ½ page up           " + col1[4] + "\n"
	s += "d        ½ page down         " + col1[5] + "\n"
	s += "                             " + col1[6] + "\n"
	s += "\n"
	for _, item := range col2 {
		s += item + "\n"
	}

	s = indent(s, 2)

	// Fill up empty cells with spaces for background coloring
	if m.common.width > 0 {
		lines := strings.Split(s, "\n")
		for i := 0; i < len(lines); i++ {
			l := runewidth.StringWidth(lines[i])
			n := max(m.common.width-l, 0)
			lines[i] += strings.Repeat(" ", n)
		}

		s = strings.Join(lines, "\n")
	}

	return helpViewStyle(s)
}

// COMMANDS

func renderWithGlamour(m pagerModel, md string) tea.Cmd {
	return func() tea.Msg {
		s, err := glamourRender(m, md)
		if err != nil {
			log.Error("error rendering with Glamour", "error", err)
			return errMsg{err}
		}
		return contentRenderedMsg(s)
	}
}

// This is where the magic happens.
func glamourRender(m pagerModel, markdown string) (string, error) {
	trunc := lipgloss.NewStyle().MaxWidth(m.viewport.Width - lineNumberWidth).Render

	if !m.common.cfg.GlamourEnabled {
		return markdown, nil
	}

	isCode := !utils.IsMarkdownFile(m.currentDocument.Note)
	width := max(0, min(int(m.common.cfg.GlamourMaxWidth), m.viewport.Width)) //nolint:gosec

	// Use the injected renderer
	out, err := m.common.renderer.Render(
		markdown,
		width,
		m.common.cfg.GlamourStyle,
		m.currentDocument.Note,
		m.common.cfg.PreserveNewLines,
	)
	if err != nil {
		return "", err
	}

	// trim lines
	lines := strings.Split(out, "\n")

	var content strings.Builder
	for i, s := range lines {
		if isCode || m.common.cfg.ShowLineNumbers {
			content.WriteString(lineNumberStyle(fmt.Sprintf("%"+fmt.Sprint(lineNumberWidth)+"d", i+1)))
			content.WriteString(trunc(s))
		} else {
			content.WriteString(s)
		}

		// don't add an artificial newline after the last split
		if i+1 < len(lines) {
			content.WriteRune('\n')
		}
	}

	return content.String(), nil
}

func (m *pagerModel) initWatcher() {
	var err error
	m.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		log.Error("error creating fsnotify watcher", "error", err)
	}
}

func (m *pagerModel) watchFile() tea.Msg {
	dir := m.localDir()

	if err := m.watcher.Add(dir); err != nil {
		log.Error("error adding dir to fsnotify watcher", "error", err)
		return nil
	}

	log.Info("fsnotify watching dir", "dir", dir)

	for {
		select {
		case event, ok := <-m.watcher.Events:
			if !ok || event.Name != m.currentDocument.localPath {
				continue
			}

			if !event.Has(fsnotify.Write) && !event.Has(fsnotify.Create) {
				continue
			}

			log.Debug("fsnotify event", "file", event.Name, "event", event.Op)
			return reloadMsg{}
		case err, ok := <-m.watcher.Errors:
			if !ok {
				continue
			}
			log.Debug("fsnotify error", "dir", dir, "error", err)
		}
	}
}

func (m *pagerModel) unwatchFile() {
	dir := m.localDir()

	err := m.watcher.Remove(dir)
	if err == nil {
		log.Debug("fsnotify dir unwatched", "dir", dir)
	} else {
		log.Error("fsnotify fail to unwatch dir", "dir", dir, "error", err)
	}
}

func (m *pagerModel) localDir() string {
	return filepath.Dir(m.currentDocument.localPath)
}
