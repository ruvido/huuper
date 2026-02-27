<script>
	import { apiFetch, pb } from '../lib/pocketbase';
	import { currentRoute, navigate } from '../lib/router';
	import DashboardLayout from '../components/DashboardLayout.svelte';
	import Card from '../components/Card.svelte';

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
		<Card variant="state">{error}</Card>
	{:else if loading}
		<Card variant="state">Loading...</Card>
	{:else if !canViewRequests}
		<Card variant="state">Not allowed.</Card>
	{:else}
		<Card variant="admin">
			<h2>{requests.length} requests</h2>
		</Card>

		<Card variant="admin">
			<div class="list">
				{#if requests.length === 0}
					<p class="empty">No requests.</p>
				{:else}
					{#each requests as item}
						<button class="row-action" type="button" on:click={() => openRequest(item)}>
							<Card variant="item">
								<p class="name">{displayName(item)}</p>
								<p class="status">{displayStatus(item)}</p>
							</Card>
						</button>
					{/each}
				{/if}
			</div>
		</Card>
	{/if}
</DashboardLayout>

<style>
	h2 { margin: 0; }

	.list {
		display: grid;
		gap: 0.5rem;
	}

	.row-action {
		width: 100%;
		border: none;
		background: transparent;
		padding: 0;
		cursor: pointer;
	}

	.name {
		margin: 0;
		font-weight: 700;
	}

	.status {
		margin: 0.2rem 0 0;
		font-size: var(--ui-font-size);
	}

	.row-action + .row-action {
		margin-top: 0.5rem;
	}

	.empty {
		margin: 0;
	}
</style>
