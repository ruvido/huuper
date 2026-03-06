<script>
	import { Check, ChevronDown } from 'lucide-svelte';
	import { renderContent } from '../../lib/markdown';

	export let event;
	export let registered = false;
	export let canRegister = false;
	export let registering = false;
	export let showStatus = true;
	export let canUnsubscribe = false;
	export let unsubscribing = false;
	export let onRegister = () => {};
	export let onUnsubscribe = () => {};
	export let open = false;
	export let onToggle = () => {};
	export let selectable = false;
	export let onSelect = null;


	function formatDatePart(raw) {
		if (!raw) return '';
		const parts = raw.split('-');
		if (parts.length !== 3) return raw;
		return `${parts[2]}/${parts[1]}/${parts[0]}`;
	}

	function formatTimePart(raw) {
		if (!raw) return '';
		const clean = raw.replace('Z', '').trim();
		const time = clean.split('.')[0] || '';
		return time.slice(0, 5);
	}

	function formatEventDate(value, includeTime) {
		if (!value || typeof value !== 'string') return '';
		const normalized = value.replace('T', ' ');
		const [dateRaw, timeRaw] = normalized.split(' ');
		const dateText = formatDatePart(dateRaw);
		if (!dateText) return '';
		if (!includeTime) return dateText;
		const timeText = formatTimePart(timeRaw);
		return timeText ? `${dateText} ${timeText}` : dateText;
	}

	function stringValue(value) {
		if (Array.isArray(value)) return value.filter(Boolean).join(', ');
		if (typeof value === 'string') return value.trim();
		return '';
	}

	$: eventData = event?.data && typeof event.data === 'object' ? event.data : {};
	$: showTime = eventData?.show_time !== false;
	$: durationText = stringValue(eventData?.duration);
	$: locationText = stringValue(eventData?.location);
	$: dateText = durationText || formatEventDate(event?.event_date, showTime);
	$: summaryMeta = [dateText, locationText].filter(Boolean).join(' · ');

	$: registerLabel = stringValue(eventData?.register_label) || 'Register';
	$: unsubscribeLabel = stringValue(eventData?.unsubscribe_label) || 'Unsubscribe';
	$: registeredLabel = stringValue(eventData?.registered_label) || 'Registered';

	function contentString(value) {
		if (!value) return '';
		if (typeof value === 'string') return value;
		if (typeof value === 'object') {
			if (typeof value.md === 'string') return `md: ${value.md}`;
			if (typeof value.html === 'string') return value.html;
			if (typeof value.text === 'string') return value.text;
		}
		return '';
	}

	$: contentValue = contentString(eventData?.content);
	$: hasDetails = contentValue.trim().length > 0;

	function handleRegister(eventItem, e) {
		e?.preventDefault?.();
		onRegister(eventItem);
	}

	function handleUnsubscribe(eventItem, e) {
		e?.preventDefault?.();
		onUnsubscribe(eventItem);
	}

	function handleSummaryClick(e) {
		if (!selectable || typeof onSelect !== 'function') return;
		e?.preventDefault?.();
		onSelect(event);
	}
</script>

<details id={`event-${event?.id || ""}`} class="event-card" open={open} tabindex="-1" on:toggle={(event) => onToggle(!!event.currentTarget?.open)}>
	<summary on:click={handleSummaryClick}>
		<div class="summary-left">
			<h3>{event?.title}</h3>
			{#if summaryMeta}
				<p class="event-date">{summaryMeta}</p>
			{/if}
		</div>
		<div class="summary-right">
			{#if showStatus && registered}
				<div class="status-check" aria-label={registeredLabel}>
					<Check size={20} />
				</div>
			{/if}
			<span class="chevron" aria-hidden="true"><ChevronDown size={18} /></span>
		</div>
	</summary>

	{#if hasDetails || canUnsubscribe || canRegister}
		<div class="details">
			{#if hasDetails}
				<div class="details-content">{@html renderContent(contentValue)}</div>
			{/if}
			{#if canRegister || canUnsubscribe}
				<div class="details-actions">
					{#if canRegister}
						<button
							class="register"
							on:click|preventDefault|stopPropagation={handleRegister.bind(null, event)}
							disabled={registering}
						>
							{registerLabel}
						</button>
					{:else if canUnsubscribe}
						<button
							class="unsubscribe"
							on:click|preventDefault|stopPropagation={handleUnsubscribe.bind(null, event)}
							disabled={unsubscribing}
						>
							{unsubscribeLabel}
						</button>
					{/if}
				</div>
			{/if}
		</div>
	{/if}
</details>

<style>
	.event-card {
		border: 1px solid #000;
		border-radius: var(--card-radius);
		background: #fff;
		padding: 0;
	}

	summary {
		list-style: none;
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: clamp(0.75rem, 3vw, 1.5rem);
		padding: clamp(0.75rem, 2.5vw, 1rem);
		cursor: pointer;
	}

	summary::-webkit-details-marker {
		display: none;
	}

	.summary-left {
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
		min-width: 0;
	}

	h3 {
		margin: 0;
		font-size: var(--ui-font-size);
		color: #000;
		font-weight: 700;
		word-break: break-word;
	}

	.event-date {
		margin: 0;
		font-size: var(--ui-font-size);
		color: #000;
		font-weight: 600;
	}

	.summary-right {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		flex-shrink: 0;
	}

	.status-check {
		width: 2.75rem;
		height: 2.75rem;
		border-radius: 999px;
		background: #14a44d;
		color: #fff;
		display: inline-flex;
		align-items: center;
		justify-content: center;
		flex-shrink: 0;
	}

	.register {
		padding: 0.65rem 1rem;
		border: 2px solid #000;
		background: #000;
		color: #fff;
		font-weight: 600;
		cursor: pointer;
		transition: background 0.2s, color 0.2s;
	}

	.register:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	.register:hover:not(:disabled) {
		background: #fff;
		color: #000;
	}

	.unsubscribe {
		padding: 0.5rem 0.75rem;
		border: 2px solid #b00020;
		background: #b00020;
		color: #fff;
		font-weight: 600;
		cursor: pointer;
		transition: background 0.2s, color 0.2s;
	}

	.unsubscribe:hover:not(:disabled) {
		background: #fff;
		color: #b00020;
	}

	.unsubscribe:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	.chevron {
		display: inline-flex;
		transition: transform 0.2s ease;
	}

	details[open] .chevron {
		transform: rotate(180deg);
	}

	.details {
		padding: clamp(0.75rem, 2.5vw, 1rem);
		padding-top: clamp(0.75rem, 2.5vw, 1rem);
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}

	.details-content {
		font-size: var(--ui-font-size);
		color: #000;
		line-height: 1.5;
	}

	.details-actions {
		display: flex;
		justify-content: center;
	}
</style>
