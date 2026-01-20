<script>
	import Card from '../Card.svelte';
	import { Clock, Check } from 'lucide-svelte';

	export let status = 'active'; // 'active' | 'suspended'
</script>

<Card>
	<h3 class="section-title">Account Status</h3>

	{#if status === 'suspended'}
		<div class="status-suspended">
			<p class="suspended-message"><Clock size={20} /> Account sospeso</p>
			<p class="help-text">Contatta un amministratore per riattivare il tuo accesso.</p>
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

	.status-suspended,
	.status-active {
		text-align: center;
	}

	.suspended-message {
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
