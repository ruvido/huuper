<script>
	import { X, ArrowLeft, ArrowRight } from 'lucide-svelte';

	export let isFirstStep = false;
	export let isLastStep = false;
	export let canProceed = false;
	export let loading = false;
	export let onBack;
	export let onNext;
	export let onClose;
	export let onComplete;
</script>

<nav class="top-nav">
	{#if isFirstStep}
		<button class="nav-btn close" on:click={onClose} disabled={loading}>
			<X size={20} />
		</button>
	{:else}
		<button class="nav-btn back" on:click={onBack} disabled={loading}>
			<ArrowLeft size={20} />
		</button>
	{/if}
	<div class="nav-spacer"></div>
	{#if isLastStep}
		<button class="nav-btn next" on:click={onComplete} disabled={!canProceed || loading}>
			Completa
		</button>
	{:else}
		<button class="nav-btn next" on:click={onNext} disabled={!canProceed || loading}>
			Avanti <ArrowRight size={20} />
		</button>
	{/if}
</nav>

<style>
	.top-nav {
		position: fixed;
		top: 3px;
		left: 0;
		right: 0;
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: clamp(1rem, 3vw, 1.5rem);
		background: #fff;
		z-index: 100;
	}

	.nav-spacer {
		flex: 1;
	}

	.nav-btn {
		background: transparent;
		border: 2px solid #000;
		padding: clamp(0.5rem, 2vw, 0.75rem) clamp(1rem, 3vw, 1.5rem);
		font-size: clamp(0.875rem, 2.5vw, 1rem);
		font-weight: 600;
		cursor: pointer;
		transition: all 0.2s ease;
		color: #000;
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}

	.nav-btn:hover:not(:disabled) {
		background: #000;
		color: #fff;
	}

	.nav-btn:disabled {
		opacity: 0.3;
		cursor: not-allowed;
	}

	.nav-btn.next {
		background: #000;
		color: #fff;
	}

	.nav-btn.next:hover:not(:disabled) {
		background: #333;
	}

	.nav-btn.next:disabled {
		background: #ccc;
		border-color: #ccc;
		color: #666;
	}

	.nav-btn.back {
		border: none;
		padding: clamp(0.25rem, 1.5vw, 0.5rem);
	}

	.nav-btn.back:hover:not(:disabled) {
		background: transparent;
		color: #666;
	}

	.nav-btn.close {
		font-size: clamp(1.25rem, 3.5vw, 1.5rem);
		padding: clamp(0.25rem, 1.5vw, 0.5rem) clamp(0.75rem, 2.5vw, 1rem);
		line-height: 1;
	}
</style>
