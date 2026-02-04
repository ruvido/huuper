<script>
	import { onMount } from 'svelte';
	import { queryParams, navigate } from '../lib/router';
	import ConfirmationPage from '../components/onboarding/ConfirmationPage.svelte';

	let status = 'loading'; // loading | accepted | already_accepted | expired | invalid | missing
	let title = 'Loading...';
	let text = 'Please wait.';

	const copy = {
		accepted: {
			title: 'Accepted.',
			text: 'The request has been confirmed.'
		},
		already_accepted: {
			title: 'Already accepted.',
			text: 'This request was already confirmed.'
		},
		expired: {
			title: 'Link expired.',
			text: 'The link is no longer valid.'
		},
		invalid: {
			title: 'Invalid link.',
			text: 'The link is invalid or expired.'
		},
		missing: {
			title: 'Missing token.',
			text: 'The link is incomplete.'
		}
	};

	function applyCopy(key) {
		const next = copy[key] || copy.invalid;
		title = next.title;
		text = next.text;
		status = key;
	}

	function readTokenFromHash() {
		const hash = window.location.hash || '';
		const clean = hash.startsWith('#') ? hash.slice(1) : hash;
		const [, queryString] = clean.split('?');
		return new URLSearchParams(queryString || '').get('token') || '';
	}

	onMount(async () => {
		const token = $queryParams?.token || readTokenFromHash();
		if (!token) {
			applyCopy('missing');
			return;
		}

		try {
			const res = await fetch(`/api/events/accept?token=${encodeURIComponent(token)}`);
			if (!res.ok) {
				let message = '';
				try {
					const data = await res.json();
					message = data?.message || '';
				} catch {
					message = '';
				}
				if (message === 'token_expired') {
					applyCopy('expired');
				} else {
					applyCopy('invalid');
				}
				return;
			}

			const data = await res.json();
			if (data?.status === 'already_accepted') {
				applyCopy('already_accepted');
			} else {
				applyCopy('accepted');
			}
		} catch {
			applyCopy('invalid');
		}
	});

	function goToLogin() {
		navigate('login');
	}
</script>

<ConfirmationPage
	{title}
	{text}
	buttonText="Go to login"
	onSubmit={goToLogin}
	showButton={status !== 'loading'}
	showCheckmark={status === 'accepted' || status === 'already_accepted'}
/>
