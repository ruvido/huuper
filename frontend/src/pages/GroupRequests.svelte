<script>
	import { apiFetch, pb } from '../lib/pocketbase';
	import { currentRoute, navigate } from '../lib/router';
	import DashboardLayout from '../components/DashboardLayout.svelte';
	import StateCard from '../components/StateCard.svelte';
	import AdminCard from '../components/AdminCard.svelte';

	let loading = true;
	let error = '';
	let group = null;
	let members = [];
	let requests = [];
	let lastGroupId = '';

	$: groupId = parseGroupId($currentRoute);
	$: actorId = pb.authStore.record?.id || '';
	$: isAdmin = !!pb.authStore.record?.admin;
	$: isAssistant = !!(group?.assistant && actorId && group.assistant === actorId);
	$: meInMembers = members.find((item) => item?.id === actorId) || null;
	$: isGuardian = !!meInMembers?.is_guardian;
	$: canViewRequests = isAdmin || isAssistant || isGuardian;

	$: if (groupId && groupId !== lastGroupId) {
		lastGroupId = groupId;
		void loadAll();
	}

	function parseGroupId(route) {
		if (typeof route !== 'string') return '';
		const match = route.match(/^app\/groups\/([^/]+)\/requests$/);
		return match?.[1] || '';
	}

	function displayName(item) {
		if (typeof item?.data?.full_name === 'string' && item.data.full_name.trim()) return item.data.full_name.trim();
		return item?.email || 'Unknown';
	}

	function formatStatus(status) {
		if (!status || typeof status !== 'string') return '';
		const clean = status.replace(/^\d+-/, '').replaceAll('_', ' ');
		if (!clean) return '';
		return clean.charAt(0).toUpperCase() + clean.slice(1);
	}

	function displayStatus(item) {
		if (item?.rejected) return 'Rejected';
		return formatStatus(item?.status);
	}

	async function loadAll() {
		if (!groupId) return;
		loading = true;
		error = '';
		try {
			const fetchJSONOrThrow = async (path) => {
				const response = await apiFetch(path);
				if (!response.ok) throw new Error('Failed to load requests');
				return response.json();
			};

			const [groupData, membersData, requestsData] = await Promise.all([
				pb.collection('groups').getOne(groupId),
				fetchJSONOrThrow(`/api/groups/${encodeURIComponent(groupId)}/members`),
				fetchJSONOrThrow(`/api/requests?group_id=${encodeURIComponent(groupId)}`)
			]);

			group = groupData || null;
			members = Array.isArray(membersData?.items) ? membersData.items : [];
			requests = Array.isArray(requestsData?.items) ? requestsData.items : [];
		} catch (err) {
			error = err?.message || err?.toString() || 'Failed to load requests';
		} finally {
			loading = false;
		}
	}

	function openRequest(item) {
		if (!item?.id) return;
		navigate(`app/requests/${encodeURIComponent(item.id)}`);
	}
</script>

<DashboardLayout title={group?.name || 'Requests'}>
	{#if error}
		<StateCard>{error}</StateCard>
	{:else if loading}
		<StateCard>Loading...</StateCard>
	{:else if !canViewRequests}
		<StateCard>Not allowed.</StateCard>
	{:else}
		<AdminCard>
			<h2>{requests.length} requests</h2>
		</AdminCard>

		<AdminCard>
			<div class="list">
				{#if requests.length === 0}
					<p class="empty">No requests.</p>
				{:else}
					{#each requests as item}
						<button class="row" type="button" on:click={() => openRequest(item)}>
							<p class="name">{displayName(item)}</p>
							<p class="status">{displayStatus(item)}</p>
						</button>
					{/each}
				{/if}
			</div>
		</AdminCard>
	{/if}
</DashboardLayout>

<style>
	h2 {
		margin: 0;
		font-size: 1.1rem;
	}

	.list {
		display: grid;
		gap: 0.5rem;
	}

	.row {
		width: 100%;
		text-align: left;
		border: 1px solid #000;
		background: #fff;
		padding: 0.75rem;
		cursor: pointer;
	}

	.name {
		margin: 0;
		font-weight: 700;
	}

	.status {
		margin: 0.2rem 0 0;
		font-size: 0.9rem;
	}

	.empty {
		margin: 0;
	}
</style>
