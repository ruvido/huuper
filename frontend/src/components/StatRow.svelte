<script>
	export let items = [];
	export let center = false;
</script>

<div class="stat-row" class:center={center}>
	{#each items as item}
		<div class="stat">
			<span class="label">{item.label}</span>
			<span class="value" style={`color: ${item.color || '#000'}`}>{item.value}</span>
		</div>
	{/each}
</div>

<style>
	.stat-row {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: clamp(0.5rem, 2vw, 1rem);
		flex-wrap: nowrap;
	}

	.stat-row.center {
		justify-content: center;
	}

	.stat {
		display: flex;
		flex-direction: column;
		gap: 0.35rem;
		flex: 1 1 0;
		min-width: 0;
		align-items: center;
		text-align: center;
	}

	.label {
		font-size: 0.7rem;
		letter-spacing: 0.08em;
		text-transform: uppercase;
		font-weight: 700;
		color: #333;
	}

	.value {
		font-size: clamp(1.5rem, 4vw, 2.1rem);
		font-weight: 700;
	}

	@media (max-width: 480px) {
		.label {
			font-size: 0.6rem;
		}

		.value {
			font-size: 1.35rem;
		}
	}
</style>
