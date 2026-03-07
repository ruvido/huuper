<script>
	import { Check } from 'lucide-svelte';
	import Card from '../Card.svelte';

	export let group;
	export let isMember = false;
	export let onOpen = null;
	export let requestsCount = 0;
	export let showRequestsBadge = false;
	export let inviteLink = '';
	export let onInviteClick = null;
	export let showInviteForDefault = true;
	$: clickable = typeof onOpen === 'function';
	$: isDefaultGroup = typeof group?.type === 'string' && group.type === 'default';
	$: showInviteLink = showInviteForDefault && !isMember && isDefaultGroup;

	function handleOpen() {
		if (clickable) {
			onOpen(group);
		}
	}

	function handleInviteClick(e) {
		e?.preventDefault?.();
		e?.stopPropagation?.();
		if (typeof onInviteClick === 'function') {
			onInviteClick(group, inviteLink);
		}
	}
</script>

{#if clickable}
	<button class="group-card clickable" type="button" aria-label={`Open ${group.name}`} on:click={handleOpen}>
		<div class="card-shell">
			{#if showRequestsBadge && requestsCount > 0}
				<span class="requests-badge" aria-label="Pending requests"></span>
			{/if}
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
							{#if showInviteLink}
								<button
									type="button"
									class="invite-badge"
									aria-label={`Open invite link for ${group.name}`}
									on:click={handleInviteClick}
								>
									invite link
								</button>
							{/if}
						</h3>
						{#if group.description}
							<p class="description">{group.description}</p>
						{/if}
					</div>
					{#if isDefaultGroup}
						<span class="default-badge" aria-label="Default group">
							<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="currentColor" viewBox="0 0 16 16" aria-hidden="true">
								<path fill-rule="evenodd" d="M2 15.5V2a2 2 0 0 1 2-2h8a2 2 0 0 1 2 2v13.5a.5.5 0 0 1-.74.439L8 13.069l-5.26 2.87A.5.5 0 0 1 2 15.5M8.16 4.1a.178.178 0 0 0-.32 0l-.634 1.285a.18.18 0 0 1-.134.098l-1.42.206a.178.178 0 0 0-.098.303L6.58 6.993c.042.041.061.1.051.158L6.39 8.565a.178.178 0 0 0 .258.187l1.27-.668a.18.18 0 0 1 .165 0l1.27.668a.178.178 0 0 0 .257-.187L9.368 7.15a.18.18 0 0 1 .05-.158l1.028-1.001a.178.178 0 0 0-.098-.303l-1.42-.206a.18.18 0 0 1-.134-.098z"/>
							</svg>
						</span>
					{/if}
				</div>
			</Card>
		</div>
	</button>
{:else}
	<div class="card-shell">
		{#if showRequestsBadge && requestsCount > 0}
			<span class="requests-badge" aria-label="Pending requests"></span>
		{/if}
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
						{#if showInviteLink}
							<button
								type="button"
								class="invite-badge"
								aria-label={`Open invite link for ${group.name}`}
								on:click={handleInviteClick}
							>
								invite link
							</button>
						{/if}
					</h3>
					{#if group.description}
						<p class="description">{group.description}</p>
					{/if}
				</div>
				{#if isDefaultGroup}
					<span class="default-badge" aria-label="Default group">
						<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="currentColor" viewBox="0 0 16 16" aria-hidden="true">
							<path fill-rule="evenodd" d="M2 15.5V2a2 2 0 0 1 2-2h8a2 2 0 0 1 2 2v13.5a.5.5 0 0 1-.74.439L8 13.069l-5.26 2.87A.5.5 0 0 1 2 15.5M8.16 4.1a.178.178 0 0 0-.32 0l-.634 1.285a.18.18 0 0 1-.134.098l-1.42.206a.178.178 0 0 0-.098.303L6.58 6.993c.042.041.061.1.051.158L6.39 8.565a.178.178 0 0 0 .258.187l1.27-.668a.18.18 0 0 1 .165 0l1.27.668a.178.178 0 0 0 .257-.187L9.368 7.15a.18.18 0 0 1 .05-.158l1.028-1.001a.178.178 0 0 0-.098-.303l-1.42-.206a.18.18 0 0 1-.134-.098z"/>
						</svg>
					</span>
				{/if}
			</div>
		</Card>
	</div>
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

	.card-shell {
		position: relative;
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
		display: inline-flex;
		align-items: center;
		gap: 0.25rem;
		margin-left: clamp(0.5rem, 2vw, 0.75rem);
		font-size: var(--ui-font-size);
		color: #0a0;
		font-weight: 600;
	}

	.requests-badge {
		position: absolute;
		top: 0.6rem;
		right: 0.6rem;
		z-index: 2;
		width: 0.65rem;
		height: 0.65rem;
		border-radius: 50%;
		background: #d50000;
		box-shadow: 0 1px 3px rgba(0, 0, 0, 0.25);
		pointer-events: none;
	}

	.invite-badge {
		display: inline-flex;
		align-items: center;
		margin-left: clamp(0.5rem, 2vw, 0.75rem);
		color: #1296db;
		font-size: var(--ui-font-size);
		font-weight: 600;
		text-decoration: underline;
		border: none;
		background: transparent;
		padding: 0;
		cursor: pointer;
	}

	.default-badge {
		margin-left: auto;
		color: #000;
		display: inline-flex;
		align-items: center;
		justify-content: center;
	}
</style>
