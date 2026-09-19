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
	return styleForEmail(out.String()), true
}

// Type scale for emails. Mail clients throw away <style> blocks, so a design
// system here is a set of inline styles applied to the rendered HTML — one place
// that decides what a heading, a name and a figure look like, rather than the
// person writing the template reaching for bold to fake a hierarchy.
//
// Three levels and no more: a section label (small, spaced, quiet), the body,
// and inside a table the label column and the figure column. Nothing inside a
// table is bigger than the text around it, which is what makes a table read as
// data instead of as a shout.
const (
	emailBodyStyle     = "font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,Helvetica,Arial,sans-serif;font-size:15px;line-height:1.5;color:#111"
	emailTitleStyle    = "margin:0 0 2px;font-size:19px;font-weight:700;line-height:1.25;color:#111"
	emailHeadingStyle  = "margin:22px 0 6px;font-size:12px;font-weight:700;letter-spacing:.09em;text-transform:uppercase;color:#888"
	emailSubtitleStyle = "margin:0 0 4px;color:#777"
	emailTableStyle    = "border-collapse:collapse;margin:14px 0;width:100%;font-size:15px"
	emailCellStyle     = "border-bottom:1px solid #e6e6e6;padding:7px 0"
	emailParaStyle     = "margin:6px 0"
	emailListStyle     = "list-style:none;margin:0;padding:0"
	emailItemStyle     = "border-bottom:1px solid #e6e6e6;padding:9px 0"
)

// styleForEmail inlines that scale onto the rendered HTML.
func styleForEmail(html string) string {
	// The subject of the email gets a title; everything below it is a section
	// label. One title, one level of section: an email is not a document.
	html = titlePattern.ReplaceAllString(html, `<div style="`+emailTitleStyle+`">$1</div>`)
	html = headingPattern.ReplaceAllString(html, `<div style="`+emailHeadingStyle+`">$2</div>`)
	html = strings.ReplaceAll(html, "<p>", `<p style="`+emailParaStyle+`">`)
	html = subtitlePattern.ReplaceAllString(html, `$1<p style="`+emailSubtitleStyle+`">$2</p>`)
	html = strings.ReplaceAll(html, "<table>", `<table style="`+emailTableStyle+`">`)
	// People are a list, and one person is ruled off from the next.
	html = strings.ReplaceAll(html, "<ul>", `<ul style="`+emailListStyle+`">`)
	html = strings.ReplaceAll(html, "<li>", `<li style="`+emailItemStyle+`">`)
	html = cellPattern.ReplaceAllStringFunc(html, func(tag string) string {
		m := cellPattern.FindStringSubmatch(tag)
		name, existing := m[1], m[2]
		style := emailCellStyle
		switch {
		case name == "th":
			// The header is a label, not a title: same size, quiet colour.
			style += ";font-weight:600;color:#888;border-bottom:2px solid #ddd"
		case existing != "":
			// The aligned column is the figures one.
			style += ";font-variant-numeric:tabular-nums"
		}
		if existing != "" {
			style += ";" + existing
		} else if name == "th" {
			style += ";text-align:left"
		}
		return fmt.Sprintf(`<%s style="%s">`, name, style)
	})
	return `<div style="` + emailBodyStyle + `">` + html + `</div>`
}

// headingPattern matches any rendered heading, whatever its level: in an email
// they are all the same thing, a label over a block.
var headingPattern = regexp.MustCompile(`(?s)<h([2-6])[^>]*>(.*?)</h[2-6]>`)

// subtitlePattern matches the paragraph that follows the title: the dates, the
// numbers, whatever sits under the name of the thing — said quietly.
var subtitlePattern = regexp.MustCompile(`(?s)(<div style="` + regexp.QuoteMeta(emailTitleStyle) + `">.*?</div>\s*)<p[^>]*>(.*?)</p>`)

// titlePattern matches the one heading that is the email's own title.
var titlePattern = regexp.MustCompile(`(?s)<h1[^>]*>(.*?)</h1>`)

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
