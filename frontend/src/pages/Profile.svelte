<script>
	import { onMount } from 'svelte';
	import { pb } from '../lib/pocketbase';
	import { navigate } from '../lib/router';

	let user = pb.authStore.record;
	let telegramData = user?.telegram;
	let loading = false;
	let error = '';
	let successMessage = '';

	const BOT_NAME = import.meta.env.VITE_TELEGRAM_BOT_NAME || '@branco_realmen_bot';

	// Log auth URL for debugging
	$: if (user?.id) {
		const authUrl = `${window.location.origin}/api/telegram/callback?user_id=${user.id}`;
		console.log('=== TELEGRAM AUTH URL ===', authUrl);
	}

	onMount(async () => {
		// Check if returning from Telegram callback
		const urlParams = new URLSearchParams(window.location.hash.split('?')[1]);
		if (urlParams.get('telegram_linked') === 'true') {
			// Reload user data
			try {
				user = await pb.collection('users').getOne(user.id);
				telegramData = user?.telegram;
				successMessage = 'Telegram account linked successfully!';

				// Clean URL
				window.history.replaceState({}, '', '/#profile');

				// Clear success message after 3 seconds
				setTimeout(() => {
					successMessage = '';
				}, 3000);
			} catch (err) {
				error = 'Failed to load updated user data';
			}
		}
	});

	function goToGroups() {
		navigate('groups');
	}
</script>

<div class="page">
	<div class="container">
		<h1 class="page-title">Profile</h1>

		<div class="card">
			<div class="avatar">
				{user?.email?.charAt(0).toUpperCase()}
			</div>
			<h2>{user?.email}</h2>
			<p class="status">Admin</p>
		</div>

		<!-- Telegram Connection Section -->
		<div class="card">
			<h3>Telegram Account</h3>

			{#if telegramData}
				<!-- Show Telegram info when connected -->
				<div class="telegram-connected">
					<p class="telegram-info">
						{#if telegramData.username}
							<strong>@{telegramData.username}</strong>
						{:else}
							<strong>{telegramData.first_name} {telegramData.last_name || ''}</strong>
						{/if}
					</p>
					<p class="telegram-status">✓ Connected</p>
				</div>
			{:else}
				<!-- Show Telegram Login Widget when not connected -->
				<div class="telegram-login">
					<p>Connect your Telegram account to access private groups</p>
					<div id="telegram-login-container">
						<script
							async
							src="https://telegram.org/js/telegram-widget.js?22"
							data-telegram-login={BOT_NAME.replace('@', '')}
							data-size="large"
							data-auth-url={`${window.location.origin}/api/telegram/callback?user_id=${user.id}`}
							data-request-access="write"
						></script>
					</div>

					{#if error}
						<div class="error">{error}</div>
					{/if}

					{#if successMessage}
						<div class="success">{successMessage}</div>
					{/if}
				</div>
			{/if}
		</div>

		<button class="btn-primary" on:click={goToGroups}>
			View Groups
		</button>
	</div>
</div>

<style>
	.page {
		min-height: 100vh;
		background: #fff;
		padding: 5rem clamp(1rem, 4vw, 2rem) clamp(1rem, 4vw, 2rem);
	}

	.container {
		max-width: 50rem;
		margin: 0 auto;
		display: flex;
		flex-direction: column;
		gap: clamp(1.5rem, 4vw, 2rem);
	}

	.page-title {
		margin: 0 0 clamp(1rem, 3vw, 1.5rem) 0;
		font-size: 2rem;
		font-weight: bold;
		color: #000;
	}

	/* Card profilo */
	.card {
		background: #fff;
		border: 2px solid #000;
		padding: clamp(2rem, 5vw, 3rem) clamp(1.5rem, 4vw, 2rem);
		text-align: center;
	}

	.avatar {
		width: clamp(5rem, 15vw, 6.25rem);
		height: clamp(5rem, 15vw, 6.25rem);
		border: 3px solid #000;
		color: #000;
		font-size: clamp(2.5rem, 8vw, 3rem);
		display: flex;
		align-items: center;
		justify-content: center;
		margin: 0 auto clamp(1rem, 3vw, 1.5rem);
		font-weight: bold;
	}

	h2 {
		margin: 0 0 clamp(0.375rem, 1vw, 0.5rem) 0;
		color: #000;
		font-size: clamp(1.25rem, 4vw, 1.5rem);
		font-weight: bold;
		word-break: break-word;
	}

	.status {
		color: #000;
		font-weight: 600;
		margin: 0;
		font-size: clamp(0.875rem, 2.5vw, 1rem);
	}

	/* Button - touch target 48px */
	.btn-primary {
		width: 100%;
		padding: clamp(0.875rem, 3vw, 1rem);
		background: #000;
		color: #fff;
		border: 2px solid #000;
		font-size: clamp(1rem, 3vw, 1.125rem);
		font-weight: 600;
		cursor: pointer;
		touch-action: manipulation;
		transition: background 0.2s, color 0.2s;
	}

	.btn-primary:hover {
		background: #fff;
		color: #000;
	}

	/* Telegram Section */
	h3 {
		margin: 0 0 clamp(1rem, 3vw, 1.5rem) 0;
		font-size: clamp(1.125rem, 3.5vw, 1.25rem);
		color: #000;
		font-weight: bold;
	}

	.telegram-connected {
		text-align: center;
	}

	.telegram-info {
		margin: 0 0 clamp(0.5rem, 2vw, 0.75rem) 0;
		font-size: clamp(1rem, 3vw, 1.125rem);
		color: #000;
	}

	.telegram-status {
		margin: 0;
		font-size: clamp(0.875rem, 2.5vw, 0.9rem);
		color: #000;
		font-weight: 600;
	}

	.telegram-login {
		text-align: center;
	}

	.telegram-login p {
		margin: 0 0 clamp(1rem, 3vw, 1.5rem) 0;
		font-size: clamp(0.875rem, 2.5vw, 1rem);
		color: #000;
	}

	#telegram-login-container {
		display: flex;
		justify-content: center;
		margin: clamp(1rem, 3vw, 1.5rem) 0;
	}

	.error {
		color: #d00;
		font-size: clamp(0.875rem, 2.5vw, 1rem);
		margin-top: clamp(1rem, 3vw, 1.5rem);
	}

	.success {
		color: #070;
		font-size: clamp(0.875rem, 2.5vw, 1rem);
		margin-top: clamp(1rem, 3vw, 1.5rem);
		font-weight: 600;
	}
</style>
