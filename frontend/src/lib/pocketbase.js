import PocketBase from 'pocketbase';
import { writable } from 'svelte/store';

export const pb = new PocketBase('/');

// Reactive auth store for Svelte
export const isAuthenticated = writable(pb.authStore.isValid);
export const authRecord = writable(pb.authStore.record);

// Auto refresh auth state
pb.authStore.onChange(() => {
	isAuthenticated.set(pb.authStore.isValid);
	authRecord.set(pb.authStore.record);
});

export function apiFetch(path, options = {}) {
	const headers = new Headers(options.headers || {});
	const token = pb.authStore.token;
	if (token) {
		headers.set('Authorization', token);
	}

	return fetch(path, {
		...options,
		headers,
	});
}

export function fetchSetting(name) {
	return apiFetch(`/api/settings/${name}`);
}
