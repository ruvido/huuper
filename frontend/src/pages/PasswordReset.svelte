<script>
	import { onMount } from 'svelte';
	import { pb, fetchSetting } from '../lib/pocketbase';
	import { navigate, queryParams } from '../lib/router';
	import AuthLayout from '../components/AuthLayout.svelte';
	import FormGroup from '../components/FormGroup.svelte';
	import Button from '../components/Button.svelte';
	import ErrorMessage from '../components/ErrorMessage.svelte';

	let email = '';
	let emailError = '';
	let password = '';
	let passwordConfirm = '';
	let passwordError = '';
	let confirmError = '';
	let error = '';
	let loading = false;
	let token = '';
	let requestSent = false;
	let resetDone = false;
	let copy = {
		request: {
			title: '',
			helper: '',
			email_label: '',
			submit_label: '',
			submitting_label: '',
			footer_prompt: '',
			footer_action: '',
			confirmation_title: '',
			confirmation_message: '',
			confirmation_hint: '',
			confirmation_back_to_login: '',
		},
		reset: {
			title: '',
			helper: '',
			password_label: '',
			confirm_label: '',
			submit_label: '',
			submitting_label: '',
			success_title: '',
			success_message: '',
			success_back_to_login: '',
		},
		errors: {
			required_email: '',
			send_failed: '',
			load_failed: '',
			invalid_token: '',
			required_passwords: '',
			reset_failed: '',
			reset_invalid: '',
		},
	};

	$: token = $queryParams?.token || '';
	$: isResetMode = !!token;

	onMount(async () => {
		try {
			const response = await fetchSetting('password_reset');
			if (response.ok) {
				const data = await response.json();
				if (data?.data) {
					copy = data.data;
				}
			} else {
				error = copy.errors.load_failed;
			}
		} catch (err) {
			error = copy.errors.load_failed;
		}
	});

	async function handleRequest() {
		if (!email) {
			error = copy.errors.required_email;
			return;
		}

		loading = true;
		error = '';
		emailError = '';

		try {
			await pb.collection('users').requestPasswordReset(email);
			requestSent = true;
		} catch (err) {
			if (err?.status === 400) {
				requestSent = true;
			} else {
				error = copy.errors.send_failed;
			}
		} finally {
			loading = false;
		}
	}

	async function handleReset() {
		if (!token) {
			error = copy.errors.invalid_token;
			return;
		}

		if (!password || !passwordConfirm) {
			error = copy.errors.required_passwords;
			return;
		}

		loading = true;
		error = '';
		passwordError = '';
		confirmError = '';

		try {
			await pb.collection('users').confirmPasswordReset(token, password, passwordConfirm);
			resetDone = true;
		} catch (err) {
			if (err?.status === 400) {
				error = copy.errors.reset_invalid;
			} else {
				error = copy.errors.reset_failed;
			}
		} finally {
			loading = false;
		}
	}

	function goToLogin() {
		navigate('login');
	}
</script>

<AuthLayout>
	{#if isResetMode}
		{#if resetDone}
			<div class="checkmark-container">
				<svg class="checkmark" viewBox="0 0 52 52">
					<circle class="checkmark-circle" cx="26" cy="26" r="25" fill="none" />
					<path class="checkmark-check" fill="none" d="M14.1 27.2l7.1 7.2 16.7-16.8" />
				</svg>
			</div>
			<h1>{copy.reset.success_title}</h1>
			<p class="confirmation-text">{copy.reset.success_message}</p>
			<div class="footer">
				<Button variant="link" on:click={goToLogin} disabled={loading}>
					{copy.reset.success_back_to_login}
				</Button>
			</div>
		{:else}
			<h1>{copy.reset.title}</h1>
			{#if copy.reset.helper}
				<p class="helper">{copy.reset.helper}</p>
			{/if}

			<form on:submit|preventDefault={handleReset}>
				<FormGroup
					id="password"
					type="password"
					label={copy.reset.password_label}
					name="password"
					bind:value={password}
					bind:error={passwordError}
					disabled={loading}
					required
				/>

				<FormGroup
					id="passwordConfirm"
					type="password"
					label={copy.reset.confirm_label}
					name="passwordConfirm"
					bind:value={passwordConfirm}
					bind:error={confirmError}
					disabled={loading}
					required
					matchField="password"
					matchValue={password}
				/>

				<ErrorMessage {error} />

				<Button variant="submit" type="submit" disabled={loading}>
					{loading ? copy.reset.submitting_label : copy.reset.submit_label}
				</Button>
			</form>

			<div class="footer">
				<Button variant="link" on:click={goToLogin} disabled={loading}>
					{copy.reset.success_back_to_login}
				</Button>
			</div>
		{/if}
	{:else}
		{#if requestSent}
			<div class="checkmark-container">
				<svg class="checkmark" viewBox="0 0 52 52">
					<circle class="checkmark-circle" cx="26" cy="26" r="25" fill="none" />
					<path class="checkmark-check" fill="none" d="M14.1 27.2l7.1 7.2 16.7-16.8" />
				</svg>
			</div>
			<h1>{copy.request.confirmation_title}</h1>
			<p class="confirmation-text">{copy.request.confirmation_message}</p>
			<p class="helper">{copy.request.confirmation_hint}</p>
			<div class="footer">
				<Button variant="link" on:click={goToLogin} disabled={loading}>
					{copy.request.confirmation_back_to_login}
				</Button>
			</div>
		{:else}
			<h1>{copy.request.title}</h1>
			<p class="helper">{copy.request.helper}</p>

			<form on:submit|preventDefault={handleRequest}>
				<FormGroup
					id="email"
					type="email"
					label={copy.request.email_label}
					name="email"
					bind:value={email}
					bind:error={emailError}
					disabled={loading}
					required
				/>

				<ErrorMessage {error} />

				<Button variant="submit" type="submit" disabled={loading}>
					{loading ? copy.request.submitting_label : copy.request.submit_label}
				</Button>
			</form>

			<div class="footer">
				{copy.request.footer_prompt}
				<Button variant="link" on:click={goToLogin} disabled={loading}>
					{copy.request.footer_action}
				</Button>
			</div>
		{/if}
	{/if}
</AuthLayout>

<style>
	h1 {
		margin: 0 0 1.5rem 0;
		font-size: 1.5rem;
		text-align: center;
		font-weight: bold;
	}

	.checkmark-container {
		width: 90px;
		height: 90px;
		margin: 0 auto 1.5rem auto;
	}

	.confirmation-text {
		margin: 0;
		text-align: center;
		font-size: 1.05rem;
		font-weight: 600;
	}

	.helper {
		margin: 0 0 1.5rem 0;
		text-align: center;
		font-size: 0.95rem;
	}

	.footer {
		margin-top: 1rem;
		text-align: center;
		font-size: 0.9rem;
	}

	.checkmark {
		width: 100%;
		height: 100%;
	}

	.checkmark-circle {
		stroke: #22c55e;
		stroke-width: 2;
		stroke-dasharray: 166;
		stroke-dashoffset: 166;
		animation: stroke 0.6s cubic-bezier(0.65, 0, 0.45, 1) forwards;
	}

	.checkmark-check {
		stroke: #22c55e;
		stroke-width: 3;
		stroke-linecap: round;
		stroke-linejoin: round;
		stroke-dasharray: 48;
		stroke-dashoffset: 48;
		animation: stroke 0.3s cubic-bezier(0.65, 0, 0.45, 1) 0.6s forwards;
	}

	@keyframes stroke {
		100% {
			stroke-dashoffset: 0;
		}
	}
</style>
