<script>
	export let title;
</script>

<div class="dashboard-page">
	<div class="dashboard-container">
		<h1 class="dashboard-title">{title}</h1>
		<slot />
	</div>
</div>

<style>
	.dashboard-page {
		min-height: 100vh;
		background: #fff;
		padding: 5rem clamp(1rem, 4vw, 2rem) clamp(1rem, 4vw, 2rem);
	}

	.dashboard-container {
		max-width: 50rem;
		margin: 0 auto;
		display: flex;
		flex-direction: column;
		gap: clamp(1.5rem, 4vw, 2rem);
	}

	.dashboard-title {
		margin: 0 0 clamp(1rem, 3vw, 1.5rem) 0;
		font-size: 2rem;
		font-weight: bold;
		color: #000;
	}
</style>
