import { writable } from 'svelte/store';

export const currentRoute = writable('login');
export const queryParams = writable({});

// Valid routes in the application
const validRoutes = ['login', 'signup', 'welcome', 'onboarding', 'pending-approval', 'profile', 'groups'];

function updateRoute() {
	const hash = window.location.hash.slice(1) || 'login'; // Remove #
	const routeWithParams = hash.startsWith('/') ? hash.slice(1) : hash; // Remove leading /
	const [route, queryString] = routeWithParams.split('?');
	const cleanRoute = route || 'login';

	// Redirect invalid routes to login
	if (!validRoutes.includes(cleanRoute)) {
		window.location.hash = 'login';
		return;
	}

	currentRoute.set(cleanRoute);
	queryParams.set(Object.fromEntries(new URLSearchParams(queryString || '')));
}

// Listen to hash changes
window.addEventListener('hashchange', updateRoute);
window.addEventListener('load', updateRoute);

export function navigate(route) {
	window.location.hash = route;
}
