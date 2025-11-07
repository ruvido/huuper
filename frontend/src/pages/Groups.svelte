<script>
	import { onMount, onDestroy } from 'svelte';
	import { pb } from '../lib/pocketbase';
	import { navigate } from '../lib/router';

	let groups = [];
	let loading = true;
	let error = '';
	let memberGroups = [];
	let unsubscribe;

	onMount(async () => {
		const currentUser = pb.authStore.record;

		try {
			// Load groups
			const result = await pb.collection('groups').getList(1, 500, {
				sort: '-created',
			});
			groups = result.items;

			// Load user_groups for current user
			const userGroupsResult = await pb.collection('user_groups').getList(1, 500, {
				filter: `user = "${currentUser.id}"`,
			});
			memberGroups = userGroupsResult.items.map(ug => ug.group);

			// Subscribe to user_groups changes
			unsubscribe = await pb.collection('user_groups').subscribe('*', (e) => {
				// Only update if it's for current user
				if (e.record.user !== currentUser.id) return;

				if (e.action === 'create') {
					// Add group to memberGroups
					if (!memberGroups.includes(e.record.group)) {
						memberGroups = [...memberGroups, e.record.group];
					}
				} else if (e.action === 'delete') {
					// Remove group from memberGroups
					memberGroups = memberGroups.filter(g => g !== e.record.group);
				}
			});
		} catch (err) {
			error = err.message || err.toString() || 'Failed to load groups';
		} finally {
			loading = false;
		}
	});

	onDestroy(() => {
		if (unsubscribe) {
			unsubscribe();
		}
	});

	function goToProfile() {
		navigate('profile');
	}
</script>

<div class="dashboard-page">
	<div class="dashboard-container">
		<h1 class="dashboard-title">Groups</h1>

		{#if loading}
			<div class="state-card">Loading groups...</div>
		{:else if error}
			<div class="state-card">{error}</div>
		{:else if groups.length === 0}
			<div class="state-card">
				<p>No groups found</p>
			</div>
		{:else}
			<div class="groups-list">
				{#each groups as group}
					<div class="group-card">
						<div class="group-icon">
						{group.name.charAt(0).toUpperCase()}
						</div>
						<div class="group-info">
							<h3>
								{group.name}
								{#if memberGroups.includes(group.id)}
									<span class="member-badge">✓ Member</span>
								{/if}
							</h3>
							{#if group.description}
								<p class="description">{group.description}</p>
							{/if}
						</div>
						{#if group.invite_link}
							<a
								href={group.invite_link}
								target="_blank"
								rel="noopener noreferrer"
								class="btn-join"
							>
								Join
							</a>
						{/if}
					</div>
				{/each}
			</div>
		{/if}
		<button class="btn-primary" on:click={goToProfile}>
			Profile
		</button>
	</div>
</div>

<style>
	/* Component-specific styles only */
	/* Grid list - responsive automatico */
	.groups-list {
		display: grid;
		gap: clamp(1rem, 3vw, 1.5rem);
		grid-template-columns: 1fr;
	}

	/* Single group card - Grid layout */
	.group-card {
		background: #fff;
		border: 2px solid #000;
		padding: clamp(0.75rem, 2.5vw, 1rem);
		display: flex;
		align-items: center;
		gap: clamp(0.75rem, 2vw, 1rem);
	}

	.group-icon {
		width: clamp(3.5rem, 10vw, 4rem);
		height: clamp(3.5rem, 10vw, 4rem);
		border: 2px solid #000;
		display: flex;
		align-items: center;
		justify-content: center;
		flex-shrink: 0;
		font-size: clamp(1.75rem, 5vw, 2rem);
		font-weight: bold;
		color: #000;
	}

	.group-info {
		flex: 1;
		min-width: 0;
		display: flex;
		flex-direction: column;
		gap: clamp(0.125rem, 0.5vw, 0.25rem);
	}

	h3 {
		margin: 0;
		font-size: clamp(1.125rem, 3.5vw, 1.25rem);
		color: #000;
		font-weight: bold;
		word-break: break-word;
	}

	.description {
		margin: 0;
		color: #000;
		font-weight: 600;
		font-size: clamp(0.8125rem, 2.5vw, 0.875rem);
		overflow: hidden;
		text-overflow: ellipsis;
		display: -webkit-box;
		-webkit-line-clamp: 2;
		-webkit-box-orient: vertical;
	}

	/* Join button */
	.btn-join {
		background: #000;
		color: #fff;
		border: 2px solid #000;
		padding: clamp(0.875rem, 3vw, 1rem);
		text-decoration: none;
		font-weight: 600;
		text-align: center;
		white-space: nowrap;
		touch-action: manipulation;
		transition: background 0.2s, color 0.2s;
		flex-shrink: 0;
	}

	.btn-join:hover {
		background: #fff;
		color: #000;
	}

	/* Member badge */
	.member-badge {
		display: inline-block;
		margin-left: clamp(0.5rem, 2vw, 0.75rem);
		font-size: clamp(0.75rem, 2vw, 0.875rem);
		color: #0a0;
		font-weight: 600;
	}
</style>
