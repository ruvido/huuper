import PocketBase from 'pocketbase';
import { writable } from 'svelte/store';

export const pb = new PocketBase('/');

// Reactive auth store for Svelte
export const isAuthenticated = writable(pb.authStore.isValid);

// Auto refresh auth state
pb.authStore.onChange(() => {
	console.log('Auth state changed:', pb.authStore.isValid);
	isAuthenticated.set(pb.authStore.isValid);
});
