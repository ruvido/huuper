<script>
	import { pb } from '../lib/pocketbase';
	import { navigate } from '../lib/router';

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
			// Create user account
			await pb.collection('users').create({
				email,
				password,
				passwordConfirm,
			});

			// Auto login after signup
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

<div class="page">
	<div class="card">
		<h1>Create Account</h1>

		<form on:submit|preventDefault={handleSignup}>
			<div class="form-group">
				<label for="email">Email</label>
				<input
					id="email"
					type="email"
					bind:value={email}
					placeholder="your@email.com"
					disabled={loading}
					required
				/>
			</div>

			<div class="form-group">
				<label for="password">Password</label>
				<input
					id="password"
					type="password"
					bind:value={password}
					placeholder="••••••••"
					disabled={loading}
					required
				/>
			</div>

			<div class="form-group">
				<label for="passwordConfirm">Confirm Password</label>
				<input
					id="passwordConfirm"
					type="password"
					bind:value={passwordConfirm}
					placeholder="••••••••"
					disabled={loading}
					required
				/>
			</div>

			{#if error}
				<div class="error">{error}</div>
			{/if}

			<button type="submit" disabled={loading}>
				{loading ? 'Creating account...' : 'Sign Up'}
			</button>
		</form>

		<div class="footer">
			Already have an account?
			<button class="link-btn" on:click={goToLogin} disabled={loading}>
				Login
			</button>
		</div>
	</div>
</div>
