<script>
	import { onMount } from 'svelte';
	import { slide } from 'svelte/transition';
	import { pb } from '../lib/pocketbase';
	import { navigate } from '../lib/router';
	import Button from '../components/Button.svelte';
	import ErrorMessage from '../components/ErrorMessage.svelte';
	import { X, ArrowLeft, ArrowRight, Circle, CircleDot } from 'lucide-svelte';

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

	// Progress percentage
	$: progressPercentage = realSteps.length > 0 ? (realStepIndex / realSteps.length) * 100 : 0;

	// Check if current step is complete
	$: canProceed = (() => {
		const step = steps[currentStep];
		if (!step) return false;

		if (step.type === 'start') {
			return true;
		} else if (step.type === 'text') {
			return !!formData[step.field]?.trim();
		} else if (step.type === 'textarea') {
			return !!formData[step.field]?.trim();
		} else if (step.type === 'file') {
			return !!formData[step.field];
		} else if (step.type === 'select') {
			const value = formData[step.field];
			if (step.min) {
				// Multiple selection
				if (!value || value.length < step.min) return false;
				// Check if any selected option needs custom input
				const hasInputOption = value.some(v => v.includes(':input'));
				if (hasInputOption) {
					return !!formData[step.field + '_custom']?.trim();
				}
				return true;
			} else if (step.max === 1) {
				// Single selection
				const needsCustom = value?.includes(':input');
				if (needsCustom) {
					return !!formData[step.field + '_custom']?.trim();
				}
				return !!value;
			}
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

	function toggleOption(field, value, isMultiple) {
		if (isMultiple) {
			if (!formData[field]) formData[field] = [];
			const index = formData[field].indexOf(value);
			if (index > -1) {
				formData[field] = formData[field].filter(v => v !== value);
			} else {
				formData[field] = [...formData[field], value];
			}
		} else {
			formData[field] = value;
		}

		// If option needs custom input, scroll to show it
		if (value?.includes(':input')) {
			setTimeout(() => {
				const customInput = document.querySelector('.custom-input');
				if (customInput) {
					customInput.scrollIntoView({ behavior: 'smooth', block: 'center' });
				}
			}, 350); // After slide transition
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
					const value = formData[step.field];

					// Handle arrays (multiple selection)
					if (Array.isArray(value)) {
						dataFields[step.field] = value.map(v => {
							if (v.includes(':input')) {
								return formData[step.field + '_custom'] || v.split(':')[0];
							}
							return v;
						});
					}
					// Handle single values
					else if (value.includes?.(':input')) {
						dataFields[step.field] = formData[step.field + '_custom'] || value.split(':')[0];
					} else {
						dataFields[step.field] = value;
					}
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
		<div class="progress-bar">
			<div class="progress-fill" style="width: {progressPercentage}%"></div>
		</div>
		<nav class="top-nav">
			{#if currentStep === 0}
				<button class="nav-btn close" on:click={handleClose} disabled={loading}>
					<X size={20} />
				</button>
			{:else}
				<button class="nav-btn back" on:click={prevStep} disabled={loading}>
					<ArrowLeft size={20} />
				</button>
			{/if}
			<div class="nav-spacer"></div>
			{#if currentStep < steps.length - 1}
				<button class="nav-btn next" on:click={nextStep} disabled={!canProceed || loading}>
					Avanti <ArrowRight size={20} />
				</button>
			{:else}
				<button class="nav-btn next" on:click={handleSubmit} disabled={!canProceed || loading}>
					{loading ? 'Salvataggio...' : 'Completa'}
				</button>
			{/if}
		</nav>
	{/if}

	<div class="step-container"
	     class:is-start={steps[currentStep]?.type === 'start'}
	     class:is-list={steps[currentStep]?.type === 'select'}>
		{#if steps[currentStep]}
			{@const step = steps[currentStep]}
			<div class="step-content" class:is-start={step.type === 'start'}>
				{#if step.type === 'start'}
					<button class="close-btn-start" on:click={handleClose} disabled={loading}>
						<X size={24} />
					</button>
					<h1>{step.title}</h1>
					<p class="start-text">{@html step.text}</p>
					<Button variant="submit" on:click={nextStep}>
						{step.button}
					</Button>
				{:else}
					<h1>{step.title}</h1>
					{#if step.type === 'text'}
						<label for={step.id}>{step.label}</label>
						<input
							id={step.id}
							type="text"
							bind:value={formData[step.field]}
							on:keydown={(e) => {
								if (e.key === 'Enter' && canProceed) {
									e.preventDefault();
									nextStep();
								}
							}}
							disabled={loading}
							placeholder="Scrivi qui..."
						/>
					{:else if step.type === 'textarea'}
						<label for={step.id}>{step.label}</label>
						<textarea
							id={step.id}
							bind:value={formData[step.field]}
							disabled={loading}
							rows="8"
							placeholder="Scrivi qui..."
						></textarea>
					{:else if step.type === 'file'}
						<h2 class="file-label">{step.label}</h2>
						<input
							id={step.id}
							type="file"
							accept="image/*"
							on:change={(e) => formData[step.field] = e.target.files[0]}
							disabled={loading}
							style="display: none;"
						/>
						<button
							type="button"
							class="file-button"
							on:click={() => document.getElementById(step.id).click()}
							disabled={loading}
						>
							{formData[step.field] ? 'Cambia foto' : 'Carica foto'}
						</button>
						{#if formData[step.field]}
							<p class="file-name">{formData[step.field].name}</p>
						{/if}
					{:else if step.type === 'select'}
						{@const isMultiple = !!step.min}
						{@const selectedCount = formData[step.field]?.length || 0}
						{@const remaining = step.min ? Math.max(0, step.min - selectedCount) : 0}
						<p class="field-label" class:invisible={remaining === 0}>
							Seleziona almeno {remaining || step.min}
						</p>
						<div class="grid-container" on:keydown={(e) => {
							if (e.key === 'Enter' && canProceed) {
								e.preventDefault();
								if (currentStep < steps.length - 1) {
									nextStep();
								} else {
									handleSubmit();
								}
							}
						}}>
							{#each step.options as option}
								{@const needsInput = option.includes(':input')}
								{@const displayText = needsInput ? option.split(':')[0] : option}
								{@const isSelected = isMultiple
									? formData[step.field]?.includes(option)
									: formData[step.field] === option}
								<button
									type="button"
									class="grid-box"
									class:selected={isSelected}
									on:click={() => toggleOption(step.field, option, isMultiple)}
									disabled={loading}
								>
									{displayText}
								</button>
							{/each}
						</div>
						{#key step.field}
							{@const hasInputOption = isMultiple
								? formData[step.field]?.some(v => v.includes(':input'))
								: formData[step.field]?.includes(':input')}
							{#if hasInputOption}
								<div class="custom-input-container" transition:slide={{ duration: 300 }}>
									<input
										type="text"
										class="custom-input"
										bind:value={formData[step.field + '_custom']}
										placeholder="Specifica..."
										on:keydown={(e) => {
											if (e.key === 'Enter' && canProceed) {
												e.preventDefault();
												if (currentStep < steps.length - 1) {
													nextStep();
												} else {
													handleSubmit();
												}
											}
										}}
										autofocus
										disabled={loading}
									/>
								</div>
							{/if}
						{/key}
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

	.progress-bar {
		position: fixed;
		top: 0;
		left: 0;
		right: 0;
		height: 3px;
		background: #f0f0f0;
		z-index: 101;
	}

	.progress-fill {
		height: 100%;
		background: #000;
		transition: width 0.3s ease;
	}

	.top-nav {
		position: fixed;
		top: 3px;
		left: 0;
		right: 0;
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: clamp(1rem, 3vw, 1.5rem);
		background: #fff;
		z-index: 100;
	}

	.nav-spacer {
		flex: 1;
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
		display: flex;
		align-items: center;
		gap: 0.5rem;
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

	.nav-btn.back {
		border: none;
		padding: clamp(0.25rem, 1.5vw, 0.5rem);
	}

	.nav-btn.back:hover:not(:disabled) {
		background: transparent;
		color: #666;
	}

	.step-container {
		flex: 1;
		display: flex;
		align-items: center;
		justify-content: center;
		padding: clamp(1rem, 3vw, 2rem);
		padding-top: clamp(5rem, 12vw, 6rem);
		overflow-y: auto;
	}

	.step-container.is-start {
		padding-top: 0;
		min-height: 100vh;
	}

	.step-container.is-list {
		align-items: flex-start;
	}

	.step-content {
		width: 100%;
		max-width: 32rem;
		background: #fff;
		padding: clamp(1.5rem, 4vw, 2.5rem);
		position: relative;
	}

	.step-content.is-start {
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

	.field-label.invisible {
		visibility: hidden;
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

	.file-label {
		display: block;
		margin-bottom: 1.5rem;
		font-weight: 600;
		color: #000;
		font-size: clamp(0.875rem, 2.5vw, 1rem);
	}

	.file-button {
		width: 100%;
		padding: clamp(1rem, 3vw, 1.25rem);
		background: #000;
		color: #fff;
		border: 2px solid #000;
		font-size: clamp(0.875rem, 2.5vw, 1rem);
		font-weight: 600;
		font-family: inherit;
		cursor: pointer;
		transition: all 0.2s ease;
	}

	.file-button:hover:not(:disabled) {
		background: #333;
	}

	.file-button:disabled {
		background: #ccc;
		border-color: #ccc;
		color: #666;
		cursor: not-allowed;
	}

	.file-name {
		font-size: 0.875rem;
		color: #666;
		margin-top: 0.75rem;
		text-align: center;
	}

	input[type="text"] {
		width: 100%;
		padding: clamp(0.75rem, 2vw, 1rem);
		border: 2px solid #000;
		font-size: clamp(0.875rem, 2.5vw, 1rem);
		font-family: inherit;
	}

	input[type="text"]:focus {
		outline: 2px solid #000;
		outline-offset: -2px;
	}

	.custom-input-container {
		margin-top: clamp(1rem, 3vw, 1.5rem);
	}

	.custom-input {
		width: 100%;
		padding: clamp(0.75rem, 2vw, 1rem);
		border: 2px solid #000;
		font-size: clamp(0.875rem, 2.5vw, 1rem);
		font-family: inherit;
	}

	.custom-input:focus {
		outline: 2px solid #000;
		outline-offset: -2px;
	}

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
		background: #fff;
		color: #000;
		font-size: clamp(0.875rem, 2.5vw, 1rem);
		font-weight: 500;
		font-family: inherit;
		line-height: 1.3;
		cursor: pointer;
		transition: all 0.15s ease;
		text-align: center;
		height: clamp(4rem, 10vw, 5rem);
		display: flex;
		align-items: center;
		justify-content: center;
		word-wrap: break-word;
		hyphens: auto;
	}

	.grid-box:hover:not(:disabled) {
		background: #f5f5f5;
	}

	.grid-box.selected {
		background: #000;
		color: #fff;
	}

	.grid-box:disabled {
		opacity: 0.5;
		cursor: not-allowed;
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
		border: none;
		padding: clamp(0.25rem, 2vw, 0.5rem);
		line-height: 1;
		cursor: pointer;
		color: #000;
		transition: color 0.2s ease;
		z-index: 1000;
	}

	.close-btn-start:hover:not(:disabled) {
		color: #666;
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
