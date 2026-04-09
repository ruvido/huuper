package internal

import (
	"fmt"
	"html"
	"sort"
	"strings"
)

func RenderTemplate(raw string, replacements map[string]string) string {
	out := raw
	if len(replacements) == 0 {
		return out
	}

	keys := make([]string, 0, len(replacements))
	for key := range replacements {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		if len(keys[i]) == len(keys[j]) {
			return keys[i] < keys[j]
		}
		return len(keys[i]) > len(keys[j])
	})

	for _, key := range keys {
		out = strings.ReplaceAll(out, "{{"+key+"}}", replacements[key])
	}

	return out
}

func RenderDataText(data map[string]any) string {
	return renderData(data, "\n", func(key string, value any) string {
		return key + ": " + stringify(value)
	})
}

func RenderDataHTML(data map[string]any) string {
	return renderData(data, "<br>", func(key string, value any) string {
		return html.EscapeString(key) + ": " + html.EscapeString(stringify(value))
	})
}

func renderData(data map[string]any, sep string, renderPair func(string, any) string) string {
	if data == nil {
		return ""
	}

	var b strings.Builder
	for i, key := range sortedKeys(data) {
		if i > 0 {
			b.WriteString(sep)
		}
		b.WriteString(renderPair(key, data[key]))
	}
	return b.String()
}

func sortedKeys(data map[string]any) []string {
	keys := make([]string, 0, len(data))
	for key := range data {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func stringify(value any) string {
	return fmt.Sprintf("%v", value)
}
