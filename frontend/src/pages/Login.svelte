<script>
	import { pb } from '../lib/pocketbase';
	import { navigate } from '../lib/router';

	let email = '';
	let password = '';
	let error = '';
	let loading = false;

	async function handleLogin() {
		if (!email || !password) {
			error = 'Email and password are required';
			return;
		}

		loading = true;
		error = '';

		try {
			await pb.collection('users').authWithPassword(email, password);
			navigate('profile');
		} catch (err) {
			error = err.message || 'Login failed';
		} finally {
			loading = false;
		}
	}

	function goToSignup() {
		navigate('signup');
	}
</script>

<div class="page">
	<div class="card">
		<h1>Login</h1>

		<form on:submit|preventDefault={handleLogin}>
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

			{#if error}
				<div class="error">{error}</div>
			{/if}

			<button type="submit" disabled={loading}>
				{loading ? 'Logging in...' : 'Login'}
			</button>
		</form>

		<div class="footer">
			Don't have an account?
			<button class="link-btn" on:click={goToSignup} disabled={loading}>
				Sign Up
			</button>
		</div>
	</div>
</div>
