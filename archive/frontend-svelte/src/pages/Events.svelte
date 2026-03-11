<script>
	import { onMount } from 'svelte';
	import { apiFetch, pb } from '../lib/pocketbase';
	import { navigate } from '../lib/router';
	import DashboardLayout from '../components/DashboardLayout.svelte';
	import EventCard from '../components/cards/EventCard.svelte';
	import ConfirmModal from '../components/modals/ConfirmModal.svelte';
	import Card from '../components/Card.svelte';

	export let adminMode = false;

	let events = [];
	let eventsLoading = false;
	let eventsError = '';
	let registeredById = {};
	let registeringById = {};
	let unsubscribingById = {};
	let openById = {};
	let confirmState = null;
	let confirmConfig = null;

	function isFutureEvent(eventDate) {
		if (!eventDate || typeof eventDate !== 'string') return false;
		const normalized = eventDate.replace('T', ' ');
		const [dateRaw] = normalized.split(' ');
		const parts = dateRaw.split('-');
		if (parts.length !== 3) return false;
		const [year, month, day] = parts.map(Number);
		if (!year || !month || !day) return false;
		const now = new Date();
		const today = new Date(now.getFullYear(), now.getMonth(), now.getDate());
		const eventDay = new Date(year, month - 1, day);
		return eventDay > today;
	}

	async function loadEvents() {
		eventsLoading = true;
		eventsError = '';
		try {
			const result = await pb.collection('events').getList(1, 200, {
				filter: 'active = true',
				sort: 'event_date'
			});
			const items = result.items || [];
			events = items.filter((event) => isFutureEvent(event.event_date));
			if (!adminMode) {
				await loadStatuses();
			}
		} catch (err) {
			eventsError = err.message || err.toString() || 'Failed to load events';
		} finally {
			eventsLoading = false;
		}
	}

	async function loadStatuses() {
		if (!events.length) return;
		const updates = {};
		await Promise.all(events.map(async (event) => {
			try {
				const res = await apiFetch(`/api/events/${encodeURIComponent(event.slug)}/status`);
				if (!res.ok) return;
				const data = await res.json();
				updates[event.id] = !!data?.registered;
			} catch {
				// Ignore status errors
			}
		}));
		registeredById = { ...registeredById, ...updates };
	}

	function setCardOpen(eventId, value) {
		if (!eventId) return;
		openById = { ...openById, [eventId]: !!value };
	}

	function scrollToEventCard(eventId) {
		if (!eventId || typeof document === 'undefined') return;
		const el = document.getElementById(`event-${eventId}`);
		if (!el) return;
		el.scrollIntoView({ behavior: 'smooth', block: 'start' });
		requestAnimationFrame(() => {
			window.scrollBy({ top: -120, behavior: 'smooth' });
			el.focus?.();
		});
	}

	async function registerEvent(event) {
		await mutateRegistration('register', event);
	}

	function openConfirm(event, kind) {
		confirmState = { event, kind };
	}

	function closeConfirm() {
		confirmState = null;
	}

	async function confirmRegister() {
		if (!confirmState?.event) return;
		await registerEvent(confirmState.event);
		closeConfirm();
	}

	async function confirmUnsubscribe() {
		if (!confirmState?.event) return;
		await unsubscribeEvent(confirmState.event);
		closeConfirm();
	}

	async function unsubscribeEvent(event) {
		await mutateRegistration('unsubscribe', event);
	}

	async function mutateRegistration(kind, event) {
		if (!event) return;
		const isRegister = kind === 'register';
		const isUnsubscribe = kind === 'unsubscribe';

		if (isRegister && registeringById[event.id]) return;
		if (isUnsubscribe && unsubscribingById[event.id]) return;

		if (isRegister) {
			registeringById = { ...registeringById, [event.id]: true };
		} else {
			unsubscribingById = { ...unsubscribingById, [event.id]: true };
		}

		try {
			if (isRegister) {
				const email = pb.authStore.record?.email || '';
				if (!email) return;

				const payload = { email };
				const res = await apiFetch(`/api/events/${encodeURIComponent(event.slug)}/register`, {
					method: 'POST',
					headers: { 'Content-Type': 'application/json' },
					body: JSON.stringify(payload)
				});

				if (!res.ok) return;
				registeredById = { ...registeredById, [event.id]: true };
				setCardOpen(event.id, false);
				scrollToEventCard(event.id);
			}

			if (isUnsubscribe) {
				const res = await apiFetch(`/api/events/${encodeURIComponent(event.slug)}/unsubscribe`, {
					method: 'POST'
				});
				if (!res.ok) return;
				registeredById = { ...registeredById, [event.id]: false };
			}
		} finally {
			if (isRegister) {
				registeringById = { ...registeringById, [event.id]: false };
			} else {
				unsubscribingById = { ...unsubscribingById, [event.id]: false };
			}
		}
	}

	$: confirmConfig = (() => {
		if (!confirmState?.event) return null;
		const data = confirmState.event.data || {};
		if (confirmState.kind === 'unsubscribe') {
			return {
				title: data.unsubscribe_title || 'Confirm',
				message: data.unsubscribe_message || '',
				confirmLabel: data.unsubscribe_confirm_label || 'Unsubscribe',
				loading: !!unsubscribingById[confirmState.event.id],
				onConfirm: confirmUnsubscribe
			};
		}
		return {
			title: data.confirm_title || 'Confirm registration',
			message: data.confirmation_message || '',
			confirmLabel: data.confirm_label || 'Confirm',
			loading: !!registeringById[confirmState.event.id],
			onConfirm: confirmRegister
		};
	})();

	onMount(() => {
		loadEvents();
	});

	function openAdminEvent(event) {
		if (!adminMode || !event?.id) return;
		navigate(`admin/event?id=${encodeURIComponent(event.id)}`);
	}
</script>

<DashboardLayout title={adminMode ? 'Admin Events' : 'Eventi'}>
	<section class="section">
		<h2 class="section-title">Upcoming Events</h2>
		{#if eventsError}
			<Card variant="state">{eventsError}</Card>
		{:else if eventsLoading}
			<Card variant="state">Loading events...</Card>
		{:else if events.length === 0}
			<Card variant="state">No upcoming events.</Card>
		{:else}
			<div class="stack-list">
				{#each events as event}
					<EventCard
						{event}
						selectable={adminMode}
						onSelect={openAdminEvent}
						registered={!!registeredById[event.id]}
						canRegister={!adminMode && !registeredById[event.id]}
						showStatus={!adminMode}
						registering={!!registeringById[event.id]}
						canUnsubscribe={!adminMode && !!registeredById[event.id]}
						unsubscribing={!!unsubscribingById[event.id]}
						open={!!openById[event.id]}
						onToggle={(value) => setCardOpen(event.id, value)}
						onRegister={(evt) => openConfirm(evt, 'register')}
						onUnsubscribe={(evt) => openConfirm(evt, 'unsubscribe')}
					/>
				{/each}
			</div>
		{/if}
	</section>
</DashboardLayout>

{#if !adminMode}
	<ConfirmModal
		show={!!confirmConfig}
		title={confirmConfig?.title || 'Confirm'}
		message={confirmConfig?.message || ''}
		confirmLabel={confirmConfig?.confirmLabel || 'Confirm'}
		loading={confirmConfig?.loading || false}
		onConfirm={confirmConfig?.onConfirm}
		onCancel={closeConfirm}
	/>
{/if}

<style>
	.section {
		display: flex;
		flex-direction: column;
		gap: clamp(1rem, 3vw, 1.25rem);
	}

	.section-title {
		margin: 0;
		font-weight: 700;
		color: #000;
	}
</style>
