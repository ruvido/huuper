<script>
	import { onMount } from 'svelte';
	import { pb, fetchSetting } from '../lib/pocketbase';
	import { navigate } from '../lib/router';
	import ProgressBar from '../components/ui/ProgressBar.svelte';
	import OnboardingNavigation from '../components/onboarding/OnboardingNavigation.svelte';
	import OnboardingStep from '../components/onboarding/OnboardingStep.svelte';
	import ConfirmationPage from '../components/onboarding/ConfirmationPage.svelte';
	import ErrorMessage from '../components/ErrorMessage.svelte';

	let steps = [];
	let currentStep = 0;
	let formData = {};
	let error = '';
	let emailError = '';
	let loading = false;
	let submitted = false;
	let emailCheck = { value: '', status: 'idle' };

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
		} else if (step.type === 'select') {
			const value = formData[step.field];
			if (step.min) {
				if (!value || value.length < step.min) return false;
				const hasInputOption = value.some(v => v.includes(':input'));
				if (hasInputOption) {
					return !!formData[step.field + '_custom']?.trim();
				}
				return true;
			} else if (step.max === 1 || !step.max) {
				const needsCustom = value?.includes?.(':input');
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
			const response = await fetchSetting('signup');
			if (response.ok) {
				const data = await response.json();
				steps = await hydrateSteps(data.data.steps || []);
			} else {
				error = 'Failed to load signup configuration';
			}
		} catch (err) {
			error = 'Failed to load signup configuration';
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
					value: record[valueField],
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

	async function handleSubmit() {
		loading = true;
		error = '';

		try {
			const payload = {};

			formSteps.forEach(step => {
				if (step.type === 'start') return;
				if (!step.field || !formData[step.field]) return;

				payload[step.field] = normalizeValue(step.field, formData[step.field]);
			});

			payload.status = '0-pending';

			await pb.collection('requests').create(payload);
			submitted = true;
		} catch (err) {
			const fieldErrors = err?.data?.data || err?.data || {};
			const emailStepIndex = formSteps.findIndex(step => step.field === 'email');
			const emailError = fieldErrors?.email?.message || '';

			if (emailError && emailStepIndex >= 0) {
				error = formSteps[emailStepIndex]?.error || err.message || '';
				currentStep = emailStepIndex;
			} else {
				error = err.message || 'Failed to send request';
			}
		} finally {
			loading = false;
		}
	}

	function handleClose() {
		navigate('login');
	}

</script>

<div class="onboarding-page">
	{#if formSteps[currentStep] && formSteps[currentStep].type !== 'start' && !submitted}
		<ProgressBar percentage={progressPercentage} />
		<OnboardingNavigation
			isFirstStep={currentStep === 0}
			isLastStep={currentStep === formSteps.length - 1}
			{canProceed}
			{loading}
			onBack={prevStep}
			onNext={nextStep}
			onClose={handleClose}
			onComplete={handleSubmit}
		/>
	{/if}

	{#if !submitted}
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
					onComplete={handleSubmit}
					onClose={handleClose}
					onToggleOption={toggleOption}
				/>
			{/if}
		</div>
	{/if}

	{#if submitted}
		<ConfirmationPage
			title={confirmationStep?.title || 'Request sent'}
			text={confirmationStep?.text || 'We will review your request and contact you soon.'}
			showButton={false}
			{loading}
		/>
	{/if}
</div>

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
