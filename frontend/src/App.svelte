<script>
	import { isAuthenticated } from './lib/pocketbase';
	import { currentRoute, navigate } from './lib/router';
	import Header from './components/Header.svelte';
	import Menu from './components/Menu.svelte';
	import Login from './pages/Login.svelte';
	import Signup from './pages/Signup.svelte';
	import Profile from './pages/Profile.svelte';
	import Groups from './pages/Groups.svelte';

	let menuOpen = false;

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
		<Signup />
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
