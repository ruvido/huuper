import { writable } from 'svelte/store';

export const currentRoute = writable('login');
export const queryParams = writable({});

const publicRoutes = ['login', 'signup', 'signup-direct', 'password-reset', 'event-accept'];
const authOnlyRoutes = ['onboarding', 'pending-approval', 'telegram-connect'];
const appPrefix = 'app/';
const adminPrefix = 'admin/';
export const defaultAppRoute = 'app/events';
const defaultAdminRoute = 'admin';
const appRoutes = [
	defaultAppRoute,
	'app/groups',
	'app/profile',
	'app/requests'
];
const adminRoutes = [
	defaultAdminRoute,
	'admin/events',
	'admin/event',
	'admin/groups',
	'admin/requests',
	'admin/profile'
];

function scrollToTop() {
	requestAnimationFrame(() => {
		window.scrollTo(0, 0);
		if (document.scrollingElement) {
			document.scrollingElement.scrollTop = 0;
		}
	});
}

function updateRoute() {
	const hash = window.location.hash.slice(1) || 'login'; // Remove #
	const routeWithParams = hash.startsWith('/') ? hash.slice(1) : hash; // Remove leading /
	const [route, queryString] = routeWithParams.split('?');
	const cleanRoute = route || 'login';

	currentRoute.set(cleanRoute);
	queryParams.set(Object.fromEntries(new URLSearchParams(queryString || '')));
	scrollToTop();
}

// Listen to hash changes
window.addEventListener('hashchange', updateRoute);
window.addEventListener('load', updateRoute);

export function navigate(route) {
	window.location.hash = route;
}

export function resolveNextRoute(next) {
	if (!next || typeof next !== 'string') return defaultAppRoute;
	if (next === 'app') return defaultAppRoute;
	if (next === 'admin') return defaultAdminRoute;
	if (next.startsWith('app/groups/')) return next;
	if (next.startsWith('admin/groups/')) return next;
	if (next.startsWith('app/requests/')) return next;
	if (next.startsWith('admin/requests/')) return next;
	if (next.startsWith(appPrefix) && appRoutes.includes(next)) return next;
	if (next.startsWith(adminPrefix) && adminRoutes.includes(next)) return next;
	return defaultAppRoute;
}

export function getTargetRoute(isAuthenticated, user, currentRoute) {
	if (currentRoute === 'event-accept') {
		return 'event-accept';
	}

	if (!isAuthenticated) {
		if (publicRoutes.includes(currentRoute)) return currentRoute;
		return `login?next=${encodeURIComponent(currentRoute)}`;
	}

	const status = user?.status;
	const hasData = user?.data && Object.keys(user.data).length > 0;
	const hasTelegram = user?.telegram && Object.keys(user.telegram).length > 0;

	if (status === 'pending') {
		return hasData ? 'pending-approval' : 'onboarding';
	}
	if (status === 'active') {
		if (!hasData) return 'onboarding';
		if (!hasTelegram) return 'telegram-connect';
	}

	const isAdmin = !!user?.admin;

	if ((currentRoute === defaultAdminRoute || currentRoute.startsWith(adminPrefix)) && !isAdmin) {
		return defaultAppRoute;
	}
	if (currentRoute.startsWith('app/groups/')) return currentRoute;
	if (currentRoute.startsWith('admin/groups/')) return isAdmin ? currentRoute : defaultAppRoute;
	if (currentRoute.startsWith('app/requests/')) return currentRoute;
	if (currentRoute.startsWith('admin/requests/')) return currentRoute;
	if (currentRoute === 'app') return isAdmin ? defaultAdminRoute : defaultAppRoute;
	if (currentRoute.startsWith(appPrefix)) {
		if (appRoutes.includes(currentRoute)) return currentRoute;
		return isAdmin ? defaultAdminRoute : defaultAppRoute;
	}
	if (currentRoute === defaultAdminRoute) return isAdmin ? currentRoute : defaultAppRoute;
	if (currentRoute.startsWith(adminPrefix)) {
		if (adminRoutes.includes(currentRoute)) return currentRoute;
		return isAdmin ? defaultAdminRoute : defaultAppRoute;
	}
	if (authOnlyRoutes.includes(currentRoute)) return isAdmin ? defaultAdminRoute : defaultAppRoute;
	if (publicRoutes.includes(currentRoute)) return isAdmin ? defaultAdminRoute : defaultAppRoute;

	return isAdmin ? defaultAdminRoute : defaultAppRoute;
}
