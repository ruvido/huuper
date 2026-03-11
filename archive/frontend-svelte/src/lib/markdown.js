import { marked } from 'marked';

const MARKDOWN_PREFIX = 'md:';

marked.setOptions({
	gfm: true,
	breaks: false,
	headerIds: false,
	mangle: false,
});

function escapeHtml(value) {
	return value
		.replace(/&/g, '&amp;')
		.replace(/</g, '&lt;')
		.replace(/>/g, '&gt;')
		.replace(/"/g, '&quot;')
		.replace(/'/g, '&#39;');
}

function looksLikeHtml(value) {
	return /<\/?[a-z][\s\S]*>/i.test(value);
}

function addLinkTargets(html) {
	return html.replace(/<a\s+([^>]*href=["'][^"']+["'][^>]*)>/gi, (match, attrs) => {
		const hasTarget = /\btarget=/.test(attrs);
		const hasRel = /\brel=/.test(attrs);
		let updated = `<a ${attrs}`;
		if (!hasTarget) {
			updated += ' target="_blank"';
		}
		if (!hasRel) {
			updated += ' rel="noopener"';
		}
		return `${updated}>`;
	});
}

export function renderContent(input) {
	if (typeof input !== 'string' || input.length === 0) {
		return '';
	}

	const prefixMatch = input.match(/^md:\s*/);
	if (prefixMatch) {
		const markdownBody = input.slice(prefixMatch[0].length);
		return addLinkTargets(marked.parse(markdownBody));
	}

	if (looksLikeHtml(input)) {
		return addLinkTargets(input);
	}

	return escapeHtml(input);
}
