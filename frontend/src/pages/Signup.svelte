<script>
	import { pb } from '../lib/pocketbase';
	import { navigate } from '../lib/router';
	import OnboardingWizard from '../components/onboarding/OnboardingWizard.svelte';

	async function handleSubmit({ values, formSteps }) {
		try {
			const payload = {
				...values,
				status: '0-pending',
			};

			await pb.collection('requests').create(payload);
			return null;
		} catch (err) {
			const fieldErrors = err?.data?.data || err?.data || {};
			const emailStepIndex = formSteps.findIndex(step => step.field === 'email');
			const emailError = fieldErrors?.email?.message || '';

			if (emailError && emailStepIndex >= 0) {
				return {
					emailError: formSteps[emailStepIndex]?.error || err.message || '',
					focusStepIndex: emailStepIndex,
				};
			}

			return { error: err.message || 'Failed to send request' };
		}
	}

	function handleClose() {
		navigate('login');
	}
</script>

<OnboardingWizard
	settingsKey="signup"
	onSubmit={handleSubmit}
	onClose={handleClose}
/>
