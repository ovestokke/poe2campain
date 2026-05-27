package importer

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"golang.org/x/net/html"

	"poe2campain/internal/campaign"
)

// parseDomistaeHTML reads a Domistae act HTML file and extracts structured zones and steps.
func parseDomistaeHTML(filePath string, actNum int) ([]domistaeZone, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("opening %s: %w", filePath, err)
	}
	defer f.Close()

	doc, err := html.Parse(f)
	if err != nil {
		return nil, fmt.Errorf("parsing %s: %w", filePath, err)
	}

	return extractZonesFromHTML(doc), nil
}

// extractZonesFromHTML walks the parsed HTML tree and extracts zone structures.
func extractZonesFromHTML(doc *html.Node) []domistaeZone {
	var zones []domistaeZone
	var currentZone *domistaeZone

	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "div" {
			cls := getHTMLAttr(n, "class")

			// Zone header
			if containsClass(cls, "zone-header") {
				if currentZone != nil && len(currentZone.Steps) > 0 {
					zones = append(zones, *currentZone)
				}
				id := getHTMLAttr(n, "id")
				text := extractTextFromHTMLNode(n)
				currentZone = &domistaeZone{
					RawHeader: text,
					HTMLID:    id,
				}
				return // don't recurse into zone-header children
			}

			// Step content
			if containsClass(cls, "step-content") {
				if currentZone != nil {
					text, spans := extractTextAndSpansFromHTMLNode(n)
					if text != "" {
						// Find data-step from parent step div
						var order int
						for p := n.Parent; p != nil; p = p.Parent {
							if p.Type == html.ElementNode && p.Data == "div" {
								if containsClass(getHTMLAttr(p, "class"), "step") {
									ds := getHTMLAttr(p, "data-step")
									fmt.Sscanf(ds, "%d", &order)
									break
								}
							}
						}
						currentZone.Steps = append(currentZone.Steps, domistaeStep{
							Order:   order,
							RawText: text,
							Spans:   spans,
						})
					}
				}
				return
			}

			// Note (but not a step-note)
			if containsClass(cls, "note") && !containsClass(cls, "step") {
				if currentZone != nil {
					text := extractTextFromHTMLNode(n)
					if text != "" {
						currentZone.Notes = append(currentZone.Notes, domistaeNote{
							RawText: text,
							Class:   cls,
						})
					}
				}
				return
			}
		}

		// Recurse into children
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}

	walk(doc)

	// Flush last zone
	if currentZone != nil && len(currentZone.Steps) > 0 {
		zones = append(zones, *currentZone)
	}

	return zones
}

// extractTextFromHTMLNode extracts all text content from an HTML node and its children.
// It handles label spans specially: they get converted to "Label — " separators.
func extractTextFromHTMLNode(n *html.Node) string {
	var buf strings.Builder
	extractTextInner(n, &buf, true)
	text := buf.String()
	text = strings.ReplaceAll(text, "\u00a0", " ")
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")
	text = regexp.MustCompile(`\s*— —\s*`).ReplaceAllString(text, ": ")
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")
	return strings.TrimSpace(text)
}

// extractTextInner recursively walks the HTML tree collecting text.
// When insideLabel is true, label spans are converted to "Label — " separators.
// When insideLabel is false, label spans are skipped entirely.
func extractTextInner(n *html.Node, buf *strings.Builder, insideLabel bool) {
	switch n.Type {
	case html.TextNode:
		if buf.Len() > 0 && n.Data != "" && !strings.HasPrefix(n.Data, " ") && !strings.HasPrefix(n.Data, "\n") {
			last := buf.String()[buf.Len()-1:]
			if last != " " && last != "\n" && last != "—" {
				buf.WriteString(" ")
			}
		}
		buf.WriteString(n.Data)
	case html.ElementNode:
		if n.Data == "br" {
			buf.WriteString("\n")
			return
		}
		if n.Data == "span" {
			cls := getHTMLAttr(n, "class")
			// Label spans: extract inner text and add as separator
			if insideLabel && (cls == "label" || cls == "label new" || cls == "label tip" || cls == "label success" || cls == "label warning") {
				var innerBuf strings.Builder
				for c := n.FirstChild; c != nil; c = c.NextSibling {
					extractTextInner(c, &innerBuf, false) // don't process labels inside labels
				}
				innerText := innerBuf.String()
				if innerText != "" {
					buf.WriteString(innerText)
					buf.WriteString(" — ")
				}
				return
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			extractTextInner(c, buf, insideLabel)
		}
	}
}

// extractTextAndSpansFromHTMLNode extracts plain text and all <span> annotations.
func extractTextAndSpansFromHTMLNode(n *html.Node) (string, []domistaeSpan) {
	var spans []domistaeSpan
	var buf strings.Builder

	var walkNode func(*html.Node)
	walkNode = func(node *html.Node) {
		switch node.Type {
		case html.TextNode:
			buf.WriteString(node.Data)
		case html.ElementNode:
			if node.Data == "br" {
				buf.WriteString("\n")
				return
			}
			if node.Data == "span" {
				cls := getHTMLAttr(node, "class")
				// Use extractTextInner to get span text (handles nested elements)
				var innerBuf strings.Builder
				for c := node.FirstChild; c != nil; c = c.NextSibling {
					extractTextInner(c, &innerBuf, true)
				}
				spanText := normalizeWhitespace(innerBuf.String())
				if spanText != "" && cls != "" {
					spans = append(spans, domistaeSpan{Text: spanText, Classes: cls})
				}
				// Also add span text to main buffer
				buf.WriteString(spanText)
				return
			}
			for c := node.FirstChild; c != nil; c = c.NextSibling {
				walkNode(c)
			}
		}
	}

	walkNode(n)

	text := buf.String()
	text = strings.ReplaceAll(text, "\u00a0", " ")
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")
	text = strings.TrimSpace(text)

	return text, spans
}

// getHTMLAttr returns the value of the named attribute on an HTML node, or "" if not found.
func getHTMLAttr(n *html.Node, key string) string {
	for _, attr := range n.Attr {
		if attr.Key == key {
			return attr.Val
		}
	}
	return ""
}

// containsClass checks if a class string contains the given class name.
func containsClass(classStr, className string) bool {
	for _, c := range strings.Fields(classStr) {
		if c == className {
			return true
		}
	}
	return false
}

func normalizeWhitespace(s string) string {
	s = regexp.MustCompile(`\s+`).ReplaceAllString(s, " ")
	return strings.TrimSpace(s)
}

// --- Exile-UI zone loading (uses existing functions from exileui.go) ---

func loadExileUIZones(sourceDir string) ([]campaign.Zone, map[string]campaign.Zone, error) {
	areas, err := readAreas(sourceDir + "/exile-ui/areas_2.json")
	if err != nil {
		return nil, nil, err
	}
	zones, zoneByID := buildZones(areas)
	return zones, zoneByID, nil
}