package internal

import (
	"bytes"
	"strings"

	"github.com/yuin/goldmark"
)

func RenderMarkdownHTML(raw string) (string, bool) {
	clean := strings.TrimSpace(raw)
	if strings.HasPrefix(clean, "md:") {
		clean = strings.TrimSpace(strings.TrimPrefix(clean, "md:"))
	}
	if clean == "" {
		return "", false
	}

	var out bytes.Buffer
	if err := goldmark.Convert([]byte(clean), &out); err != nil {
		return "", false
	}
	return out.String(), true
}

// RenderInlineMarkdown renders a single-line string as inline markdown,
// stripping the surrounding <p> wrapper that goldmark adds by default.
// Useful for headings/titles where block-level wrapping is unwanted.
func RenderInlineMarkdown(raw string) (string, bool) {
	html, ok := RenderMarkdownHTML(raw)
	if !ok {
		return "", false
	}
	html = strings.TrimSpace(html)
	html = strings.TrimPrefix(html, "<p>")
	html = strings.TrimSuffix(html, "</p>")
	return strings.TrimSpace(html), true
}
