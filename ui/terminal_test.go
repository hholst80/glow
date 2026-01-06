package ui

import (
	"errors"
	"testing"
)

// TestTerminal is a test double for Terminal that records calls
// and allows configuring return values.
type TestTerminal struct {
	// Configuration
	DarkBackground   bool
	ClipboardError   error
	ClipboardContent string

	// Call tracking
	OSC52Calls     []string
	ClipboardCalls []string
}

// NewTestTerminal creates a TestTerminal with default settings
// (light background, no clipboard errors).
func NewTestTerminal() *TestTerminal {
	return &TestTerminal{
		DarkBackground: false,
		ClipboardError: nil,
	}
}

// HasDarkBackground returns the configured DarkBackground value.
func (t *TestTerminal) HasDarkBackground() bool {
	return t.DarkBackground
}

// CopyOSC52 records the copy call for later assertion.
func (t *TestTerminal) CopyOSC52(s string) {
	t.OSC52Calls = append(t.OSC52Calls, s)
}

// CopyClipboard records the copy call and returns the configured error.
func (t *TestTerminal) CopyClipboard(s string) error {
	t.ClipboardCalls = append(t.ClipboardCalls, s)
	if t.ClipboardError == nil {
		t.ClipboardContent = s
	}
	return t.ClipboardError
}

// Ensure TestTerminal implements Terminal.
var _ Terminal = (*TestTerminal)(nil)

func TestTestTerminal_HasDarkBackground(t *testing.T) {
	tests := []struct {
		name     string
		dark     bool
		expected bool
	}{
		{"dark background", true, true},
		{"light background", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			term := &TestTerminal{DarkBackground: tt.dark}
			if got := term.HasDarkBackground(); got != tt.expected {
				t.Errorf("HasDarkBackground() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestTestTerminal_CopyOSC52(t *testing.T) {
	term := NewTestTerminal()

	term.CopyOSC52("hello")
	term.CopyOSC52("world")

	if len(term.OSC52Calls) != 2 {
		t.Errorf("expected 2 OSC52 calls, got %d", len(term.OSC52Calls))
	}
	if term.OSC52Calls[0] != "hello" {
		t.Errorf("first call = %q, want %q", term.OSC52Calls[0], "hello")
	}
	if term.OSC52Calls[1] != "world" {
		t.Errorf("second call = %q, want %q", term.OSC52Calls[1], "world")
	}
}

func TestTestTerminal_CopyClipboard(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		term := NewTestTerminal()

		err := term.CopyClipboard("test content")

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if term.ClipboardContent != "test content" {
			t.Errorf("ClipboardContent = %q, want %q", term.ClipboardContent, "test content")
		}
		if len(term.ClipboardCalls) != 1 {
			t.Errorf("expected 1 clipboard call, got %d", len(term.ClipboardCalls))
		}
	})

	t.Run("error", func(t *testing.T) {
		term := NewTestTerminal()
		term.ClipboardError = errors.New("clipboard unavailable")

		err := term.CopyClipboard("test content")

		if err == nil {
			t.Error("expected error, got nil")
		}
		if term.ClipboardContent != "" {
			t.Errorf("ClipboardContent should be empty on error, got %q", term.ClipboardContent)
		}
	})
}

func TestRealTerminal_ImplementsInterface(t *testing.T) {
	// This test verifies RealTerminal implements Terminal at compile time.
	// The actual var _ Terminal = RealTerminal{} in terminal.go does this,
	// but this test makes it explicit.
	var term Terminal = RealTerminal{}
	_ = term
}

func TestNewModel_DarkBackground(t *testing.T) {
	term := &TestTerminal{DarkBackground: true}
	renderer := &TestMarkdownRenderer{}
	cfg := Config{GlamourStyle: "auto"}

	m := newModel(cfg, "", term, renderer).(model)

	// When terminal has dark background, style should be set to dark
	if m.common.cfg.GlamourStyle != "dark" {
		t.Errorf("GlamourStyle = %q, want %q for dark background", m.common.cfg.GlamourStyle, "dark")
	}
}

func TestNewModel_LightBackground(t *testing.T) {
	term := &TestTerminal{DarkBackground: false}
	renderer := &TestMarkdownRenderer{}
	cfg := Config{GlamourStyle: "auto"}

	m := newModel(cfg, "", term, renderer).(model)

	// When terminal has light background, style should be set to light
	if m.common.cfg.GlamourStyle != "light" {
		t.Errorf("GlamourStyle = %q, want %q for light background", m.common.cfg.GlamourStyle, "light")
	}
}

func TestNewModel_ExplicitStyle(t *testing.T) {
	term := &TestTerminal{DarkBackground: true}
	renderer := &TestMarkdownRenderer{}
	cfg := Config{GlamourStyle: "dracula"} // Explicit style, not "auto"

	m := newModel(cfg, "", term, renderer).(model)

	// When an explicit style is set, it should be preserved
	if m.common.cfg.GlamourStyle != "dracula" {
		t.Errorf("GlamourStyle = %q, want %q for explicit style", m.common.cfg.GlamourStyle, "dracula")
	}
}

func TestNewModel_DependencyInjection(t *testing.T) {
	term := NewTestTerminal()
	renderer := &TestMarkdownRenderer{}
	cfg := Config{}

	m := newModel(cfg, "", term, renderer).(model)

	// Verify both dependencies are properly injected into commonModel
	if m.common.terminal != term {
		t.Error("terminal was not properly injected into commonModel")
	}
	if m.common.renderer != renderer {
		t.Error("renderer was not properly injected into commonModel")
	}
}
