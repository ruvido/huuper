<script>
	import { onMount } from 'svelte';
	import { authRecord, apiFetch, pb } from '../lib/pocketbase';
	import { currentRoute, navigate } from '../lib/router';
	import DashboardLayout from '../components/DashboardLayout.svelte';
	import StateCard from '../components/StateCard.svelte';
	import AdminCard from '../components/AdminCard.svelte';

	let loading = true;
	let error = '';
	let group = null;
	let members = [];
	let guardians = [];
	let requests = [];
	let assigningById = {};
	let selectedGuardianByReq = {};
	let lastGroupId = '';

	$: groupId = parseGroupId($currentRoute);
	$: isAdmin = !!$authRecord?.admin;
	$: canManage = isAdmin || (group?.assistant && group.assistant === pb.authStore.record?.id);

	$: if (groupId && groupId !== lastGroupId) {
		lastGroupId = groupId;
		loadAll();
	}

	function parseGroupId(route) {
		if (typeof route !== 'string') return '';
		const prefix = 'app/group/';
		if (!route.startsWith(prefix)) return '';
		return route.slice(prefix.length).trim();
	}

	function displayName(item) {
		if (typeof item?.full_name === 'string' && item.full_name.trim()) return item.full_name.trim();
		if (typeof item?.data?.full_name === 'string' && item.data.full_name.trim()) return item.data.full_name.trim();
		return item?.email || 'Unknown';
	}

	function formatStatus(status) {
		if (!status || typeof status !== 'string') return 'Unknown';
		const clean = status.replace(/^\d+-/, '').replaceAll('_', ' ');
		if (!clean) return status;
		return clean.charAt(0).toUpperCase() + clean.slice(1);
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

			const [groupData, membersData, guardiansData, requestsData] = await Promise.all([
				pb.collection('groups').getOne(groupId),
				fetchJSONOrThrow(`/api/groups/${encodeURIComponent(groupId)}/members`),
				fetchJSONOrThrow(`/api/groups/${encodeURIComponent(groupId)}/guardians`),
				fetchJSONOrThrow(`/api/requests?group_id=${encodeURIComponent(groupId)}`)
			]);

			if (!groupData) throw new Error('Group not found');
			group = groupData;
			members = Array.isArray(membersData?.items) ? membersData.items : [];
			guardians = Array.isArray(guardiansData?.items) ? guardiansData.items : [];
			requests = Array.isArray(requestsData?.items) ? requestsData.items : [];
		} catch (err) {
			error = err?.message || err?.toString() || 'Failed to load group';
		} finally {
			loading = false;
		}
	}

	function openRequest(item) {
		if (!item?.id) return;
		if (isAdmin) {
			navigate(`app/admin/requests/${encodeURIComponent(item.id)}?group_id=${encodeURIComponent(groupId)}`);
			return;
		}
		navigate(`app/requests/${encodeURIComponent(item.id)}?group_id=${encodeURIComponent(groupId)}`);
	}

	$: guardiansById = Object.fromEntries(guardians.map((guardian) => [guardian.id, guardian]));
	$: memberOptions = members.map((member) => {
		const existing = guardiansById[member.id];
		const load = existing ? ` (${existing.proteges_count})` : ' (new guardian)';
		return {
			id: member.id,
			label: `${displayName(member)}${load}`
		};
	});

	async function assignGuardian(item) {
		if (!item?.id || !canManage) return;
		const guardianId = selectedGuardianByReq[item.id];
		if (!guardianId) return;

		assigningById = { ...assigningById, [item.id]: true };
		error = '';
		try {
			const response = await apiFetch(`/api/requests/${encodeURIComponent(item.id)}/action`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					action: 'transition',
					target_status: '3-guardian_assigned',
					guardian: guardianId
				})
			});
			if (!response.ok) {
				throw new Error('Failed to assign guardian');
			}
			await loadAll();
		} catch (err) {
			error = err?.message || err?.toString() || 'Failed to assign guardian';
		} finally {
			assigningById = { ...assigningById, [item.id]: false };
		}
	}

	onMount(() => {
		if (groupId) {
			lastGroupId = groupId;
			loadAll();
		}
	});
</script>

<DashboardLayout title={group?.name || 'Group'}>
	{#if error}
		<StateCard>{error}</StateCard>
	{:else if loading}
		<StateCard>Loading group...</StateCard>
	{:else}
		<AdminCard>
			<h2>Group Requests</h2>
			{#if requests.length === 0}
				<p class="empty">No requests.</p>
			{:else}
				<div class="list">
					{#each requests as item}
						<div class="row">
							<button class="row-main" on:click={() => openRequest(item)}>
								<p class="name">{displayName(item)}</p>
								<p class="meta">{item.email}</p>
								<p class="meta">{formatStatus(item.status)}</p>
							</button>
							{#if canManage && item.status === '2-group_assigned' && memberOptions.length > 0}
								<div class="assign">
									<select bind:value={selectedGuardianByReq[item.id]}>
										<option value="">Select guardian</option>
										{#each memberOptions as member}
											<option value={member.id}>{member.label}</option>
										{/each}
									</select>
									<button disabled={!selectedGuardianByReq[item.id] || assigningById[item.id]} on:click={() => assignGuardian(item)}>
										{assigningById[item.id] ? 'Assigning...' : 'Assign'}
									</button>
								</div>
							{/if}
						</div>
					{/each}
				</div>
			{/if}
		</AdminCard>

		<AdminCard>
			<h2>Guardians ({guardians.length})</h2>
			{#if guardians.length === 0}
				<p class="empty">No guardians yet.</p>
			{:else}
				<div class="list">
					{#each guardians as guardian}
						<div class="row-static">
							<p class="name">{displayName(guardian)}</p>
							<p class="meta">{guardian.email}</p>
							<p class="meta">{guardian.proteges_count} protegees</p>
						</div>
					{/each}
				</div>
			{/if}
		</AdminCard>

		<AdminCard>
			<h2>Members ({members.length})</h2>
			{#if members.length === 0}
				<p class="empty">No members.</p>
			{:else}
				<div class="list">
					{#each members as member}
						<div class="row-static">
							<div>
								<p class="name">{displayName(member)}</p>
								<p class="meta">{member.email}</p>
							</div>
							<div>
								{#if member.is_guardian}
									<span class="badge">Guardian</span>
								{/if}
							</div>
						</div>
					{/each}
				</div>
			{/if}
		</AdminCard>
	{/if}
</DashboardLayout>

<style>
	h2 {
		margin: 0 0 0.75rem;
		font-size: 1.1rem;
	}

	.list {
		display: grid;
		gap: 0.5rem;
	}

	.row,
	.row-static {
		display: flex;
		justify-content: space-between;
		align-items: center;
		gap: 0.75rem;
		padding: 0.75rem;
		border: 2px solid #000;
		background: #fff;
	}

	.row-main {
		border: none;
		background: transparent;
		text-align: left;
		cursor: pointer;
		padding: 0;
		flex: 1;
	}

	.assign {
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}

	.name {
		margin: 0;
		font-weight: 700;
	}

	.meta {
		margin: 0.2rem 0 0;
		font-size: 0.9rem;
	}

	.badge {
		display: inline-block;
		padding: 0.25rem 0.5rem;
		border: 2px solid #000;
		font-size: 0.8rem;
		font-weight: 700;
	}

	.empty {
		margin: 0;
	}
</style>
