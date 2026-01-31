package ascii

// Global configuration variables used by the renderer.
// These are package-level defaults that affect rendering behavior.

var (
	// Verbose enables debug logging.
	Verbose bool
	// Coords enables coordinate display in output.
	Coords bool
	// boxBorderPadding is the padding between text and border in nodes.
	boxBorderPadding = 1
	// paddingBetweenX is horizontal space between nodes.
	paddingBetweenX = 5
	// paddingBetweenY is vertical space between nodes.
	paddingBetweenY = 5
	// graphDirection is the default direction for flowcharts ("LR" or "TD").
	graphDirection = "LR"
	// useAscii disables extended Unicode characters when true.
	useAscii = false
)
