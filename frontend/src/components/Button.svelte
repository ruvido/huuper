<script>
	export let variant = 'primary';
	export let type = 'button';
	export let disabled = false;
</script>

<button class="btn btn-{variant}" {type} {disabled} on:click>
	<slot />
</button>

<style>
	.btn {
		font-weight: 600;
		cursor: pointer;
		transition: background 0.2s, color 0.2s;
	}

	.btn-primary {
		width: 100%;
		padding: clamp(0.875rem, 3vw, 1rem);
		background: #000;
		color: #fff;
		border: 2px solid #000;
		font-size: clamp(1rem, 3vw, 1.125rem);
	}

	.btn-primary:hover:not(:disabled) {
		background: #fff;
		color: #000;
	}

	.btn-submit {
		width: 100%;
		padding: 0.875rem;
		background: #000;
		color: #fff;
		border: 2px solid #000;
		font-size: 1rem;
	}

	.btn-submit:hover:not(:disabled) {
		background: #fff;
		color: #000;
	}

	.btn-submit:disabled {
		opacity: 0.4;
		cursor: not-allowed;
	}

	.btn-link {
		background: transparent !important;
		border: none !important;
		text-decoration: underline;
		padding: 0 !important;
		margin-left: 0.25rem;
		width: auto !important;
		color: #000;
	}

	.btn-link:hover {
		opacity: 0.7;
		background: transparent !important;
	}
</style>
