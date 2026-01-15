<script>
	import { onMount, onDestroy } from 'svelte';
	import { pb, fetchSetting } from '../lib/pocketbase';
	import { generateTelegramDeepLink } from '../lib/telegram';
	import { navigate } from '../lib/router';
	import DashboardLayout from '../components/DashboardLayout.svelte';
	import Card from '../components/Card.svelte';
	import Button from '../components/Button.svelte';
	import WelcomeModal from '../components/modals/WelcomeModal.svelte';
	import AccountStatusCard from '../components/cards/AccountStatusCard.svelte';
	import { Clock, Check, X } from 'lucide-svelte';

	let user = pb.authStore.record;
	let telegramData = user?.telegram;
	let userStatus = user?.status;
	let connecting = false;
	let error = '';
	let botName = '';
	let showWelcomeModal = false;
	let welcomeContent = '';
	let welcomeFetchInProgress = false;
	const WELCOME_STORAGE_KEY = 'profile_welcome_seen';

	let unsubscribe;

	onMount(async () => {
		// Refresh auth to get latest user data
		try {
			await pb.collection('users').authRefresh();
			user = pb.authStore.record;
			telegramData = user?.telegram;
			userStatus = user?.status;
		} catch (err) {
			// Silently fail - auth refresh is optional enhancement
		}

		maybeShowWelcomePopup();

		try {
			const response = await fetchSetting('telegram');
			if (response.ok) {
				const data = await response.json();
				botName = data.data.name;
			}
		} catch (err) {
			// Silently fail - bot name is optional
		}

		try {
			unsubscribe = await pb.collection('users').subscribe(user.id, (e) => {
				user = e.record;
				telegramData = e.record.telegram;
				userStatus = e.record.status;
				connecting = false;
				maybeShowWelcomePopup();
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
			const deepLink = await generateTelegramDeepLink(botName);
			window.open(deepLink, '_blank');

		} catch (err) {
			error = err.message || 'Failed to connect Telegram';
			connecting = false;
		}
	}

	function goToGroups() {
		navigate('app/groups');
	}

	function hasSeenWelcome() {
		if (typeof window === 'undefined') return true;
		return window.localStorage.getItem(WELCOME_STORAGE_KEY) === 'true';
	}

	function markWelcomeSeen() {
		if (typeof window === 'undefined') return;
		window.localStorage.setItem(WELCOME_STORAGE_KEY, 'true');
	}

	async function loadWelcomeMessage() {
		if (welcomeFetchInProgress) return;
		welcomeFetchInProgress = true;

		try {
			const response = await fetchSetting('welcome');
			if (!response.ok) return;
			const data = await response.json();
			const content = data?.data?.content;
			if (content) {
				welcomeContent = content;
				showWelcomeModal = true;
				markWelcomeSeen();
			}
		} catch (err) {
			// Ignore welcome fetch errors
		} finally {
			welcomeFetchInProgress = false;
		}
	}

	function maybeShowWelcomePopup() {
		if (userStatus !== 'active') return;
		if (showWelcomeModal) return;
		if (hasSeenWelcome()) return;
		loadWelcomeMessage();
	}

	function closeWelcomeModal() {
		showWelcomeModal = false;
		markWelcomeSeen();
	}
</script>

<DashboardLayout title="Profile">
	<Card>
		<div class="avatar">
			{user?.email?.charAt(0).toUpperCase()}
		</div>
		<h2>{user?.email}</h2>
		{#if user?.admin}
			<p class="status">Admin</p>
		{/if}
	</Card>

	<AccountStatusCard status={userStatus} />

	{#if userStatus === 'active'}
		<Card>
			<h3 class="section-title">Telegram Account</h3>

			{#if telegramData}
				<div class="telegram-connected">
					<p class="telegram-info">
						{#if telegramData.username}
							<strong>@{telegramData.username}</strong>
						{:else}
							<strong>{telegramData.first_name} {telegramData.last_name || ''}</strong>
						{/if}
					</p>
					<p class="telegram-status"><Check size={16} /> Connected</p>
				</div>
			{:else}
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
		</Card>
	{/if}

	<Button variant="primary" on:click={goToGroups}>
		View Groups
	</Button>

	<WelcomeModal
		show={showWelcomeModal}
		content={welcomeContent}
		onClose={closeWelcomeModal}
	/>
</DashboardLayout>

<style>
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

	.section-title {
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
