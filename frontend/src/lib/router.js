import { writable } from 'svelte/store';

export const currentRoute = writable('login');

function updateRoute() {
	const hash = window.location.hash.slice(1) || 'login'; // Remove #
	const routeWithParams = hash.startsWith('/') ? hash.slice(1) : hash; // Remove leading /
	const route = routeWithParams.split('?')[0] || 'login'; // Extract route before ?
	currentRoute.set(route);
}

// Listen to hash changes
window.addEventListener('hashchange', updateRoute);
window.addEventListener('load', updateRoute);

export function navigate(route) {
	window.location.hash = route;
}
