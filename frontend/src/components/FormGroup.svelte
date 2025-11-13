<script>
	export let label;
	export let type = 'text';
	export let name;
	export let value;
	export let id;
	export let disabled = false;
	export let required = false;
	export let error = ''; // Expose error as bindable prop
	export let matchField = ''; // Optional: field name to match (for password confirmation)
	export let matchValue = ''; // Optional: value to match against

	let touched = false;

	// Real-time validation for password match only - but only after field is touched
	$: if (matchField && value !== undefined && touched) {
		validateField();
	}

	function handleBlur() {
		touched = true;
		validateField();
	}

	function validateField() {
		error = '';

		// Required check
		if (required && !value) {
			error = `${label} is required`;
			return;
		}

		// Email validation
		if (type === 'email' && value) {
			const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
			if (!emailRegex.test(value)) {
				error = 'Please enter a valid email address';
				return;
			}
		}

		// Password validation
		if (type === 'password' && value && name === 'password') {
			if (value.length < 8) {
				error = 'Password must be at least 8 characters';
				return;
			}
		}

		// Password match validation
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
		bind:value
		on:input
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
