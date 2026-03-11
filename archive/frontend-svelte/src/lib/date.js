export function formatDatePart(raw) {
	if (!raw) return '';
	const parts = raw.split('-');
	if (parts.length !== 3) return raw;
	return `${parts[2]}/${parts[1]}/${parts[0]}`;
}

export function formatEventDate(value) {
	if (!value || typeof value !== 'string') return '';
	const normalized = value.replace('T', ' ');
	const [dateRaw] = normalized.split(' ');
	return formatDatePart(dateRaw);
}
