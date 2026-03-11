<script>
	import { onMount, onDestroy } from 'svelte';
	import { apiFetch, pb } from '../lib/pocketbase';
	import { navigate } from '../lib/router';
	import DashboardLayout from '../components/DashboardLayout.svelte';
	import GroupCard from '../components/cards/GroupCard.svelte';
	import Card from '../components/Card.svelte';

	export let adminMode = false;

	let groups = [];
	let visibleGroups = [];
	let requestsCountByGroupId = {};
	let loaded = false;
	let error = '';
	let unsubscribeUserGroups;
	let inviteDialogOpen = false;
	let inviteDialogLink = '';
	let inviteDialogName = '';
	let inviteDialogMessage = '';
	let inviteDialogBusy = false;
	let syncBusy = false;
	let syncMessage = '';

	onMount(async () => {
		await loadGroups();
		const currentUser = pb.authStore.record;
		if (!currentUser?.id) return;

		try {
			unsubscribeUserGroups = await pb.collection('user_groups').subscribe('*', (e) => {
				if (e.record.user !== currentUser.id) return;
				void loadGroups();
			});
		} catch {
			// Ignore realtime errors
		}
	});

	onDestroy(() => {
		if (unsubscribeUserGroups) {
			unsubscribeUserGroups();
		}
	});

	async function loadGroups() {
		const currentUser = pb.authStore.record;
		if (!currentUser?.id) return;

		loaded = false;
		error = '';
		try {
			const userGroupsResult = await pb.collection('user_groups').getList(1, 500, {
				filter: `user = "${currentUser.id}"`,
				sort: '-created',
				expand: 'group'
			});
			const memberGroups = userGroupsResult.items
				.map((item) => item?.expand?.group || null)
				.filter(Boolean);
			groups = memberGroups;

			let defaultGroup = null;
			try {
				const defaultResult = await pb.collection('groups').getList(1, 1, {
					filter: `type = "default"`,
					sort: 'created'
				});
				defaultGroup = Array.isArray(defaultResult?.items) ? defaultResult.items[0] || null : null;
			} catch {
				// Ignore default group lookup errors
			}

			const byId = new Map();
			for (const group of memberGroups) {
				if (group?.id) byId.set(group.id, { group, isMember: true });
			}
			if (defaultGroup?.id && !byId.has(defaultGroup.id)) {
				byId.set(defaultGroup.id, { group: defaultGroup, isMember: false });
			}

			const rank = (type) => {
				const value = typeof type === 'string' ? type.trim() : '';
				if (value === 'default') return 0;
				if (value === 'local') return 1;
				return 2;
			};

			visibleGroups = Array.from(byId.values()).sort((a, b) => {
				const r = rank(a.group?.type) - rank(b.group?.type);
				if (r !== 0) return r;
				const aName = (a.group?.name || '').toLowerCase();
				const bName = (b.group?.name || '').toLowerCase();
				return aName.localeCompare(bName);
			});

			const countEntries = await Promise.all(
				memberGroups.map(async (group) => {
					if (!group?.id) {
						return [group?.id || '', 0];
					}

					const response = await apiFetch(`/api/requests?group_id=${encodeURIComponent(group.id)}&per_page=1`);
					if (!response.ok) {
						return [group.id, 0];
					}
					const payload = await response.json();
					const items = Array.isArray(payload?.items) ? payload.items : [];
					return [group.id, items.length];
				})
			);
			requestsCountByGroupId = Object.fromEntries(countEntries.filter(([id]) => !!id));
		} catch (err) {
			error = err.message || err.toString() || 'Failed to load groups';
		} finally {
			loaded = true;
		}
	}

	function goToGroup(group) {
		if (!group?.id) return;
		const scope = adminMode ? 'admin' : 'app';
		navigate(`${scope}/groups/${encodeURIComponent(group.id)}`);
	}

	function openInviteDialog(group) {
		inviteDialogLink = '';
		inviteDialogName = group?.name || 'Default group';
		inviteDialogMessage = '';
		inviteDialogOpen = true;
	}

	function closeInviteDialog() {
		inviteDialogOpen = false;
		inviteDialogBusy = false;
	}

	async function fetchFreshInviteLink() {
		const response = await apiFetch('/api/groups/default-invite');
		if (!response.ok) {
			throw new Error('Failed to generate invite link');
		}
		const payload = await response.json();
		const link = typeof payload?.invite_link === 'string' ? payload.invite_link.trim() : '';
		if (!link) {
			throw new Error('Invite link is empty');
		}
		inviteDialogLink = link;
		return link;
	}

	async function openJoinPage() {
		inviteDialogBusy = true;
		inviteDialogMessage = '';
		try {
			const link = await fetchFreshInviteLink();
			window.open(link, '_blank', 'noopener');
		} catch {
			inviteDialogMessage = 'Failed to open invite link';
		} finally {
			inviteDialogBusy = false;
		}
	}

	async function copyInviteLink() {
		inviteDialogBusy = true;
		inviteDialogMessage = '';
		try {
			const link = inviteDialogLink || await fetchFreshInviteLink();
			await navigator.clipboard.writeText(link);
			inviteDialogMessage = 'Link copied';
		} catch {
			inviteDialogMessage = 'Copy failed';
		} finally {
			inviteDialogBusy = false;
		}
	}

	async function syncMemberships() {
		if (!adminMode || syncBusy) return;
		syncBusy = true;
		syncMessage = '';
		try {
			const response = await apiFetch('/api/admin/groups/sync-memberships', { method: 'POST' });
			if (!response.ok) {
				throw new Error('Sync failed');
			}
			syncMessage = 'Membership sync completed';
			await loadGroups();
		} catch {
			syncMessage = 'Membership sync failed';
		} finally {
			syncBusy = false;
		}
	}
</script>

<DashboardLayout title={adminMode ? 'Admin Groups' : 'Groups'}>
	{#if error}
		<Card variant="state">{error}</Card>
	{:else if loaded && visibleGroups.length === 0}
		<Card variant="state">
			<p>No groups found</p>
		</Card>
	{:else}
		<div class="stack-list">
			{#each visibleGroups as item}
				<GroupCard
					group={item.group}
					isMember={item.isMember}
					onOpen={item.isMember ? goToGroup : null}
					onInviteClick={openInviteDialog}
					inviteLink=""
					requestsCount={requestsCountByGroupId[item.group.id] || 0}
					showRequestsBadge={item.isMember}
				/>
			{/each}
		</div>
	{/if}

	{#if adminMode}
		<div class="admin-sync">
			<button type="button" class="sync-button" on:click={syncMemberships} disabled={syncBusy}>
				{syncBusy ? 'Syncing...' : 'Sync memberships'}
			</button>
			{#if syncMessage}
				<p class="sync-message">{syncMessage}</p>
			{/if}
		</div>
	{/if}
</DashboardLayout>

{#if inviteDialogOpen}
	<div
		class="invite-overlay"
		role="button"
		tabindex="0"
		on:click={closeInviteDialog}
		on:keydown={(e) => {
			if (e.key === 'Enter' || e.key === ' ') closeInviteDialog();
		}}
	>
		<div
			class="invite-dialog"
			role="dialog"
			aria-modal="true"
			aria-label="Open invite link"
			tabindex="-1"
			on:click|stopPropagation
			on:keydown|stopPropagation={() => {}}
		>
			<h3>{inviteDialogName}</h3>
			<p>Open the invite page, then tap Join Group in Telegram.</p>
			<div class="invite-actions">
				<button type="button" class="invite-primary" on:click={openJoinPage} disabled={inviteDialogBusy}>Open Join Group page</button>
				<button type="button" class="invite-secondary" on:click={copyInviteLink} disabled={inviteDialogBusy}>Copy invite link</button>
				<button type="button" class="invite-secondary" on:click={closeInviteDialog} disabled={inviteDialogBusy}>Close</button>
			</div>
			{#if inviteDialogMessage}
				<p class="invite-message">{inviteDialogMessage}</p>
			{/if}
		</div>
	</div>
{/if}

<style>
	.admin-sync {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
		align-items: center;
	}

	.sync-button {
		border: 2px solid #000;
		background: #000;
		color: #fff;
		padding: 0.6rem 0.8rem;
		font-size: 0.95rem;
		font-weight: 700;
		cursor: pointer;
		width: min(100%, var(--sync-button-width));
	}

	.sync-button:disabled {
		opacity: 0.6;
		cursor: default;
	}

	.sync-message {
		margin: 0;
		font-size: 0.85rem;
		color: #333;
	}

	@media (min-width: 768px) and (max-width: 1024px) {
		.sync-button {
			width: min(100%, var(--sync-button-width));
		}
	}

	.invite-overlay {
		position: fixed;
		inset: 0;
		background: rgba(0, 0, 0, 0.55);
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 1rem;
		z-index: 2000;
	}

	.invite-dialog {
		background: #fff;
		border: 2px solid #000;
		width: min(28rem, 100%);
		padding: 1rem;
	}

	.invite-dialog h3 {
		margin: 0 0 0.75rem;
		font-size: 1rem;
	}

	.invite-dialog p {
		margin: 0 0 0.75rem;
	}

	.invite-actions {
		display: grid;
		gap: 0.5rem;
	}

	.invite-primary,
	.invite-secondary {
		border: 2px solid #000;
		padding: 0.65rem 0.75rem;
		font-weight: 600;
		cursor: pointer;
	}

	.invite-primary {
		background: #000;
		color: #fff;
	}

	.invite-secondary {
		background: #fff;
		color: #000;
	}

	.invite-message {
		margin-top: 0.75rem;
		font-size: 0.9rem;
	}
</style>
