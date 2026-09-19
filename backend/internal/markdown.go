package internal

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
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
	return styleTables(out.String()), true
}

// styleTables gives a rendered table the borders and padding an email client
// will actually honour. Mail clients drop <style> blocks and have no default
// table styling worth anything, so a table arrives looking like loose text
// unless every cell carries its own inline style. Goldmark writes the column
// alignment as a style of its own, so the borders are merged into whatever is
// already there instead of replacing it.
func styleTables(html string) string {
	html = strings.ReplaceAll(html, "<table>",
		`<table style="border-collapse:collapse;margin:12px 0;font-size:15px;width:100%">`)
	html = cellPattern.ReplaceAllStringFunc(html, func(tag string) string {
		m := cellPattern.FindStringSubmatch(tag)
		name, existing := m[1], m[2]
		style := "border:1px solid #ddd;padding:6px 10px"
		if name == "th" {
			style += ";background:#f6f6f6"
		}
		if existing != "" {
			style += ";" + existing
		} else if name == "th" {
			style += ";text-align:left"
		}
		return fmt.Sprintf(`<%s style="%s">`, name, style)
	})
	return html
}

// cellPattern matches a table cell's opening tag, with or without the style
// goldmark adds for a right-aligned column.
var cellPattern = regexp.MustCompile(`<(td|th)(?: style="([^"]*)")?>`)

// Everything rendered here is typed by a person into a textarea — email
// bodies, mentoring notes, settings copy — not authored as a Markdown
// document. Plain Markdown joins consecutive lines into one paragraph, so an
// address or a list of dates written one per line arrived as a single run-on
// line. WithHardWraps makes a newline mean a newline, which is what whoever
// pressed Enter meant.
var markdownRenderer = goldmark.New(
	// Tables, so an email can lay figures out as a table instead of as a column of
	// sentences the reader has to add up themselves.
	goldmark.WithExtensions(extension.Table),
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
