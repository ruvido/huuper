<script>
	export let error = '';
</script>

{#if error}
	<div class="error">{error}</div>
{/if}

<style>
	.error {
		color: #d32f2f;
		background: #ffebee;
		border: 2px solid #d32f2f;
		padding: 0.75rem;
		margin-bottom: 1rem;
		font-size: 0.9rem;
		font-weight: 600;
	}
</style>
