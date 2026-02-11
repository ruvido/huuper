<script>
	export let label;
	export let type = 'text';
	export let name;
	export let value;
	export let id;
	export let disabled = false;
	export let required = false;
	export let minLength;
	export let error = ''; // Expose error as bindable prop
	export let matchField = ''; // Optional: field name to match (for password confirmation)
	export let matchValue = ''; // Optional: value to match against

	let touched = false;

	$: if (value !== undefined && touched) {
		validateField();
	}

	function handleInput() {
		if (touched) {
			validateField();
		}
	}

	function handleBlur() {
		touched = true;
		validateField();
	}

	function validateField() {
		error = '';

		const normalizedValue = typeof value === 'string' ? value.trim() : value;

		if (required && !normalizedValue) {
			error = `${label} is required`;
			return;
		}

		if (type === 'email' && typeof value === 'string' && value !== '') {
			const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
			if (!emailRegex.test(value)) {
				error = 'Please enter a valid email address';
				return;
			}
		}

		if (minLength && typeof value === 'string' && value.length > 0 && value.length < minLength) {
			error = `Must be at least ${minLength} characters`;
			return;
		}

		if (matchField && value !== matchValue) {
			error = 'Passwords do not match';
			return;
		}
	}

	export function isValid() {
		touched = true;
		validateField();
		return !error;
	}
</script>

<div class="form-group">
	<label for={id}>{label}</label>
	<input
		{id}
		{type}
		{name}
		{disabled}
		{required}
		minlength={minLength}
		bind:value
		on:input={handleInput}
		on:blur={handleBlur}
		class:error
	/>
	{#if error}
		<div class="error-message">
			<span class="error-icon">⚠</span>
			{error}
		</div>
	{/if}
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
		border: 2px solid #000;
		font-size: 1rem;
		background: #fff;
	}

	input.error {
		border-color: #d32f2f;
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

	.error-message {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		margin-top: 0.5rem;
		color: #d32f2f;
		font-size: 0.875rem;
		line-height: 1.4;
	}

	.error-icon {
		font-size: 1rem;
		flex-shrink: 0;
	}
</style>
