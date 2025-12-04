<script>
	import { pb } from '../lib/pocketbase';
	import { navigate } from '../lib/router';
	import AuthLayout from '../components/AuthLayout.svelte';
	import FormGroup from '../components/FormGroup.svelte';
	import Button from '../components/Button.svelte';
	import ErrorMessage from '../components/ErrorMessage.svelte';

	let email = '';
	let password = '';
	let passwordConfirm = '';
	let error = '';
	let loading = false;
	let emailError = '';
	let passwordError = '';
	let confirmError = '';

	async function handleSignup() {
		// Reset errors
		error = '';
		emailError = '';
		passwordError = '';
		confirmError = '';

		// Basic validation
		if (!email || !password || !passwordConfirm) {
			error = 'All fields are required';
			return;
		}

		if (password !== passwordConfirm) {
			confirmError = 'Passwords do not match';
			return;
		}

		if (password.length < 8) {
			passwordError = 'Password must be at least 8 characters';
			return;
		}

		loading = true;

		try {
			// Create user with active status
			const formData = new FormData();
			formData.append('email', email);
			formData.append('password', password);
			formData.append('passwordConfirm', passwordConfirm);
			formData.append('status', 'active');

			await pb.collection('users').create(formData);

			// Auto-login
			await pb.collection('users').authWithPassword(email, password);

			// Redirect to profile (will check for empty data there)
			navigate('profile');
		} catch (err) {
			// Parse PocketBase field-specific errors
			const fieldErrors = err.data?.data || err.data || {};

			if (fieldErrors.email) {
				const msg = fieldErrors.email.message || '';
				if (msg.includes('must be unique') || msg.includes('already exists')) {
					emailError = 'This email is already registered';
				} else if (msg.includes('invalid')) {
					emailError = 'Please enter a valid email address';
				} else {
					emailError = msg;
				}
			}
			if (fieldErrors.password) {
				passwordError = fieldErrors.password.message;
			}
			if (fieldErrors.passwordConfirm) {
				confirmError = fieldErrors.passwordConfirm.message;
			}
			// General error if no field-specific errors
			if (!emailError && !passwordError && !confirmError) {
				error = err.message || 'Signup failed';
			}
		} finally {
			loading = false;
		}
	}
</script>

<AuthLayout>
	<h1>Sign Up (beta mk2)</h1>

	<form on:submit|preventDefault={handleSignup}>
		<FormGroup
			id="email"
			type="email"
			label="Email"
			name="email"
			bind:value={email}
			bind:error={emailError}
			disabled={loading}
			required={true}
		/>

		<FormGroup
			id="password"
			type="password"
			label="Password"
			name="password"
			bind:value={password}
			bind:error={passwordError}
			disabled={loading}
			required={true}
		/>

		<FormGroup
			id="passwordConfirm"
			type="password"
			label="Confirm Password"
			name="passwordConfirm"
			bind:value={passwordConfirm}
			bind:error={confirmError}
			disabled={loading}
			required={true}
			matchField="password"
			matchValue={password}
		/>

		<ErrorMessage {error} />

		<Button variant="submit" type="submit" disabled={loading}>
			{loading ? 'Creating account...' : 'Sign Up'}
		</Button>
	</form>
</AuthLayout>

<style>
	h1 {
		margin: 0 0 1.5rem 0;
		font-size: 1.5rem;
		text-align: center;
		font-weight: bold;
	}
</style>
