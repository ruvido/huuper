<div class="card">
	<slot />
</div>

<style>
	.card {
		background: #fff;
		border: 2px solid #000;
		padding: clamp(2rem, 5vw, 3rem) clamp(1.5rem, 4vw, 2rem);
		text-align: center;
		width: 100%;
	}
</style>
