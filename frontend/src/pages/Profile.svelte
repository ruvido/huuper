<script>
	import { pb } from '../lib/pocketbase';
	import { navigate } from '../lib/router';

	const user = pb.authStore.record;

	function goToGroups() {
		navigate('groups');
	}
</script>

<div class="page">
	<div class="container">
		<h1 class="page-title">Profile</h1>

		<div class="card">
			<div class="avatar">
				{user?.email?.charAt(0).toUpperCase()}
			</div>
			<h2>{user?.email}</h2>
			<p class="status">Admin</p>
		</div>

		<button class="btn-primary" on:click={goToGroups}>
			View Groups
		</button>
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
		display: flex;
		flex-direction: column;
		gap: clamp(1.5rem, 4vw, 2rem);
	}

	.page-title {
		margin: 0 0 clamp(1rem, 3vw, 1.5rem) 0;
		font-size: 2rem;
		font-weight: bold;
		color: #000;
	}

	/* Card profilo */
	.card {
		background: #fff;
		border: 2px solid #000;
		padding: clamp(2rem, 5vw, 3rem) clamp(1.5rem, 4vw, 2rem);
		text-align: center;
	}

	.avatar {
		width: clamp(5rem, 15vw, 6.25rem);
		height: clamp(5rem, 15vw, 6.25rem);
		border: 3px solid #000;
		color: #000;
		font-size: clamp(2.5rem, 8vw, 3rem);
		display: flex;
		align-items: center;
		justify-content: center;
		margin: 0 auto clamp(1rem, 3vw, 1.5rem);
		font-weight: bold;
	}

	h2 {
		margin: 0 0 clamp(0.375rem, 1vw, 0.5rem) 0;
		color: #000;
		font-size: clamp(1.25rem, 4vw, 1.5rem);
		font-weight: bold;
		word-break: break-word;
	}

	.status {
		color: #000;
		font-weight: 600;
		margin: 0;
		font-size: clamp(0.875rem, 2.5vw, 1rem);
	}

	/* Button - touch target 48px */
	.btn-primary {
		width: 100%;
		padding: clamp(0.875rem, 3vw, 1rem);
		background: #000;
		color: #fff;
		border: 2px solid #000;
		font-size: clamp(1rem, 3vw, 1.125rem);
		font-weight: 600;
		cursor: pointer;
		touch-action: manipulation;
		transition: background 0.2s, color 0.2s;
	}

	.btn-primary:hover {
		background: #fff;
		color: #000;
	}
</style>
