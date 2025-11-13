import { writable } from 'svelte/store';

export const currentRoute = writable('login');
export const queryParams = writable({});

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
