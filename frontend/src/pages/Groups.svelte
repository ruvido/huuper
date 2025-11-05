<script>
	import { onMount } from 'svelte';
	import { pb } from '../lib/pocketbase';

	let groups = [];
	let loading = true;
	let error = '';

	onMount(async () => {
		try {
			const result = await pb.collection('groups').getList(1, 500, {
				sort: '-created',
			});
			groups = result.items;
		} catch (err) {
			console.error('Error loading groups:', err);
			error = err.message || err.toString() || 'Failed to load groups';
		} finally {
			loading = false;
		}
	});
</script>

<div class="page">
	<div class="container">
		<h1 class="page-title">Groups</h1>

		{#if loading}
			<div class="loading">Loading groups...</div>
		{:else if error}
			<div class="error">{error}</div>
		{:else if groups.length === 0}
			<div class="empty">
				<p>No groups found</p>
			</div>
		{:else}
			<div class="groups-list">
				{#each groups as group}
					<div class="group-card">
						<div class="group-icon">
							{#if group.type === 'telegram'}
								<span class="icon">📱</span>
							{:else}
								<span class="icon">💬</span>
							{/if}
						</div>
						<div class="group-info">
							<h3>{group.name}</h3>
							<p class="type">{group.type}</p>
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
	</div>
</div>

<style>
	.page {
		min-height: 100vh;
		background: #fff;
		padding: 5rem clamp(1rem, 4vw, 2rem) clamp(1rem, 4vw, 2rem);
	}

	.container {
		max-width: 50rem;
		margin: 0 auto;
	}

	.page-title {
		margin: 0 0 clamp(1rem, 3vw, 1.5rem) 0;
		font-size: 2rem;
		font-weight: bold;
		color: #000;
	}

	/* Stati: loading, error, empty */
	.loading, .error, .empty {
		background: #fff;
		border: 2px solid #000;
		padding: clamp(2rem, 5vw, 3rem) clamp(1.5rem, 4vw, 2rem);
		text-align: center;
		color: #000;
		font-size: clamp(1rem, 3vw, 1.125rem);
	}

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
		padding: clamp(1rem, 3vw, 1.5rem);
		display: grid;
		grid-template-columns: auto 1fr;
		grid-template-rows: auto auto;
		gap: clamp(0.75rem, 2vw, 1rem);
		align-items: start;
	}

	.group-icon {
		grid-row: 1 / 2;
		width: clamp(3.5rem, 10vw, 4rem);
		height: clamp(3.5rem, 10vw, 4rem);
		border: 2px solid #000;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.icon {
		font-size: clamp(1.75rem, 5vw, 2rem);
	}

	.group-info {
		grid-row: 1 / 2;
		min-width: 0;
		display: flex;
		flex-direction: column;
		gap: clamp(0.25rem, 1vw, 0.375rem);
	}

	h3 {
		margin: 0;
		font-size: clamp(1.125rem, 3.5vw, 1.25rem);
		color: #000;
		font-weight: bold;
		word-break: break-word;
	}

	.type {
		margin: 0;
		color: #000;
		font-weight: 600;
		font-size: clamp(0.8125rem, 2.5vw, 0.875rem);
		text-transform: capitalize;
	}

	.description {
		margin: 0;
		color: #000;
		font-size: clamp(0.875rem, 2.5vw, 0.9rem);
		overflow: hidden;
		text-overflow: ellipsis;
		display: -webkit-box;
		-webkit-line-clamp: 2;
		-webkit-box-orient: vertical;
	}

	/* Join button - full width su mobile, touch target 48px */
	.btn-join {
		grid-column: 1 / -1;
		grid-row: 2 / 3;
		background: #000;
		color: #fff;
		border: 2px solid #000;
		padding: clamp(0.875rem, 3vw, 1rem);
		text-decoration: none;
		font-weight: 600;
		text-align: center;
		display: inline-block;
		touch-action: manipulation;
		transition: background 0.2s, color 0.2s;
	}

	.btn-join:hover {
		background: #fff;
		color: #000;
	}

	/* Tablet e desktop: card più complessa */
	@media (min-width: 48em) {
		.group-card {
			grid-template-columns: auto 1fr auto;
			grid-template-rows: 1fr;
		}

		.group-icon {
			grid-row: 1 / 2;
		}

		.group-info {
			grid-row: 1 / 2;
		}

		.btn-join {
			grid-column: 3 / 4;
			grid-row: 1 / 2;
			white-space: nowrap;
			padding-inline: clamp(1.25rem, 3vw, 1.5rem);
		}
	}
</style>
