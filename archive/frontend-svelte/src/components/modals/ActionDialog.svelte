<script>
	export let show = false;
	export let title = 'Confirm';
	export let message = '';
	export let confirmLabel = 'Confirm';
	export let loading = false;
	export let showTextField = false;
	export let textValue = '';
	export let textPlaceholder = 'Reason...';
	export let textRows = 4;
	export let onConfirm;
	export let onCancel;
	export let onTextChange;
	export let closeOnConfirm = false;

	function handleOverlayClick() {
		if (loading) return;
		onCancel?.();
	}

	function handleKeydown(event) {
		if (loading) return;
		if (event.key === 'Escape') onCancel?.();
	}

	function handleInput(event) {
		onTextChange?.(event.currentTarget.value);
	}

	async function handleConfirm() {
		const result = await onConfirm?.();
		if (closeOnConfirm && result !== false) {
			onCancel?.();
		}
	}
</script>

{#if show}
	<div class="action-overlay" role="presentation" on:click={handleOverlayClick}>
		<div
			class="action-dialog"
			role="dialog"
			aria-modal="true"
			tabindex="0"
			on:click|stopPropagation
			on:keydown={handleKeydown}
		>
			<button type="button" class="action-close" on:click={onCancel} aria-label="Close dialog" disabled={loading}>
				<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="currentColor" viewBox="0 0 16 16" aria-hidden="true">
					<path d="M2.146 2.854a.5.5 0 1 1 .708-.708L8 7.293l5.146-5.147a.5.5 0 0 1 .708.708L8.707 8l5.147 5.146a.5.5 0 0 1-.708.708L8 8.707l-5.146 5.147a.5.5 0 0 1-.708-.708L7.293 8z"/>
				</svg>
			</button>
			<h3>{title}</h3>
			{#if message}
				<p>{message}</p>
			{/if}
			{#if showTextField}
				<textarea
					value={textValue}
					rows={textRows}
					placeholder={textPlaceholder}
					disabled={loading}
					on:input={handleInput}
				></textarea>
			{/if}
			<div class="action-cta">
				<button type="button" class="action-confirm" on:click={handleConfirm} disabled={loading}>
					{loading ? 'Saving...' : confirmLabel}
				</button>
			</div>
		</div>
	</div>
{/if}

<style>
	.action-overlay {
		position: fixed;
		inset: 0;
		background: rgba(0, 0, 0, 0.5);
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 1rem;
		z-index: 2100;
	}

	.action-dialog {
		width: min(32rem, 100%);
		background: #fff;
		border: 2px solid #000;
		padding: 1rem;
		position: relative;
		display: grid;
		gap: 0.75rem;
	}

	.action-close {
		position: absolute;
		top: 0.6rem;
		right: 0.6rem;
		border: none;
		background: transparent;
		padding: 0.15rem;
		cursor: pointer;
		color: #000;
	}

	h3 {
		margin: 0;
		font-size: var(--ui-font-size);
		font-weight: 700;
	}

	p {
		margin: 0;
		font-size: var(--ui-font-size);
		color: #333;
	}

	textarea {
		width: 100%;
		border: 1px solid #000;
		padding: 0.6rem;
		resize: vertical;
		font-family: inherit;
		font-size: var(--ui-font-size);
	}

	.action-cta {
		display: flex;
		justify-content: center;
	}

	.action-confirm {
		border: 2px solid #000;
		background: #000;
		color: #fff;
		padding: 0.5rem 0.9rem;
		font-size: var(--ui-font-size);
		font-weight: 700;
		cursor: pointer;
	}
</style>
