<script>
	import { onMount } from 'svelte';
	import { apiFetch, pb } from '../../lib/pocketbase';
	import { navigate, queryParams } from '../../lib/router';
	import DashboardLayout from '../DashboardLayout.svelte';
	import Card from '../Card.svelte';

	export let title = 'Requests';
	export let adminMode = false;

	let loading = true;
	let error = '';
	let requests = [];
	let lastGroupFilter = '';

	let groupOptions = [];
	let groupOptionsLoaded = false;
	let groupOptionsLoading = false;
	let guardianOptionsByGroup = {};
	let guardianLoadingByGroup = {};

	let selectedGroupById = {};
	let selectedGuardianById = {};
	let mentoringNotesById = {};
	let rejectReasonById = {};
	let actionLoadingById = {};

	$: groupFilter = ($queryParams?.group_id || '').trim();
	$: isAdmin = !!pb.authStore.record?.admin;

	$: if (groupFilter !== lastGroupFilter) {
		lastGroupFilter = groupFilter;
		void loadRequests();
	}

	$: {
		for (const item of requests) {
			if (!canTransition(item)) continue;
			if (needsGroup(item)) void ensureGroupOptions();
			if (needsGuardian(item) && asTrimmedString(item?.group)) {
				void ensureGuardianOptions(item.group);
			}
		}
	}

	function asTrimmedString(value) {
		return typeof value === 'string' ? value.trim() : '';
	}

	function asObject(value) {
		return value && typeof value === 'object' ? value : {};
	}

	function asArray(value) {
		return Array.isArray(value) ? value : [];
	}

	function displayName(item) {
		const value = item?.data?.full_name;
		if (typeof value === 'string' && value.trim()) return value.trim();
		return item?.email || 'Unknown';
	}

	function displayRegion(item) {
		const value = item?.data?.region;
		if (typeof value === 'string' && value.trim()) return value.trim();
		return '';
	}

	function mapOption(record) {
		const id = asTrimmedString(record?.id) || asTrimmedString(record?.user) || asTrimmedString(record?.value);
		return {
			id,
			label: asTrimmedString(record?.name) || asTrimmedString(record?.full_name) || asTrimmedString(record?.email) || record.id
		};
	}

	function workflow(item) {
		return asObject(item?.workflow);
	}

	function totalSteps(item) {
		const total = Number(workflow(item)?.total_steps ?? 0);
		return Number.isFinite(total) && total > 0 ? total : 0;
	}

	function visualStepIndex(item) {
		const total = totalSteps(item);
		if (total <= 0) return 0;
		const raw = Number(item?.step_index ?? 0);
		const step = Number.isFinite(raw) ? raw + 1 : 1;
		return Math.max(1, Math.min(total, step));
	}

	function progressPercent(item) {
		const total = totalSteps(item);
		if (total <= 0) return 0;
		return Math.min(100, Math.round((visualStepIndex(item) / total) * 100));
	}

	function currentStepLabel(item) {
		const wf = workflow(item);
		const nextLabel = asTrimmedString(wf?.next_action_label);
		return nextLabel;
	}

	function daysInCurrentStep(item) {
		const raw = asTrimmedString(item?.updated);
		if (!raw) return 0;
		const dt = new Date(raw);
		if (Number.isNaN(dt.getTime())) return 0;
		const ms = Date.now() - dt.getTime();
		if (ms <= 0) return 0;
		return Math.floor(ms / 86400000);
	}

	function currentStepMeta(item) {
		const days = daysInCurrentStep(item);
		const label = currentStepLabel(item);
		return label ? `${label} · ${days}d` : `${days}d`;
	}

	function canTransition(item) {
		const wf = workflow(item);
		return !item?.rejected && !!wf?.has_next_step && !!wf?.can_advance;
	}

	function nextAction(item) {
		return asTrimmedString(workflow(item)?.next_action);
	}

	function requiredField(item) {
		return asTrimmedString(workflow(item)?.required_field);
	}

	function needsGroup(item) {
		return requiredField(item) === 'group';
	}

	function needsGuardian(item) {
		return requiredField(item) === 'guardian';
	}

	function needsMentoringNotes(item) {
		const wf = workflow(item);
		return !!wf?.has_next_step && nextAction(item) === 'mentoring';
	}

	function isPromoteStep(item) {
		const wf = workflow(item);
		return !!wf?.has_next_step && nextAction(item) === 'admin_approved' && !!wf?.can_advance;
	}

	function canPromote(item) {
		const wf = workflow(item);
		return isAdmin && !item?.rejected && (!wf?.has_next_step || isPromoteStep(item));
	}

	function canReject(item) {
		return isAdmin && !item?.rejected;
	}

	function nextLabel(item) {
		const label = asTrimmedString(workflow(item)?.next_action_label);
		if (label) return label;
		return 'Advance';
	}

	function selectedGroup(item) {
		return asTrimmedString(selectedGroupById[item?.id]) || asTrimmedString(item?.group);
	}

	function selectedGuardian(item) {
		return asTrimmedString(selectedGuardianById[item?.id]);
	}

	function mentoringNotes(item) {
		return asTrimmedString(mentoringNotesById[item?.id]);
	}

	function rejectReason(item) {
		return asTrimmedString(rejectReasonById[item?.id]);
	}

	function canSubmitNext(item) {
		if (!canTransition(item)) return false;
		if (needsGroup(item) && !selectedGroup(item)) return false;
		if (needsGuardian(item) && !selectedGuardian(item)) return false;
		if (needsMentoringNotes(item) && !mentoringNotes(item)) return false;
		return true;
	}

	function isActionLoading(item) {
		return !!actionLoadingById[item?.id];
	}

	function setActionLoading(id, value) {
		actionLoadingById = { ...actionLoadingById, [id]: !!value };
	}

	function openRequest(item) {
		if (!item?.id) return;
		const base = adminMode ? 'admin/requests' : 'app/requests';
		navigate(`${base}/${encodeURIComponent(item.id)}`);
	}

	function setSelectedGroup(id, value) {
		selectedGroupById = { ...selectedGroupById, [id]: value };
	}

	function setSelectedGuardian(id, value) {
		const nextValue = asTrimmedString(value);
		selectedGuardianById = { ...selectedGuardianById, [id]: nextValue };
	}

	function setMentoringNotes(id, value) {
		mentoringNotesById = { ...mentoringNotesById, [id]: value };
	}

	function setRejectReason(id, value) {
		rejectReasonById = { ...rejectReasonById, [id]: value };
	}

	function primeSelectionState(items) {
		const nextGroup = { ...selectedGroupById };
		const nextGuardian = { ...selectedGuardianById };
		for (const item of items) {
			if (!nextGroup[item.id]) nextGroup[item.id] = asTrimmedString(item.group);
			if (nextGuardian[item.id] === undefined) nextGuardian[item.id] = '';
		}
		selectedGroupById = nextGroup;
		selectedGuardianById = nextGuardian;
	}

	async function ensureGroupOptions() {
		if (groupOptionsLoaded || groupOptionsLoading) return;
		groupOptionsLoading = true;
		try {
			const result = await pb.collection('groups').getList(1, 500, { sort: 'name' });
			groupOptions = asArray(result?.items).map(mapOption);
			groupOptionsLoaded = true;
		} catch {
			groupOptions = [];
		} finally {
			groupOptionsLoading = false;
		}
	}

	async function ensureGuardianOptions(groupId) {
		const id = asTrimmedString(groupId);
		if (!id || guardianOptionsByGroup[id] || guardianLoadingByGroup[id]) return;
		guardianLoadingByGroup = { ...guardianLoadingByGroup, [id]: true };
		try {
			const response = await apiFetch(`/api/groups/${encodeURIComponent(id)}/members`);
			if (!response.ok) throw new Error('failed_guardians');
			const payload = await response.json();
			const items = asArray(payload?.items).map(mapOption);
			guardianOptionsByGroup = { ...guardianOptionsByGroup, [id]: items };
		} catch {
			guardianOptionsByGroup = { ...guardianOptionsByGroup, [id]: [] };
		} finally {
			guardianLoadingByGroup = { ...guardianLoadingByGroup, [id]: false };
		}
	}

	async function postAction(item, payload) {
		if (!item?.id) return;
		setActionLoading(item.id, true);
		error = '';
		try {
			const response = await apiFetch(`/api/requests/${encodeURIComponent(item.id)}/action`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(payload)
			});
			if (!response.ok) throw new Error('Action failed');
			await loadRequests();
		} catch (err) {
			error = err?.message || err?.toString() || 'Action failed';
		} finally {
			setActionLoading(item.id, false);
		}
	}

	async function handleNext(item) {
		if (!canSubmitNext(item) || isActionLoading(item)) return;
		const payload = { action: 'advance' };
		const group = selectedGroup(item);
		if (needsGroup(item) && group) payload.group = group;
		const guardian = selectedGuardian(item);
		if (needsGuardian(item) && guardian) payload.guardian = guardian;
		if (needsMentoringNotes(item)) payload.mentoring_notes = mentoringNotes(item);
		await postAction(item, payload);
	}

	async function handlePromote(item) {
		if (!canPromote(item) || isActionLoading(item)) return;
		await postAction(item, { action: 'promote' });
	}

	async function handleReject(item) {
		const reason = rejectReason(item);
		if (!canReject(item) || !reason || isActionLoading(item)) return;
		await postAction(item, { action: 'reject', reason });
	}

	async function loadRequests() {
		loading = true;
		error = '';
		const params = new URLSearchParams();
		if (groupFilter) params.set('group_id', groupFilter);
		if (adminMode) params.set('include_rejected', '1');

		const query = params.toString();
		try {
			const response = await apiFetch(`/api/requests${query ? `?${query}` : ''}`);
			if (!response.ok) throw new Error('Failed to load requests');
			const data = await response.json();
			requests = asArray(data?.items);
			primeSelectionState(requests);
		} catch (err) {
			error = err?.message || err?.toString() || 'Failed to load requests';
		} finally {
			loading = false;
		}
	}

	onMount(loadRequests);

	$: visibleRequests = requests.filter((item) => !item?.rejected);
	$: rejectedRequests = requests.filter((item) => !!item?.rejected);
</script>

<DashboardLayout {title}>
	{#if error}
		<Card variant="state">{error}</Card>
	{:else if loading}
		<Card variant="state">Loading requests...</Card>
	{:else if requests.length === 0}
		<Card variant="state">No requests.</Card>
	{:else}
		<div class="list">
			{#each visibleRequests as item (item.id)}
				<Card variant="item">
					<div class="head">
						<div>
							<p class="name">{displayName(item)}</p>
							{#if displayRegion(item)}
								<p class="meta">{displayRegion(item)}</p>
							{/if}
						</div>
						<button class="details" type="button" on:click={() => openRequest(item)}>Details</button>
					</div>

					{#if totalSteps(item) > 0}
						<div class="progress-track" role="presentation">
							<div class="progress-fill" style={`width: ${progressPercent(item)}%`}></div>
						</div>
						<div class="steps" aria-hidden="true">
							{#each Array(totalSteps(item)) as _, index}
								<span class="step" class:active={index < visualStepIndex(item)}>{index + 1}</span>
							{/each}
						</div>
						<p class="step-meta">{currentStepMeta(item)}</p>
					{/if}

					{#if canTransition(item) || canPromote(item) || canReject(item)}
						<div class="actions">
							{#if canTransition(item) && !isPromoteStep(item)}
								{#if needsGroup(item)}
									<select
										value={selectedGroup(item)}
										on:change={(e) => setSelectedGroup(item.id, e.currentTarget.value)}
										disabled={isActionLoading(item)}
									>
										<option value="">Select group</option>
										{#each groupOptions as option}
											<option value={option.id}>{option.label}</option>
										{/each}
									</select>
								{/if}

								{#if needsGuardian(item)}
									<select
										bind:value={selectedGuardianById[item.id]}
										on:change={(e) => setSelectedGuardian(item.id, e.currentTarget.value)}
										disabled={isActionLoading(item)}
									>
										<option value="">Select guardian</option>
										{#each (guardianOptionsByGroup[asTrimmedString(item.group)] || []) as option}
											<option value={option.id}>{option.label}</option>
										{/each}
									</select>
								{/if}

								{#if needsMentoringNotes(item)}
									<textarea
										rows="3"
										placeholder="Mentoring notes"
										value={mentoringNotes(item)}
										on:input={(e) => setMentoringNotes(item.id, e.currentTarget.value)}
										disabled={isActionLoading(item)}
									></textarea>
								{/if}

								<button type="button" disabled={!canSubmitNext(item) || isActionLoading(item)} on:click={() => handleNext(item)}>
									{nextLabel(item)}
								</button>
							{/if}

							{#if canPromote(item)}
								<button type="button" disabled={isActionLoading(item)} on:click={() => handlePromote(item)}>Promote</button>
							{/if}

							{#if canReject(item)}
								<div class="reject-row">
									<input
										type="text"
										placeholder="Reject reason"
										value={rejectReason(item)}
										on:input={(e) => setRejectReason(item.id, e.currentTarget.value)}
										disabled={isActionLoading(item)}
									/>
									<button type="button" disabled={!rejectReason(item) || isActionLoading(item)} on:click={() => handleReject(item)}>
										Reject
									</button>
								</div>
							{/if}
						</div>
					{/if}
				</Card>
			{/each}

			{#if rejectedRequests.length > 0}
				<div class="divider" aria-hidden="true">---</div>
				{#each rejectedRequests as item (item.id)}
					<div class="rejected">
						<Card variant="item">
							<div class="head">
								<div>
									<p class="name">{displayName(item)}</p>
									{#if displayRegion(item)}
										<p class="meta">{displayRegion(item)}</p>
									{/if}
								</div>
								<button class="details" type="button" on:click={() => openRequest(item)}>Details</button>
							</div>

							{#if totalSteps(item) > 0}
								<div class="progress-track" role="presentation">
									<div class="progress-fill" style={`width: ${progressPercent(item)}%`}></div>
								</div>
								<div class="steps" aria-hidden="true">
									{#each Array(totalSteps(item)) as _, index}
										<span class="step" class:active={index < visualStepIndex(item)}>{index + 1}</span>
									{/each}
								</div>
								<p class="step-meta">{currentStepMeta(item)}</p>
							{/if}
						</Card>
					</div>
				{/each}
			{/if}
		</div>
	{/if}
</DashboardLayout>

<style>
	.list {
		display: grid;
		gap: 0.65rem;
	}

	.head {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		gap: 0.75rem;
	}

	.details {
		border: 1px solid #000;
		background: #fff;
		padding: 0.35rem 0.6rem;
		font-size: 0.8rem;
		font-weight: 700;
		cursor: pointer;
	}

	.name {
		margin: 0;
		font-weight: 700;
	}

	.meta {
		margin: 0.2rem 0 0;
		font-size: 0.9rem;
	}

	.progress-track {
		width: 100%;
		height: 0.35rem;
		border: 1px solid #000;
		background: #fff;
	}

	.progress-fill {
		height: 100%;
		background: #000;
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

	.step-meta {
		margin: 0;
		font-size: 0.85rem;
		font-weight: 600;
		color: #222;
	}

	.actions {
		display: grid;
		gap: 0.5rem;
		padding-top: 0.5rem;
		border-top: 1px solid #000;
	}

	select,
	textarea,
	input,
	.actions button {
		border: 1px solid #000;
		background: #fff;
		padding: 0.6rem 0.7rem;
		font-size: 0.9rem;
	}

	textarea {
		resize: vertical;
		font-family: inherit;
	}

	.reject-row {
		display: grid;
		gap: 0.5rem;
	}

	.divider {
		text-align: center;
		font-weight: 700;
		letter-spacing: 0.1em;
		padding: 0.25rem 0;
	}

	.rejected {
		opacity: 0.7;
	}
</style>
