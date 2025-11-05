<script>
	export let onMenuClick;
</script>

<header class="header">
	<h1 class="app-name">Members</h1>
	<button class="hamburger" on:click={onMenuClick} aria-label="Menu">
		<span class="hamburger-line"></span>
		<span class="hamburger-line"></span>
		<span class="hamburger-line"></span>
	</button>
</header>

<style>
	/* Header fixed - best practices */
	.header {
		position: fixed;
		top: 0;
		left: 0;
		right: 0;
		width: 100%;
		height: 4rem;
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 0 1rem;
		background: #fff;
		border-bottom: 2px solid #000;
		z-index: 100;
	}

	.app-name {
		margin: 0;
		font-size: 1.25rem;
		font-weight: bold;
		color: #000;
		white-space: nowrap;
		user-select: none;
	}

	/* Hamburger button - 48px touch target */
	.hamburger {
		display: flex;
		flex-direction: column;
		justify-content: center;
		align-items: center;
		gap: 0.375rem;
		width: 3rem;
		height: 3rem;
		padding: 0.625rem;
		background: transparent;
		border: 2px solid #000;
		cursor: pointer;
		touch-action: manipulation;
	}

	.hamburger:hover {
		background: #000;
	}

	.hamburger:hover .hamburger-line {
		background: #fff;
	}

	.hamburger-line {
		width: 100%;
		height: 2px;
		background: #000;
		transition: background 0.2s;
	}
</style>
