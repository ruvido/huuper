<script>
	import { Check } from 'lucide-svelte';
	import Card from '../Card.svelte';

	export let group;
	export let isMember = false;
	export let onOpen = null;
	$: clickable = typeof onOpen === 'function';

	function handleOpen() {
		if (clickable) {
			onOpen(group);
		}
	}
</script>

{#if clickable}
	<button class="group-card clickable" type="button" aria-label={`Open ${group.name}`} on:click={handleOpen}>
		<Card variant="item">
			<div class="group-content">
				<div class="group-icon">
					{group.name.charAt(0).toUpperCase()}
				</div>
				<div class="group-info">
					<h3>
						{group.name}
						{#if isMember}
							<span class="member-badge"><Check size={14} /> Member</span>
						{/if}
					</h3>
					{#if group.description}
						<p class="description">{group.description}</p>
					{/if}
				</div>
			</div>
		</Card>
	</button>
{:else}
	<Card variant="item">
		<div class="group-content">
			<div class="group-icon">
				{group.name.charAt(0).toUpperCase()}
			</div>
			<div class="group-info">
				<h3>
					{group.name}
					{#if isMember}
						<span class="member-badge"><Check size={14} /> Member</span>
					{/if}
				</h3>
				{#if group.description}
					<p class="description">{group.description}</p>
				{/if}
			</div>
		</div>
	</Card>
{/if}

<style>
	.group-card {
		width: 100%;
		text-align: left;
		background: transparent;
		border: none;
		padding: 0;
	}

	.group-card :global(.card.item) {
		width: 100%;
	}

	.group-content {
		display: flex;
		align-items: center;
		gap: clamp(0.75rem, 2vw, 1rem);
	}

	.group-card.clickable {
		cursor: pointer;
	}

	.group-icon {
		width: clamp(3.5rem, 10vw, 4rem);
		height: clamp(3.5rem, 10vw, 4rem);
		border: 2px solid #000;
		display: flex;
		align-items: center;
		justify-content: center;
		flex-shrink: 0;
		font-size: var(--ui-font-size);
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
		font-size: var(--ui-font-size);
		color: #000;
		font-weight: bold;
		word-break: break-word;
	}

	.description {
		margin: 0;
		color: #000;
		font-weight: 600;
		font-size: var(--ui-font-size);
		overflow: hidden;
		text-overflow: ellipsis;
		display: -webkit-box;
		-webkit-line-clamp: 2;
		-webkit-box-orient: vertical;
	}

	.member-badge {
		display: inline-block;
		margin-left: clamp(0.5rem, 2vw, 0.75rem);
		font-size: var(--ui-font-size);
		color: #0a0;
		font-weight: 600;
	}
</style>
