<script>
	import { onMount } from 'svelte';
	import { apiFetch } from '../lib/pocketbase';
	import { formatEventDate } from '../lib/date';
	import { navigate } from '../lib/router';
	import DashboardLayout from '../components/DashboardLayout.svelte';
	import StatRow from '../components/StatRow.svelte';
	import Card from '../components/Card.svelte';

	let loading = true;
	let error = '';
	let summary = {
		users: { total: null, noTelegram: null, notActive: null },
		groups: { total: null },
		events: { total: null, next: null }
	};

	function goToNextEvent() {
		if (!summary?.events?.next?.id) return;
		navigate(`admin/events?id=${encodeURIComponent(summary.events.next.id)}`);
	}

	async function loadSummary() {
		loading = true;
		error = '';
		try {
			const response = await apiFetch('/api/admin/summary');
			if (!response.ok) {
				throw new Error('Failed to load admin data');
			}
			const data = await response.json();
			summary = {
				users: {
					total: Number(data?.users?.total ?? 0),
					noTelegram: Number(data?.users?.noTelegram ?? 0),
					notActive: Number(data?.users?.notActive ?? 0)
				},
				groups: {
					total: Number(data?.groups?.total ?? 0)
				},
				events: {
					total: Number(data?.events?.total ?? 0),
					next: data?.events?.next || null
				}
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
			<Card variant="state">{error}</Card>
		{:else if loading}
			<Card variant="state">Caricamento...</Card>
		{:else}
			<Card variant="admin">
				<button class="event-card" on:click={goToNextEvent} disabled={!summary.events?.next?.id}>
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
									{ label: 'Da approvare', value: summary.events.next.pending ?? '—' }
								]}
							/>
						</div>
					{/if}
				</button>
			</Card>
		{/if}
	</section>

	<section class="section">
		<Card variant="admin">
			<StatRow
				items={[
					{ label: 'Utenti', value: summary.users.total ?? '—', color: '#000' },
					{ label: 'No Telegram', value: summary.users.noTelegram ?? '—', color: '#1e88e5' },
					{ label: '!Active', value: summary.users.notActive ?? '—', color: '#d32f2f' }
				]}
			/>
		</Card>
	</section>

	<section class="section">
		<div class="two-col">
			<Card variant="admin">
				<StatRow center items={[{ label: 'Gruppi', value: summary.groups.total ?? '—' }]} />
			</Card>
			<Card variant="admin">
				<StatRow center items={[{ label: 'Eventi', value: summary.events.total ?? '—' }]} />
			</Card>
		</div>
	</section>
</DashboardLayout>

<style>
	.section {
		display: grid;
		gap: 0.75rem;
	}

	.event-card {
		border: none;
		background: transparent;
		padding: 0;
		width: 100%;
		text-align: left;
		cursor: pointer;
	}

	.event-card:disabled {
		cursor: default;
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
