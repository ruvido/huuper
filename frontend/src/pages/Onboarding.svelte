<script>
	import { pb } from '../lib/pocketbase';
	import { navigate, defaultAppRoute } from '../lib/router';
	import OnboardingWizard from '../components/onboarding/OnboardingWizard.svelte';

	async function handleSubmit({ values, formSteps, formData }) {
		try {
			const user = pb.authStore.record;
			if (!user) {
				return { error: 'User not authenticated' };
			}

			const formDataToSend = new FormData();
			const dataFields = {};

			formSteps.forEach(step => {
				if (step.type === 'start' || step.type === 'file') return;
				if (!step.field || !values[step.field]) return;
				dataFields[step.field] = values[step.field];
			});

			if (Object.keys(dataFields).length > 0) {
				formDataToSend.append('data', JSON.stringify(dataFields));
			}

			if (formData.avatar) {
				formDataToSend.append('avatar', formData.avatar);
			}

			await pb.collection('users').update(user.id, formDataToSend);
			await pb.collection('users').authRefresh();

			navigate(defaultAppRoute);
			return null;
		} catch (err) {
			return { error: err.message || 'Failed to save profile' };
		}
	}

	function handleClose() {
		pb.authStore.clear();
		navigate('login');
	}
</script>

<OnboardingWizard
	settingsKey="onboarding"
	onSubmit={handleSubmit}
	onClose={handleClose}
/>
