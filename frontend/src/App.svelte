<script>
	import { onMount } from 'svelte';
	import { isAuthenticated, pb } from './lib/pocketbase';
	import { currentRoute, navigate, queryParams } from './lib/router';
	import Header from './components/Header.svelte';
	import Menu from './components/Menu.svelte';
	import Login from './pages/Login.svelte';
	import Signup from './pages/Signup.svelte';
	import SignupSimple from './pages/SignupSimple.svelte';
	import Welcome from './pages/Welcome.svelte';
	import Onboarding from './pages/Onboarding.svelte';
	import PendingApproval from './pages/PendingApproval.svelte';
	import Profile from './pages/Profile.svelte';
	import Groups from './pages/Groups.svelte';

	let menuOpen = false;
	let authReady = false;

	// Check if simple signup requested
	$: showSimpleSignup = $queryParams.simple === 'true';

	// Refresh auth on app load to sync with server
	onMount(async () => {
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

	// Unified auth + status + onboarding guard - waits for auth to be ready
	$: if (authReady) {
		if (!$isAuthenticated) {
			// Not authenticated → login/signup only
			if ($currentRoute !== 'login' && $currentRoute !== 'signup') {
				navigate('login');
			}
		} else {
			// Authenticated → check status + data
			const user = pb.authStore.record;
			const status = user?.status;
			const hasData = user?.data && Object.keys(user.data).length > 0;

			if (status === 'pending') {
				// Pending users: onboarding → pending-approval
				if (!hasData && !['welcome', 'onboarding'].includes($currentRoute)) {
					navigate('welcome');
				} else if (hasData && $currentRoute !== 'pending-approval') {
					navigate('pending-approval');
				}
			} else if (status === 'active') {
				// Active users: cannot stay on pending-approval
				if ($currentRoute === 'pending-approval') {
					navigate('profile');
				}
				// Active users without data: onboarding required
				else if (!hasData && !['welcome', 'onboarding'].includes($currentRoute)) {
					navigate('welcome');
				}
			}
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

<!-- Header: only visible when authenticated and not on welcome/onboarding/pending-approval -->
{#if $isAuthenticated && !['welcome', 'onboarding', 'pending-approval'].includes($currentRoute)}
	<Header onMenuClick={toggleMenu} />
	<Menu isOpen={menuOpen} onClose={closeMenu} />
{/if}

<main>
	{#if $currentRoute === 'login'}
		<Login />
	{:else if $currentRoute === 'signup'}
		{#if showSimpleSignup}
			<SignupSimple />
		{:else}
			<Signup />
		{/if}
	{:else if $currentRoute === 'welcome'}
		<Welcome />
	{:else if $currentRoute === 'onboarding'}
		<Onboarding />
	{:else if $currentRoute === 'pending-approval'}
		<PendingApproval />
	{:else if $currentRoute === 'profile'}
		<Profile />
	{:else if $currentRoute === 'groups'}
		<Groups />
	{:else}
		<Login />
	{/if}
</main>

<style>
	:global(body) {
		overflow-x: hidden;
	}

	main {
		width: 100%;
		max-width: 100vw;
	}
</style>
