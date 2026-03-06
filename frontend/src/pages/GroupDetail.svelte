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
	let currentScope = 'app';

	$: routeContext = parseRouteContext($currentRoute);
	$: groupId = routeContext.id;
	$: currentScope = routeContext.scope;
	$: actorId = pb.authStore.record?.id || '';
	$: isAdmin = !!pb.authStore.record?.admin;
	$: isAssistant = !!(group?.assistant && actorId && group.assistant === actorId);
	$: meInMembers = members.find((item) => item?.id === actorId) || null;
	$: isGuardian = !!meInMembers?.is_guardian;
	$: canViewRequests = isAdmin || isAssistant || isGuardian;
	$: visibleRequests = requests;

	$: if (groupId && groupId !== lastGroupId) {
		lastGroupId = groupId;
		void loadAll();
	}

	function parseRouteContext(route) {
		if (typeof route !== 'string') return { scope: 'app', id: '' };
		const match = route.match(/^(app|admin)\/groups\/([^/]+)$/);
		if (!match) return { scope: 'app', id: '' };
		return { scope: match[1], id: match[2] || '' };
	}

	function goBack() {
		window.history.back();
	}

	function displayName(item) {
		if (typeof item?.full_name === 'string' && item.full_name.trim()) return item.full_name.trim();
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

	function openRequest(item) {
		if (!item?.id) return;
		navigate(`${currentScope}/requests/${encodeURIComponent(item.id)}`);
	}

	async function loadAll() {
		if (!groupId) return;
		loading = true;
		error = '';
		try {
			const fetchJSONOrThrow = async (path) => {
				const response = await apiFetch(path);
				if (!response.ok) throw new Error('Failed to load group data');
				return response.json();
			};

			const requestsPath = (() => {
				const base = `/api/requests?group_id=${encodeURIComponent(groupId)}`;
				if (isGuardian && !isAssistant && !isAdmin && actorId) {
					return `${base}&guardian=${encodeURIComponent(actorId)}`;
				}
				return base;
			})();

			const [groupData, membersData, requestsData] = await Promise.all([
				pb.collection('groups').getOne(groupId),
				fetchJSONOrThrow(`/api/groups/${encodeURIComponent(groupId)}/members`),
				fetchJSONOrThrow(requestsPath)
			]);

			if (!groupData) throw new Error('Group not found');
			group = groupData;
			members = Array.isArray(membersData?.items) ? membersData.items : [];
			requests = Array.isArray(requestsData?.items) ? requestsData.items : [];
		} catch (err) {
			error = err?.message || err?.toString() || 'Failed to load group';
		} finally {
			loading = false;
		}
	}
</script>

<DashboardLayout title={group?.name || 'Group'} showBack={true} onBack={goBack}>

	{#if error}
		<Card variant="state">{error}</Card>
	{:else if loading}
		<Card variant="state">Loading group...</Card>
	{:else}
		{#if canViewRequests}
			<Card variant="admin">
				<div class="list">
					<div class="list-section">
						<h2>Requests</h2>
						{#if visibleRequests.length === 0}
							<p class="empty">No requests to manage.</p>
						{:else}
							{#each visibleRequests as item}
								<button class="request-action" type="button" on:click={() => openRequest(item)}>
									<Card variant="item">
										<p class="name">{displayName(item)}</p>
										<p class="status">{displayStatus(item)}</p>
									</Card>
								</button>
							{/each}
						{/if}
					</div>
				</div>
			</Card>
		{/if}

		<Card variant="admin">
			<div class="list">
				<div class="list-section">
					<h2>Members</h2>
					{#if members.length === 0}
						<p class="empty">No members.</p>
					{:else}
						{#each members as member}
							<Card variant="item">
								<p class="name">{displayName(member)}</p>
							</Card>
						{/each}
					{/if}
				</div>
			</div>
		</Card>
	{/if}
</DashboardLayout>

<style>
	h2 {
		margin: 0;
		font-size: var(--ui-font-size);
	}

	.list {
		display: grid;
		gap: 1.25rem;
	}

	.list-section {
		display: grid;
		gap: 0.5rem;
	}

	.request-action {
		border: none;
		background: transparent;
		padding: 0;
		width: 100%;
		cursor: pointer;
	}

	.status {
		margin: 0.2rem 0 0;
		font-size: var(--ui-font-size);
		font-weight: 500;
	}

	.list-section :global(.card.item) + :global(.card.item),
	.list-section .request-action + .request-action {
		margin-top: 0.5rem;
	}

	.name {
		margin: 0;
		font-size: var(--ui-font-size);
		font-weight: 700;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.empty {
		margin: 0;
	}
</style>
