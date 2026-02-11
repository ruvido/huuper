<script>
	import { onMount } from 'svelte';
	import { apiFetch, pb } from '../lib/pocketbase';
	import { navigate } from '../lib/router';
	import DashboardLayout from '../components/DashboardLayout.svelte';
	import StateCard from '../components/StateCard.svelte';
	import EventCard from '../components/cards/EventCard.svelte';
	import GroupCard from '../components/cards/GroupCard.svelte';
	import Card from '../components/Card.svelte';
	import ConfirmModal from '../components/modals/ConfirmModal.svelte';

	let events = [];
	let groups = [];
	let eventsLoading = false;
	let groupsLoading = false;
	let eventsError = '';
	let groupsError = '';
	let registeredById = {};
	let registeringById = {};
	let unsubscribingById = {};
	let confirmState = null;
	let confirmConfig = null;

	const user = pb.authStore.record;

	function hasTelegram() {
		const telegram = user?.telegram;
		return telegram && Object.keys(telegram).length > 0;
	}

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
			await loadStatuses();
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

	async function loadGroups() {
		groupsLoading = true;
		groupsError = '';
		try {
			const currentUser = pb.authStore.record;
			const result = await pb.collection('user_groups').getList(1, 200, {
				filter: `user = "${currentUser.id}"`,
				expand: 'group'
			});
			groups = result.items
				.map((item) => item?.expand?.group)
				.filter(Boolean);
		} catch (err) {
			groupsError = err.message || err.toString() || 'Failed to load groups';
		} finally {
			groupsLoading = false;
		}
	}

	function goToGroupRequests(group) {
		if (!group?.id) return;
		navigate(`app/requests?group_id=${encodeURIComponent(group.id)}`);
	}

	function goToGroup(group) {
		if (!group?.id) return;
		navigate(`app/group/${encodeURIComponent(group.id)}`);
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

				const payload = {
					email,
					data: pb.authStore.record?.data || {}
				};

				const res = await apiFetch(`/api/events/${encodeURIComponent(event.slug)}/register`, {
					method: 'POST',
					headers: {
						'Content-Type': 'application/json'
					},
					body: JSON.stringify(payload)
				});

				if (!res.ok) return;
				registeredById = { ...registeredById, [event.id]: true };
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
		loadGroups();
	});
</script>

<DashboardLayout title="Home">
	<section class="section">
		<h2 class="section-title">Upcoming Events</h2>
		{#if eventsError}
			<StateCard>{eventsError}</StateCard>
		{:else if eventsLoading}
			<StateCard>Loading events...</StateCard>
		{:else if events.length === 0}
			<StateCard>No upcoming events.</StateCard>
		{:else}
			<div class="events-list">
				{#each events as event}
					<EventCard
						{event}
						registered={!!registeredById[event.id]}
						canRegister={!registeredById[event.id]}
						showStatus={true}
						registering={!!registeringById[event.id]}
						canUnsubscribe={!!registeredById[event.id]}
						unsubscribing={!!unsubscribingById[event.id]}
						onRegister={(evt) => openConfirm(evt, 'register')}
						onUnsubscribe={(evt) => openConfirm(evt, 'unsubscribe')}
					/>
				{/each}
			</div>
		{/if}
	</section>

	<section class="section">
		<h2 class="section-title">Your Groups</h2>
		{#if groupsError}
			<StateCard>{groupsError}</StateCard>
		{:else if groupsLoading}
			<StateCard>Loading groups...</StateCard>
		{:else if groups.length === 0}
			<StateCard>You are not in any groups yet.</StateCard>
		{:else}
			<div class="groups-list">
				{#each groups as group}
					<GroupCard {group} isMember={true} onRequests={goToGroupRequests} onOpen={goToGroup} />
				{/each}
			</div>
		{/if}
	</section>

	<section class="section">
		<h2 class="section-title">Profile</h2>
		<Card>
			<p class="profile-email">{user?.email}</p>
			<p class="profile-telegram">
				{hasTelegram() ? 'Telegram connected' : 'Telegram missing'}
			</p>
		</Card>
	</section>
</DashboardLayout>

<ConfirmModal
	show={!!confirmConfig}
	title={confirmConfig?.title || 'Confirm'}
	message={confirmConfig?.message || ''}
	confirmLabel={confirmConfig?.confirmLabel || 'Confirm'}
	loading={confirmConfig?.loading || false}
	onConfirm={confirmConfig?.onConfirm}
	onCancel={closeConfirm}
/>

<style>
	.section {
		display: flex;
		flex-direction: column;
		gap: clamp(1rem, 3vw, 1.25rem);
	}

	.section + .section {
		margin-top: clamp(1.5rem, 5vw, 2.5rem);
	}

	.section-title {
		margin: 0;
		font-size: clamp(1.1rem, 3vw, 1.25rem);
		font-weight: 700;
		color: #000;
	}

	.events-list,
	.groups-list {
		display: grid;
		gap: clamp(1rem, 3vw, 1.5rem);
		grid-template-columns: 1fr;
	}

	.profile-email {
		margin: 0 0 0.5rem 0;
		font-weight: 700;
		color: #000;
		word-break: break-word;
	}

	.profile-telegram {
		margin: 0;
		font-weight: 600;
		color: #000;
	}
</style>
