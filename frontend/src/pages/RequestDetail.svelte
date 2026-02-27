<script>
	import { apiFetch, pb } from '../lib/pocketbase';
	import { currentRoute } from '../lib/router';
	import DashboardLayout from '../components/DashboardLayout.svelte';
	import Card from '../components/Card.svelte';

	const FACT_KEYS = ['region', 'birth_year', 'marital_status', 'children'];
	const RESERVED_DATA_KEYS = new Set([
		'full_name',
		'mobile',
		'motivation',
		'guardian',
		'rejected',
		'__flow_version',
		'__step_index',
		...FACT_KEYS
	]);

	let loading = true;
	let error = '';
	let request = null;
	let lastRequestId = '';

	function asTrimmedString(value) {
		return typeof value === 'string' ? value.trim() : '';
	}

	function asObject(value) {
		return value && typeof value === 'object' ? value : {};
	}

	function parseRequestId(route) {
		if (typeof route !== 'string') return '';
		const match = route.match(/^(?:app|admin)\/requests\/([^/]+)$/);
		return match?.[1] || '';
	}

	function formatStatus(status) {
		const value = asTrimmedString(status);
		if (!value) return '';
		const clean = value.replace(/^\d+-/, '').replaceAll('_', ' ');
		return clean ? clean.charAt(0).toUpperCase() + clean.slice(1) : '';
	}

	function formatDate(value) {
		const dateRaw = asTrimmedString(value);
		if (!dateRaw) return '-';
		const parsed = new Date(dateRaw);
		if (Number.isNaN(parsed.getTime())) return dateRaw;
		return parsed.toLocaleString('it-IT', {
			year: 'numeric',
			month: '2-digit',
			day: '2-digit',
			hour: '2-digit',
			minute: '2-digit'
		});
	}

	function prettyKey(key) {
		return asTrimmedString(key).replaceAll('_', ' ');
	}

	function hasValue(value) {
		if (value === null || value === undefined) return false;
		if (typeof value === 'string') return value.trim() !== '';
		if (Array.isArray(value)) return value.length > 0;
		return true;
	}

	function formatValue(value) {
		if (!hasValue(value)) return '-';
		if (typeof value === 'boolean') return value ? 'true' : 'false';
		if (Array.isArray(value)) return value.join(', ');
		if (typeof value === 'object') return JSON.stringify(value);
		return String(value);
	}

	async function fetchRequestById(id) {
		const response = await apiFetch(`/api/requests/${encodeURIComponent(id)}`);
		if (!response.ok) throw new Error('Failed to load request');
		return response.json();
	}

	$: requestId = parseRequestId($currentRoute);
	$: actorId = asTrimmedString(pb.authStore.record?.id);
	$: isAdmin = !!pb.authStore.record?.admin;
	$: displayName = (() => {
		const fullName = asTrimmedString(request?.data?.full_name);
		if (fullName) return fullName;
		const email = asTrimmedString(request?.email);
		if (email) return email;
		return 'Request';
	})();
	$: requestData = asObject(request?.data);
	$: emailValue = asTrimmedString(request?.email);
	$: createdValue = formatDate(request?.created);
	$: facts = FACT_KEYS.map((key) => asTrimmedString(requestData?.[key])).filter(Boolean);
	$: mobileValue = asTrimmedString(requestData?.mobile);
	$: motivationValue = asTrimmedString(requestData?.motivation);
	$: extraEntries = Object.entries(requestData).filter(([key, value]) => !RESERVED_DATA_KEYS.has(key) && hasValue(value));

	$: if (requestId && requestId !== lastRequestId) {
		lastRequestId = requestId;
		void loadAll();
	}

	function goBack() {
		window.history.back();
	}

	async function loadAll() {
		if (!requestId) return;
		loading = true;
		error = '';
		request = null;
		try {
			request = await fetchRequestById(requestId);
		} catch (err) {
			error = err?.message || err?.toString() || 'Failed to load request';
		} finally {
			loading = false;
		}
	}
</script>

<DashboardLayout title="Request" showBack={true} onBack={goBack}>

	{#if error}
		<Card variant="state">{error}</Card>
	{:else if loading}
		<Card variant="state">Loading...</Card>
	{:else if !request}
		<Card variant="state">Request not found.</Card>
	{:else}
		<Card variant="admin">
			<div class="data">
				<p class="name">{displayName}</p>
				<p class="status">{formatStatus(request.status)}</p>
				<div class="meta">
					{#if emailValue}
						<span class="meta-item">{emailValue}</span>
					{/if}
					<span class="meta-item">{createdValue}</span>
				</div>

				{#if facts.length > 0}
					<div class="facts">
						{#each facts as fact}
							<span class="fact">{fact}</span>
						{/each}
					</div>
				{/if}

				{#if mobileValue}
					<p class="mobile">{mobileValue}</p>
				{/if}

				{#if motivationValue}
					<p class="motivation">{motivationValue}</p>
				{/if}

				{#if extraEntries.length > 0}
					<div class="extra">
						{#each extraEntries as [key, value]}
							<p class="extra-row">{prettyKey(key)}: {formatValue(value)}</p>
						{/each}
					</div>
				{/if}
			</div>
		</Card>
	{/if}
</DashboardLayout>

<style>
	.data {
		display: grid;
		gap: 0.8rem;
	}

	.name {
		margin: 0;
		font-size: 1.3rem;
		font-weight: 800;
	}

	.status {
		margin: 0;
		font-size: 0.95rem;
		font-weight: 700;
	}

	.meta {
		display: flex;
		flex-wrap: wrap;
		gap: 0.45rem;
	}

	.meta-item {
		font-size: 0.88rem;
		font-weight: 600;
	}

	.facts {
		display: flex;
		flex-wrap: wrap;
		gap: 0.45rem;
	}

	.fact {
		font-size: 0.88rem;
		font-weight: 600;
	}

	.mobile {
		margin: 0;
		font-size: 1rem;
		font-weight: 700;
	}

	.motivation {
		margin: 0;
		font-size: 0.95rem;
		line-height: 1.5;
		white-space: pre-wrap;
		word-break: break-word;
	}

	.extra {
		display: grid;
		gap: 0.3rem;
	}

	.extra-row {
		margin: 0;
		font-size: 0.92rem;
		line-height: 1.4;
	}
</style>
