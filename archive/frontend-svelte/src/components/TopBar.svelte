<script>
	import { currentRoute, navigate } from '../lib/router';

	export let title = '';
	export let showBack = false;
	export let onBack = null;
	$: route = typeof $currentRoute === 'string' ? $currentRoute : '';
	$: showBuildDate = route === 'admin' || route.startsWith('admin/');
	const buildDateLabel = (() => {
		const raw = typeof __BUILD_DATE__ === 'string' ? __BUILD_DATE__.trim() : '';
		if (!raw) return '';
		const dt = new Date(raw);
		if (Number.isNaN(dt.getTime())) return raw;
		return dt.toLocaleString('it-IT', {
			year: 'numeric',
			month: '2-digit',
			day: '2-digit',
			hour: '2-digit',
			minute: '2-digit'
		});
	})();

	function handleBack() {
		if (typeof onBack === 'function') onBack();
	}

	function goToProfile() {
		const route = typeof $currentRoute === 'string' ? $currentRoute : '';
		const scope = route === 'admin' || route.startsWith('admin/') ? 'admin' : 'app';
		navigate(`${scope}/profile`);
	}
</script>

<header class="top-bar">
	<div class="top-bar-inner">
		<div class="top-left">
			{#if showBack}
				<button class="back" type="button" aria-label="Go back" on:click={handleBack}>
					<svg xmlns="http://www.w3.org/2000/svg" fill="currentColor" viewBox="0 0 16 16" aria-hidden="true">
						<path fill-rule="evenodd" d="M11.354 1.646a.5.5 0 0 1 0 .708L5.707 8l5.647 5.646a.5.5 0 0 1-.708.708l-6-6a.5.5 0 0 1 0-.708l6-6a.5.5 0 0 1 .708 0"/>
					</svg>
				</button>
			{/if}
		</div>
			<div class="top-center">
				<h1 class="top-title" title={title}>{title}</h1>
				{#if showBuildDate && buildDateLabel}
					<div class="build-date" title={__BUILD_DATE__}>Build {buildDateLabel}</div>
				{/if}
			</div>
		<div class="top-right">
			<button class="profile" type="button" aria-label="Profile" on:click={goToProfile}>
				<svg xmlns="http://www.w3.org/2000/svg" fill="currentColor" viewBox="0 0 16 16" aria-hidden="true">
					<path d="M3 14s-1 0-1-1 1-4 6-4 6 3 6 4-1 1-1 1zm5-6a3 3 0 1 0 0-6 3 3 0 0 0 0 6"/>
				</svg>
			</button>
		</div>
	</div>
</header>

<style>
	.top-bar {
		position: fixed;
		top: 0;
		left: 0;
		right: 0;
		background: #fff;
		border-bottom: 1px solid #e8e8e8;
		z-index: 110;
	}

	.top-bar-inner {
		max-width: 50rem;
		margin: 0 auto;
		height: 3.25rem;
		display: grid;
		grid-template-columns: 2.25rem minmax(0, 1fr) 2.25rem;
		align-items: center;
		gap: 0.5rem;
		padding: 0 0.25rem;
	}

	.top-left,
	.top-right {
		display: flex;
		align-items: center;
		justify-content: center;
		min-width: 0;
	}

	.top-title {
		margin: 0;
		font-size: var(--ui-font-size);
		font-weight: 700;
		color: #000;
		text-align: center;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.top-center {
		min-width: 0;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		line-height: 1.1;
	}

	.build-date {
		font-size: 0.66rem;
		color: #666;
		white-space: nowrap;
	}

	.back,
	.profile {
		border: none;
		background: transparent;
		padding: 0.2rem 0.4rem;
		cursor: pointer;
		display: inline-flex;
		align-items: center;
		justify-content: center;
	}

	.back svg,
	.profile svg {
		width: var(--icon-size);
		height: var(--icon-size);
	}
</style>
