<script>
	import { renderContent } from '../../lib/markdown';

	export let show = false;
	export let content = '';
	export let onClose;
</script>

{#if show}
	<div class="welcome-overlay">
			<div class="welcome-modal">
				<div class="welcome-content" aria-live="polite">
					{@html renderContent(content)}
				</div>
			<button class="welcome-action" on:click={onClose}>
				Ok
			</button>
		</div>
	</div>
{/if}

<style>
	.welcome-overlay {
		position: fixed;
		top: 0;
		left: 0;
		right: 0;
		bottom: 0;
		background: rgba(0, 0, 0, 0.6);
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 1.5rem;
		z-index: 2000;
	}

	.welcome-modal {
		background: #fff;
		border: 2px solid #000;
		max-width: 28rem;
		width: 100%;
		padding: clamp(1.25rem, 4vw, 2rem);
		position: relative;
		box-shadow: 0 20px 40px rgba(0, 0, 0, 0.25);
	}

	.welcome-content {
		max-height: 50vh;
		overflow-y: auto;
		font-size: clamp(0.95rem, 3vw, 1rem);
		color: #000;
		line-height: 1.5;
		margin-bottom: 1.5rem;
	}

	.welcome-action {
		width: 100%;
		padding: 0.85rem;
		background: #000;
		color: #fff;
		border: 2px solid #000;
		font-size: 1rem;
		font-weight: 600;
		cursor: pointer;
	}

	.welcome-action:hover {
		background: #333;
	}
</style>
