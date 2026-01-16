<script>
	import Card from '../Card.svelte';
	import { Clock, Check } from 'lucide-svelte';

	export let status = 'pending'; // 'pending' | 'active'
</script>

<Card>
	<h3 class="section-title">Account Status</h3>

	{#if status === 'pending'}
		<div class="status-pending">
			<p class="pending-message"><Clock size={20} /> In attesa di approvazione</p>
			<p class="help-text">Il tuo account è in attesa di essere approvato da un amministratore.</p>
		</div>
	{:else}
		<div class="status-active">
			<p class="active-message"><Check size={20} /> Account attivo</p>
		</div>
	{/if}
</Card>

<style>
	.section-title {
		margin: 0 0 clamp(1rem, 3vw, 1.5rem) 0;
		font-size: clamp(1.125rem, 3.5vw, 1.25rem);
		color: #000;
		font-weight: bold;
	}

	.status-pending,
	.status-active {
		text-align: center;
	}

	.pending-message {
		font-weight: 600;
		color: #ff8c00;
		font-size: clamp(1rem, 3vw, 1.125rem);
		margin: 0 0 clamp(0.5rem, 2vw, 0.75rem) 0;
	}

	.active-message {
		font-weight: 600;
		color: #008000;
		font-size: clamp(1rem, 3vw, 1.125rem);
		margin: 0;
	}

	.help-text {
		font-size: clamp(0.75rem, 2vw, 0.875rem);
		color: #666;
	}
</style>
