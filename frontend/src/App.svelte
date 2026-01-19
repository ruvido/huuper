<script>
	import { onMount } from 'svelte';
	import { isAuthenticated, pb, authRecord, fetchSetting } from './lib/pocketbase';
	import { currentRoute, navigate, queryParams, getTargetRoute } from './lib/router';
	import Header from './components/Header.svelte';
	import Menu from './components/Menu.svelte';
	import Login from './pages/Login.svelte';
	import Signup from './pages/Signup.svelte';
	import SignupDirect from './pages/SignupDirect.svelte';
	import PasswordReset from './pages/PasswordReset.svelte';
	import Onboarding from './pages/Onboarding.svelte';
	import PendingApproval from './pages/PendingApproval.svelte';
	import TelegramConnect from './pages/TelegramConnect.svelte';
	import Profile from './pages/Profile.svelte';
	import Groups from './pages/Groups.svelte';

	let menuOpen = false;
	let authReady = false;
	let renderReady = false;
	let appTitle = 'Members';
	const version = __APP_VERSION__;

	// Refresh auth on app load to sync with server
	onMount(async () => {
		try {
			const response = await fetchSetting('title');
			if (response.ok) {
				const data = await response.json();
				if (data?.data?.name) {
					appTitle = data.data.name;
					document.title = appTitle;
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

		if ($isAuthenticated && $authRecord?.status === 'suspended') {
			pb.authStore.clear();
			navigate('login');
			shouldRedirect = true;
		}

		const targetRoute = getTargetRoute($isAuthenticated, $authRecord, $currentRoute);
		if (targetRoute !== $currentRoute) {
			navigate(targetRoute);
			shouldRedirect = true;
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
	{#if $isAuthenticated && $currentRoute.startsWith('app/')}
		<Header onMenuClick={toggleMenu} title={appTitle} />
		<Menu isOpen={menuOpen} onClose={closeMenu} />
	{/if}

	<main>
		{#if $currentRoute === 'login'}
			<Login />
		{:else if $currentRoute === 'signup'}
			<Signup />
		{:else if $currentRoute === 'signup-direct'}
			<SignupDirect defaultStatus="active" showFooter={false} pageTitle="Sign Up (beta direct)" />
		{:else if $currentRoute === 'password-reset'}
			<PasswordReset />
		{:else if $currentRoute === 'onboarding'}
			<Onboarding />
		{:else if $currentRoute === 'pending-approval'}
			<PendingApproval />
		{:else if $currentRoute === 'telegram-connect'}
			<TelegramConnect />
		{:else if $currentRoute === 'app/profile'}
			<Profile />
		{:else if $currentRoute === 'app/groups'}
			<Groups />
		{:else}
			<Login />
		{/if}
	</main>
{/if}

<div class="version">{version}</div>

<style>
	:global(body) {
		overflow-x: hidden;
	}

	main {
		width: 100%;
		max-width: 100%;
	}

	.version {
		position: fixed;
		top: 8px;
		left: 8px;
		font-size: 10px;
		color: #999;
		z-index: 9999;
		pointer-events: none;
	}
</style>
