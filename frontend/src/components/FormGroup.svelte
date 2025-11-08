<script>
	export let label;
	export let type = 'text';
	export let name;
	export let value;
	export let id;
	export let disabled = false;
</script>

<div class="form-group">
	<label for={id}>{label}</label>
	<input {id} {type} {name} {disabled} bind:value on:input />
</div>

<style>
	.form-group {
		margin-bottom: 1rem;
	}

	label {
		display: block;
		margin-bottom: 0.5rem;
		font-weight: 600;
		font-size: 0.9rem;
	}

	input {
		width: 100%;
		padding: 0.75rem;
		border: 1px solid #000;
		font-size: 1rem;
		background: #fff;
	}

	input:focus {
		outline: 2px solid #000;
		outline-offset: -2px;
	}

	input:disabled {
		background: #f0f0f0;
		opacity: 0.6;
		cursor: not-allowed;
	}
</style>
