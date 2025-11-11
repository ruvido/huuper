<script>
	import { onMount } from 'svelte';
	import { pb } from '../lib/pocketbase';
	import { navigate } from '../lib/router';
	import FormGroup from '../components/FormGroup.svelte';
	import Button from '../components/Button.svelte';
	import ErrorMessage from '../components/ErrorMessage.svelte';

	let steps = [];
	let currentStep = 0;
	let formData = {};
	let error = '';
	let loading = false;
	let container;

	// Parse URL parameters
	const urlParams = new URLSearchParams(window.location.hash.split('?')[1] || '');
	const directParam = urlParams.get('direct');
	const isDirect = directParam === 'true';

	onMount(async () => {
		// Fetch signup config from settings
		try {
			const response = await fetch('/api/settings/signup');
			if (response.ok) {
				const data = await response.json();
				steps = data.data.steps || [];
			}
		} catch (err) {
			error = 'Failed to load signup configuration';
		}

		// Add touch/swipe support
		let startY = 0;
		let startTime = 0;

		const handleTouchStart = (e) => {
			startY = e.touches[0].clientY;
			startTime = Date.now();
		};

		const handleTouchEnd = (e) => {
			const endY = e.changedTouches[0].clientY;
			const deltaY = startY - endY;
			const deltaTime = Date.now() - startTime;

			// Swipe threshold: 50px and less than 300ms
			if (Math.abs(deltaY) > 50 && deltaTime < 300) {
				if (deltaY > 0 && canGoNext()) {
					nextStep();
				} else if (deltaY < 0 && currentStep > 0) {
					prevStep();
				}
			}
		};

		if (container) {
			container.addEventListener('touchstart', handleTouchStart, { passive: true });
			container.addEventListener('touchend', handleTouchEnd, { passive: true });
		}

		return () => {
			if (container) {
				container.removeEventListener('touchstart', handleTouchStart);
				container.removeEventListener('touchend', handleTouchEnd);
			}
		};
	});

	function handleKeyDown(e) {
		// Skip if typing in fields (except Enter/Arrows)
		if ((e.target.tagName === 'INPUT' || e.target.tagName === 'TEXTAREA') &&
		    !['Enter', 'ArrowUp', 'ArrowDown'].includes(e.key)) {
			return;
		}

		if ((e.key === 'ArrowDown' || e.key === 'Enter') && currentStep < steps.length - 1 && canGoNext()) {
			e.preventDefault();
			nextStep();
		} else if (e.key === 'ArrowUp' && currentStep > 0) {
			e.preventDefault();
			prevStep();
		} else if (e.key === 'Enter' && currentStep === steps.length - 1 && canGoNext()) {
			e.preventDefault();
			handleSubmit();
		}
	}

	function canGoNext() {
		const step = steps[currentStep];
		if (!step) return false;

		if (step.type === 'form') {
			return step.fields.every(f => formData[f.name]);
		} else if (step.type === 'textarea') {
			return !!formData[step.field];
		} else if (step.type === 'file') {
			return !!formData[step.field];
		} else if (step.type === 'checkboxes') {
			return formData[step.field] && formData[step.field].length > 0;
		}
		return false;
	}

	function nextStep() {
		if (currentStep < steps.length - 1) {
			currentStep++;
			scrollToStep(currentStep);
		}
	}

	function prevStep() {
		if (currentStep > 0) {
			currentStep--;
			scrollToStep(currentStep);
		}
	}

	function scrollToStep(index) {
		if (container) {
			container.children[index]?.scrollIntoView({ behavior: 'smooth', block: 'start' });
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

	async function handleSubmit() {
		loading = true;
		error = '';

		try {
			// Build form data with all fields including avatar
			const formDataToSend = new FormData();
			formDataToSend.append('email', formData.email);
			formDataToSend.append('password', formData.password);
			formDataToSend.append('passwordConfirm', formData.password);
			formDataToSend.append('status', isDirect ? 'active' : 'pending');

			// Add data fields (why, hobbies, etc)
			const dataFields = {};
			steps.forEach(step => {
				if (step.field && step.field.startsWith('data.')) {
					const fieldName = step.field.replace('data.', '');
					dataFields[fieldName] = formData[step.field];
				}
			});
			if (Object.keys(dataFields).length > 0) {
				formDataToSend.append('data', JSON.stringify(dataFields));
			}

			// Add avatar if present
			if (formData.avatar) {
				formDataToSend.append('avatar', formData.avatar);
			}

			// Create user with all data in one request
			await pb.collection('users').create(formDataToSend);

			// Login
			await pb.collection('users').authWithPassword(formData.email, formData.password);

			navigate('profile');
		} catch (err) {
			error = err.message || 'Signup failed';
		} finally {
			loading = false;
		}
	}
</script>

<svelte:window on:keydown={handleKeyDown} />

<div class="container" bind:this={container}>
	{#each steps as step, i}
		<div class="step">
			<div class="step-content">
				<h1>{step.title}</h1>

				{#if step.type === 'form'}
					{#each step.fields as field}
						<FormGroup
							id={field.name}
							type={field.type}
							label={field.label}
							name={field.name}
							bind:value={formData[field.name]}
							disabled={loading}
						/>
					{/each}
				{:else if step.type === 'textarea'}
					<label for={step.id}>{step.label}</label>
					<textarea
						id={step.id}
						bind:value={formData[step.field]}
						disabled={loading}
						rows="6"
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

				<ErrorMessage {error} />

				{#if i === steps.length - 1}
					<Button variant="submit" on:click={handleSubmit} disabled={loading || !canGoNext()}>
						{loading ? 'Creating account...' : 'Complete Signup'}
					</Button>
				{/if}
			</div>
		</div>
	{/each}
</div>

<!-- Fixed arrow at bottom - only show if not last step and can go next -->
{#if currentStep < steps.length - 1 && canGoNext()}
	<button class="arrow-down-fixed" on:click={nextStep} aria-label="Next step">
		⬇
	</button>
{/if}

<style>
	.container {
		height: 100vh;
		overflow-y: auto;
		scroll-snap-type: y mandatory;
		scroll-behavior: smooth;
	}

	.step {
		height: 100vh;
		scroll-snap-align: start;
		display: flex;
		align-items: center;
		justify-content: center;
		padding: clamp(1rem, 3vw, 2rem);
	}

	.step-content {
		width: 100%;
		max-width: 28rem;
		background: #fff;
		border: 2px solid #000;
		padding: clamp(1.5rem, 4vw, 2rem);
		position: relative;
	}

	h1 {
		margin: 0 0 clamp(1.5rem, 4vw, 2rem) 0;
		font-size: clamp(1.5rem, 5vw, 2rem);
		text-align: center;
		font-weight: bold;
	}

	label,
	.field-label {
		display: block;
		margin-bottom: 0.5rem;
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
		margin-bottom: 1rem;
	}

	input[type="file"] {
		width: 100%;
		padding: 0.5rem;
		margin-bottom: 0.5rem;
	}

	.file-name {
		font-size: 0.875rem;
		color: #666;
		margin-bottom: 1rem;
	}

	.checkboxes {
		display: flex;
		flex-direction: column;
		gap: clamp(0.75rem, 2vw, 1rem);
		margin-bottom: clamp(1rem, 3vw, 1.5rem);
	}

	.checkbox-label {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		cursor: pointer;
		font-weight: normal;
	}

	.checkbox-label input[type="checkbox"] {
		width: clamp(1.25rem, 3vw, 1.5rem);
		height: clamp(1.25rem, 3vw, 1.5rem);
		cursor: pointer;
	}

	.arrow-down-fixed {
		position: fixed;
		bottom: clamp(2rem, 5vw, 3rem);
		left: 50%;
		transform: translateX(-50%);
		background: #000;
		color: #fff;
		border: none;
		width: clamp(3.5rem, 10vw, 4.5rem);
		height: clamp(3.5rem, 10vw, 4.5rem);
		border-radius: 50%;
		font-size: clamp(1.75rem, 5vw, 2.25rem);
		cursor: pointer;
		transition: transform 0.3s ease;
		display: flex;
		align-items: center;
		justify-content: center;
		z-index: 1000;
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
	}

	.arrow-down-fixed:hover {
		transform: translateX(-50%) scale(1.1);
	}

	.arrow-down-fixed:active {
		transform: translateX(-50%) scale(0.95);
	}
</style>
