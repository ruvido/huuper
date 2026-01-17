import { writable } from 'svelte/store';

export const currentRoute = writable('login');
export const queryParams = writable({});

const publicRoutes = ['login', 'signup', 'signup-direct'];
const authOnlyRoutes = ['onboarding', 'pending-approval', 'telegram-connect'];
const appPrefix = 'app/';
export const defaultAppRoute = 'app/profile';
const appRoutes = [defaultAppRoute, 'app/groups'];

function updateRoute() {
	const hash = window.location.hash.slice(1) || 'login'; // Remove #
	const routeWithParams = hash.startsWith('/') ? hash.slice(1) : hash; // Remove leading /
	const [route, queryString] = routeWithParams.split('?');
	const cleanRoute = route || 'login';

	currentRoute.set(cleanRoute);
	queryParams.set(Object.fromEntries(new URLSearchParams(queryString || '')));
}

// Listen to hash changes
window.addEventListener('hashchange', updateRoute);
window.addEventListener('load', updateRoute);

export function navigate(route) {
	window.location.hash = route;
}

export function getTargetRoute(isAuthenticated, user, currentRoute) {
	if (!isAuthenticated) {
		return publicRoutes.includes(currentRoute) ? currentRoute : 'login';
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

	if (currentRoute === 'app') return defaultAppRoute;
	if (currentRoute.startsWith(appPrefix)) {
		return appRoutes.includes(currentRoute) ? currentRoute : defaultAppRoute;
	}
	if (authOnlyRoutes.includes(currentRoute)) return defaultAppRoute;
	if (publicRoutes.includes(currentRoute)) return defaultAppRoute;

	return defaultAppRoute;
}
