export function normalizeEmailInput(value) {
	return typeof value === 'string' ? value.trim().toLowerCase() : '';
}
