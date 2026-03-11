<script>
	export let total = 4;
	export let current = 0;
</script>

<div class="dots">
	{#each Array(total) as _, i}
		<div class="dot" class:active={i === current}></div>
	{/each}
</div>

<style>
	.dots {
		position: fixed;
		right: clamp(1rem, 3vw, 1.5rem);
		top: 50%;
		transform: translateY(-50%);
		display: flex;
		flex-direction: column;
		gap: clamp(0.75rem, 2vw, 1rem);
		z-index: 100;
	}

	.dot {
		width: clamp(0.5rem, 1.5vw, 0.625rem);
		height: clamp(0.5rem, 1.5vw, 0.625rem);
		border-radius: 50%;
		background: #ddd;
		border: 2px solid #000;
		transition: all 0.3s ease;
	}

	.dot.active {
		background: #000;
		transform: scale(1.3);
	}
</style>
