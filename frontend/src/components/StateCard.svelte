<div class="state-card">
	<slot />
</div>

<style>
	.state-card {
		background: #fff;
		border: 2px solid #000;
		padding: clamp(2rem, 5vw, 3rem) clamp(1.5rem, 4vw, 2rem);
		text-align: center;
		color: #000;
		font-size: clamp(1rem, 3vw, 1.125rem);
		width: 100%;
	}
</style>
