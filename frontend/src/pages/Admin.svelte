<script>
	import { onMount } from 'svelte';
	import { pb } from '../lib/pocketbase';
	import DashboardLayout from '../components/DashboardLayout.svelte';
	import StateCard from '../components/StateCard.svelte';
	import AdminCard from '../components/AdminCard.svelte';
	import StatRow from '../components/StatRow.svelte';

	let loading = true;
	let error = '';
	let summary = {
		users: { total: null, noTelegram: null, notActive: null },
		groups: { total: null },
		events: { total: null, next: null },
	};

	function formatDatePart(raw) {
		if (!raw) return '';
		const parts = raw.split('-');
		if (parts.length !== 3) return raw;
		return `${parts[2]}/${parts[1]}/${parts[0]}`;
	}

	function formatEventDate(value) {
		if (!value || typeof value !== 'string') return '';
		const normalized = value.replace('T', ' ');
		const [dateRaw] = normalized.split(' ');
		return formatDatePart(dateRaw);
	}

	async function loadSummary() {
		loading = true;
		error = '';
		try {
			const response = await fetch('/api/admin/summary', {
				headers: {
					Authorization: pb.authStore.token,
				},
			});
			if (!response.ok) {
				throw new Error('Failed to load admin data');
			}
			const data = await response.json();
			summary = {
				users: {
					total: Number(data?.users?.total ?? 0),
					noTelegram: Number(data?.users?.noTelegram ?? 0),
					notActive: Number(data?.users?.notActive ?? 0),
				},
				groups: {
					total: Number(data?.groups?.total ?? 0),
				},
				events: {
					total: Number(data?.events?.total ?? 0),
					next: data?.events?.next || null,
				},
			};
		} catch (err) {
			error = err.message || err.toString() || 'Failed to load admin data';
		} finally {
			loading = false;
		}
	}

	onMount(loadSummary);
</script>

<DashboardLayout title="Admin">
	<section class="section">
		{#if error}
			<StateCard>{error}</StateCard>
		{:else if loading}
			<StateCard>Caricamento...</StateCard>
		{:else}
			<AdminCard>
				<div class="title-row">
					<p class="title">{summary.events?.next?.title || 'Nessun evento'}</p>
					<p class="date">{formatEventDate(summary.events?.next?.event_date)}</p>
				</div>
				{#if summary.events?.next}
					<div class="center-row">
						<StatRow
							center
							items={[
								{ label: 'Iscritti', value: summary.events.next.registrations ?? '—' },
								{ label: 'Da approvare', value: summary.events.next.pending ?? '—' },
							]}
						/>
					</div>
				{/if}
			</AdminCard>
		{/if}
	</section>

	<section class="section">
		<AdminCard>
			<StatRow
				items={[
					{ label: 'Utenti', value: summary.users.total ?? '—', color: '#000' },
					{ label: 'No Telegram', value: summary.users.noTelegram ?? '—', color: '#1e88e5' },
					{ label: '!Active', value: summary.users.notActive ?? '—', color: '#d32f2f' },
				]}
			/>
		</AdminCard>
	</section>

	<section class="section">
		<div class="two-col">
			<AdminCard>
				<StatRow center items={[{ label: 'Gruppi', value: summary.groups.total ?? '—' }]} />
			</AdminCard>
			<AdminCard>
				<StatRow center items={[{ label: 'Eventi', value: summary.events.total ?? '—' }]} />
			</AdminCard>
		</div>
	</section>
</DashboardLayout>

<style>
	.section {
		display: grid;
		gap: 0.75rem;
	}

	.title-row {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 1rem;
	}

	.title {
		margin: 0;
		font-size: clamp(1rem, 3vw, 1.2rem);
		font-weight: 700;
		color: #000;
		flex: 1 1 auto;
	}

	.date {
		margin: 0;
		font-size: 0.9rem;
		font-weight: 600;
		color: #000;
		white-space: nowrap;
	}

	.center-row {
		margin-top: 0.75rem;
	}

	.two-col {
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: clamp(0.75rem, 2.5vw, 1rem);
	}

	@media (max-width: 360px) {
		.two-col {
			grid-template-columns: 1fr;
		}
	}
</style>
