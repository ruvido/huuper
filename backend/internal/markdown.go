package internal

import (
	"bytes"
	"strings"

	"github.com/yuin/goldmark"
	goldmarkhtml "github.com/yuin/goldmark/renderer/html"
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
	if err := markdownRenderer.Convert([]byte(clean), &out); err != nil {
		return "", false
	}
	return out.String(), true
}

// Everything rendered here is typed by a person into a textarea — email
// bodies, mentoring notes, settings copy — not authored as a Markdown
// document. Plain Markdown joins consecutive lines into one paragraph, so an
// address or a list of dates written one per line arrived as a single run-on
// line. WithHardWraps makes a newline mean a newline, which is what whoever
// pressed Enter meant.
var markdownRenderer = goldmark.New(
	goldmark.WithRendererOptions(goldmarkhtml.WithHardWraps()),
)

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
