<script>
	export let title;
	export let showBack = false;
	export let onBack = null;
	import TopBar from './TopBar.svelte';
</script>

<div class="dashboard-page">
	<TopBar {title} {showBack} {onBack} />

	<div class="dashboard-container">
		<slot />
	</div>
</div>

<style>
	.dashboard-page {
		min-height: 100%;
		background: #fff;
		padding: 4.25rem clamp(1rem, 4vw, 2rem) 5.5rem;
	}

	.dashboard-container {
		max-width: 50rem;
		margin: 0 auto;
		display: flex;
		flex-direction: column;
		gap: clamp(1.5rem, 4vw, 2rem);
	}
</style>
