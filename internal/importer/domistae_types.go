package importer

// Domistae HTML data structures.

type domistaeZone struct {
	RawHeader string          // Full text content of zone-header div
	HTMLID    string          // id attribute from zone-header, e.g. "riverbank"
	Steps     []domistaeStep
	Notes     []domistaeNote
}

type domistaeStep struct {
	Order   int             // data-step number (global within an act)
	RawText string          // Plain text of step-content
	Spans   []domistaeSpan  // Semantic annotations from <span> tags
}

type domistaeSpan struct {
	Text    string // Text inside the span
	Classes string // Space-separated CSS classes, e.g. "reward-tag gem"
}

type domistaeNote struct {
	RawText string // Plain text of the note
	Class   string // CSS class string, e.g. "note tip"
}