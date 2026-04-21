package requests

import (
	"strings"
	"time"

	"github.com/pocketbase/pocketbase/core"
)

const mentoringDataKey = "mentoring"

func mentoringBlock(data map[string]any) map[string]any {
	block, _ := data[mentoringDataKey].(map[string]any)
	if block == nil {
		block = map[string]any{}
	}
	return block
}

func mentoringNotes(data map[string]any) []map[string]any {
	raw, ok := mentoringBlock(data)["notes"].([]any)
	if !ok {
		return nil
	}
	out := make([]map[string]any, 0, len(raw))
	for _, item := range raw {
		if note, ok := item.(map[string]any); ok && note != nil {
			out = append(out, note)
		}
	}
	return out
}

func mentoringDoneAt(data map[string]any) string {
	value, _ := mentoringBlock(data)["done_at"].(string)
	return strings.TrimSpace(value)
}

func appendMentoringNote(data map[string]any, text string, actor *core.Record) {
	block := mentoringBlock(data)
	notes, _ := block["notes"].([]any)
	entry := map[string]any{
		"text": text,
		"at":   time.Now().UTC().Format(time.RFC3339),
	}
	if actor != nil {
		entry["by"] = actorDisplayName(actor)
	}
	block["notes"] = append(notes, entry)
	data[mentoringDataKey] = block
}

func markMentoringDone(data map[string]any, actor *core.Record) {
	block := mentoringBlock(data)
	block["done_at"] = time.Now().UTC().Format(time.RFC3339)
	if actor != nil {
		block["done_by"] = actorDisplayName(actor)
	}
	data[mentoringDataKey] = block
}

// MentoringNotesJoined returns all note texts joined with a blank line.
// Used by API helpers and notification templates that render a single blob.
func MentoringNotesJoined(data map[string]any) string {
	notes := mentoringNotes(data)
	if len(notes) == 0 {
		return ""
	}
	parts := make([]string, 0, len(notes))
	for _, note := range notes {
		text, _ := note["text"].(string)
		text = strings.TrimSpace(text)
		if text == "" {
			continue
		}
		parts = append(parts, text)
	}
	return strings.Join(parts, "\n\n")
}
