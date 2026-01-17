<script>
	import { pb } from '../lib/pocketbase';
	import { navigate, defaultAppRoute } from '../lib/router';
	import AuthLayout from '../components/AuthLayout.svelte';
	import FormGroup from '../components/FormGroup.svelte';
	import Button from '../components/Button.svelte';
	import ErrorMessage from '../components/ErrorMessage.svelte';

	let email = '';
	let password = '';
	let emailError = '';
	let passwordError = '';
	let error = '';
	let loading = false;

	async function handleLogin() {
		if (!email || !password) {
			error = 'Email and password are required';
			return;
		}

		loading = true;
		error = '';
		emailError = '';
		passwordError = '';

		try {
			await pb.collection('users').authWithPassword(email, password);
			// Navigate to default app route - App.svelte will handle onboarding redirect if needed
			navigate(defaultAppRoute);
		} catch (err) {
			// Parse PocketBase errors - err.data is alias for err.response
			if (err.status === 400) {
				error = 'Invalid email or password';
			} else {
				error = err.message || 'Login failed';
			}
		} finally {
			loading = false;
		}
	}

	function goToSignup() {
		navigate('signup');
	}
</script>

<AuthLayout>
	<h1>Login</h1>

	<form on:submit|preventDefault={handleLogin}>
		<FormGroup
			id="email"
			type="email"
			label="Email"
			name="email"
			bind:value={email}
			bind:error={emailError}
			disabled={loading}
		/>

		<FormGroup
			id="password"
			type="password"
			label="Password"
			name="password"
			bind:value={password}
			bind:error={passwordError}
			disabled={loading}
		/>

		<ErrorMessage {error} />

		<Button variant="submit" type="submit" disabled={loading}>
			{loading ? 'Logging in...' : 'Login'}
		</Button>
	</form>

	<div class="footer">
		Don't have an account?
		<Button variant="link" on:click={goToSignup} disabled={loading}>
			Sign Up
		</Button>
	</div>
</AuthLayout>

<style>
	h1 {
		margin: 0 0 1.5rem 0;
		font-size: 1.5rem;
		text-align: center;
		font-weight: bold;
	}

	.footer {
		margin-top: 1rem;
		text-align: center;
		font-size: 0.9rem;
	}
</style>
