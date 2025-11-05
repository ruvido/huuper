import { writable } from 'svelte/store';

export const currentRoute = writable('login');

function updateRoute() {
	const hash = window.location.hash.slice(1) || 'login';
	currentRoute.set(hash);
}

// Listen to hash changes
window.addEventListener('hashchange', updateRoute);
window.addEventListener('load', updateRoute);

export function navigate(route) {
	window.location.hash = route;
}
