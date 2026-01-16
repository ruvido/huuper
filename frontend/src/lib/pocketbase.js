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

export function fetchSetting(name) {
	return fetch(`/api/settings/${name}`, {
		headers: {
			Authorization: pb.authStore.token,
		},
	});
}
