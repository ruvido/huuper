<script>
	import { onMount, onDestroy } from 'svelte';
	import { pb } from '../lib/pocketbase';
	import { navigate } from '../lib/router';

	let user = pb.authStore.record;
	let telegramData = user?.telegram;
	let connecting = false;
	let error = '';
	let botName = '';

	let unsubscribe;

	onMount(async () => {
		// Fetch bot name from settings
		try {
			const response = await fetch('/api/settings/telegram');
			if (response.ok) {
				const data = await response.json();
				botName = data.data.name;
			}
		} catch (err) {
			// Silently fail - bot name is optional
		}

		// Subscribe to user record changes for realtime updates
		try {
			unsubscribe = await pb.collection('users').subscribe(user.id, (e) => {
				user = e.record;
				telegramData = e.record.telegram;
				connecting = false;
			});
		} catch (err) {
			// Silently fail - subscription is optional enhancement
		}
	});

	onDestroy(() => {
		if (unsubscribe) {
			unsubscribe();
		}
	});

	async function connectTelegram() {
		connecting = true;
		error = '';

		try {
			// Generate token
			const response = await fetch('/api/telegram/generate-token', {
				method: 'POST',
				headers: {
					'Authorization': pb.authStore.token,
				},
			});

			if (!response.ok) {
				throw new Error('Failed to generate connection token');
			}

			const data = await response.json();
			const token = data.token;

			// Open Telegram bot with deep link
			const cleanBotName = botName.replace('@', '');
			const deepLink = `https://t.me/${cleanBotName}?start=${token}`;
			window.open(deepLink, '_blank');

		} catch (err) {
			error = err.message || 'Failed to connect Telegram';
			connecting = false;
		}
	}

	function goToGroups() {
		navigate('groups');
	}
</script>

<div class="dashboard-page">
	<div class="dashboard-container">
		<h1 class="dashboard-title">Profile</h1>

		<div class="dashboard-card">
			<div class="avatar">
				{user?.email?.charAt(0).toUpperCase()}
			</div>
			<h2>{user?.email}</h2>
			<p class="status">Admin</p>
		</div>

		<!-- Telegram Connection Section -->
		<div class="dashboard-card">
			<h3 class="section-title">Telegram Account</h3>

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
				<!-- Telegram not connected -->
				<div class="telegram-connect">
					{#if connecting}
						<p class="connecting-message">Waiting for connection...</p>
						<p class="help-text">Complete the connection in Telegram</p>
					{:else}
						<p>Connect your Telegram account to access private groups</p>
						<button class="btn-telegram" on:click={connectTelegram}>
							Connect Telegram
						</button>
					{/if}

					{#if error}
						<p class="error-message">{error}</p>
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
	/* Component-specific styles only */
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

	/* Telegram Section */
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

	.telegram-connect {
		text-align: center;
	}

	.telegram-connect p {
		margin: 0 0 clamp(1rem, 3vw, 1.5rem) 0;
		font-size: clamp(0.875rem, 2.5vw, 1rem);
		color: #000;
	}

	.btn-telegram {
		width: 100%;
		padding: clamp(0.875rem, 3vw, 1rem);
		background: #0088cc;
		color: #fff;
		border: 2px solid #0088cc;
		font-size: clamp(1rem, 3vw, 1.125rem);
		font-weight: 600;
		cursor: pointer;
		touch-action: manipulation;
		transition: background 0.2s, color 0.2s;
	}

	.btn-telegram:hover {
		background: #006699;
		border-color: #006699;
	}

	.connecting-message {
		font-weight: 600;
		color: #0088cc;
	}

	.help-text {
		font-size: clamp(0.75rem, 2vw, 0.875rem) !important;
		color: #666 !important;
	}

	.error-message {
		color: #d00;
		font-size: clamp(0.875rem, 2.5vw, 1rem);
		margin-top: clamp(1rem, 3vw, 1.5rem) !important;
	}

</style>
