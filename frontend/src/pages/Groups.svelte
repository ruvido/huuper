<script>
	import { onMount, onDestroy } from 'svelte';
	import { pb } from '../lib/pocketbase';
	import { navigate } from '../lib/router';
	import DashboardLayout from '../components/DashboardLayout.svelte';
	import GroupCard from '../components/cards/GroupCard.svelte';
	import Card from '../components/Card.svelte';

	let groups = [];
	let loaded = false;
	let error = '';
	let unsubscribeUserGroups;

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
			groups = userGroupsResult.items
				.map((item) => item?.expand?.group || null)
				.filter(Boolean);
		} catch (err) {
			error = err.message || err.toString() || 'Failed to load groups';
		} finally {
			loaded = true;
		}
	}

	function goToGroup(group) {
		if (!group?.id) return;
		navigate(`app/groups/${encodeURIComponent(group.id)}`);
	}
</script>

<DashboardLayout title="Groups">
	{#if error}
		<Card variant="state">{error}</Card>
	{:else if loaded && groups.length === 0}
		<Card variant="state">
			<p>No groups found</p>
		</Card>
	{:else}
		<div class="stack-list">
			{#each groups as group}
				<GroupCard {group} isMember={true} onOpen={goToGroup} />
			{/each}
		</div>
	{/if}
</DashboardLayout>
