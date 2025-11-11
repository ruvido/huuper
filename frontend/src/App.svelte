<script>
	import { isAuthenticated } from './lib/pocketbase';
	import { currentRoute, navigate } from './lib/router';
	import Header from './components/Header.svelte';
	import Menu from './components/Menu.svelte';
	import Login from './pages/Login.svelte';
	import Signup from './pages/Signup.svelte';
	import SignupMultistep from './pages/SignupMultistep.svelte';
	import Profile from './pages/Profile.svelte';
	import Groups from './pages/Groups.svelte';

	let menuOpen = false;

	// Check if signup should be multistep - reactive to both route and hash changes
	$: isMultiStepSignup = (() => {
		if ($currentRoute !== 'signup') return false;
		const hashParts = window.location.hash.split('?');
		const params = hashParts[1] || '';
		const urlParams = new URLSearchParams(params);
		return urlParams.get('multi') === 'true';
	})();

	// Check auth and redirect if needed
	$: {
		if (!$isAuthenticated && $currentRoute !== 'login' && $currentRoute !== 'signup') {
			navigate('login');
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

<!-- Header: only visible when authenticated -->
{#if $isAuthenticated}
	<Header onMenuClick={toggleMenu} />
	<Menu isOpen={menuOpen} onClose={closeMenu} />
{/if}

<main>
	{#if $currentRoute === 'login'}
		<Login />
	{:else if $currentRoute === 'signup'}
		{#if isMultiStepSignup}
			<SignupMultistep />
		{:else}
			<Signup />
		{/if}
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
