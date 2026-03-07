<script>
	import { onMount } from 'svelte';
	import { isAuthenticated, pb, authRecord, fetchSetting, apiFetch } from './lib/pocketbase';
	import { currentRoute, navigate, getTargetRoute } from './lib/router';
	import AppBottomBar from './components/AppBottomBar.svelte';
	import AdminBottomBar from './components/AdminBottomBar.svelte';

	let authReady = false;
	let renderReady = false;
	let appTitle = 'Members';
	let eventsAlert = false;
	let groupsAlert = false;
	let refreshingBadges = false;
	let lastBadgeRoute = '';
	let ActivePage = null;
	let activePageProps = {};
	let routeLoadToken = 0;
	let lastRouteViewKey = '';
	const version = __APP_VERSION__;
	const pageCache = new Map();
	const pageLoaders = {
		Login: () => import('./pages/Login.svelte'),
		Signup: () => import('./pages/SignupDirect.svelte'),
		PasswordReset: () => import('./pages/PasswordReset.svelte'),
		Onboarding: () => import('./pages/Onboarding.svelte'),
		PendingApproval: () => import('./pages/PendingApproval.svelte'),
		EventAccept: () => import('./pages/EventAccept.svelte'),
		TelegramConnect: () => import('./pages/TelegramConnect.svelte'),
		Events: () => import('./pages/Events.svelte'),
		Profile: () => import('./pages/Profile.svelte'),
		AdminProfile: () => import('./pages/AdminProfile.svelte'),
		Groups: () => import('./pages/Groups.svelte'),
		GroupDetail: () => import('./pages/GroupDetail.svelte'),
		GroupRequests: () => import('./pages/GroupRequests.svelte'),
		Admin: () => import('./pages/Admin.svelte'),
		AdminEvent: () => import('./pages/AdminEvent.svelte'),
		AdminRequests: () => import('./pages/AdminRequests.svelte'),
		Requests: () => import('./pages/Requests.svelte'),
		RequestDetail: () => import('./pages/RequestDetail.svelte'),
	};

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
	$: showMainNav = $isAuthenticated && (appRoute || adminRoute);

	$: if (!$isAuthenticated) {
		eventsAlert = false;
		groupsAlert = false;
		lastBadgeRoute = '';
	}

	$: if (authReady && $isAuthenticated && showMainNav && $currentRoute !== lastBadgeRoute) {
		lastBadgeRoute = $currentRoute;
		void refreshBadges();
	}

	function resolveRouteView(route) {
		if (route === 'login') return { page: 'Login', props: {} };
		if (route === 'signup') return { page: 'Signup', props: {} };
		if (route === 'signup-direct') return { page: 'Signup', props: { defaultStatus: 'active', showFooter: false, pageTitle: 'Sign Up (beta direct)' } };
		if (route === 'password-reset') return { page: 'PasswordReset', props: {} };
		if (route === 'onboarding') return { page: 'Onboarding', props: {} };
		if (route === 'pending-approval') return { page: 'PendingApproval', props: {} };
		if (route === 'event-accept') return { page: 'EventAccept', props: {} };
		if (route === 'telegram-connect') return { page: 'TelegramConnect', props: {} };
		if (route === 'app/events') return { page: 'Events', props: {} };
		if (route === 'app/profile') return { page: 'Profile', props: {} };
		if (route === 'admin/profile') return { page: 'AdminProfile', props: {} };
		if (route === 'app/groups') return { page: 'Groups', props: {} };
		if (/^admin\/groups\/[^/]+\/requests$/.test(route)) return { page: 'GroupRequests', props: {} };
		if (/^app\/groups\/[^/]+\/requests$/.test(route)) return { page: 'GroupRequests', props: {} };
		if (route.startsWith('admin/groups/')) return { page: 'GroupDetail', props: {} };
		if (route.startsWith('app/groups/')) return { page: 'GroupDetail', props: {} };
		if (route === 'admin') return { page: 'Admin', props: {} };
		if (route === 'admin/events') return { page: 'Events', props: { adminMode: true } };
		if (route === 'admin/event') return { page: 'AdminEvent', props: {} };
		if (route === 'admin/groups') return { page: 'Groups', props: { adminMode: true } };
		if (route === 'admin/requests') return { page: 'AdminRequests', props: {} };
		if (/^admin\/requests\/[^/]+$/.test(route)) return { page: 'RequestDetail', props: {} };
		if (/^app\/requests\/[^/]+$/.test(route)) return { page: 'RequestDetail', props: {} };
		if (route === 'app/requests' || route.startsWith('app/requests/')) return { page: 'Requests', props: {} };
		return { page: 'Login', props: {} };
	}

	async function loadActivePage(view) {
		const token = ++routeLoadToken;
		ActivePage = null;
		activePageProps = view.props;

		if (pageCache.has(view.page)) {
			ActivePage = pageCache.get(view.page);
			return;
		}

		try {
			const loader = pageLoaders[view.page];
			if (!loader) throw new Error('Missing page loader');
			const mod = await loader();
			if (token !== routeLoadToken) return;
			pageCache.set(view.page, mod.default);
			ActivePage = mod.default;
		} catch {
			if (token !== routeLoadToken) return;
			const mod = await pageLoaders.Login();
			pageCache.set('Login', mod.default);
			activePageProps = {};
			ActivePage = mod.default;
		}
	}

	$: routeView = resolveRouteView($currentRoute || '');
	$: routeViewKey = routeView ? `${routeView.page}:${JSON.stringify(routeView.props || {})}` : '';
	$: if (renderReady && routeView && routeViewKey !== lastRouteViewKey) {
		lastRouteViewKey = routeViewKey;
		void loadActivePage(routeView);
	}
</script>

<!-- Only render when guards have validated -->
{#if renderReady}
	<div class="app-shell">
		<div class="cards-block">
			<main>
				{#if ActivePage}
					<svelte:component this={ActivePage} {...activePageProps} />
				{:else}
					<div class="page-loading" aria-live="polite">Loading...</div>
				{/if}
			</main>
		</div>

		<div class="version" class:with-nav={showMainNav}>{version}</div>
	</div>

	{#if appRoute}
		<AppBottomBar {eventsAlert} {groupsAlert} />
	{:else if adminRoute}
		<AdminBottomBar {eventsAlert} requestsAlert={groupsAlert} />
	{/if}
{/if}

<style>
	:global(body) {
		overflow-x: hidden;
	}

	.app-shell {
		min-height: 100vh;
		display: flex;
		flex-direction: column;
	}

	.cards-block {
		flex: 1 0 auto;
		display: flex;
		flex-direction: column;
	}

	main {
		width: 100%;
		max-width: 100%;
	}

	.page-loading {
		padding: 2rem 1rem;
		text-align: center;
		font-size: 0.95rem;
		color: #666;
	}

	.version {
		margin-top: 0;
		padding: 0.75rem 0;
		text-align: center;
		font-size: 10px;
		color: #999;
	}

	.version.with-nav {
		margin-bottom: 4.75rem;
	}
</style>
