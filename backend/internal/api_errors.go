package internal

// RawErrorValue passes a raw value through PocketBase's safeErrorsData
// transform without being rewritten as {code, message}. PocketBase treats
// every entry in an ApiError data map as a validation error by default;
// implementing SafeErrorResolver lets us return the raw value verbatim.
type RawErrorValue struct {
	Value any
}

func (r RawErrorValue) Resolve(_ map[string]any) any {
	return r.Value
}
