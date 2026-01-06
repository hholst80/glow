package ui

import (
	"github.com/atotto/clipboard"
	"github.com/muesli/termenv"
)

// Terminal abstracts terminal I/O operations for testability.
// This interface allows injecting test doubles that don't require
// an actual terminal.
type Terminal interface {
	// HasDarkBackground returns true if the terminal has a dark background.
	// Used to select appropriate color schemes.
	HasDarkBackground() bool

	// CopyOSC52 copies the given string to the clipboard using OSC 52
	// escape sequences. This works over SSH and in terminals that support it.
	CopyOSC52(s string)

	// CopyClipboard copies the given string to the native system clipboard.
	// Returns an error if clipboard access fails.
	CopyClipboard(s string) error
}

// RealTerminal implements Terminal using actual terminal operations.
type RealTerminal struct{}

// HasDarkBackground returns true if the terminal has a dark background.
func (RealTerminal) HasDarkBackground() bool {
	return termenv.HasDarkBackground()
}

// CopyOSC52 copies the string to clipboard using OSC 52 escape sequences.
func (RealTerminal) CopyOSC52(s string) {
	termenv.Copy(s)
}

// CopyClipboard copies the string to the native system clipboard.
func (RealTerminal) CopyClipboard(s string) error {
	return clipboard.WriteAll(s)
}

// Ensure RealTerminal implements Terminal.
var _ Terminal = RealTerminal{}
