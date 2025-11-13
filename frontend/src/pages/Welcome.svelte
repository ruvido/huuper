<script>
	import { navigate } from '../lib/router';
	import Button from '../components/Button.svelte';

	function startOnboarding() {
		navigate('onboarding');
	}
</script>

<div class="welcome-page">
	<div class="content">
		<h1>Benvenuto in realmen</h1>
		<p>Per iniziare a usare i gruppi<br />completa il tuo profilo</p>
		<Button variant="submit" on:click={startOnboarding}>
			INIZIA
		</Button>
	</div>
</div>

<style>
	.welcome-page {
		height: 100vh;
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 2rem;
		background: #fff;
	}

	.content {
		text-align: center;
		max-width: 28rem;
	}

	h1 {
		margin: 0 0 1.5rem 0;
		font-size: clamp(1.75rem, 5vw, 2.5rem);
		font-weight: bold;
		color: #000;
	}

	p {
		margin: 0 0 2rem 0;
		font-size: clamp(1rem, 3vw, 1.25rem);
		color: #333;
		line-height: 1.6;
	}
</style>
