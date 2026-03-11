<script>
	import { apiFetch, pb } from '../lib/pocketbase';
import { currentRoute, navigate } from '../lib/router';
	import DashboardLayout from '../components/DashboardLayout.svelte';
	import Card from '../components/Card.svelte';
	import GroupCard from '../components/cards/GroupCard.svelte';
	import ActionDialog from '../components/modals/ActionDialog.svelte';
	import { renderContent } from '../lib/markdown';

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
	let groupName = '';
	let groupMembers = [];
	let actionLoading = false;
	let actionError = '';
	let rejectDialogOpen = false;
	let rejectReason = '';
	let guardianAssignDialogOpen = false;
	let pendingGuardian = null;
	let groupPickerOpen = false;
	let assignDialogOpen = false;
	let assignTargetGroup = null;
	let groups = [];
	let groupsLoading = false;
	let groupsError = '';
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

	function guardianDisplayName(item) {
		const data = asObject(item?.data);
		const guardianData = asObject(data?.guardian);
		const fromData = asTrimmedString(guardianData?.name);
		if (fromData) return fromData;
		const fromString = asTrimmedString(data?.guardian);
		if (fromString) return fromString;
		return '';
	}

	function workflow(item) {
		return asObject(item?.workflow);
	}

	function nextAction(item) {
		return asTrimmedString(workflow(item)?.next_action);
	}

	function nextActionNotes(item) {
		return asTrimmedString(workflow(item)?.next_action_notes);
	}

	function canAdvance(item) {
		const wf = workflow(item);
		return !item?.rejected && !!wf?.has_next_step && !!wf?.can_advance;
	}

	function canReject(item) {
		return isAdmin && !item?.rejected;
	}

	function canPromote(item) {
		if (!isAdmin || item?.rejected) return false;
		const wf = workflow(item);
		const action = nextAction(item);
		if (!wf?.has_next_step) return true;
		return action === 'admin_approved' && !!wf?.can_advance;
	}

	function canAccept(item) {
		if (canPromote(item)) return true;
		if (!canAdvance(item)) return false;
		return requiresGroupSelection(item);
	}

	function canSetGuardian(item) {
		return !isAdmin && canAdvance(item) && nextAction(item) === 'assign_guardian';
	}

	function requiresGroupSelection(item) {
		return nextAction(item) === 'assign_group';
	}

	function openGuardianAssignDialog(member) {
		if (!member?.id || actionLoading) return;
		pendingGuardian = member;
		guardianAssignDialogOpen = true;
	}

	function closeGuardianAssignDialog() {
		if (actionLoading) return;
		guardianAssignDialogOpen = false;
		pendingGuardian = null;
	}

	async function confirmGuardianAssign() {
		const guardianId = asTrimmedString(pendingGuardian?.id);
		if (!guardianId || actionLoading) return;
		const ok = await postAction({ action: 'set_guardian', guardian: guardianId });
		if (ok) {
			closeGuardianAssignDialog();
		}
	}

	async function postAction(payload) {
		const reqId = asTrimmedString(request?.id);
		if (!reqId || actionLoading) return false;
		actionLoading = true;
		actionError = '';
		try {
			const response = await apiFetch(`/api/requests/${encodeURIComponent(reqId)}/action`, {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json'
				},
				body: JSON.stringify(payload)
			});
			if (!response.ok) {
				const errorPayload = await response.json().catch(() => ({}));
				throw new Error(asTrimmedString(errorPayload?.message) || 'Action failed');
			}
			await loadAll();
			return true;
		} catch (err) {
			actionError = err?.message || err?.toString() || 'Action failed';
			return false;
		} finally {
			actionLoading = false;
		}
	}

	async function handleAcceptAction() {
		if (!canAccept(request)) return;
		if (canPromote(request)) {
			await postAction({ action: 'promote' });
			return;
		}
		if (requiresGroupSelection(request)) {
			await openGroupPicker();
			return;
		}
		await postAction({ action: 'advance' });
	}

	async function openGroupPicker() {
		groupPickerOpen = true;
		await loadGroups();
	}

	function closeGroupPicker() {
		if (actionLoading) return;
		groupPickerOpen = false;
		assignDialogOpen = false;
		assignTargetGroup = null;
	}

	function openAssignDialog(group) {
		if (!group?.id || actionLoading) return;
		assignTargetGroup = group;
		assignDialogOpen = true;
	}

	function closeAssignDialog() {
		if (actionLoading) return;
		assignDialogOpen = false;
		assignTargetGroup = null;
	}

	async function confirmAssignGroup() {
		const groupId = asTrimmedString(assignTargetGroup?.id);
		if (!groupId || actionLoading) return;
		const ok = await postAction({ action: 'advance', group: groupId });
		if (ok) {
			assignDialogOpen = false;
			assignTargetGroup = null;
			groupPickerOpen = false;
		}
	}

	async function handleRejectAction() {
		if (!canReject(request) || actionLoading) return;
		rejectReason = '';
		rejectDialogOpen = true;
	}

	function closeRejectDialog() {
		if (actionLoading) return;
		rejectDialogOpen = false;
		rejectReason = '';
	}

	function setRejectReason(value) {
		rejectReason = typeof value === 'string' ? value : '';
	}

	async function confirmRejectAction() {
		const reason = asTrimmedString(rejectReason);
		if (!reason || !canReject(request) || actionLoading) return;
		const ok = await postAction({ action: 'reject', reason });
		if (ok) {
			rejectDialogOpen = false;
			rejectReason = '';
			const scope = typeof $currentRoute === 'string' && $currentRoute.startsWith('admin/')
				? 'admin'
				: 'app';
			navigate(`${scope}/requests`);
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
	$: emailValue = asTrimmedString(request?.email);
	$: createdValue = formatDate(request?.created);
	$: facts = FACT_KEYS.map((key) => asTrimmedString(requestData?.[key])).filter(Boolean);
	$: mobileValue = asTrimmedString(requestData?.mobile);
	$: motivationValue = asTrimmedString(requestData?.motivation);
	$: extraEntries = Object.entries(requestData).filter(([key, value]) => !RESERVED_DATA_KEYS.has(key) && hasValue(value));
	$: guardianNotes = nextActionNotes(request);
	$: showGuardianNotes = !isAdmin && guardianNotes !== '';

	$: if (requestId && requestId !== lastRequestId) {
		lastRequestId = requestId;
		void loadAll();
	}

	function goBack() {
		if (groupPickerOpen) {
			closeGroupPicker();
			return;
		}
		window.history.back();
	}

	async function loadAll() {
		if (!requestId) return;
		loading = true;
		error = '';
		actionError = '';
		request = null;
		groupName = '';
		groupMembers = [];
		try {
			request = await fetchRequestById(requestId);
			const currentGroupId = asTrimmedString(request?.group);
			if (currentGroupId) {
				const [groupData, membersData] = await Promise.all([
					pb.collection('groups').getOne(currentGroupId),
					fetchJSONOrThrow(`/api/groups/${encodeURIComponent(currentGroupId)}/members`)
				]);
				groupName = asTrimmedString(groupData?.name);
				groupMembers = Array.isArray(membersData?.items) ? membersData.items : [];
			}
		} catch (err) {
			error = err?.message || err?.toString() || 'Failed to load request';
		} finally {
			loading = false;
		}
	}

	async function loadGroups() {
		if (groupsLoading) return;
		groupsLoading = true;
		groupsError = '';
		try {
			const result = await pb.collection('groups').getList(1, 500, { sort: 'name' });
			groups = Array.isArray(result?.items) ? result.items : [];
		} catch (err) {
			groups = [];
			groupsError = err?.message || err?.toString() || 'Failed to load groups';
		} finally {
			groupsLoading = false;
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
		{#if groupPickerOpen}
			<Card variant="admin">
				<div class="group-picker">
					<h3>Select group</h3>
					{#if groupsError}
						<p class="action-error">{groupsError}</p>
					{:else if groupsLoading}
						<p class="row">Loading groups...</p>
					{:else if groups.length === 0}
						<p class="row">No groups found.</p>
					{:else}
						<div class="stack-list">
							{#each groups as group}
								<GroupCard group={group} isMember={false} onOpen={openAssignDialog} showInviteForDefault={false} />
							{/each}
						</div>
					{/if}
				</div>
			</Card>
		{:else}
			<Card variant="admin">
			<div class="data">
				<p class="name">{requestTitle}</p>
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

				<div class="section">
					<h3>Group</h3>
					<p class="row">{groupName || '-'}</p>
				</div>

				<div class="section">
					<h3>Guardian</h3>
					<p class="row">{guardianDisplayName(request) || 'Not yet assigned'}</p>
					{#if showGuardianNotes}
						<div class="guide markdown" aria-label="Guardian notes text">
							{@html renderContent(guardianNotes)}
						</div>
					{/if}
					{#if canSetGuardian(request) && !guardianDisplayName(request)}
						{#if groupMembers.length === 0}
							<p class="row">No members available.</p>
						{:else}
							<div class="guardian-list">
								{#each groupMembers as member}
									<button
										type="button"
										class="guardian-item"
										on:click={() => openGuardianAssignDialog(member)}
										disabled={actionLoading}
									>
										{displayName(member)}
									</button>
								{/each}
							</div>
						{/if}
					{/if}
				</div>

				{#if canAccept(request) || canReject(request)}
					<div class="section">
						{#if actionError}
							<p class="action-error">{actionError}</p>
						{/if}
						<div class="actions-row" class:single={!canAccept(request) || !canReject(request)}>
							{#if canAccept(request)}
								<button type="button" class="accept-btn" disabled={actionLoading} on:click={handleAcceptAction}>Accept</button>
							{/if}
							{#if canReject(request)}
								<button type="button" class="reject-btn" disabled={actionLoading} on:click={handleRejectAction}>Reject</button>
							{/if}
						</div>
					</div>
				{/if}

			</div>
			</Card>
		{/if}
	{/if}
</DashboardLayout>

<ActionDialog
	show={guardianAssignDialogOpen}
	title={`Assign request to ${displayName(pendingGuardian)} as guardian?`}
	confirmLabel="Assign"
	loading={actionLoading}
	onConfirm={confirmGuardianAssign}
	onCancel={closeGuardianAssignDialog}
/>

<ActionDialog
	show={assignDialogOpen}
	title={`Assign request to ${asTrimmedString(assignTargetGroup?.name) || 'group'}?`}
	confirmLabel="Assign"
	loading={actionLoading}
	onConfirm={confirmAssignGroup}
	onCancel={closeAssignDialog}
/>

<ActionDialog
	show={rejectDialogOpen}
	title="Reject request"
	message="Reject motivation"
	confirmLabel="Reject"
	loading={actionLoading}
	showTextField={true}
	textValue={rejectReason}
	textPlaceholder="Write reject motivation"
	onTextChange={setRejectReason}
	onConfirm={confirmRejectAction}
	onCancel={closeRejectDialog}
/>

<style>
	.group-picker {
		display: grid;
		gap: 0.75rem;
	}

	.data {
		display: grid;
		gap: 0.8rem;
	}

	.name {
		margin: 0;
		font-size: 1.3rem;
		font-weight: 800;
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

	.guide {
		font-size: var(--ui-font-size);
		color: #111;
		line-height: 1.4;
	}

	.guide :global(h1),
	.guide :global(h2),
	.guide :global(h3),
	.guide :global(h4),
	.guide :global(h5),
	.guide :global(h6) {
		margin: 0.35rem 0 0;
		font-size: 0.96rem;
		font-weight: 800;
		line-height: 1.4;
	}

	.guide :global(p),
	.guide :global(ul),
	.guide :global(ol) {
		margin: 0.3rem 0 0;
		font-size: var(--ui-font-size);
		line-height: 1.4;
	}

	.guide :global(li) {
		margin: 0.15rem 0;
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

	.row {
		margin: 0;
		font-size: 0.9rem;
	}

	.guardian-list {
		display: grid;
		gap: 0.45rem;
		padding-top: 0.35rem;
	}

	.guardian-item {
		border: 1px solid #000;
		background: #fff;
		text-align: left;
		padding: 0.55rem 0.65rem;
		font: inherit;
		font-weight: 600;
		cursor: pointer;
	}

	.action-error {
		margin: 0;
		font-size: 0.85rem;
		color: #b00020;
		font-weight: 700;
	}

	.actions-row {
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: 0.6rem;
		margin-top: 2rem;
		justify-content: center;
	}

	.actions-row.single {
		grid-template-columns: minmax(0, 14rem);
		justify-content: center;
	}

	.accept-btn,
	.reject-btn {
		border: 1px solid #000;
		padding: 0.55rem 0.65rem;
		font: inherit;
		font-weight: 700;
		cursor: pointer;
	}

	.accept-btn {
		background: #000;
		color: #fff;
	}

	.reject-btn {
		background: #fff;
		color: #000;
	}

	@media (min-width: 768px) and (max-width: 1024px) {
		.actions-row {
			grid-template-columns: repeat(2, minmax(0, 14rem));
			justify-content: center;
		}
	}
</style>
