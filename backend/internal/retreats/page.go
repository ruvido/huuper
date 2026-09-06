package retreats

import (
	"html"
	"strings"

	"github.com/pocketbase/pocketbase/core"
)

// maxMetaDescription keeps the share-card description within what the
// crawlers actually display before cutting it themselves.
const maxMetaDescription = 200

// RenderPageHead rewrites the <head> of the built retreat page so it carries
// this retreat's own title and description.
//
// One static file serves every slug under /retreat/{slug}, and its head ships
// the placeholder title the build wrote. Social crawlers and search engines do
// not run the JS that fills in the hero, so without this every shared link
// would show the same generic card whatever retreat it points at.
//
// The substitution is deliberately dumb: it takes the title and description
// the build put in the head and swaps every occurrence of those exact strings
// (they appear once each in <title>/<meta name="description"> and again in the
// og: and twitter: tags), inside the head only. Nothing is parsed, nothing is
// written to disk, and an unrecognised head is returned untouched.
func RenderPageHead(page []byte, retreat *core.Record) []byte {
	if retreat == nil {
		return page
	}

	raw := string(page)
	end := strings.Index(strings.ToLower(raw), "</head>")
	if end < 0 {
		return page
	}
	head, rest := raw[:end], raw[end:]

	head = replaceMeta(head, between(head, "<title>", "</title>"), retreat.GetString("title"))
	head = replaceMeta(head, attributeValue(head, `<meta name="description" content="`), metaDescription(retreat))

	return []byte(head + rest)
}

// replaceMeta swaps a value the build wrote into the head for the record's
// own. A missing placeholder or an empty replacement leaves the head as it
// is — a generic card beats a broken one.
func replaceMeta(head string, placeholder string, value string) string {
	placeholder = strings.TrimSpace(placeholder)
	value = strings.TrimSpace(value)
	if placeholder == "" || value == "" {
		return head
	}
	return strings.ReplaceAll(head, placeholder, html.EscapeString(value))
}

// metaDescription is the retreat's own lead paragraph — the same sentence the
// page shows under the title — falling back to the tagline.
func metaDescription(retreat *core.Record) string {
	data := parseData(retreat)
	text := strings.TrimSpace(DataString(data, "lead"))
	if text == "" {
		text = strings.TrimSpace(retreat.GetString("tagline"))
	}
	return truncateOnWord(text, maxMetaDescription)
}

func between(raw string, prefix string, suffix string) string {
	start := strings.Index(raw, prefix)
	if start < 0 {
		return ""
	}
	start += len(prefix)
	end := strings.Index(raw[start:], suffix)
	if end < 0 {
		return ""
	}
	return raw[start : start+end]
}

func attributeValue(raw string, prefix string) string {
	return between(raw, prefix, `"`)
}

// truncateOnWord cuts at the last space before the limit so the description
// never ends mid-word. Counts runes, not bytes: the copy is Italian.
func truncateOnWord(text string, limit int) string {
	runes := []rune(text)
	if len(runes) <= limit {
		return text
	}
	cut := string(runes[:limit])
	if space := strings.LastIndex(cut, " "); space > 0 {
		cut = cut[:space]
	}
	return strings.TrimRight(cut, " ,.;:—-") + "…"
}
