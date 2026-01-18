<script>
	import { onMount, onDestroy } from 'svelte';
	import { pb } from '../lib/pocketbase';
	import { navigate, defaultAppRoute } from '../lib/router';
	import DashboardLayout from '../components/DashboardLayout.svelte';
	import StateCard from '../components/StateCard.svelte';
	import Button from '../components/Button.svelte';
	import GroupCard from '../components/cards/GroupCard.svelte';

	let groups = [];
	let loaded = false;
	let error = '';
	let memberGroups = [];
	let unsubscribeUserGroups;
	let unsubscribeGroups;

	onMount(async () => {
		const currentUser = pb.authStore.record;

		try {
			const result = await pb.collection('groups').getList(1, 500, {
				sort: '-created',
			});
			groups = result.items;

			const userGroupsResult = await pb.collection('user_groups').getList(1, 500, {
				filter: `user = "${currentUser.id}"`,
			});
			memberGroups = userGroupsResult.items.map(ug => ug.group);

			unsubscribeUserGroups = await pb.collection('user_groups').subscribe('*', (e) => {
				if (e.record.user !== currentUser.id) return;

				if (e.action === 'create') {
					if (!memberGroups.includes(e.record.group)) {
						memberGroups = [...memberGroups, e.record.group];
					}
				} else if (e.action === 'delete') {
					memberGroups = memberGroups.filter(g => g !== e.record.group);
				}
			});

			unsubscribeGroups = await pb.collection('groups').subscribe('*', (e) => {
				if (e.action === 'create') {
					groups = [...groups, e.record];
				} else if (e.action === 'update') {
					groups = groups.map(g => g.id === e.record.id ? e.record : g);
				} else if (e.action === 'delete') {
					groups = groups.filter(g => g.id !== e.record.id);
				}
			});
		} catch (err) {
			error = err.message || err.toString() || 'Failed to load groups';
		} finally {
			loaded = true;
		}
	});

	onDestroy(() => {
		if (unsubscribeUserGroups) {
			unsubscribeUserGroups();
		}
		if (unsubscribeGroups) {
			unsubscribeGroups();
		}
	});

	function goToProfile() {
		navigate(defaultAppRoute);
	}
</script>

<DashboardLayout title="Groups">
	{#if error}
		<StateCard>{error}</StateCard>
	{:else if loaded && groups.length === 0}
		<StateCard>
			<p>No groups found</p>
		</StateCard>
	{:else}
		<div class="groups-list">
			{#each groups as group}
				<GroupCard {group} isMember={memberGroups.includes(group.id)} />
			{/each}
		</div>
	{/if}
	<Button variant="primary" on:click={goToProfile}>
		Profile
	</Button>
</DashboardLayout>

<style>
	.groups-list {
		display: grid;
		gap: clamp(1rem, 3vw, 1.5rem);
		grid-template-columns: 1fr;
	}
</style>
