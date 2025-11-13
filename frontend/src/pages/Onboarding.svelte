<script>
	import { onMount } from 'svelte';
	import { pb } from '../lib/pocketbase';
	import { navigate } from '../lib/router';
	import Button from '../components/Button.svelte';
	import ErrorMessage from '../components/ErrorMessage.svelte';

	let steps = [];
	let currentStep = 0;
	let formData = {};
	let error = '';
	let loading = false;

	// Count only non-start steps
	$: realSteps = steps.filter(s => s.type !== 'start');
	$: realStepIndex = (() => {
		let count = 0;
		for (let i = 0; i < currentStep; i++) {
			if (steps[i].type !== 'start') count++;
		}
		return count + 1; // 1-indexed
	})();

	// Check if current step is complete
	$: canProceed = (() => {
		const step = steps[currentStep];
		if (!step) return false;

		if (step.type === 'start') {
			return true; // Start step is always complete
		} else if (step.type === 'textarea') {
			return !!formData[step.field];
		} else if (step.type === 'file') {
			return !!formData[step.field];
		} else if (step.type === 'checkboxes') {
			return formData[step.field] && formData[step.field].length > 0;
		}
		return false;
	})();

	onMount(async () => {
		// Fetch onboarding config from settings
		try {
			const response = await fetch('/api/settings/onboarding');
			if (response.ok) {
				const data = await response.json();
				steps = data.data.steps || [];
			} else {
				error = 'Failed to load onboarding configuration';
			}
		} catch (err) {
			error = 'Failed to load onboarding configuration';
		}
	});

	function nextStep() {
		if (currentStep < steps.length - 1 && canProceed) {
			currentStep++;
		}
	}

	function prevStep() {
		if (currentStep > 0) {
			currentStep--;
		}
	}

	function toggleCheckbox(field, value) {
		if (!formData[field]) formData[field] = [];
		const index = formData[field].indexOf(value);
		if (index > -1) {
			formData[field] = formData[field].filter(v => v !== value);
		} else {
			formData[field] = [...formData[field], value];
		}
	}

	function handleClose() {
		pb.authStore.clear();
		navigate('login');
	}

	async function handleSubmit() {
		loading = true;
		error = '';

		try {
			const user = pb.authStore.record;
			if (!user) {
				error = 'User not authenticated';
				return;
			}

			// Build form data with avatar and data fields
			const formDataToSend = new FormData();

			// Collect all data fields (excluding avatar and start step)
			const dataFields = {};
			steps.forEach(step => {
				if (step.type !== 'start' && step.field !== 'avatar' && formData[step.field]) {
					dataFields[step.field] = formData[step.field];
				}
			});

			// Add data as JSON
			if (Object.keys(dataFields).length > 0) {
				formDataToSend.append('data', JSON.stringify(dataFields));
			}

			// Add avatar if present
			if (formData.avatar) {
				formDataToSend.append('avatar', formData.avatar);
			}

			// Update user record
			await pb.collection('users').update(user.id, formDataToSend);

			// Refresh auth to get updated user data
			await pb.collection('users').authRefresh();

			// Redirect based on user status
			const updatedUser = pb.authStore.record;
			if (updatedUser.status === 'pending') {
				navigate('pending-approval');
			} else {
				navigate('profile');
			}
		} catch (err) {
			error = err.message || 'Failed to save profile';
		} finally {
			loading = false;
		}
	}
</script>

<div class="onboarding-page">
	{#if steps[currentStep] && steps[currentStep].type !== 'start'}
		<nav class="top-nav">
			{#if currentStep === 0}
				<button class="nav-btn close" on:click={handleClose} disabled={loading}>
					✕
				</button>
			{:else}
				<button class="nav-btn" on:click={prevStep} disabled={loading}>
					← Indietro
				</button>
			{/if}
			<div class="step-counter">
				{realStepIndex}/{realSteps.length}
			</div>
			{#if currentStep < steps.length - 1}
				<button class="nav-btn next" on:click={nextStep} disabled={!canProceed || loading}>
					Avanti →
				</button>
			{:else}
				<button class="nav-btn next" on:click={handleSubmit} disabled={!canProceed || loading}>
					{loading ? 'Salvataggio...' : 'Completa'}
				</button>
			{/if}
		</nav>
	{/if}

	<div class="step-container" class:is-start={steps[currentStep]?.type === 'start'}>
		{#if steps[currentStep]}
			{@const step = steps[currentStep]}
			<div class="step-content" class:is-start={step.type === 'start'}>
				{#if step.type === 'start'}
					<button class="close-btn-start" on:click={handleClose} disabled={loading}>
						✕
					</button>
					<h1>{step.title}</h1>
					<p class="start-text">{@html step.text}</p>
					<Button variant="submit" on:click={nextStep}>
						{step.button}
					</Button>
				{:else}
					<h1>{step.title}</h1>
					{#if step.type === 'textarea'}
					<label for={step.id}>{step.label}</label>
					<textarea
						id={step.id}
						bind:value={formData[step.field]}
						on:keydown={(e) => {
							if (e.key === 'Enter' && !e.shiftKey && canProceed) {
								e.preventDefault();
								if (currentStep < steps.length - 1) {
									nextStep();
								} else {
									handleSubmit();
								}
							}
						}}
						enterkeyhint={currentStep < steps.length - 1 ? 'next' : 'done'}
						disabled={loading}
						rows="8"
						placeholder="Scrivi qui..."
					></textarea>
				{:else if step.type === 'file'}
					<label for={step.id}>{step.label}</label>
					<input
						id={step.id}
						type="file"
						accept="image/*"
						on:change={(e) => formData[step.field] = e.target.files[0]}
						disabled={loading}
					/>
					{#if formData[step.field]}
						<p class="file-name">{formData[step.field].name}</p>
					{/if}
				{:else if step.type === 'checkboxes'}
					<p class="field-label">{step.label}</p>
					<div class="checkboxes">
						{#each step.options as option}
							<label class="checkbox-label">
								<input
									type="checkbox"
									checked={formData[step.field]?.includes(option)}
									on:change={() => toggleCheckbox(step.field, option)}
									disabled={loading}
								/>
								{option}
							</label>
						{/each}
					</div>
				{/if}
				{/if}

				{#if step.type !== 'start'}
					<ErrorMessage {error} />
				{/if}
			</div>
		{/if}
	</div>
</div>

<style>
	.onboarding-page {
		min-height: 100vh;
		display: flex;
		flex-direction: column;
		background: #fff;
	}

	.top-nav {
		position: fixed;
		top: 0;
		left: 0;
		right: 0;
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: clamp(1rem, 3vw, 1.5rem);
		border-bottom: 2px solid #000;
		background: #fff;
		z-index: 100;
	}

	.step-counter {
		font-size: clamp(1rem, 3vw, 1.25rem);
		font-weight: bold;
		color: #000;
	}

	.nav-btn {
		background: transparent;
		border: 2px solid #000;
		padding: clamp(0.5rem, 2vw, 0.75rem) clamp(1rem, 3vw, 1.5rem);
		font-size: clamp(0.875rem, 2.5vw, 1rem);
		font-weight: 600;
		cursor: pointer;
		transition: all 0.2s ease;
		color: #000;
	}

	.nav-btn:hover:not(:disabled) {
		background: #000;
		color: #fff;
	}

	.nav-btn:disabled {
		opacity: 0.3;
		cursor: not-allowed;
	}

	.nav-btn.next {
		background: #000;
		color: #fff;
	}

	.nav-btn.next:hover:not(:disabled) {
		background: #333;
	}

	.nav-btn.next:disabled {
		background: #ccc;
		border-color: #ccc;
		color: #666;
	}

	.step-container {
		flex: 1;
		display: flex;
		align-items: center;
		justify-content: center;
		padding: clamp(1rem, 3vw, 2rem);
		padding-top: calc(70px + clamp(1rem, 3vw, 2rem));
		overflow-y: auto;
	}

	.step-container.is-start {
		padding-top: 0;
		min-height: 100vh;
	}

	.step-content {
		width: 100%;
		max-width: 32rem;
		background: #fff;
		border: 2px solid #000;
		padding: clamp(1.5rem, 4vw, 2.5rem);
		position: relative;
	}

	.step-content.is-start {
		border: none;
		text-align: center;
		max-width: 28rem;
	}

	h1 {
		margin: 0 0 clamp(1.5rem, 4vw, 2rem) 0;
		font-size: clamp(1.5rem, 5vw, 2rem);
		text-align: center;
		font-weight: bold;
		color: #000;
	}

	label,
	.field-label {
		display: block;
		margin-bottom: 0.75rem;
		font-weight: 600;
		color: #000;
		font-size: clamp(0.875rem, 2.5vw, 1rem);
	}

	textarea {
		width: 100%;
		padding: clamp(0.75rem, 2vw, 1rem);
		border: 2px solid #000;
		font-size: clamp(0.875rem, 2.5vw, 1rem);
		font-family: inherit;
		resize: vertical;
		min-height: 12rem;
	}

	textarea:focus {
		outline: 2px solid #000;
		outline-offset: -2px;
	}

	input[type="file"] {
		width: 100%;
		padding: 0.5rem;
		margin-bottom: 0.5rem;
		border: 2px solid #000;
		font-size: clamp(0.875rem, 2.5vw, 1rem);
	}

	.file-name {
		font-size: 0.875rem;
		color: #666;
		margin-top: 0.5rem;
	}

	.checkboxes {
		display: flex;
		flex-direction: column;
		gap: clamp(0.75rem, 2vw, 1rem);
	}

	.checkbox-label {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		cursor: pointer;
		font-weight: normal;
		font-size: clamp(0.875rem, 2.5vw, 1rem);
	}

	.checkbox-label input[type="checkbox"] {
		width: clamp(1.25rem, 3vw, 1.5rem);
		height: clamp(1.25rem, 3vw, 1.5rem);
		cursor: pointer;
	}

	.nav-btn.close {
		font-size: clamp(1.25rem, 3.5vw, 1.5rem);
		padding: clamp(0.25rem, 1.5vw, 0.5rem) clamp(0.75rem, 2.5vw, 1rem);
		line-height: 1;
	}

	.close-btn-start {
		position: fixed;
		top: clamp(1rem, 3vw, 1.5rem);
		left: clamp(1rem, 3vw, 1.5rem);
		background: transparent;
		border: 2px solid #000;
		font-size: clamp(1.5rem, 4vw, 2rem);
		padding: clamp(0.25rem, 2vw, 0.5rem) clamp(0.75rem, 3vw, 1rem);
		line-height: 1;
		cursor: pointer;
		font-weight: 600;
		color: #000;
		transition: all 0.2s ease;
		z-index: 1000;
	}

	.close-btn-start:hover:not(:disabled) {
		background: #000;
		color: #fff;
	}

	.close-btn-start:disabled {
		opacity: 0.3;
		cursor: not-allowed;
	}

	.start-text {
		text-align: center;
		font-size: clamp(1rem, 3vw, 1.25rem);
		color: #333;
		line-height: 1.6;
		margin: 0 0 2rem 0;
	}

	.step-content.is-start h1 {
		margin-top: 0;
		font-size: clamp(1.75rem, 5vw, 2.5rem);
	}
</style>
