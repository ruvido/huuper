<script>
	import { onMount } from 'svelte';
	import { pb, fetchSetting } from '../../lib/pocketbase';
	import ProgressBar from '../ui/ProgressBar.svelte';
	import OnboardingNavigation from './OnboardingNavigation.svelte';
	import OnboardingStep from './OnboardingStep.svelte';
	import CropModal from '../modals/CropModal.svelte';
	import ConfirmationPage from './ConfirmationPage.svelte';

	export let settingsKey = '';
	export let onSubmit;
	export let onClose;

	let steps = [];
	let currentStep = 0;
	let formData = {};
	let error = '';
	let emailError = '';
	let loading = false;
	let showConfirmation = false;
	let emailCheck = { value: '', status: 'idle' };

	// Crop state
	let showCropModal = false;
	let cropImage = '';
	let cropFile = null;
	let cropField = '';
	let cropModalRef;

	$: confirmationStep = steps.find(s => s.type === 'confirmation');
	$: formSteps = steps.filter(s => s.type !== 'confirmation');
	$: emailStep = formSteps.find(step => step.field === 'email' && step.check_unique);
	$: realSteps = formSteps.filter(s => s.type !== 'start');
	$: realStepIndex = (() => {
		let count = 0;
		for (let i = 0; i < currentStep; i++) {
			if (formSteps[i].type !== 'start') count++;
		}
		return count + 1;
	})();
	$: progressPercentage = realSteps.length > 0 ? (realStepIndex / realSteps.length) * 100 : 0;
	$: confirmationButtonVisible = !!confirmationStep?.button?.trim();
	$: requiresConfirmation = confirmationButtonVisible;

	$: displayError = (() => {
		if (emailStep?.field === formSteps[currentStep]?.field && emailStep?.check_unique) {
			return emailError;
		}
		return error;
	})();

	$: canProceed = (() => {
		const step = formSteps[currentStep];
		if (!step) return false;

		if (step.type === 'start') {
			return true;
		} else if (step.type === 'text') {
			const value = formData[step.field]?.trim();
			return !!value;
		} else if (step.type === 'textarea') {
			return !!formData[step.field]?.trim();
		} else if (step.type === 'file') {
			return !!formData[step.field];
		} else if (step.type === 'select') {
			const value = formData[step.field];
			if (step.min) {
				if (!value || value.length < step.min) return false;
				const hasInputOption = value.some(v => v.includes(':input'));
				if (hasInputOption) {
					return !!formData[step.field + '_custom']?.trim();
				}
				return true;
			}

			const needsCustom = value?.includes?.(':input');
			if (needsCustom) {
				return !!formData[step.field + '_custom']?.trim();
			}
			return !!value;
		}
		return false;
	})();

	onMount(async () => {
		try {
			const response = await fetchSetting(settingsKey);
			if (response.ok) {
				const data = await response.json();
				steps = await hydrateSteps(data.data.steps || []);
			} else {
				error = `Failed to load ${settingsKey} configuration`;
			}
		} catch (err) {
			error = `Failed to load ${settingsKey} configuration`;
		}
	});

	async function hydrateSteps(rawSteps) {
		const hydrated = [];
		for (const step of rawSteps) {
			if (!step?.options_source) {
				hydrated.push(step);
				continue;
			}

			const options = await loadOptions(step.options_source);
			hydrated.push({
				...step,
				options: options.length ? options : step.options || [],
			});
		}

		return hydrated;
	}

	async function loadOptions(source) {
		let collection = '';
		let field = 'name';
		let valueField = 'id';
		let sort = 'name';

		if (typeof source === 'string') {
			collection = source;
		} else if (source && typeof source === 'object') {
			collection = source.collection || '';
			field = source.field || field;
			valueField = source.value_field || valueField;
			sort = source.sort || sort;
		}

		if (!collection) return [];

		try {
			const records = await pb.collection(collection).getFullList({ sort });
			return records
				.map(record => ({
					label: record[field],
					value: record[valueField] ?? record[field],
				}))
				.filter(option => typeof option.label === 'string' && option.label.trim() !== '');
		} catch (err) {
			return [];
		}
	}

	function nextStep() {
		if (currentStep < formSteps.length - 1 && canProceed) {
			const step = formSteps[currentStep];
			if (step?.check_unique && step.field === 'email') {
				handleEmailCheck(step);
				return;
			}
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

		if (value?.includes(':input')) {
			setTimeout(() => {
				const customInput = document.querySelector('.custom-input');
				if (customInput) {
					customInput.scrollIntoView({ behavior: 'smooth', block: 'center' });
				}
			}, 350);
		}
	}

	function normalizeValue(field, value) {
		if (Array.isArray(value)) {
			return value.map(v => {
				if (v.includes(':input')) {
					return formData[field + '_custom'] || v.split(':')[0];
				}
				return v;
			});
		}

		if (value?.includes?.(':input')) {
			return formData[field + '_custom'] || value.split(':')[0];
		}

		return value;
	}

	function collectValues() {
		const values = {};
		formSteps.forEach(step => {
			if (step.type === 'start') return;
			if (!step.field || !formData[step.field]) return;
			values[step.field] = normalizeValue(step.field, formData[step.field]);
		});
		return values;
	}

	function looksLikeEmail(value) {
		return /.+@.+\..+/.test(value);
	}

	async function handleEmailCheck(step) {
		const value = formData[step.field]?.trim() || '';
		if (!value) return;

		if (!looksLikeEmail(value)) {
			emailCheck = { value, status: 'invalid' };
			emailError = step?.error_invalid || '';
			return;
		}

		if (emailCheck.value === value && emailCheck.status === 'valid') {
			currentStep++;
			return;
		}

		loading = true;
		emailError = '';
		emailCheck = { value, status: 'checking' };

		try {
			const response = await fetch('/api/signup/check-email', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
				},
				body: JSON.stringify({ email: value }),
			});

			if (!response.ok) {
				throw new Error('Email check failed');
			}

			const data = await response.json();
			if (data?.unique) {
				emailCheck = { value, status: 'valid' };
				emailError = '';
				currentStep++;
			} else {
				const message = step?.error || '';
				emailCheck = { value, status: 'invalid' };
				emailError = message;
			}
		} catch (err) {
			const message = step?.error_unavailable || '';
			emailCheck = { value, status: 'invalid' };
			emailError = message;
		} finally {
			loading = false;
		}
	}

	function handleComplete() {
		if (requiresConfirmation) {
			showConfirmation = true;
			return;
		}
		handleSubmit();
	}

	async function handleSubmit() {
		loading = true;
		error = '';
		emailError = '';

		try {
			const values = collectValues();
			const result = await onSubmit?.({ values, formSteps, formData, normalizeValue });
			if (result?.emailError) {
				emailError = result.emailError;
			}
			if (result?.focusField) {
				const index = formSteps.findIndex(step => step.field === result.focusField);
				if (index >= 0) currentStep = index;
			}
			if (result?.focusStepIndex != null) {
				currentStep = result.focusStepIndex;
			}
			if (result?.emailError && !result?.error) {
				return;
			}
			if (result?.error) {
				error = result.error;
				return;
			}

			showConfirmation = !!confirmationStep;
		} catch (err) {
			error = err?.message || 'Submission failed';
		} finally {
			loading = false;
		}
	}

	function handleFileSelect(file, field) {
		if (file?.type?.startsWith?.('image/')) {
			openCropModal(file, field);
			return;
		}
		formData[field] = file;
	}

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

			if (currentStep < formSteps.length - 1) {
				currentStep++;
			} else {
				handleComplete();
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
			onClose={onClose}
			onComplete={handleComplete}
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
					error={displayError}
					{canProceed}
					isLastStep={currentStep === formSteps.length - 1}
					onNext={nextStep}
					onComplete={handleComplete}
					onClose={onClose}
					onToggleOption={toggleOption}
					onFileSelect={handleFileSelect}
				/>
			{/if}
		</div>
	{/if}

	{#if showConfirmation && confirmationStep}
		<ConfirmationPage
			title={confirmationStep.title}
			text={confirmationStep.text}
			buttonText={confirmationStep.button}
			showButton={confirmationButtonVisible}
			{loading}
			onSubmit={handleSubmit}
		/>
	{/if}
</div>

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
