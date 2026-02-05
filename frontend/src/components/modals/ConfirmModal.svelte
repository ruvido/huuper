<script>
	export let show = false;
	export let title = 'Confirm';
	export let message = '';
	export let confirmLabel = 'Confirm';
	export let loading = false;
	export let onConfirm;
	export let onCancel;

	function handleOverlayClick() {
		if (loading) return;
		onCancel?.();
	}

	function handleKeydown(event) {
		if (loading) return;
		if (event.key === 'Escape') onCancel?.();
	}
</script>

{#if show}
	<div class="confirm-overlay" role="presentation" on:click={handleOverlayClick}>
		<div
			class="confirm-modal"
			role="dialog"
			aria-modal="true"
			tabindex="0"
			on:click|stopPropagation
			on:keydown={handleKeydown}
		>
			<button class="btn-close" on:click={onCancel} aria-label="Close" disabled={loading}>×</button>
			<h3 class="confirm-title">{title}</h3>
			{#if message}
				<p class="confirm-message">{message}</p>
			{/if}
			<div class="confirm-actions">
				<button class="btn-confirm" on:click={onConfirm} disabled={loading}>
					{#if loading}
						<span class="spinner" aria-hidden="true"></span>
					{/if}
					<span>{confirmLabel}</span>
				</button>
			</div>
		</div>
	</div>
{/if}

<style>
	.confirm-overlay {
		position: fixed;
		inset: 0;
		background: rgba(0, 0, 0, 0.6);
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 1.5rem;
		z-index: 2100;
	}

	.confirm-modal {
		background: #fff;
		border: 2px solid #000;
		max-width: 28rem;
		width: 100%;
		padding: clamp(1.25rem, 4vw, 2rem);
		box-shadow: 0 20px 40px rgba(0, 0, 0, 0.25);
		position: relative;
	}

	.confirm-title {
		margin: 0 0 0.75rem 0;
		font-size: 1.2rem;
		font-weight: 700;
		color: #000;
	}

	.confirm-message {
		margin: 0 0 1.5rem 0;
		font-size: 0.95rem;
		line-height: 1.5;
		color: #000;
	}

	.confirm-actions {
		display: flex;
		gap: 0.75rem;
		justify-content: center;
	}

	.confirm-actions button {
		padding: 0.7rem 1rem;
		font-size: 0.95rem;
		font-weight: 600;
		cursor: pointer;
	}

	.btn-confirm {
		background: #000;
		border: 2px solid #000;
		color: #fff;
		display: inline-flex;
		align-items: center;
		gap: 0.5rem;
	}

	.btn-confirm:hover {
		background: #333;
		border-color: #333;
	}

	.spinner {
		width: 1rem;
		height: 1rem;
		border: 2px solid rgba(255, 255, 255, 0.4);
		border-top-color: #fff;
		border-radius: 50%;
		animation: spin 0.7s linear infinite;
	}

	@keyframes spin {
		to {
			transform: rotate(360deg);
		}
	}

	.btn-close {
		position: absolute;
		top: 0.75rem;
		right: 0.75rem;
		width: 2rem;
		height: 2rem;
		border: 2px solid #000;
		background: #fff;
		color: #000;
		font-size: 1.25rem;
		line-height: 1;
		cursor: pointer;
	}

	.btn-close:hover {
		background: #000;
		color: #fff;
	}
</style>
