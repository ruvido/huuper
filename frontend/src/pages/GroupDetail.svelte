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
	let requestsCount = 0;
	let lastGroupId = '';

	$: groupId = parseGroupId($currentRoute);
	$: actorId = pb.authStore.record?.id || '';
	$: isAdmin = !!pb.authStore.record?.admin;
	$: isAssistant = !!(group?.assistant && actorId && group.assistant === actorId);
	$: meInMembers = members.find((item) => item?.id === actorId) || null;
	$: isGuardian = !!meInMembers?.is_guardian;
	$: canOpenRequests = isAdmin || isAssistant || isGuardian;

	$: if (groupId && groupId !== lastGroupId) {
		lastGroupId = groupId;
		void loadAll();
	}

	function parseGroupId(route) {
		if (typeof route !== 'string') return '';
		const prefix = 'app/groups/';
		if (!route.startsWith(prefix)) return '';
		const tail = route.slice(prefix.length).trim();
		return tail.split('/')[0] || '';
	}

	function displayName(item) {
		if (typeof item?.full_name === 'string' && item.full_name.trim()) return item.full_name.trim();
		if (typeof item?.data?.full_name === 'string' && item.data.full_name.trim()) return item.data.full_name.trim();
		return item?.email || 'Unknown';
	}

	function openRequests() {
		if (!canOpenRequests || !groupId) return;
		navigate(`app/groups/${encodeURIComponent(groupId)}/requests`);
	}

	async function loadAll() {
		if (!groupId) return;
		loading = true;
		error = '';
		try {
			const fetchJSONOrThrow = async (path) => {
				const response = await apiFetch(path);
				if (!response.ok) {
					throw new Error('Failed to load group data');
				}
				return response.json();
			};

			const [groupData, membersData, requestsCountData] = await Promise.all([
				pb.collection('groups').getOne(groupId),
				fetchJSONOrThrow(`/api/groups/${encodeURIComponent(groupId)}/members`),
				fetchJSONOrThrow(`/api/groups/${encodeURIComponent(groupId)}/requests-count`)
			]);

			if (!groupData) throw new Error('Group not found');
			group = groupData;
			members = Array.isArray(membersData?.items) ? membersData.items : [];
			requestsCount = Number(requestsCountData?.count ?? 0);
		} catch (err) {
			error = err?.message || err?.toString() || 'Failed to load group';
		} finally {
			loading = false;
		}
	}

</script>

<DashboardLayout title={group?.name || 'Group'}>
	{#if error}
		<StateCard>{error}</StateCard>
	{:else if loading}
		<StateCard>Loading group...</StateCard>
	{:else}
		<AdminCard>
			<button class="requests-card" type="button" disabled={!canOpenRequests} on:click={openRequests}>
				<h2>{requestsCount} requests</h2>
			</button>
		</AdminCard>

		<AdminCard>
			<div class="list">
				<div class="list-section">
					{#if members.length === 0}
						<p class="empty">No members.</p>
					{:else}
						{#each members as member}
							<div class="row">
								<div class="row-info">
									<p class="name">{displayName(member)}</p>
								</div>
							</div>
						{/each}
					{/if}
				</div>
			</div>
		</AdminCard>
	{/if}
</DashboardLayout>

<style>
	h2 {
		margin: 0;
		font-size: 1.1rem;
	}

	.requests-card {
		width: 100%;
		text-align: left;
		border: none;
		padding: 0;
		background: transparent;
		cursor: pointer;
	}

	.requests-card:disabled {
		cursor: default;
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
		grid-template-columns: minmax(0, 1fr);
		align-items: center;
		gap: 1rem;
		padding: 0.75rem 0;
		border-top: 1px solid #000;
	}

	.row-info {
		min-width: 0;
		overflow: hidden;
	}

	.name {
		margin: 0;
		font-weight: 700;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.empty {
		margin: 0;
	}
</style>
