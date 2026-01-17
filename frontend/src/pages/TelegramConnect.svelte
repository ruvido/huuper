<script>
	import { onMount, onDestroy } from 'svelte';
	import { pb, fetchSetting } from '../lib/pocketbase';
	import { generateTelegramDeepLink } from '../lib/telegram';
	import { navigate, defaultAppRoute } from '../lib/router';
	import { renderContent } from '../lib/markdown';
	import Button from '../components/Button.svelte';
	import { X } from 'lucide-svelte';

	let connecting = false;
	let error = '';
	let botName = '';
	let unsubscribe;
	let config = null;
	let primaryLink = '';
	let fallbackLink = '';
	let helperTextHtml = '';

	onMount(async () => {
		// Fetch bot name
		try {
			const response = await fetchSetting('telegram');
			if (response.ok) {
				const data = await response.json();
				botName = data.data.name;
			}
		} catch (err) {
			// Silently fail - bot name is optional
		}

		// Fetch telegram_connect config
		try {
			const response = await fetchSetting('telegram_connect');
			if (response.ok) {
				const data = await response.json();
				config = data.data;
			}
		} catch (err) {
			// Silently fail
		}

		// Pre-generate links so the fallback is always visible
		if (botName) {
			await prepareLinks();
		}

		// Subscribe to user changes to detect when Telegram is connected
		try {
			const user = pb.authStore.record;
			unsubscribe = await pb.collection('users').subscribe(user.id, (e) => {
				const hasTelegram = e.record.telegram && Object.keys(e.record.telegram).length > 0;
				if (hasTelegram) {
					// Telegram connected! Sync to authStore immediately (official PocketBase pattern)
					pb.authStore.save(pb.authStore.token, e.record);
					// Now navigate - authStore updated synchronously
					navigate(defaultAppRoute);
				}
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

	function handleClose() {
		pb.authStore.clear();
		navigate('login');
	}

	async function handleConnect() {
		connecting = true;
		error = '';

		try {
			if (!primaryLink || !fallbackLink) {
				await prepareLinks();
			}
			if (!primaryLink) {
				throw new Error('Missing Telegram link');
			}
			window.open(primaryLink, '_blank', 'noopener');

		} catch (err) {
			error = err.message || 'Failed to connect Telegram';
			connecting = false;
		}
	}

	async function prepareLinks() {
		const { primary, fallback } = await generateTelegramDeepLink(botName);
		primaryLink = primary;
		fallbackLink = fallback;
		updateHelperText();
	}

	function updateHelperText() {
		if (!config?.helper_text) {
			helperTextHtml = '';
			return;
		}
		const linkValue = fallbackLink || '#';
		helperTextHtml = config.helper_text.replaceAll('{fallback_link}', linkValue);
	}

	$: if (config && fallbackLink) {
		updateHelperText();
	}
</script>

{#if config}
<div class="telegram-page">
	<button class="close-btn" on:click={handleClose}>
		<X size={24} />
	</button>
	<div class="content">
		<h1>{config.title}</h1>

		<div class="message">
			<div class="main">{@html renderContent(config.main_text)}</div>

			<div class="text">{@html renderContent(config.description)}</div>
		</div>

		{#if error}
			<p class="error">{error}</p>
		{/if}

		<Button variant="submit" on:click={handleConnect} disabled={connecting}>
			{connecting ? config.loading : config.button}
		</Button>
		{#if config.helper_text}
			<div class="helper">
				{@html renderContent(helperTextHtml)}
			</div>
		{/if}
	</div>
</div>
{/if}

<style>
	.telegram-page {
		height: 100vh;
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 2rem;
		background: #fff;
	}

	.content {
		text-align: center;
		max-width: 32rem;
	}

	h1 {
		margin: 0 0 2rem 0;
		font-size: clamp(1.5rem, 5vw, 2rem);
		font-weight: bold;
		color: #000;
		line-height: 1.3;
	}

	.message {
		margin-bottom: 2rem;
	}

	.message .main,
	.message .text {
		margin: 0 0 1.25rem 0;
		font-size: clamp(1rem, 3vw, 1.125rem);
		color: #333;
		line-height: 1.6;
	}

	.message .main {
		font-size: clamp(1.125rem, 3.5vw, 1.25rem);
		font-weight: 600;
		color: #000;
		margin-bottom: 1.5rem;
	}

	.close-btn {
		position: fixed;
		top: clamp(1rem, 3vw, 1.5rem);
		left: clamp(1rem, 3vw, 1.5rem);
		background: transparent;
		border: none;
		padding: clamp(0.25rem, 2vw, 0.5rem);
		line-height: 1;
		cursor: pointer;
		color: #000;
		transition: color 0.2s ease;
		z-index: 1000;
	}

	.close-btn:hover {
		color: #666;
	}

	.error {
		color: #d32f2f;
		font-size: clamp(0.875rem, 2.5vw, 1rem);
		margin: 1rem 0;
		text-align: center;
	}

	.helper {
		margin: 1rem 0 0.5rem;
		font-size: clamp(0.875rem, 2.5vw, 1rem);
		color: #444;
	}
</style>
