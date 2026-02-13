<script>
	import { apiFetch, fetchSetting, pb } from '../lib/pocketbase';
	import { currentRoute, navigate } from '../lib/router';
	import DashboardLayout from '../components/DashboardLayout.svelte';
	import StateCard from '../components/StateCard.svelte';
	import AdminCard from '../components/AdminCard.svelte';

	const FACT_KEYS = ['region', 'birth_year', 'marital_status', 'children'];
	const RESERVED_DATA_KEYS = new Set(['full_name', 'mobile', 'motivation', ...FACT_KEYS]);

	let loading = true;
	let actionLoading = false;
	let error = '';
	let request = null;
	let groupRecord = null;
	let flowStatuses = [];
	let setStatusBy = {};
	let groupOptions = [];
	let guardianOptions = [];
	let selectedGroup = '';
	let selectedGuardian = '';
	let rejectReason = '';
	let lastRequestId = '';

	function asTrimmedString(value) {
		return typeof value === 'string' ? value.trim() : '';
	}

	function asObject(value) {
		return value && typeof value === 'object' ? value : {};
	}

	function parseRequestId(route) {
		if (typeof route !== 'string') return '';
		const match = route.match(/^app\/requests\/([^/]+)$/);
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

	function nextStatusInFlow(statuses, currentStatus) {
		const idx = statuses.indexOf(currentStatus);
		return idx >= 0 && idx < statuses.length - 1 ? statuses[idx + 1] : '';
	}

	function isRoleAllowed(role, isAdmin, isAssistant, isGuardian) {
		if (isAdmin) return true;
		if (role === 'assistant') return isAssistant;
		if (role === 'guardian') return isGuardian;
		return false;
	}

	function mapOption(record) {
		return {
			id: record.id,
			label: asTrimmedString(record?.name) || asTrimmedString(record?.full_name) || asTrimmedString(record?.email) || record.id
		};
	}

	async function fetchRequestById(id) {
		const response = await apiFetch(`/api/requests/${encodeURIComponent(id)}`);
		if (!response.ok) throw new Error('Failed to load request');
		return response.json();
	}

	async function fetchFlowConfig() {
		const response = await fetchSetting('requests_flow');
		if (!response.ok) return { statuses: [], setStatusBy: {} };
		const payload = await response.json();
		const statuses = Array.isArray(payload?.data?.statuses)
			? payload.data.statuses.map(asTrimmedString).filter(Boolean)
			: [];
		const setStatusBy = asObject(payload?.data?.set_status_by);
		return { statuses, setStatusBy };
	}

	async function loadGroupRecord(groupId) {
		const id = asTrimmedString(groupId);
		if (!id) return null;
		try {
			return await pb.collection('groups').getOne(id);
		} catch {
			return null;
		}
	}

	async function loadGroupOptions() {
		const groupsResult = await pb.collection('groups').getList(1, 500, { sort: 'name' });
		return Array.isArray(groupsResult?.items) ? groupsResult.items.map(mapOption) : [];
	}

	async function loadGuardianOptions(groupId) {
		const response = await apiFetch(`/api/groups/${encodeURIComponent(groupId)}/members`);
		if (!response.ok) return [];
		const payload = await response.json();
		const items = Array.isArray(payload?.items) ? payload.items : [];
		return items.map(mapOption);
	}

	function resetViewState() {
		request = null;
		groupRecord = null;
		groupOptions = [];
		guardianOptions = [];
		selectedGroup = '';
		selectedGuardian = '';
	}

	$: requestId = parseRequestId($currentRoute);
	$: actorId = asTrimmedString(pb.authStore.record?.id);
	$: isAdmin = !!pb.authStore.record?.admin;
	$: isAssistant = asTrimmedString(groupRecord?.assistant) !== '' && asTrimmedString(groupRecord?.assistant) === actorId;
	$: isGuardian = asTrimmedString(request?.guardian) !== '' && asTrimmedString(request?.guardian) === actorId;
	$: currentStatus = asTrimmedString(request?.status);
	$: nextStatus = nextStatusInFlow(flowStatuses, currentStatus);
	$: requiredRole = asTrimmedString(setStatusBy[nextStatus]);
	$: canTransition = !request?.rejected && !!nextStatus && isRoleAllowed(requiredRole, isAdmin, isAssistant, isGuardian);
	$: needsGroup = nextStatus === '2-group_assigned';
	$: needsGuardian = nextStatus === '3-guardian_assigned';
	$: canReject = isAdmin && !request?.rejected;
	$: finalStatus = flowStatuses.length > 0 ? flowStatuses[flowStatuses.length - 1] : '';
	$: canPromote = isAdmin && !request?.rejected && currentStatus === finalStatus;
	$: nextLabel = needsGroup ? 'Assign group' : (needsGuardian ? 'Assign guardian' : 'Next');
	$: selectedGroupValue = asTrimmedString(selectedGroup);
	$: selectedGuardianValue = asTrimmedString(selectedGuardian);
	$: canSubmitNext = canTransition &&
		(!needsGroup || selectedGroupValue !== '') &&
		(!needsGuardian || selectedGuardianValue !== '');
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
		if (window.history.length > 1) {
			window.history.back();
			return;
		}
		if (request?.group) {
			navigate(`app/groups/${encodeURIComponent(request.group)}/requests`);
			return;
		}
		navigate('app/groups');
	}

	async function postAction(payload) {
		actionLoading = true;
		error = '';
		try {
			const response = await apiFetch(`/api/requests/${encodeURIComponent(requestId)}/action`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(payload)
			});
			if (!response.ok) throw new Error('Action failed');
			const data = await response.json();
			if (data?.promoted) {
				navigate('app/groups');
				return;
			}
			await loadAll();
		} catch (err) {
			error = err?.message || err?.toString() || 'Action failed';
		} finally {
			actionLoading = false;
		}
	}

	async function handleNext() {
		if (!canSubmitNext) return;
		const payload = {
			action: 'transition',
			target_status: nextStatus
		};
		if (needsGroup) payload.group = selectedGroupValue;
		if (needsGuardian) payload.guardian = selectedGuardianValue;
		await postAction(payload);
	}

	async function handleReject() {
		const reason = asTrimmedString(rejectReason);
		if (!reason) return;
		await postAction({ action: 'reject', reason });
	}

	async function handlePromote() {
		await postAction({ action: 'promote' });
	}

	async function loadAll() {
		if (!requestId) return;
		loading = true;
		error = '';
		resetViewState();
		try {
			const [requestPayload, flowConfig] = await Promise.all([
				fetchRequestById(requestId),
				fetchFlowConfig()
			]);
			request = requestPayload;
			flowStatuses = flowConfig.statuses;
			setStatusBy = flowConfig.setStatusBy;

			groupRecord = await loadGroupRecord(requestPayload?.group);

			const loadedNextStatus = nextStatusInFlow(flowStatuses, asTrimmedString(requestPayload?.status));
			if (loadedNextStatus === '2-group_assigned' && isAdmin) {
				groupOptions = await loadGroupOptions();
				selectedGroup = asTrimmedString(requestPayload?.group);
			}
			if (loadedNextStatus === '3-guardian_assigned' && asTrimmedString(requestPayload?.group)) {
				guardianOptions = await loadGuardianOptions(requestPayload.group);
				selectedGuardian = asTrimmedString(requestPayload?.guardian);
			}
		} catch (err) {
			error = err?.message || err?.toString() || 'Failed to load request';
		} finally {
			loading = false;
		}
	}
</script>

<DashboardLayout title="Request">
	{#if error}
		<StateCard>{error}</StateCard>
	{:else if loading}
		<StateCard>Loading...</StateCard>
	{:else if !request}
		<StateCard>Request not found.</StateCard>
	{:else}
		<AdminCard>
			<button class="back" type="button" on:click={goBack}>Back</button>
		</AdminCard>

		<AdminCard>
			<div class="flow">
				<p class="name">{displayName}</p>
				<p class="current">{formatStatus(request.status)}</p>
				{#if flowStatuses.length > 0}
					<div class="steps">
						{#each flowStatuses as status, index}
							<span class="step" class:active={status === request.status}>{index + 1}</span>
						{/each}
					</div>
				{/if}
			</div>
		</AdminCard>

		{#if canTransition || canReject || canPromote}
			<AdminCard>
				<div class="actions">
					{#if canTransition}
						{#if needsGroup}
							<select bind:value={selectedGroup} disabled={actionLoading}>
								<option value="">Group</option>
								{#each groupOptions as option}
									<option value={option.id}>{option.label}</option>
								{/each}
							</select>
						{/if}
						{#if needsGuardian}
							<select bind:value={selectedGuardian} disabled={actionLoading}>
								<option value="">Guardian</option>
								{#each guardianOptions as option}
									<option value={option.id}>{option.label}</option>
								{/each}
							</select>
						{/if}
						<button type="button" disabled={!canSubmitNext || actionLoading} on:click={handleNext}>
							{nextLabel}
						</button>
					{/if}

					{#if canReject}
						<input type="text" bind:value={rejectReason} placeholder="Reason" disabled={actionLoading} />
						<button type="button" disabled={!asTrimmedString(rejectReason) || actionLoading} on:click={handleReject}>
							Reject
						</button>
					{/if}

					{#if canPromote}
						<button type="button" disabled={actionLoading} on:click={handlePromote}>
							Promote
						</button>
					{/if}
				</div>
			</AdminCard>
		{/if}

		<AdminCard>
			<div class="data">
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
		</AdminCard>
	{/if}
</DashboardLayout>

<style>
	.back {
		border: 1px solid #000;
		background: #fff;
		padding: 0.45rem 0.7rem;
		cursor: pointer;
	}

	.flow {
		display: grid;
		gap: 0.55rem;
	}

	.current {
		margin: 0;
		font-size: 1rem;
		font-weight: 700;
	}

	.name {
		margin: 0;
		font-size: 1.1rem;
		font-weight: 800;
	}

	.steps {
		display: flex;
		gap: 0.35rem;
		flex-wrap: wrap;
	}

	.step {
		width: 1.4rem;
		height: 1.4rem;
		border: 1px solid #000;
		display: inline-flex;
		align-items: center;
		justify-content: center;
		font-size: 0.75rem;
		font-weight: 700;
	}

	.step.active {
		background: #000;
		color: #fff;
	}

	.actions {
		display: grid;
		gap: 0.5rem;
	}

	select,
	input,
	.actions button {
		border: 1px solid #000;
		background: #fff;
		padding: 0.55rem 0.7rem;
		font-size: 0.95rem;
	}

	.data {
		display: grid;
		gap: 0.65rem;
	}

	.meta {
		display: flex;
		flex-wrap: wrap;
		gap: 0.35rem;
	}

	.meta-item {
		font-size: 0.78rem;
		font-weight: 600;
	}

	.facts {
		display: flex;
		flex-wrap: wrap;
		gap: 0.35rem;
	}

	.fact {
		font-size: 0.78rem;
		font-weight: 600;
	}

	.mobile {
		margin: 0;
		font-size: 0.92rem;
		font-weight: 700;
	}

	.motivation {
		margin: 0;
		font-size: 0.9rem;
		white-space: pre-wrap;
		word-break: break-word;
	}

	.extra {
		display: grid;
		gap: 0.3rem;
	}

	.extra-row {
		margin: 0;
		font-size: 0.82rem;
	}
</style>
