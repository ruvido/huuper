<script>
	export let options = [];
	export let selected = [];
	export let isMultiple = false;
	export let disabled = false;
	export let onToggle;
</script>

<div class="grid-container">
	{#each options as option}
		{@const needsInput = option.includes(':input')}
		{@const displayText = needsInput ? option.split(':')[0] : option}
		{@const isSelected = isMultiple
			? selected?.includes(option)
			: selected === option}
		<button
			type="button"
			class="grid-box"
			class:selected={isSelected}
			on:click={() => onToggle(option)}
			{disabled}
		>
			{@html displayText.replace(/\n/g, '<br>')}
		</button>
	{/each}
</div>

<style>
	.grid-container {
		display: grid;
		grid-template-columns: repeat(2, 1fr);
		gap: clamp(0.75rem, 2vw, 1rem);
	}

	@media (min-width: 768px) {
		.grid-container {
			grid-template-columns: repeat(3, 1fr);
		}
	}

	@media (min-width: 1024px) {
		.grid-container {
			grid-template-columns: repeat(4, 1fr);
		}
	}

	.grid-box {
		padding: clamp(0.75rem, 2vw, 1rem);
		border: 2px solid #000;
		background-color: #fff;
		color: #000;
		font-size: clamp(0.875rem, 2.5vw, 1rem);
		font-weight: 500;
		font-family: inherit;
		line-height: 1.3;
		cursor: pointer;
		text-align: center;
		height: clamp(4rem, 10vw, 5rem);
		display: flex;
		align-items: center;
		justify-content: center;
		word-wrap: break-word;
		hyphens: auto;
		-webkit-tap-highlight-color: transparent;
	}

	.grid-box:hover:not(:disabled):not(.selected) {
		background-color: #f5f5f5;
	}

	.grid-box.selected,
	.grid-box.selected:hover,
	.grid-box.selected:focus,
	.grid-box.selected:active {
		background-color: #000;
		color: #fff;
	}

	.grid-box:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}
</style>
