<script>
	import { Check } from 'lucide-svelte';

	export let group;
	export let isMember = false;
</script>

<div class="group-card">
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
	{#if group.invite_link}
		<a
			href={group.invite_link}
			target="_blank"
			rel="noopener noreferrer"
			class="join-link"
		>
			Join
		</a>
	{/if}
</div>

<style>
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

	.join-link {
		background: #000;
		color: #fff;
		border: 2px solid #000;
		padding: clamp(0.875rem, 3vw, 1rem);
		text-decoration: none;
		font-weight: 600;
		text-align: center;
		white-space: nowrap;
		transition: background 0.2s, color 0.2s;
		flex-shrink: 0;
	}

	.join-link:hover {
		background: #fff;
		color: #000;
	}

	.member-badge {
		display: inline-block;
		margin-left: clamp(0.5rem, 2vw, 0.75rem);
		font-size: clamp(0.75rem, 2vw, 0.875rem);
		color: #0a0;
		font-weight: 600;
	}
</style>
