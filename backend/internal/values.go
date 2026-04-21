package internal

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"maps"
	"net/mail"
	"strconv"
	"strings"
	"unicode"
)

func AnyToString(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case json.Number:
		return typed.String()
	case int:
		return strconv.Itoa(typed)
	case int32:
		return strconv.FormatInt(int64(typed), 10)
	case int64:
		return strconv.FormatInt(typed, 10)
	case float64:
		return strconv.FormatInt(int64(typed), 10)
	default:
		return ""
	}
}

func AnyToInt64(value any) (int64, bool) {
	switch typed := value.(type) {
	case int:
		return int64(typed), true
	case int32:
		return int64(typed), true
	case int64:
		return typed, true
	case float64:
		return int64(typed), true
	case json.Number:
		n, err := typed.Int64()
		return n, err == nil
	case string:
		raw := strings.TrimSpace(typed)
		if raw == "" {
			return 0, false
		}
		n, err := strconv.ParseInt(raw, 10, 64)
		return n, err == nil
	default:
		return 0, false
	}
}

func ParseNormalizedEmail(raw string) (mail.Address, string, bool) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return mail.Address{}, "", false
	}
	parsed, err := mail.ParseAddress(trimmed)
	if err != nil {
		return mail.Address{}, "", false
	}
	normalized := strings.ToLower(strings.TrimSpace(parsed.Address))
	if normalized == "" {
		return mail.Address{}, "", false
	}
	return mail.Address{Name: parsed.Name, Address: normalized}, normalized, true
}

func NormalizeEmail(raw string) (string, error) {
	_, normalized, ok := ParseNormalizedEmail(raw)
	if !ok {
		return "", fmt.Errorf("missing email")
	}
	parts := strings.Split(normalized, "@")
	if len(parts) != 2 {
		return "", fmt.Errorf("missing email")
	}
	domain := strings.TrimSpace(parts[1])
	if domain == "" || !strings.Contains(domain, ".") {
		return "", fmt.Errorf("invalid email")
	}
	return normalized, nil
}

func NormalizePhone(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", fmt.Errorf("missing phone")
	}

	compact := strings.Join(strings.Fields(trimmed), "")
	if strings.HasPrefix(compact, "00") {
		compact = "+" + strings.TrimPrefix(compact, "00")
	}
	if !strings.HasPrefix(compact, "+") {
		compact = "+" + compact
	}

	digits := strings.TrimPrefix(compact, "+")
	if digits == "" {
		return "", fmt.Errorf("missing phone digits")
	}
	for _, r := range digits {
		if !unicode.IsDigit(r) {
			return "", fmt.Errorf("invalid phone")
		}
	}

	if len(digits) < 9 {
		return "", fmt.Errorf("invalid phone")
	}

	return "+" + digits[:2] + " " + digits[2:], nil
}

func DeepCopyJSONMap(value map[string]any) map[string]any {
	if value == nil {
		return map[string]any{}
	}

	raw, err := json.Marshal(value)
	if err != nil {
		return maps.Clone(value)
	}

	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		return maps.Clone(value)
	}
	if out == nil {
		return map[string]any{}
	}
	return out
}

func RandomToken() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return ""
	}
	return hex.EncodeToString(buf)
}

func NormalizePersonName(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}

	parts := strings.Fields(strings.ToLower(trimmed))
	for i, part := range parts {
		runes := []rune(part)
		if len(runes) == 0 {
			continue
		}
		runes[0] = unicode.ToUpper(runes[0])
		for j := 1; j < len(runes); j++ {
			runes[j] = unicode.ToLower(runes[j])
		}
		parts[i] = string(runes)
	}
	return strings.Join(parts, " ")
}
