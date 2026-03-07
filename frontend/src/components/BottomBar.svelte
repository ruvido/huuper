<script>
	import { currentRoute, navigate } from '../lib/router';

	export let eventsAlert = false;
	export let groupsAlert = false;

	$: route = $currentRoute || '';
	$: isAdminRoute = route === 'admin' || route.startsWith('admin/');
	$: dashboardActive = route === 'admin';
	$: eventsActive = isAdminRoute
		? route === 'admin/events' || route === 'admin/event'
		: route === 'app/events';
	$: groupsActive = isAdminRoute
		? route === 'admin/groups' || route.startsWith('admin/groups/')
		: route === 'app/groups' || route.startsWith('app/groups/');

	function goToDashboard() {
		navigate('admin');
	}

	function goToEvents() {
		navigate(isAdminRoute ? 'admin/events' : 'app/events');
	}

	function goToGroups() {
		navigate(isAdminRoute ? 'admin/groups' : 'app/groups');
	}
</script>

<nav class="bottom-bar" class:admin={isAdminRoute} aria-label="Main">
	{#if isAdminRoute}
		<button class="tab" class:active={dashboardActive} on:click={goToDashboard} aria-label="Dashboard">
			<span class="icon-wrap">
				<svg xmlns="http://www.w3.org/2000/svg" fill="currentColor" viewBox="0 0 16 16" aria-hidden="true">
					<path d="M8.354 1.146a.5.5 0 0 0-.708 0l-6 6A.5.5 0 0 0 2 8h1v6a1 1 0 0 0 1 1h3v-4h2v4h3a1 1 0 0 0 1-1V8h1a.5.5 0 0 0 .354-.854z"/>
				</svg>
			</span>
		</button>
	{/if}
	<button class="tab" class:active={eventsActive} on:click={goToEvents} aria-label="Events">
		<span class="icon-wrap">
			<svg xmlns="http://www.w3.org/2000/svg" fill="currentColor" viewBox="0 0 16 16" aria-hidden="true">
				<path d="M5.52.359A.5.5 0 0 1 6 0h4a.5.5 0 0 1 .474.658L8.694 6H12.5a.5.5 0 0 1 .395.807l-7 9a.5.5 0 0 1-.873-.454L6.823 9.5H3.5a.5.5 0 0 1-.48-.641z"/>
			</svg>
			{#if eventsAlert}
				<span class="dot" aria-hidden="true"></span>
			{/if}
		</span>
	</button>
	<button class="tab" class:active={groupsActive} on:click={goToGroups} aria-label="Groups">
		<span class="icon-wrap">
			<svg xmlns="http://www.w3.org/2000/svg" fill="currentColor" viewBox="0 0 16 16" aria-hidden="true">
				<path d="M7 14s-1 0-1-1 1-4 5-4 5 3 5 4-1 1-1 1zm4-6a3 3 0 1 0 0-6 3 3 0 0 0 0 6m-5.784 6A2.24 2.24 0 0 1 5 13c0-1.355.68-2.75 1.936-3.72A6.3 6.3 0 0 0 5 9c-4 0-5 3-5 4s1 1 1 1zM4.5 8a2.5 2.5 0 1 0 0-5 2.5 2.5 0 0 0 0 5"/>
			</svg>
			{#if groupsAlert}
				<span class="dot" aria-hidden="true"></span>
			{/if}
		</span>
	</button>
</nav>

<style>
	.bottom-bar {
		position: fixed;
		left: 0;
		right: 0;
		bottom: 0;
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		padding: 0.5rem;
		background: #fff;
		border-top: 1px solid #e6e6e6;
		z-index: 120;
	}

	.bottom-bar.admin {
		grid-template-columns: repeat(3, minmax(0, 1fr));
	}

	.tab {
		position: relative;
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 0.45rem;
		border: none;
		background: transparent;
		padding: 0.75rem 1rem;
		font-size: 0.95rem;
		font-weight: 600;
		color: var(--bottom-tab-color);
		cursor: pointer;
	}

	.tab.active {
		color: var(--bottom-tab-active-color);
	}

	.icon-wrap {
		position: relative;
		display: inline-flex;
		align-items: center;
		justify-content: center;
	}

	.icon-wrap svg {
		width: var(--icon-size);
		height: var(--icon-size);
	}

	.dot {
		position: absolute;
		top: -0.15rem;
		right: -0.3rem;
		width: 0.45rem;
		height: 0.45rem;
		border-radius: 999px;
		background: #d00;
	}
</style>
