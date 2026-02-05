<script>
	import { onMount } from 'svelte';
	import { pb } from '../lib/pocketbase';
	import { queryParams, navigate } from '../lib/router';
	import DashboardLayout from '../components/DashboardLayout.svelte';
import StateCard from '../components/StateCard.svelte';
import AdminCard from '../components/AdminCard.svelte';
import StatRow from '../components/StatRow.svelte';
	import ConfirmModal from '../components/modals/ConfirmModal.svelte';

	let loading = true;
	let error = '';
	let eventInfo = null;
	let registrations = [];
let confirmOpen = false;
let confirmLoading = false;
let confirmTarget = null;
let lastEventId = null;

	$: eventId = $queryParams?.id;

	function formatDatePart(raw) {
		if (!raw) return '';
		const parts = raw.split('-');
		if (parts.length !== 3) return raw;
		return `${parts[2]}/${parts[1]}/${parts[0]}`;
	}

	function formatEventDate(value) {
		if (!value || typeof value !== 'string') return '';
		const normalized = value.replace('T', ' ');
		const [dateRaw] = normalized.split(' ');
		return formatDatePart(dateRaw);
	}

function displayName(registration) {
	const data = registration?.data || {};
	const fullName = typeof data.full_name === 'string' ? data.full_name.trim() : '';
	if (fullName) return fullName;
	return registration?.email || 'Unknown';
}

$: pendingItems = registrations.filter((r) => !r.accepted);
$: approvedItems = registrations.filter((r) => r.accepted);
$: pendingCount = pendingItems.length;
$: approvedCount = approvedItems.length;
$: totalCount = pendingCount + approvedCount;

	async function loadDetails() {
		if (!eventId) return;
		loading = true;
		error = '';
		try {
			const response = await fetch(`/api/admin/events/${encodeURIComponent(eventId)}`, {
				headers: {
					Authorization: pb.authStore.token,
				},
			});
			if (!response.ok) {
				throw new Error('Failed to load event data');
			}
			const data = await response.json();
			eventInfo = data?.event || null;
			registrations = Array.isArray(data?.registrations) ? data.registrations : [];
		} catch (err) {
			error = err.message || err.toString() || 'Failed to load event data';
		} finally {
			loading = false;
		}
	}

	function openConfirm(registration) {
		confirmTarget = registration;
		confirmOpen = true;
	}

	function closeConfirm() {
		confirmOpen = false;
		confirmTarget = null;
	}

	async function confirmApprove() {
		if (!confirmTarget) return;
		confirmLoading = true;
		try {
			const response = await fetch(`/api/admin/registrations/${encodeURIComponent(confirmTarget.id)}/approve`, {
				method: 'POST',
				headers: {
					Authorization: pb.authStore.token,
				},
			});
			if (!response.ok) {
				throw new Error('Failed to approve');
			}
			registrations = registrations.map((item) =>
				item.id === confirmTarget.id ? { ...item, accepted: true } : item
			);
			closeConfirm();
		} catch (err) {
			error = err.message || err.toString() || 'Failed to approve';
		} finally {
			confirmLoading = false;
		}
	}

function goBack() {
	navigate('app/admin');
}

$: if (eventId && eventId !== lastEventId) {
	lastEventId = eventId;
	loadDetails();
}

onMount(() => {
	if (eventId) {
		lastEventId = eventId;
		loadDetails();
	}
});
</script>

<DashboardLayout title="Event">
	{#if !eventId}
		<StateCard>Missing event.</StateCard>
	{:else if error}
		<StateCard>{error}</StateCard>
	{:else if loading}
		<StateCard>Loading...</StateCard>
	{:else}
		<div class="header">
			<h1>{eventInfo?.title}</h1>
			<p class="date">{formatEventDate(eventInfo?.event_date)}</p>
			<StatRow
				center
				items={[
					{ label: 'Total', value: totalCount },
					{ label: 'Pending', value: pendingCount },
				]}
			/>
		</div>

		<AdminCard>
			<div class="list">
				<div class="list-section">
					<h2>Pending</h2>
					{#if pendingCount === 0}
						<p class="empty">No pending registrations.</p>
					{:else}
						{#each pendingItems as reg}
							<div class="row">
								<div>
									<p class="name">{displayName(reg)}</p>
									<p class="email">{reg.email}</p>
								</div>
								<div class="row-actions">
									<button class="approve" on:click={() => openConfirm(reg)}>Approve</button>
									{#if reg.hasUser}
										<span class="dot" aria-label="Registered user"></span>
									{/if}
								</div>
							</div>
						{/each}
					{/if}
				</div>
				<div class="list-section">
					<h2>Approved</h2>
					{#if approvedCount === 0}
						<p class="empty">No approved registrations.</p>
					{:else}
						{#each approvedItems as reg}
							<div class="row">
								<div>
									<p class="name">{displayName(reg)}</p>
									<p class="email">{reg.email}</p>
								</div>
								{#if reg.hasUser}
									<span class="dot" aria-label="Registered user"></span>
								{/if}
							</div>
						{/each}
					{/if}
				</div>
			</div>
		</AdminCard>
	{/if}

	<button class="back" on:click={goBack}>Back</button>
</DashboardLayout>

<ConfirmModal
	show={confirmOpen}
	title="Approve registration"
	message="Are you sure you want to approve this registration?"
	confirmLabel="Approve"
	loading={confirmLoading}
	onConfirm={confirmApprove}
	onCancel={closeConfirm}
/>

<style>
	h1 {
		margin: 0;
		font-size: 1.4rem;
		font-weight: 700;
	}

	h2 {
		margin: 0 0 0.5rem 0;
		font-size: 1rem;
		font-weight: 700;
		text-transform: uppercase;
		letter-spacing: 0.06em;
	}

	.header {
		display: grid;
		gap: 0.5rem;
	}

	.date {
		margin: 0;
		font-size: 0.9rem;
		font-weight: 600;
		color: #000;
		white-space: nowrap;
	}

	.list {
		display: grid;
		gap: 1.25rem;
	}

	.list-section {
		display: grid;
		gap: 0.5rem;
	}

	.row {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 1rem;
		padding: 0.75rem 0;
		border-top: 1px solid #000;
		flex-wrap: nowrap;
	}

	.row > div {
		min-width: 0;
	}

	.row-actions {
		display: inline-flex;
		align-items: center;
		gap: 0.5rem;
		justify-content: flex-end;
		flex-shrink: 0;
	}

	.dot {
		width: 0.55rem;
		height: 0.55rem;
		border-radius: 999px;
		background: #000;
		display: inline-block;
	}

	.name {
		margin: 0;
		font-weight: 700;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.email {
		margin: 0.25rem 0 0 0;
		font-size: 0.9rem;
		color: #333;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.approve {
		border: 2px solid #000;
		background: #000;
		color: #fff;
		padding: 0.5rem 0.9rem;
		font-weight: 700;
		cursor: pointer;
	}

	.approve:hover {
		background: #333;
	}

	.back {
		margin-top: 1rem;
		width: 100%;
		padding: 0.9rem 1rem;
		background: #000;
		color: #fff;
		border: 2px solid #000;
		font-weight: 700;
		cursor: pointer;
	}

	.back:hover {
		background: #333;
	}

	.empty {
		margin: 0;
		font-size: 0.9rem;
		color: #333;
	}

	@media (max-width: 480px) {
		.row {
			gap: 0.5rem;
		}
	}
</style>
