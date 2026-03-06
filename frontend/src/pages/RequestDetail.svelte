<script>
	import { apiFetch, pb } from '../lib/pocketbase';
	import { currentRoute } from '../lib/router';
	import DashboardLayout from '../components/DashboardLayout.svelte';
	import Card from '../components/Card.svelte';
	import ConfirmModal from '../components/modals/ConfirmModal.svelte';

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
	let members = [];
	let groupRequests = [];
	let guardianCountsById = {};
	let actionLoading = false;
	let actionError = '';
	let confirmOpen = false;
	let confirmTitle = '';
	let confirmMessage = '';
	let pendingGuardianId = '';
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

	async function fetchJSONOrThrow(path) {
		const response = await apiFetch(path);
		if (!response.ok) throw new Error('Failed to load group data');
		return response.json();
	}

	function displayName(item) {
		const fullName = asTrimmedString(item?.full_name ?? item?.data?.full_name);
		if (fullName) return fullName;
		const email = asTrimmedString(item?.email);
		if (email) return email;
		return 'Unknown';
	}

	function isGuardianEntry(requestItem) {
		return asTrimmedString(requestItem?.guardian) !== '';
	}

	function openGuardianConfirm(userId) {
		const guardianId = asTrimmedString(userId);
		if (!guardianId || actionLoading) return;
		const member = members.find((item) => asTrimmedString(item?.id) === guardianId);
		const name = displayName(member);

		pendingGuardianId = guardianId;
		if (selectedGuardianId === guardianId) {
			confirmTitle = 'Remove guardian';
			confirmMessage = `${name} will be removed as guardian for this request.`;
		} else {
			confirmTitle = 'Set guardian';
			confirmMessage = `${name} will become guardian for this request.`;
		}
		confirmOpen = true;
	}

	function closeGuardianConfirm() {
		if (actionLoading) return;
		confirmOpen = false;
		pendingGuardianId = '';
	}

	async function confirmGuardianAction() {
		const guardianId = asTrimmedString(pendingGuardianId);
		if (!guardianId) return;
		await setGuardian(guardianId);
	}

	async function setGuardian(userId) {
		const reqId = asTrimmedString(request?.id);
		const guardianId = asTrimmedString(userId);
		if (!reqId || !guardianId || actionLoading) return;

		actionLoading = true;
		actionError = '';
		try {
			const response = await apiFetch(`/api/requests/${encodeURIComponent(reqId)}/action`, {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json'
				},
				body: JSON.stringify({
					action: 'set_guardian',
					guardian: selectedGuardianId === guardianId ? '' : guardianId
				})
			});
			if (!response.ok) {
				const payload = await response.json().catch(() => ({}));
				throw new Error(asTrimmedString(payload?.message) || 'Failed to assign guardian');
			}
			await loadAll();
			confirmOpen = false;
			pendingGuardianId = '';
		} catch (err) {
			actionError = err?.message || err?.toString() || 'Failed to assign guardian';
		} finally {
			actionLoading = false;
		}
	}

	$: requestId = parseRequestId($currentRoute);
	$: actorId = asTrimmedString(pb.authStore.record?.id);
	$: isAdmin = !!pb.authStore.record?.admin;
	$: requestTitle = (() => {
		const fullName = asTrimmedString(request?.data?.full_name);
		if (fullName) return fullName;
		const email = asTrimmedString(request?.email);
		if (email) return email;
		return 'Request';
	})();
	$: requestData = asObject(request?.data);
	$: groupId = asTrimmedString(request?.group);
	$: selectedGuardianId = asTrimmedString(request?.guardian);
	$: emailValue = asTrimmedString(request?.email);
	$: createdValue = formatDate(request?.created);
	$: facts = FACT_KEYS.map((key) => asTrimmedString(requestData?.[key])).filter(Boolean);
	$: mobileValue = asTrimmedString(requestData?.mobile);
	$: motivationValue = asTrimmedString(requestData?.motivation);
	$: extraEntries = Object.entries(requestData).filter(([key, value]) => !RESERVED_DATA_KEYS.has(key) && hasValue(value));
	$: requestList = groupRequests.filter((item) => !isGuardianEntry(item));
	$: mentoringList = groupRequests.filter((item) => isGuardianEntry(item));

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
		actionError = '';
		request = null;
		members = [];
		groupRequests = [];
		guardianCountsById = {};
		try {
			request = await fetchRequestById(requestId);
			const currentGroupId = asTrimmedString(request?.group);
			if (currentGroupId) {
				const [membersData, guardiansData, requestsData] = await Promise.all([
					fetchJSONOrThrow(`/api/groups/${encodeURIComponent(currentGroupId)}/members`),
					fetchJSONOrThrow(`/api/groups/${encodeURIComponent(currentGroupId)}/guardians`),
					fetchJSONOrThrow(`/api/requests?group_id=${encodeURIComponent(currentGroupId)}`)
				]);

				members = Array.isArray(membersData?.items) ? membersData.items : [];
				groupRequests = Array.isArray(requestsData?.items) ? requestsData.items : [];
				const guardians = Array.isArray(guardiansData?.items) ? guardiansData.items : [];
				guardianCountsById = Object.fromEntries(
					guardians
						.map((item) => [asTrimmedString(item?.id), Number(item?.proteges_count || 0)])
						.filter(([id]) => !!id)
				);
			}
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
				<p class="name">{requestTitle}</p>
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

				{#if requestList.length > 0}
					<div class="section">
						<h3>Request</h3>
						<div class="rows">
							{#each requestList as item}
								<p class="row">{displayName(item)}</p>
							{/each}
						</div>
					</div>
				{/if}

				{#if mentoringList.length > 0}
					<div class="section">
						<h3>Mentoring</h3>
						<div class="rows">
							{#each mentoringList as item}
								<p class="row">{displayName(item)}</p>
							{/each}
						</div>
					</div>
				{/if}

				{#if members.length > 0}
					<div class="section">
						<h3>Members</h3>
						{#if actionError}
							<p class="action-error">{actionError}</p>
						{/if}
						<div class="member-list">
							{#each members as member}
								<div class="member-item">
									<div class="member-text">
										<p class="member-name">{displayName(member)}</p>
										{#if guardianCountsById[member.id] > 0 && selectedGuardianId !== member.id}
											<p class="member-meta">
												<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="currentColor" viewBox="0 0 16 16" aria-hidden="true">
													<path d="M12.5 16a3.5 3.5 0 1 0 0-7 3.5 3.5 0 0 0 0 7m.5-5v1h1a.5.5 0 0 1 0 1h-1v1a.5.5 0 0 1-1 0v-1h-1a.5.5 0 0 1 0-1h1v-1a.5.5 0 0 1 1 0m-2-6a3 3 0 1 1-6 0 3 3 0 0 1 6 0"/>
													<path d="M2 13c0 1 1 1 1 1h5.256A4.5 4.5 0 0 1 8 12.5a4.5 4.5 0 0 1 1.544-3.393Q8.844 9.002 8 9c-5 0-6 3-6 4"/>
												</svg>
												<span>{guardianCountsById[member.id]}</span>
											</p>
										{/if}
									</div>
									<button
										class="assign-btn"
										type="button"
										disabled={actionLoading}
										aria-label={selectedGuardianId === member.id ? 'Remove guardian' : 'Set guardian'}
										on:click={() => openGuardianConfirm(member.id)}
									>
										{#if selectedGuardianId === member.id}
											<svg class="icon-remove" xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="currentColor" viewBox="0 0 16 16" aria-hidden="true">
												<path d="M16 8A8 8 0 1 1 0 8a8 8 0 0 1 16 0M5.354 4.646a.5.5 0 1 0-.708.708L7.293 8l-2.647 2.646a.5.5 0 0 0 .708.708L8 8.707l2.646 2.647a.5.5 0 0 0 .708-.708L8.707 8l2.647-2.646a.5.5 0 0 0-.708-.708L8 7.293z"/>
											</svg>
										{:else}
											<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="currentColor" viewBox="0 0 16 16" aria-hidden="true">
												<path fill-rule="evenodd" d="M15.854 5.146a.5.5 0 0 1 0 .708l-3 3a.5.5 0 0 1-.708 0l-1.5-1.5a.5.5 0 0 1 .708-.708L12.5 7.793l2.646-2.647a.5.5 0 0 1 .708 0"/>
												<path d="M1 14s-1 0-1-1 1-4 6-4 6 3 6 4-1 1-1 1zm5-6a3 3 0 1 0 0-6 3 3 0 0 0 0 6"/>
											</svg>
										{/if}
									</button>
								</div>
							{/each}
						</div>
					</div>
				{/if}
			</div>
		</Card>
	{/if}
</DashboardLayout>

<ConfirmModal
	show={confirmOpen}
	title={confirmTitle}
	message={confirmMessage}
	confirmLabel="Confirm"
	loading={actionLoading}
	onConfirm={confirmGuardianAction}
	onCancel={closeGuardianConfirm}
/>

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

	.section {
		display: grid;
		gap: 0.45rem;
		padding-top: 0.2rem;
	}

	h3 {
		margin: 0;
		font-size: 0.96rem;
		font-weight: 800;
	}

	.rows {
		display: grid;
		gap: 0.35rem;
	}

	.row {
		margin: 0;
		font-size: 0.9rem;
	}

	.member-list {
		display: grid;
		gap: 0.5rem;
	}

	.member-item {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 0.6rem;
		padding: 0.55rem 0.65rem;
		border: 1px solid #d5d5d5;
		border-radius: 0.6rem;
	}

	.member-text {
		min-width: 0;
	}

	.member-name {
		margin: 0;
		font-size: 0.9rem;
		font-weight: 700;
	}

	.member-meta {
		margin: 0.15rem 0 0;
		font-size: 0.8rem;
		display: inline-flex;
		align-items: center;
		gap: 0.25rem;
	}

	.member-meta svg {
		display: block;
	}

	.assign-btn {
		border: none;
		background: transparent;
		color: #111;
		width: 2rem;
		height: 2rem;
		padding: 0;
		cursor: pointer;
		display: inline-flex;
		align-items: center;
		justify-content: center;
	}

	.assign-btn svg {
		display: block;
	}

	.icon-remove {
		color: #d50000;
	}

	.assign-btn:disabled {
		opacity: 0.55;
		cursor: default;
	}

	.action-error {
		margin: 0;
		font-size: 0.85rem;
		color: #b00020;
		font-weight: 700;
	}
</style>
