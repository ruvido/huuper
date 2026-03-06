<script>
	import { onMount, tick } from 'svelte';
	import { apiFetch } from '../lib/pocketbase';
	import { formatEventDate } from '../lib/date';
	import { queryParams, navigate } from '../lib/router';
	import DashboardLayout from '../components/DashboardLayout.svelte';
	import StatRow from '../components/StatRow.svelte';
	import ConfirmModal from '../components/modals/ConfirmModal.svelte';
	import Card from '../components/Card.svelte';

	let loading = true;
	let error = '';
	let eventInfo = null;
	let registrations = [];
	let confirmOpen = false;
	let confirmLoading = false;
	let confirmTarget = null;
	let cancelConfirmOpen = false;
	let cancelConfirmTarget = null;
	let rejectDialogOpen = false;
	let rejectTarget = null;
	let rejectNote = '';
	let rejectLoading = false;
	let lastEventId = null;
	let cancelingId = null;

	$: eventId = $queryParams?.id;

	function stringValue(value) {
		if (Array.isArray(value)) return value.filter(Boolean).join(', ');
		if (typeof value === 'string') return value.trim();
		return '';
	}

	function displayName(registration) {
		const data = registration?.data || {};
		const fullName = typeof data.full_name === 'string' ? data.full_name.trim() : '';
		if (fullName) return fullName;
		return 'Unknown';
	}

	function displayRegion(registration) {
		const data = registration?.data || {};
		const region = typeof data.region === 'string' ? data.region.trim() : '';
		if (region) return region;
		return 'Unknown';
	}

	function displayPhone(registration) {
		const data = registration?.data || {};
		const phone = typeof data.mobile === 'string' ? data.mobile.trim() : '';
		if (phone) return phone;
		return '';
	}

	function resolveStatus(registration) {
		if (registration?.status) return registration.status;
		return 'pending';
	}

	$: eventData = eventInfo?.data && typeof eventInfo.data === 'object' ? eventInfo.data : {};
	$: locationText = stringValue(eventData?.location);
	$: dateLine = [formatEventDate(eventInfo?.event_date), locationText].filter(Boolean).join(' · ');

	function parseCreated(registration) {
		const raw = registration?.created;
		if (!raw || typeof raw !== 'string') return 0;
		const parsed = Date.parse(raw);
		return Number.isNaN(parsed) ? 0 : parsed;
	}

	function formatShortDate(raw) {
		if (!raw || typeof raw !== 'string') return '';
		const parsed = new Date(raw);
		if (Number.isNaN(parsed.getTime())) return '';
		const day = String(parsed.getDate()).padStart(2, '0');
		const month = parsed.toLocaleString('en-US', { month: 'short' }).toUpperCase();
		return `${day} ${month}`;
	}

	$: pendingItems = registrations
		.filter((r) => resolveStatus(r) === 'pending')
		.sort((a, b) => parseCreated(b) - parseCreated(a));
	$: approvedItems = registrations
		.filter((r) => resolveStatus(r) === 'active')
		.sort((a, b) => parseCreated(b) - parseCreated(a));
	$: cancelledItems = registrations
		.filter((r) => resolveStatus(r) === 'cancelled')
		.sort((a, b) => parseCreated(a) - parseCreated(b));
	$: rejectedItems = registrations
		.filter((r) => resolveStatus(r) === 'rejected')
		.sort((a, b) => parseCreated(b) - parseCreated(a));
	$: pendingCount = pendingItems.length;
	$: approvedCount = approvedItems.length;
	$: cancelledCount = cancelledItems.length;
	$: rejectedCount = rejectedItems.length;
	$: totalCount = pendingCount + approvedCount;

	async function loadDetails() {
		if (!eventId) return;
		loading = true;
		error = '';
		try {
			const response = await apiFetch(`/api/admin/events/${encodeURIComponent(eventId)}`);
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

	function openCancelConfirm(registration) {
		cancelConfirmTarget = registration;
		cancelConfirmOpen = true;
	}

	function closeCancelConfirm() {
		cancelConfirmOpen = false;
		cancelConfirmTarget = null;
	}

	function openRejectDialog(registration) {
		rejectTarget = registration;
		rejectNote = '';
		rejectDialogOpen = true;
	}

	function closeRejectDialog() {
		if (rejectLoading) return;
		rejectDialogOpen = false;
		rejectTarget = null;
		rejectNote = '';
	}

	async function confirmApprove() {
		if (!confirmTarget) return;
		confirmLoading = true;
		await tick();
		const startedAt = Date.now();
		try {
			const response = await apiFetch(`/api/admin/registrations/${encodeURIComponent(confirmTarget.id)}/approve`, {
				method: 'POST'
			});
			if (!response.ok) {
				throw new Error('Failed to approve');
			}
			const elapsed = Date.now() - startedAt;
			if (elapsed < 300) {
				await new Promise((resolve) => setTimeout(resolve, 300 - elapsed));
			}
			registrations = registrations.map((item) =>
				item.id === confirmTarget.id ? { ...item, status: 'active' } : item
			);
			closeConfirm();
		} catch (err) {
			error = err.message || err.toString() || 'Failed to approve';
		} finally {
			confirmLoading = false;
		}
	}

	async function cancelRegistration() {
		if (!cancelConfirmTarget || cancelingId) return;
		cancelingId = cancelConfirmTarget.id;
		error = '';
		try {
			const response = await apiFetch(
				`/api/admin/registrations/${encodeURIComponent(cancelConfirmTarget.id)}/cancel`,
				{
					method: 'POST'
				}
			);
			if (!response.ok) {
				throw new Error('Failed to cancel');
			}
			registrations = registrations.map((item) =>
				item.id === cancelConfirmTarget.id ? { ...item, status: 'cancelled' } : item
			);
			closeCancelConfirm();
		} catch (err) {
			error = err.message || err.toString() || 'Failed to cancel';
		} finally {
			cancelingId = null;
		}
	}

	async function rejectRegistration() {
		if (!rejectTarget || rejectLoading) return;
		const note = rejectNote.trim();
		if (!note) {
			error = 'Please add a reject note';
			return;
		}
		const targetId = rejectTarget.id;
		closeRejectDialog();

		rejectLoading = true;
		error = '';
		try {
			const response = await apiFetch(
				`/api/admin/registrations/${encodeURIComponent(targetId)}/reject`,
				{
					method: 'POST',
					headers: { 'Content-Type': 'application/json' },
					body: JSON.stringify({ note })
				}
			);
			if (!response.ok) {
				throw new Error('Failed to reject');
			}
			registrations = registrations.map((item) =>
				item.id === targetId
					? { ...item, status: 'rejected', data: { ...(item.data || {}), rejected: note } }
					: item
			);
			if (typeof window !== 'undefined') {
				window.scrollTo({ top: 0, behavior: 'smooth' });
			}
		} catch (err) {
			error = err.message || err.toString() || 'Failed to reject';
		} finally {
			rejectLoading = false;
		}
	}

	function goBack() {
		navigate('admin');
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
		<Card variant="state">Missing event.</Card>
	{:else if error}
		<Card variant="state">{error}</Card>
	{:else if loading}
		<Card variant="state">Loading...</Card>
	{:else}
		<div class="header">
			<h1>{eventInfo?.title}</h1>
			{#if dateLine}
				<p class="date">{dateLine}</p>
			{/if}
			<StatRow
				center
				items={[
					{ label: 'Total', value: totalCount },
					{ label: 'Pending', value: pendingCount },
				]}
			/>
		</div>

		<Card variant="admin">
			<div class="list">
				<div class="list-section">
					<h2>Pending</h2>
					{#if pendingCount === 0}
						<p class="empty">No pending registrations.</p>
					{:else}
						{#each pendingItems as reg}
							<div class="row">
								<div class="row-info">
									<p class="meta">{formatShortDate(reg.created)}</p>
									<p class="name">{displayName(reg)}</p>
									<p class="email">{displayRegion(reg)}</p>
									{#if displayPhone(reg)}
										<p class="phone">{displayPhone(reg)}</p>
									{/if}
								</div>
								<div class="row-actions">
									<button class="approve" on:click={() => openConfirm(reg)}>Approve</button>
									<button class="reject" on:click={() => openRejectDialog(reg)}>Reject</button>
								</div>
							</div>
						{/each}
					{/if}
				</div>
			</div>
		</Card>

		<Card variant="admin">
			<div class="list">
				<div class="list-section">
					<h2>Approved</h2>
					{#if approvedCount === 0}
						<p class="empty">No approved registrations.</p>
					{:else}
						{#each approvedItems as reg}
							<div class="row">
								<div class="row-info">
									<p class="name">{displayName(reg)}</p>
									<p class="email">{displayRegion(reg)}</p>
								</div>
								<div class="row-actions">
									{#if reg.hasUser}
										<span class="user-badge" aria-label="Registered user">
											<svg viewBox="0 0 24 24" aria-hidden="true">
												<path
													d="M12 12a4 4 0 1 0-4-4 4 4 0 0 0 4 4Zm0 2c-3.33 0-7 1.67-7 5v1h14v-1c0-3.33-3.67-5-7-5Z"
												/>
											</svg>
										</span>
									{/if}
									<button
										class="remove"
										disabled={cancelingId === reg.id}
										aria-label="Remove registration"
										on:click={() => openCancelConfirm(reg)}
									>
										x
									</button>
								</div>
							</div>
						{/each}
					{/if}
				</div>
			</div>
		</Card>

		<Card variant="admin">
			<div class="list">
				<div class="list-section">
					<h2>Cancelled</h2>
					{#if cancelledCount === 0}
						<p class="empty">No cancelled registrations.</p>
					{:else}
						{#each cancelledItems as reg}
							<div class="row">
								<div class="row-info">
									<p class="name">{displayName(reg)}</p>
									<p class="email">{displayRegion(reg)}</p>
								</div>
								<div class="row-actions"></div>
							</div>
						{/each}
					{/if}
				</div>
			</div>
		</Card>

		<Card variant="admin">
			<div class="list">
				<div class="list-section">
					<h2>Rejected</h2>
					{#if rejectedCount === 0}
						<p class="empty">No rejected registrations.</p>
					{:else}
						{#each rejectedItems as reg}
							<div class="row">
								<div class="row-info">
									<p class="name">{displayName(reg)}</p>
									<p class="email">{displayRegion(reg)}</p>
									{#if typeof reg?.data?.rejected === 'string' && reg.data.rejected.trim()}
										<p class="rejected-note">{reg.data.rejected}</p>
									{/if}
								</div>
								<div class="row-actions"></div>
							</div>
						{/each}
					{/if}
				</div>
			</div>
		</Card>
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

<ConfirmModal
	show={cancelConfirmOpen}
	title="Cancel registration"
	message="Are you sure you want to cancel this registration?"
	confirmLabel="Cancel"
	loading={cancelingId !== null}
	onConfirm={cancelRegistration}
	onCancel={closeCancelConfirm}
/>

{#if rejectDialogOpen}
	<div
		class="reject-overlay"
		role="button"
		tabindex="0"
		on:click={closeRejectDialog}
		on:keydown={(e) => {
			if (e.key === 'Enter' || e.key === ' ') closeRejectDialog();
		}}
	>
		<div
			class="reject-dialog"
			role="dialog"
			aria-modal="true"
			aria-label="Reject registration"
			tabindex="-1"
			on:click|stopPropagation
			on:keydown|stopPropagation={() => {}}
		>
			<h3>Reject registration</h3>
			<p>Add a short note about why we are rejecting this registration.</p>
			<textarea
				bind:value={rejectNote}
				rows="4"
				placeholder="Reason..."
				disabled={rejectLoading}
				></textarea>
				<div class="reject-actions">
					<button type="button" class="reject-cancel" on:click={closeRejectDialog} disabled={rejectLoading}>Cancel</button>
					<button type="button" class="reject-confirm" on:click={rejectRegistration} disabled={rejectLoading}>
						{rejectLoading ? 'Saving...' : 'Reject'}
					</button>
				</div>
		</div>
	</div>
{/if}

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
		display: grid;
		grid-template-columns: minmax(0, 1fr) auto;
		align-items: center;
		gap: 1rem;
		padding: 0.75rem 0;
		border-top: 1px solid #000;
	}

	.row-info {
		min-width: 0;
		overflow: hidden;
	}

	.row-actions {
		display: inline-flex;
		align-items: center;
		gap: 0.5rem;
		justify-content: flex-end;
		min-width: 8.5rem;
	}

	.user-badge {
		width: 1.4rem;
		height: 1.4rem;
		border-radius: 999px;
		border: 2px solid #000;
		display: inline-flex;
		align-items: center;
		justify-content: center;
	}

	.user-badge svg {
		width: 0.85rem;
		height: 0.85rem;
		fill: #000;
		display: block;
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

	.phone {
		margin: 0.25rem 0 0 0;
		font-size: 0.9rem;
		color: #333;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.rejected-note {
		margin: 0.3rem 0 0 0;
		font-size: 0.85rem;
		color: #555;
	}

	.meta {
		margin: 0 0 0.2rem 0;
		font-size: 0.7rem;
		color: #999;
		letter-spacing: 0.08em;
		text-transform: uppercase;
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

	.reject {
		border: 2px solid #000;
		background: #fff;
		color: #000;
		padding: 0.5rem 0.9rem;
		font-weight: 700;
		cursor: pointer;
	}

	.reject:hover {
		background: #f1f1f1;
	}

	.remove {
		border: none;
		background: #fff;
		color: #000;
		width: 2.1rem;
		height: 2.1rem;
		font-weight: 700;
		cursor: pointer;
		padding: 0;
	}

	.remove:disabled {
		opacity: 0.5;
		cursor: default;
	}

	.remove:hover:enabled {
		background: #f1f1f1;
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

	.reject-overlay {
		position: fixed;
		inset: 0;
		background: rgba(0, 0, 0, 0.5);
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 1rem;
		z-index: 1000;
	}

	.reject-dialog {
		width: min(32rem, 100%);
		background: #fff;
		border: 2px solid #000;
		padding: 1rem;
		display: grid;
		gap: 0.75rem;
	}

	.reject-dialog h3 {
		margin: 0;
		font-size: 1rem;
	}

	.reject-dialog p {
		margin: 0;
		font-size: 0.9rem;
		color: #333;
	}

	.reject-dialog textarea {
		width: 100%;
		border: 1px solid #000;
		padding: 0.6rem;
		resize: vertical;
		font-family: inherit;
		font-size: 0.9rem;
	}

	.reject-actions {
		display: flex;
		justify-content: flex-end;
		gap: 0.5rem;
	}

	.reject-cancel,
	.reject-confirm {
		border: 2px solid #000;
		padding: 0.45rem 0.8rem;
		font-weight: 700;
		cursor: pointer;
	}

	.reject-cancel {
		background: #fff;
		color: #000;
	}

	.reject-confirm {
		background: #000;
		color: #fff;
	}

	@media (max-width: 480px) {
		.row {
			gap: 0.5rem;
		}
		.row-actions {
			min-width: 6.5rem;
		}
	}
</style>
