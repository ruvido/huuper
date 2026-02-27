<script>
	export let variant = 'default';
</script>

<div class="card" class:admin={variant === 'admin'} class:state={variant === 'state'} class:item={variant === 'item'}>
	<slot />
</div>

<style>
	.card {
		background: #fff;
		border: 2px solid #000;
		border-radius: var(--card-radius);
		padding: clamp(2rem, 5vw, 3rem) clamp(1.5rem, 4vw, 2rem);
		text-align: center;
		width: 100%;
	}

	.card.admin {
		border: none;
		padding: clamp(1.25rem, 3vw, 1.5rem);
		text-align: left;
	}

	.card.state {
		color: #000;
		font-size: clamp(1rem, 3vw, 1.125rem);
	}

	.card.item {
		border: 1px solid #000;
		padding: 0.75rem;
		text-align: left;
	}
</style>
