<script>
	export let error = '';
</script>

{#if error}
	<div class="error">{error}</div>
{/if}

<style>
	.error {
		background: #fff;
		border: 1px solid #000;
		padding: 0.75rem;
		margin-bottom: 1rem;
		font-size: 0.9rem;
	}
</style>
