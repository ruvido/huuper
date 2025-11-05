<script>
	import { onMount } from 'svelte';
	import { pb } from '../lib/pocketbase';
	import { navigate } from '../lib/router';

	let user = pb.authStore.record;
	let telegramData = user?.telegram;
	let loading = false;
	let error = '';

	const BOT_NAME = import.meta.env.VITE_TELEGRAM_BOT_NAME || '@branco_realmen_bot';

	onMount(() => {
		// Load Telegram Widget script
		const script = document.createElement('script');
		script.src = 'https://telegram.org/js/telegram-widget.js?22';
		script.async = true;
		document.body.appendChild(script);

		// Define global callback for Telegram widget
		window.onTelegramAuth = onTelegramAuth;

		return () => {
			delete window.onTelegramAuth;
		};
	});

	async function onTelegramAuth(telegramUser) {
		loading = true;
		error = '';

		try {
			const response = await fetch('/api/telegram/link', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
					'Authorization': pb.authStore.token,
				},
				body: JSON.stringify(telegramUser),
			});

			if (!response.ok) {
				const data = await response.text();
				throw new Error(data || 'Failed to link Telegram account');
			}

			const data = await response.json();
			telegramData = data.telegram;

			// Reload user data from PocketBase
			user = await pb.collection('_superusers').getOne(user.id);
		} catch (err) {
			error = err.message || 'Failed to link Telegram account';
			console.error('Telegram link error:', err);
		} finally {
			loading = false;
		}
	}

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
					{#if loading}
						<p>Connecting...</p>
					{:else}
						<p>Connect your Telegram account to access private groups</p>
						<div id="telegram-login-container">
							<script
								async
								src="https://telegram.org/js/telegram-widget.js?22"
								data-telegram-login={BOT_NAME.replace('@', '')}
								data-size="large"
								data-onauth="onTelegramAuth(user)"
								data-request-access="write"
							></script>
						</div>
					{/if}

					{#if error}
						<div class="error">{error}</div>
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
</style>
