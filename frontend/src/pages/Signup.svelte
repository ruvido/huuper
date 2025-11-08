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

	async function handleSignup() {
		if (!email || !password || !passwordConfirm) {
			error = 'All fields are required';
			return;
		}

		if (password !== passwordConfirm) {
			error = 'Passwords do not match';
			return;
		}

		if (password.length < 8) {
			error = 'Password must be at least 8 characters';
			return;
		}

		loading = true;
		error = '';

		try {
			await pb.collection('users').create({
				email,
				password,
				passwordConfirm,
			});

			await pb.collection('users').authWithPassword(email, password);

			navigate('profile');
		} catch (err) {
			error = err.message || 'Signup failed';
		} finally {
			loading = false;
		}
	}

	function goToLogin() {
		navigate('login');
	}
</script>

<AuthLayout>
	<h1>Create Account</h1>

	<form on:submit|preventDefault={handleSignup}>
		<FormGroup
			id="email"
			type="email"
			label="Email"
			name="email"
			bind:value={email}
			disabled={loading}
		/>

		<FormGroup
			id="password"
			type="password"
			label="Password"
			name="password"
			bind:value={password}
			disabled={loading}
		/>

		<FormGroup
			id="passwordConfirm"
			type="password"
			label="Confirm Password"
			name="passwordConfirm"
			bind:value={passwordConfirm}
			disabled={loading}
		/>

		<ErrorMessage {error} />

		<Button variant="submit" type="submit" disabled={loading}>
			{loading ? 'Creating account...' : 'Sign Up'}
		</Button>
	</form>

	<div class="footer">
		Already have an account?
		<Button variant="link" on:click={goToLogin} disabled={loading}>
			Login
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
