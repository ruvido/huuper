<script>
	import Button from '../components/Button.svelte';
</script>

<div class="pending-page">
	<div class="content">
		<h1>Grazie per aver chiesto di entrare nel Branco!</h1>

		<div class="message">
			<p class="main">La tua richiesta è stata inviata.</p>

			<p>Nelle prossime ore riceverai un'email con i dettagli per il prossimo passo.</p>

			<p>Un membro del gruppo locale ti contatterà per presentarti la realtà dei gruppi Realmen.</p>

			<p>Ti introdurremo gradualmente al funzionamento e allo spirito della community.</p>

			<p class="closing">A presto!</p>
		</div>

		<p class="hint">(Controlla anche spam/promozioni se non trovi l'email)</p>
	</div>
</div>

<style>
	.pending-page {
		height: 100vh;
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 2rem;
		background: #fff;
	}

	.content {
		text-align: center;
		max-width: 32rem;
	}

	h1 {
		margin: 0 0 2rem 0;
		font-size: clamp(1.5rem, 5vw, 2rem);
		font-weight: bold;
		color: #000;
		line-height: 1.3;
	}

	.message {
		margin-bottom: 2rem;
	}

	.message p {
		margin: 0 0 1.25rem 0;
		font-size: clamp(1rem, 3vw, 1.125rem);
		color: #333;
		line-height: 1.6;
	}

	.message p.main {
		font-size: clamp(1.125rem, 3.5vw, 1.25rem);
		font-weight: 600;
		color: #000;
		margin-bottom: 1.5rem;
	}

	.message p.closing {
		font-weight: 600;
		font-size: clamp(1.125rem, 3.5vw, 1.25rem);
		margin-top: 1.5rem;
		margin-bottom: 0;
	}

	.hint {
		font-size: clamp(0.875rem, 2.5vw, 1rem);
		color: #666;
		font-style: italic;
		margin: 0;
	}
</style>
