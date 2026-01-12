<script>
	import { onMount } from 'svelte';
	import { isAuthenticated, pb, authRecord } from './lib/pocketbase';
	import { currentRoute, navigate, queryParams } from './lib/router';
	import Header from './components/Header.svelte';
	import Menu from './components/Menu.svelte';
	import Login from './pages/Login.svelte';
	import Signup from './pages/Signup.svelte';
	import SignupDirect from './pages/SignupDirect.svelte';
	import Onboarding from './pages/Onboarding.svelte';
	import PendingApproval from './pages/PendingApproval.svelte';
	import TelegramConnect from './pages/TelegramConnect.svelte';
	import Profile from './pages/Profile.svelte';
	import Groups from './pages/Groups.svelte';

	let menuOpen = false;
	let authReady = false;
	let renderReady = false;

	// Refresh auth on app load to sync with server
	onMount(async () => {
		try {
			const response = await fetch('/api/settings/title');
			if (response.ok) {
				const data = await response.json();
				if (data?.data?.name) {
					document.title = data.data.name;
				}
			}
		} catch (err) {
			// Silently fail - title is optional
		}

		if (pb.authStore.isValid) {
			try {
				await pb.collection('users').authRefresh();
			} catch (err) {
				// Refresh failed - clear invalid auth
				pb.authStore.clear();
			}
		}
		authReady = true; // Signal auth is synced
	});

	// Reset renderReady when route changes to re-run guards
	$: if ($currentRoute) {
		renderReady = false;
	}

	// Guard logic - runs BEFORE allowing render
	$: if (authReady && !renderReady) {
		let shouldRedirect = false;

		if (!$isAuthenticated) {
			// Not authenticated → login/signup/signup-direct only
			if ($currentRoute !== 'login' && $currentRoute !== 'signup' && $currentRoute !== 'signup-direct') {
				navigate('login');
				shouldRedirect = true;
			}
		} else {
			// Authenticated → check status + data + telegram
			const user = $authRecord;
			const status = user?.status;
			const hasData = user?.data && Object.keys(user.data).length > 0;
			const hasTelegram = user?.telegram && Object.keys(user.telegram).length > 0;

			if (status === 'pending') {
				// Pending users: onboarding → pending-approval
				if (!hasData && $currentRoute !== 'onboarding') {
					navigate('onboarding');
					shouldRedirect = true;
				} else if (hasData && $currentRoute !== 'pending-approval') {
					navigate('pending-approval');
					shouldRedirect = true;
				}
			} else if (status === 'active') {
				// Active users: cannot stay on pending-approval
				if ($currentRoute === 'pending-approval') {
					navigate('profile');
					shouldRedirect = true;
				}
				// Active users without data: onboarding required
				else if (!hasData && $currentRoute !== 'onboarding') {
					navigate('onboarding');
					shouldRedirect = true;
				}
				// Active users with data but no telegram: telegram-connect required
				else if (hasData && !hasTelegram && $currentRoute !== 'telegram-connect') {
					navigate('telegram-connect');
					shouldRedirect = true;
				}
			}
		}

		// Only allow render if NOT redirecting
		if (!shouldRedirect) {
			renderReady = true;
		}
	}

	function toggleMenu() {
		menuOpen = !menuOpen;
	}

	function closeMenu() {
		menuOpen = false;
	}

	// Close menu when auth state changes (e.g., logout)
	$: if (!$isAuthenticated) {
		menuOpen = false;
	}
</script>

<!-- Only render when guards have validated -->
{#if renderReady}
	<!-- Header: only visible when authenticated and not on onboarding/pending-approval/telegram-connect -->
	{#if $isAuthenticated && !['onboarding', 'pending-approval', 'telegram-connect'].includes($currentRoute)}
		<Header onMenuClick={toggleMenu} />
		<Menu isOpen={menuOpen} onClose={closeMenu} />
	{/if}

	<main>
		{#if $currentRoute === 'login'}
			<Login />
		{:else if $currentRoute === 'signup'}
			<Signup />
		{:else if $currentRoute === 'signup-direct'}
			<SignupDirect />
		{:else if $currentRoute === 'onboarding'}
			<Onboarding />
		{:else if $currentRoute === 'pending-approval'}
			<PendingApproval />
		{:else if $currentRoute === 'telegram-connect'}
			<TelegramConnect />
		{:else if $currentRoute === 'profile'}
			<Profile />
		{:else if $currentRoute === 'groups'}
			<Groups />
		{:else}
			<Login />
		{/if}
	</main>
{/if}

<style>
	:global(body) {
		overflow-x: hidden;
	}

	main {
		width: 100%;
		max-width: 100%;
	}
</style>
