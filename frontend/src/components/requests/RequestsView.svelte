<script>
	import { onMount } from 'svelte';
	import { apiFetch } from '../../lib/pocketbase';
	import { navigate, queryParams } from '../../lib/router';
	import DashboardLayout from '../DashboardLayout.svelte';
	import StateCard from '../StateCard.svelte';

	export let title = 'Requests';
	export let adminMode = false; // kept for compatibility with callers
	$: adminMode;

	let loading = true;
	let error = '';
	let requests = [];
	let lastGroupFilter = '';

	$: groupFilter = ($queryParams?.group_id || '').trim();

	$: if (groupFilter !== lastGroupFilter) {
		lastGroupFilter = groupFilter;
		void loadRequests();
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

	function openRequest(item) {
		if (!item?.id) return;
		navigate(`app/requests/${encodeURIComponent(item.id)}`);
	}

	async function loadRequests() {
		loading = true;
		error = '';
		const params = new URLSearchParams();
		if (groupFilter) params.set('group_id', groupFilter);

		const query = params.toString();
		try {
			const response = await apiFetch(`/api/requests${query ? `?${query}` : ''}`);
			if (!response.ok) {
				throw new Error('Failed to load requests');
			}
			const data = await response.json();
			requests = Array.isArray(data?.items) ? data.items : [];
		} catch (err) {
			error = err?.message || err?.toString() || 'Failed to load requests';
		} finally {
			loading = false;
		}
	}

	onMount(loadRequests);
</script>

<DashboardLayout {title}>
	{#if error}
		<StateCard>{error}</StateCard>
	{:else if loading}
		<StateCard>Loading requests...</StateCard>
	{:else if requests.length === 0}
		<StateCard>No requests.</StateCard>
	{:else}
		<div class="list">
			{#each requests as item}
				<button class="row" type="button" on:click={() => openRequest(item)}>
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
</DashboardLayout>

<style>
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
</style>
