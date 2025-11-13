<script>
	import { pb } from '../lib/pocketbase';
	import { navigate } from '../lib/router';
	import { X } from 'lucide-svelte';

	export let isOpen = false;
	export let onClose;

	function handleNavigate(route) {
		navigate(route);
		onClose();
	}

	function handleLogout() {
		pb.authStore.clear();
		navigate('login');
		onClose();
	}
</script>

<!-- Backdrop -->
{#if isOpen}
	<div class="backdrop" on:click={onClose} role="presentation"></div>
{/if}

<!-- Slide-in menu -->
<nav class="menu" class:open={isOpen}>
	<button class="close-btn" on:click={onClose} aria-label="Close menu"><X size={24} /></button>

	<ul class="menu-list">
		<li>
			<button on:click={() => handleNavigate('profile')}>Profile</button>
		</li>
		<li>
			<button on:click={() => handleNavigate('groups')}>Groups</button>
		</li>
		<li>
			<button on:click={handleLogout}>Logout</button>
		</li>
	</ul>
</nav>

<style>
	/* Backdrop overlay */
	.backdrop {
		position: fixed;
		inset: 0;
		background: rgba(0, 0, 0, 0.5);
		z-index: 200;
		animation: fadeIn 0.3s;
	}

	@keyframes fadeIn {
		from { opacity: 0; }
		to { opacity: 1; }
	}

	/* Slide-in menu - mobile first */
	.menu {
		position: fixed;
		top: 0;
		right: 0;
		bottom: 0;
		width: min(80vw, 20rem);
		background: #fff;
		border-left: 2px solid #000;
		transform: translateX(100%);
		transition: transform 0.3s ease-out;
		z-index: 300;
		display: flex;
		flex-direction: column;
		padding: clamp(1rem, 4vw, 1.5rem);
		gap: clamp(1.5rem, 4vw, 2rem);
	}

	.menu.open {
		transform: translateX(0);
	}

	/* Close button - touch target 48px */
	.close-btn {
		align-self: flex-end;
		width: 3rem;
		height: 3rem;
		display: flex;
		align-items: center;
		justify-content: center;
		background: transparent;
		border: none;
		line-height: 1;
		cursor: pointer;
		touch-action: manipulation;
		color: #000;
		transition: color 0.2s ease;
	}

	.close-btn:hover {
		color: #666;
	}

	/* Menu list */
	.menu-list {
		list-style: none;
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
		margin: 0;
		padding: 0;
	}

	.menu-list button {
		width: 100%;
		padding: clamp(0.875rem, 3vw, 1rem);
		background: transparent;
		border: 2px solid #000;
		font-size: clamp(1rem, 3vw, 1.125rem);
		font-weight: 600;
		text-align: left;
		cursor: pointer;
		touch-action: manipulation;
		transition: background 0.2s, color 0.2s;
	}

	.menu-list button:hover {
		background: #000;
		color: #fff;
	}
</style>
