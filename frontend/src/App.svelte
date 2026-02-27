<script>
	import { onMount } from 'svelte';
	import { isAuthenticated, pb, authRecord, fetchSetting, apiFetch } from './lib/pocketbase';
	import { currentRoute, navigate, getTargetRoute } from './lib/router';
	import BottomBar from './components/BottomBar.svelte';
	import Login from './pages/Login.svelte';
	import Signup from './pages/SignupDirect.svelte';
	import PasswordReset from './pages/PasswordReset.svelte';
	import Onboarding from './pages/Onboarding.svelte';
	import PendingApproval from './pages/PendingApproval.svelte';
	import EventAccept from './pages/EventAccept.svelte';
	import TelegramConnect from './pages/TelegramConnect.svelte';
	import Events from './pages/Events.svelte';
	import Profile from './pages/Profile.svelte';
	import Groups from './pages/Groups.svelte';
	import GroupDetail from './pages/GroupDetail.svelte';
	import GroupRequests from './pages/GroupRequests.svelte';
	import Admin from './pages/Admin.svelte';
	import AdminEvent from './pages/AdminEvent.svelte';
	import AdminRequests from './pages/AdminRequests.svelte';
	import Requests from './pages/Requests.svelte';
	import RequestDetail from './pages/RequestDetail.svelte';

	let authReady = false;
	let renderReady = false;
	let appTitle = 'Members';
	let eventsAlert = false;
	let groupsAlert = false;
	let refreshingBadges = false;
	let lastBadgeRoute = '';
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

	function isFutureEvent(eventDate) {
		if (!eventDate || typeof eventDate !== 'string') return false;
		const normalized = eventDate.replace('T', ' ');
		const [dateRaw] = normalized.split(' ');
		const parts = dateRaw.split('-');
		if (parts.length !== 3) return false;
		const [year, month, day] = parts.map(Number);
		if (!year || !month || !day) return false;
		const now = new Date();
		const today = new Date(now.getFullYear(), now.getMonth(), now.getDate());
		const eventDay = new Date(year, month - 1, day);
		return eventDay > today;
	}

	async function refreshBadges() {
		if (refreshingBadges || !$isAuthenticated) return;
		refreshingBadges = true;
		try {
			const eventsResult = await pb.collection('events').getList(1, 200, {
				filter: 'active = true',
				sort: 'event_date'
			});
			const items = Array.isArray(eventsResult?.items) ? eventsResult.items : [];
			eventsAlert = items.some((event) => isFutureEvent(event?.event_date));
		} catch {
			// Keep previous value if events fetch fails.
		}

		try {
			const requestsResult = await apiFetch('/api/requests?per_page=1');
			if (!requestsResult.ok) {
				groupsAlert = false;
			} else {
				const data = await requestsResult.json();
				groupsAlert = Array.isArray(data?.items) && data.items.length > 0;
			}
		} catch {
			groupsAlert = false;
		} finally {
			refreshingBadges = false;
		}
	}

	// Reset renderReady when route changes to re-run guards
	$: if ($currentRoute) {
		renderReady = false;
	}

	// Guard logic - runs BEFORE allowing render
	$: if (authReady && !renderReady) {
		let shouldRedirect = false;

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

	$: appRoute = typeof $currentRoute === 'string' && $currentRoute.startsWith('app/');
	$: adminRoute = typeof $currentRoute === 'string'
		&& ($currentRoute === 'admin' || $currentRoute.startsWith('admin/'));
	$: showMainNav = $isAuthenticated && appRoute && !adminRoute;

	$: if (!$isAuthenticated) {
		eventsAlert = false;
		groupsAlert = false;
		lastBadgeRoute = '';
	}

	$: if (authReady && $isAuthenticated && showMainNav && $currentRoute !== lastBadgeRoute) {
		lastBadgeRoute = $currentRoute;
		void refreshBadges();
	}
</script>

<!-- Only render when guards have validated -->
{#if renderReady}
	<main>
		{#if $currentRoute === 'login'}
			<Login />
		{:else if $currentRoute === 'signup'}
			<Signup />
		{:else if $currentRoute === 'signup-direct'}
			<Signup defaultStatus="active" showFooter={false} pageTitle="Sign Up (beta direct)" />
		{:else if $currentRoute === 'password-reset'}
			<PasswordReset />
		{:else if $currentRoute === 'onboarding'}
			<Onboarding />
		{:else if $currentRoute === 'pending-approval'}
			<PendingApproval />
		{:else if $currentRoute === 'event-accept'}
			<EventAccept />
		{:else if $currentRoute === 'telegram-connect'}
			<TelegramConnect />
		{:else if $currentRoute === 'app/events'}
			<Events />
		{:else if $currentRoute === 'app/profile'}
			<Profile />
		{:else if $currentRoute === 'app/groups'}
			<Groups />
		{:else if /^app\/groups\/[^/]+\/requests$/.test($currentRoute)}
			<GroupRequests />
		{:else if $currentRoute.startsWith('app/groups/')}
			<GroupDetail />
		{:else if $currentRoute === 'admin'}
			<Admin />
		{:else if $currentRoute === 'admin/events'}
			<AdminEvent />
		{:else if $currentRoute === 'admin/requests'}
			<AdminRequests />
		{:else if /^admin\/requests\/[^/]+$/.test($currentRoute)}
			<RequestDetail />
		{:else if /^app\/requests\/[^/]+$/.test($currentRoute)}
			<RequestDetail />
		{:else if $currentRoute === 'app/requests' || $currentRoute.startsWith('app/requests/')}
			<Requests />
		{:else}
			<Login />
		{/if}
	</main>

	{#if showMainNav}
		<BottomBar {eventsAlert} {groupsAlert} />
	{/if}

	<div class="version">{version}</div>
{/if}

<style>
	:global(body) {
		overflow-x: hidden;
	}

	main {
		width: 100%;
		max-width: 100%;
	}

	.version {
		margin: 0 auto;
		padding: 0.75rem 0 6rem;
		text-align: center;
		font-size: 10px;
		color: #999;
	}
</style>
