<script>
	import { onMount } from 'svelte';
	import { pb, fetchSetting } from '../lib/pocketbase';
	import { navigate, defaultAppRoute } from '../lib/router';
	import ProgressBar from '../components/ui/ProgressBar.svelte';
	import OnboardingNavigation from '../components/onboarding/OnboardingNavigation.svelte';
	import OnboardingStep from '../components/onboarding/OnboardingStep.svelte';
	import CropModal from '../components/modals/CropModal.svelte';
	import ConfirmationPage from '../components/onboarding/ConfirmationPage.svelte';

	let steps = [];
	let confirmation = null;
	let currentStep = 0;
	let formData = {};
	let error = '';
	let loading = false;
	let showConfirmation = false;

	$: formSteps = steps;

	// Crop state
	let showCropModal = false;
	let cropImage = '';
	let cropFile = null;
	let cropField = '';
	let cropModalRef;
	const nameFieldKey = 'full_name';

	// Count only non-start steps
	$: realSteps = formSteps.filter(s => s.type !== 'start');
	$: realStepIndex = (() => {
		let count = 0;
		for (let i = 0; i < currentStep; i++) {
			if (formSteps[i].type !== 'start') count++;
		}
		return count + 1; // 1-indexed
	})();

	// Progress percentage
	$: progressPercentage = realSteps.length > 0 ? (realStepIndex / realSteps.length) * 100 : 0;

	// Check if current step is complete
	$: canProceed = (() => {
		const step = formSteps[currentStep];
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
		try {
			const [onboardingResponse, profileResponse] = await Promise.all([
				fetchSetting('onboarding'),
				fetchSetting('profile_schema'),
			]);

			if (!onboardingResponse.ok || !profileResponse.ok) {
				error = 'Failed to load onboarding configuration';
				return;
			}

			const onboardingPayload = await onboardingResponse.json();
			const profilePayload = await profileResponse.json();

			const onboardingData = onboardingPayload?.data || {};
			const profileData = profilePayload?.data || {};
			steps = buildOnboardingSteps(onboardingData, profileData);
			confirmation = onboardingData.confirmation || null;

			if (steps.length === 0) {
				error = 'Onboarding configuration has no steps';
			}
		} catch {
			error = 'Failed to load onboarding configuration';
		}
	});

	function normalizeFieldType(type) {
		const value = (type || '').toLowerCase();
		if (value === 'phone') return 'text';
		if (value === 'email') return 'text';
		return value;
	}

	function buildOnboardingSteps(onboardingData, profileData) {
		const out = [];
		const profileFields = Array.isArray(profileData?.fields) ? profileData.fields : [];
		const byKey = new Map();
		for (const field of profileFields) {
			const key = field?.key;
			if (typeof key === 'string' && key.trim() !== '') {
				byKey.set(key, field);
			}
		}

		const startPage = onboardingData?.start_page;
		if (startPage && typeof startPage === 'object') {
			out.push({
				id: 'start',
				type: 'start',
				title: startPage.title || '',
				text: startPage.text || '',
				button: startPage.button || 'Inizia',
			});
		}

		const rawSteps = Array.isArray(onboardingData?.steps) ? onboardingData.steps : [];
		for (const rawStep of rawSteps) {
			const key = typeof rawStep?.field === 'string' ? rawStep.field.trim() : '';
			if (!key) continue;

			const field = byKey.get(key);
			if (!field) continue;

			out.push({
				id: key,
				field: key,
				type: normalizeFieldType(field.type),
				title: rawStep.title || field.title || field.label || '',
				label: rawStep.label || field.label || '',
				options: Array.isArray(field.options) ? field.options : [],
				min: Number.isFinite(field.min) ? field.min : undefined,
				max: Number.isFinite(field.max) ? field.max : undefined,
			});
		}

		return out;
	}

	function nextStep() {
		if (currentStep < formSteps.length - 1 && canProceed) {
			currentStep++;
		}
	}

	async function completeStep() {
		if (!canProceed || loading) return;
		if (confirmation) {
			showConfirmation = true;
			return;
		}
		await handleSubmit();
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

	function normalizePersonName(value) {
		if (!value || typeof value !== 'string') return '';
		return value
			.trim()
			.toLowerCase()
			.split(/\s+/)
			.filter(Boolean)
			.map((part) => part.charAt(0).toUpperCase() + part.slice(1))
			.join(' ');
	}

	function handleClose() {
		pb.authStore.clear();
		navigate('login');
	}

	// Crop helpers
	function openCropModal(file, field) {
		cropFile = file;
		cropField = field;

		const reader = new FileReader();
		reader.onload = (e) => {
			cropImage = e.target.result;
			showCropModal = true;
		};
		reader.readAsDataURL(file);
	}

	async function handleCropConfirm() {
		if (!cropModalRef || !cropFile) return;

		loading = true;
		error = '';

		try {
			const croppedFile = await cropModalRef.processCrop(cropFile);
			formData[cropField] = croppedFile;

			showCropModal = false;
			cropImage = '';
			cropFile = null;

				// Auto-advance after crop
				if (currentStep < formSteps.length - 1) {
					currentStep++;
				} else {
					await completeStep();
				}
		} catch (err) {
			error = 'Errore nel processare l\'immagine';
		} finally {
			loading = false;
		}
	}

	function handleCropCancel() {
		showCropModal = false;
		cropImage = '';
		cropFile = null;
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
			formSteps.forEach(step => {
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

			if (typeof dataFields[nameFieldKey] === 'string') {
				const normalized = normalizePersonName(dataFields[nameFieldKey]);
				if (normalized) {
					dataFields[nameFieldKey] = normalized;
				}
			}

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
				navigate(defaultAppRoute);
			}
		} catch (err) {
			error = err.message || 'Failed to save profile';
		} finally {
			loading = false;
		}
	}
</script>

<div class="onboarding-page">
	{#if formSteps[currentStep] && formSteps[currentStep].type !== 'start' && !showConfirmation}
		<ProgressBar percentage={progressPercentage} />
		<OnboardingNavigation
			isFirstStep={currentStep === 0}
			isLastStep={currentStep === formSteps.length - 1}
			{canProceed}
			{loading}
			onBack={prevStep}
			onNext={nextStep}
				onClose={handleClose}
				onComplete={completeStep}
			/>
	{/if}

	{#if !showConfirmation}
		<div class="step-container"
		     class:is-start={formSteps[currentStep]?.type === 'start'}
		     class:is-list={formSteps[currentStep]?.type === 'select'}>
			{#if formSteps[currentStep]}
				<OnboardingStep
					step={formSteps[currentStep]}
					bind:formData
					{loading}
					{error}
					{canProceed}
					isLastStep={currentStep === formSteps.length - 1}
						onNext={nextStep}
						onComplete={completeStep}
						onClose={handleClose}
						onToggleOption={toggleOption}
						onFileSelect={openCropModal}
				/>
			{/if}
		</div>
	{/if}

	<!-- Confirmation Page -->
	{#if showConfirmation && confirmation}
		<ConfirmationPage
			title={confirmation.title}
			text={confirmation.text}
			buttonText={confirmation.button}
			{loading}
			onSubmit={handleSubmit}
		/>
	{/if}
</div>

<!-- Crop Modal -->
<CropModal
	bind:this={cropModalRef}
	show={showCropModal}
	image={cropImage}
	onConfirm={handleCropConfirm}
	onCancel={handleCropCancel}
/>

<style>
	.onboarding-page {
		min-height: 100vh;
		max-width: 100%;
		overflow-x: hidden;
		display: flex;
		flex-direction: column;
		background: #fff;
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
</style>
