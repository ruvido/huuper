<script>
	import Button from '../Button.svelte';

	export let title = '';
	export let text = '';
	export let buttonText = 'Continua';
	export let loading = false;
	export let onSubmit;
	export let showButton = true;
	export let showCheckmark = true;
</script>

<div class="confirmation-page">
	{#if showCheckmark}
		<div class="checkmark-container">
			<svg class="checkmark" viewBox="0 0 52 52">
				<circle class="checkmark-circle" cx="26" cy="26" r="25" fill="none"/>
				<path class="checkmark-check" fill="none" d="M14.1 27.2l7.1 7.2 16.7-16.8"/>
			</svg>
		</div>
	{/if}
	<h1>{title}</h1>
	<p class="confirmation-text">{@html text.replace(/\n/g, '<br>')}</p>
	{#if showButton}
		<Button variant="submit" on:click={onSubmit} disabled={loading}>
			{loading ? 'Invio...' : buttonText}
		</Button>
	{/if}
</div>

<style>
	.confirmation-page {
		min-height: 100vh;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 2rem;
		text-align: center;
	}

	.confirmation-page h1 {
		margin: 0 0 1.5rem 0;
		font-size: clamp(1.75rem, 5vw, 2.5rem);
	}

	.confirmation-text {
		max-width: 28rem;
		font-size: clamp(1rem, 3vw, 1.1rem);
		color: #333;
		line-height: 1.6;
		margin: 0 0 2rem 0;
	}

	.checkmark-container {
		width: 100px;
		height: 100px;
		margin-bottom: 2rem;
	}

	.checkmark {
		width: 100%;
		height: 100%;
	}

	.checkmark-circle {
		stroke: #22c55e;
		stroke-width: 2;
		stroke-dasharray: 166;
		stroke-dashoffset: 166;
		animation: stroke 0.6s cubic-bezier(0.65, 0, 0.45, 1) forwards;
	}

	.checkmark-check {
		stroke: #22c55e;
		stroke-width: 3;
		stroke-linecap: round;
		stroke-linejoin: round;
		stroke-dasharray: 48;
		stroke-dashoffset: 48;
		animation: stroke 0.3s cubic-bezier(0.65, 0, 0.45, 1) 0.6s forwards;
	}

	@keyframes stroke {
		100% {
			stroke-dashoffset: 0;
		}
	}
</style>
