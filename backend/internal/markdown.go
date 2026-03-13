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
