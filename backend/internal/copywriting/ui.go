package copywriting

import "strings"

const DefaultRequestButtonLabel = "Submit"

func RequestButtonLabel(cta string, label string) string {
	cta = strings.TrimSpace(cta)
	if cta != "" {
		return cta
	}

	label = strings.TrimSpace(label)
	if label != "" {
		return label
	}

	return DefaultRequestButtonLabel
}
