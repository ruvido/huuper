<script>
	import { apiFetch, pb } from '../lib/pocketbase';
	import { currentRoute, navigate } from '../lib/router';
	import DashboardLayout from '../components/DashboardLayout.svelte';
	import Card from '../components/Card.svelte';
	import ActionDialog from '../components/modals/ActionDialog.svelte';
	import { adminUsersCopy } from '../lib/copy/adminUsers';

	let loading = true;
	let error = '';
	let group = null;
	let members = [];
	let requests = [];
	let lastGroupId = '';
	let currentScope = 'app';
	let deleteDialogOpen = false;
	let memberToDelete = null;
	let deletingMember = false;
	let deleteError = '';

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

	function deleteAriaLabel(item) {
		return `${adminUsersCopy.deleteButtonAriaPrefix} ${displayName(item)}`;
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

	function openDeleteDialog(member) {
		if (!isAdmin || !member?.id) return;
		memberToDelete = member;
		deleteDialogOpen = true;
		deleteError = '';
	}

	function closeDeleteDialog() {
		if (deletingMember) return;
		deleteDialogOpen = false;
		memberToDelete = null;
		deleteError = '';
	}

	async function confirmDeleteMember() {
		const userId = memberToDelete?.id || '';
		if (!isAdmin || !userId || deletingMember) return false;
		deletingMember = true;
		deleteError = '';
		try {
			const response = await apiFetch(`/api/admin/users/${encodeURIComponent(userId)}`, {
				method: 'DELETE'
			});
			if (!response.ok) {
				const payload = await response.json().catch(() => ({}));
				throw new Error(payload?.message || adminUsersCopy.deleteErrorGeneric);
			}
			await loadAll();
			closeDeleteDialog();
			return true;
		} catch (err) {
			deleteError = err?.message || adminUsersCopy.deleteErrorGeneric;
			return false;
		} finally {
			deletingMember = false;
		}
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
								<div class="member-row">
									<p class="name">{displayName(member)}</p>
									{#if isAdmin}
										<button
											type="button"
											class="member-delete"
											on:click={() => openDeleteDialog(member)}
											aria-label={deleteAriaLabel(member)}
										>
											<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="currentColor" viewBox="0 0 16 16" aria-hidden="true">
												<path d="M5.5 5.5A.5.5 0 0 1 6 6v6a.5.5 0 0 1-1 0V6a.5.5 0 0 1 .5-.5m2.5 0a.5.5 0 0 1 .5.5v6a.5.5 0 0 1-1 0V6a.5.5 0 0 1 .5-.5m3 .5a.5.5 0 0 0-1 0v6a.5.5 0 0 0 1 0z"/>
												<path d="M14.5 3a1 1 0 0 1-1 1H13v9a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V4h-.5a1 1 0 0 1-1-1V2a1 1 0 0 1 1-1H6a1 1 0 0 1 1-1h2a1 1 0 0 1 1 1h3.5a1 1 0 0 1 1 1zM4.118 4 4 4.059V13a1 1 0 0 0 1 1h6a1 1 0 0 0 1-1V4.059L11.882 4zM2.5 3h11V2h-11z"/>
											</svg>
										</button>
									{/if}
								</div>
							</Card>
						{/each}
					{/if}
				</div>
			</div>
		</Card>
	{/if}
</DashboardLayout>

<ActionDialog
	show={deleteDialogOpen}
	title={adminUsersCopy.deleteDialogTitle}
	message={deleteError || adminUsersCopy.deleteDialogMessage}
	confirmLabel={adminUsersCopy.deleteDialogConfirmLabel}
	loading={deletingMember}
	onConfirm={confirmDeleteMember}
	onCancel={closeDeleteDialog}
/>

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

	.member-row {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 0.6rem;
	}

	.member-delete {
		border: none;
		background: transparent;
		padding: 0.2rem;
		display: inline-flex;
		align-items: center;
		justify-content: center;
		color: #c40000;
		cursor: pointer;
	}

	.empty {
		margin: 0;
	}
</style>
