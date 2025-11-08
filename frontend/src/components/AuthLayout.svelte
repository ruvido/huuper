<div class="auth-page">
	<div class="card">
		<slot />
	</div>
</div>

<style>
	.auth-page {
		height: 100vh;
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 1rem;
		background: #fff;
		overflow: auto;
	}

	.card {
		background: #fff;
		border: 2px solid #000;
		padding: 2rem 1.5rem;
		width: 100%;
		max-width: 28rem;
		margin: auto;
	}
</style>
