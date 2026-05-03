package battleplans

var allowedTransitions = map[string]map[string]bool{
	StatusActive:    {StatusCompleted: true, StatusArchived: true, StatusDraft: true},
	StatusDraft:     {StatusActive: true, StatusArchived: true},
	StatusCompleted: {StatusArchived: true},
	StatusArchived:  {},
}

// CanTransition reports whether moving a record from `from` to `to` is allowed.
func CanTransition(from, to string) bool {
	if from == to {
		return true
	}
	targets, ok := allowedTransitions[from]
	if !ok {
		return false
	}
	return targets[to]
}

// IsEditable reports whether a battleplan in the given status can have its
// content (priority/pillars/data) modified via Update.
func IsEditable(status string) bool {
	return status == StatusActive || status == StatusDraft
}
