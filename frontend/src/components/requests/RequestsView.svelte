<script>
	import { onMount } from 'svelte';
	import { apiFetch, fetchSetting } from '../../lib/pocketbase';
	import { currentRoute, navigate, queryParams } from '../../lib/router';
	import DashboardLayout from '../DashboardLayout.svelte';
	import StateCard from '../StateCard.svelte';
	import AdminCard from '../AdminCard.svelte';

	export let title = 'Requests';
	export let adminMode = false;

	let loading = true;
	let error = '';
	let statuses = [];
	let requests = [];
	let showRejected = false;
	let lastGroupFilter = '';
	let selectedId = '';
	let selectedRequest = null;
	let loadingSelected = false;

	$: groupFilter = ($queryParams?.group_id || '').trim();
	$: routeSelectedId = selectedIdFromRoute($currentRoute);
	$: selectedId = ($queryParams?.id || routeSelectedId || '').trim();

	$: activeRequests = requests.filter((item) => !item.rejected);
	$: rejectedRequests = requests.filter((item) => !!item.rejected);
	$: firstStatus = statuses.length > 0 ? statuses[0] : '';
	$: lastStatus = statuses.length > 0 ? statuses[statuses.length - 1] : '';
	$: priorityItems = activeRequests.filter((item) => item.status === firstStatus || item.status === lastStatus);
	$: otherStatuses = statuses.filter((status) => status !== firstStatus && status !== lastStatus);

	$: if (groupFilter !== lastGroupFilter) {
		lastGroupFilter = groupFilter;
		loadAll();
	}

	$: if (selectedId) {
		void loadSelected(selectedId);
	} else {
		selectedRequest = null;
	}

	function selectedIdFromRoute(route) {
		if (typeof route !== 'string') return '';
		if (route.startsWith('app/admin/requests/')) return route.slice('app/admin/requests/'.length).trim();
		if (route.startsWith('app/requests/')) return route.slice('app/requests/'.length).trim();
		return '';
	}

	function displayName(item) {
		const value = item?.data?.full_name;
		if (typeof value === 'string' && value.trim()) return value.trim();
		return item?.email || 'Unknown';
	}

	function displayRegion(item) {
		const value = item?.data?.region;
		if (typeof value === 'string' && value.trim()) return value.trim();
		return '';
	}

	function formatStatus(status) {
		if (!status || typeof status !== 'string') return 'Unknown';
		const clean = status.replace(/^\d+-/, '').replaceAll('_', ' ');
		if (!clean) return status;
		return clean.charAt(0).toUpperCase() + clean.slice(1);
	}

	function requestByStatus(status) {
		return activeRequests.filter((item) => item.status === status);
	}

	function openRequest(item) {
		if (!item?.id) return;
		const prefix = adminMode ? 'app/admin/requests/' : 'app/requests/';
		navigate(`${prefix}${encodeURIComponent(item.id)}${groupFilter ? `?group_id=${encodeURIComponent(groupFilter)}` : ''}`);
	}

	async function loadStatuses() {
		const response = await fetchSetting('requests_flow');
		if (!response.ok) {
			throw new Error('Failed to load requests flow');
		}
		const data = await response.json();
		const list = Array.isArray(data?.data?.statuses) ? data.data.statuses : [];
		statuses = list.filter((item) => typeof item === 'string' && item.trim()).map((item) => item.trim());
	}

	async function loadRequests() {
		const params = new URLSearchParams();
		if (groupFilter) params.set('group_id', groupFilter);
		if (adminMode && showRejected) params.set('include_rejected', 'true');

		const query = params.toString();
		const response = await apiFetch(`/api/requests${query ? `?${query}` : ''}`);
		if (!response.ok) {
			throw new Error('Failed to load requests');
		}
		const data = await response.json();
		requests = Array.isArray(data?.items) ? data.items : [];
	}

	async function loadSelected(id) {
		if (!id) {
			selectedRequest = null;
			return;
		}

		const existing = requests.find((item) => item.id === id);
		if (existing) {
			selectedRequest = existing;
			return;
		}

		loadingSelected = true;
		try {
			const response = await apiFetch(`/api/requests/${encodeURIComponent(id)}`);
			if (!response.ok) {
				selectedRequest = null;
				return;
			}
			selectedRequest = await response.json();
		} finally {
			loadingSelected = false;
		}
	}

	async function loadAll() {
		loading = true;
		error = '';
		try {
			await loadStatuses();
			await loadRequests();
		} catch (err) {
			error = err?.message || err?.toString() || 'Failed to load requests';
		} finally {
			loading = false;
		}
	}

	async function handleRejectedToggle(event) {
		showRejected = !!event.currentTarget.checked;
		await loadAll();
	}

	onMount(loadAll);
</script>

<DashboardLayout {title}>
	{#if error}
		<StateCard>{error}</StateCard>
	{:else if loading}
		<StateCard>Loading requests...</StateCard>
	{:else}
		<AdminCard>
			<h2>Priority</h2>
			{#if priorityItems.length === 0}
				<p class="empty">No priority requests.</p>
			{:else}
				<div class="list">
					{#each priorityItems as item}
						<button class="row" on:click={() => openRequest(item)}>
							<div>
								<p class="name">{displayName(item)}</p>
								{#if displayRegion(item)}
									<p class="meta">{displayRegion(item)}</p>
								{/if}
							</div>
							<span class="status">{formatStatus(item.status)}</span>
						</button>
					{/each}
				</div>
			{/if}
		</AdminCard>

		{#each otherStatuses as status}
			<AdminCard>
				<h2>{formatStatus(status)}</h2>
				{#if requestByStatus(status).length === 0}
					<p class="empty">No requests in this step.</p>
				{:else}
					<div class="list">
						{#each requestByStatus(status) as item}
							<button class="row" on:click={() => openRequest(item)}>
								<div>
									<p class="name">{displayName(item)}</p>
								</div>
								<span class="status">{formatStatus(item.status)}</span>
							</button>
						{/each}
					</div>
				{/if}
			</AdminCard>
		{/each}

		{#if selectedRequest || loadingSelected}
			<AdminCard>
				<h2>Request detail</h2>
				{#if loadingSelected}
					<p class="empty">Loading...</p>
				{:else if selectedRequest}
					<p class="name">{displayName(selectedRequest)}</p>
					<p class="meta">{selectedRequest.email}</p>
					<p class="meta">Status: {formatStatus(selectedRequest.status)}</p>
					{#if selectedRequest.group}
						<p class="meta">Group: {selectedRequest.group}</p>
					{/if}
					{#if selectedRequest.guardian}
						<p class="meta">Guardian: {selectedRequest.guardian}</p>
					{/if}
				{/if}
			</AdminCard>
		{/if}

		{#if adminMode && showRejected}
			<AdminCard>
				<h2>Rejected</h2>
				{#if rejectedRequests.length === 0}
					<p class="empty">No rejected requests.</p>
				{:else}
					<div class="list">
						{#each rejectedRequests as item}
							<button class="row" on:click={() => openRequest(item)}>
								<div>
									<p class="name">{displayName(item)}</p>
								</div>
								<span class="status">Rejected</span>
							</button>
						{/each}
					</div>
				{/if}
			</AdminCard>
		{/if}

		{#if adminMode || groupFilter}
			<AdminCard>
				<div class="toolbar">
					{#if groupFilter}
						<p class="context">Group filter: {groupFilter}</p>
					{/if}
					{#if adminMode}
						<label class="toggle">
							<input type="checkbox" checked={showRejected} on:change={handleRejectedToggle} />
							<span>Show rejected</span>
						</label>
					{/if}
				</div>
			</AdminCard>
		{/if}
	{/if}
</DashboardLayout>

<style>
	h2 {
		margin: 0 0 0.75rem;
		font-size: 1.1rem;
	}

	.toolbar {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 0.75rem;
	}

	.context {
		margin: 0;
		font-weight: 600;
	}

	.toggle {
		display: inline-flex;
		align-items: center;
		gap: 0.5rem;
		font-weight: 600;
	}

	.list {
		display: grid;
		gap: 0.5rem;
	}

	.row {
		display: flex;
		justify-content: space-between;
		align-items: center;
		gap: 0.75rem;
		width: 100%;
		text-align: left;
		padding: 0.75rem;
		border: 2px solid #000;
		background: #fff;
		cursor: pointer;
	}

	.name {
		margin: 0;
		font-weight: 700;
	}

	.meta {
		margin: 0.2rem 0 0;
		font-size: 0.9rem;
	}

	.status {
		font-size: 0.8rem;
		font-weight: 700;
		text-transform: uppercase;
		letter-spacing: 0.03em;
	}

	.empty {
		margin: 0;
		font-size: 0.95rem;
	}
</style>
