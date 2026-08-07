<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib/api';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
	import ErrorBanner from '$lib/components/ErrorBanner.svelte';
	import EmptyState from '$lib/components/EmptyState.svelte';

	interface Vital {
		id: number;
		vital_date: string;
		vital_type: string;
		value: string;
		unit?: string;
	}

	interface Problem {
		id: number;
		code?: string;
		description: string;
		onset_date?: string;
		status?: string;
	}

	interface Allergy {
		id: number;
		allergen: string;
		reaction?: string;
		severity?: string;
	}

	interface Medication {
		id: number;
		medication_name: string;
		dosage?: string;
		frequency?: string;
		start_date?: string;
		end_date?: string;
		status?: string;
	}

	let activeTab = $state('vitals');
	let vitals = $state<Vital[]>([]);
	let problems = $state<Problem[]>([]);
	let allergies = $state<Allergy[]>([]);
	let medications = $state<Medication[]>([]);
	let loading = $state(true);
	let error = $state('');

	const tabs = [
		{ id: 'vitals', label: 'Vitals' },
		{ id: 'problems', label: 'Problems' },
		{ id: 'allergies', label: 'Allergies' },
		{ id: 'medications', label: 'Medications' },
	];

	onMount(() => {
		loadHealthData();
	});

	async function loadHealthData() {
		error = '';
		loading = true;
		try {
			const [v, p, a, m] = await Promise.all([
				api.get<Vital[]>('/vitals').catch(() => [] as Vital[]),
				api.get<Problem[]>('/problems').catch(() => [] as Problem[]),
				api.get<Allergy[]>('/allergies').catch(() => [] as Allergy[]),
				api.get<Medication[]>('/medications').catch(() => [] as Medication[]),
			]);
			vitals = v || [];
			problems = p || [];
			allergies = a || [];
			medications = m || [];
		} catch (e: unknown) {
			error = e instanceof Error ? e.message : 'Failed to load health data';
		} finally {
			loading = false;
		}
	}
</script>

<svelte:head>
	<title>Health — FreeMED Patient Portal</title>
</svelte:head>

<div>
	<h2 class="fw-bold text-dark mb-4">Health</h2>

	{#if loading}
		<LoadingSpinner message="Loading health data…" />
	{:else if error}
		<ErrorBanner message={error} onRetry={loadHealthData} />
	{:else}
		<ul class="nav nav-tabs mb-4">
			{#each tabs as tab}
				<li class="nav-item">
					<button
						class="nav-link {activeTab === tab.id ? 'active' : ''}"
						onclick={() => activeTab = tab.id}
					>
						{tab.label}
					</button>
				</li>
			{/each}
		</ul>

		<!-- Vitals Tab -->
		{#if activeTab === 'vitals'}
			{#if vitals.length === 0}
				<EmptyState title="No vitals" message="No vitals on file." />
			{:else}
				<div class="card">
					<div class="table-responsive">
						<table class="table table-hover mb-0">
							<thead class="table-light">
								<tr>
									<th class="small">Date</th>
									<th class="small">Type</th>
									<th class="small">Value</th>
								</tr>
							</thead>
							<tbody>
								{#each vitals as v (v.id)}
									<tr>
										<td class="small">{v.vital_date}</td>
										<td class="small">{v.vital_type}</td>
										<td class="small">{v.value} {v.unit || ''}</td>
									</tr>
								{/each}
							</tbody>
						</table>
					</div>
				</div>
			{/if}
		{/if}

		<!-- Problems Tab -->
		{#if activeTab === 'problems'}
			{#if problems.length === 0}
				<EmptyState title="No problems" message="No active problems on file." />
			{:else}
				<div class="card">
					<div class="table-responsive">
						<table class="table table-hover mb-0">
							<thead class="table-light">
								<tr>
									<th class="small">Description</th>
									<th class="small">Onset</th>
									<th class="small">Status</th>
								</tr>
							</thead>
							<tbody>
								{#each problems as p (p.id)}
									<tr>
										<td class="small">{p.description}</td>
										<td class="small">{p.onset_date || '—'}</td>
										<td class="small">
											<span class="badge {p.status === 'active' ? 'bg-danger' : 'bg-secondary'}">
												{p.status || 'Unknown'}
											</span>
										</td>
									</tr>
								{/each}
							</tbody>
						</table>
					</div>
				</div>
			{/if}
		{/if}

		<!-- Allergies Tab -->
		{#if activeTab === 'allergies'}
			{#if allergies.length === 0}
				<EmptyState title="No allergies" message="No allergies on file." />
			{:else}
				<div class="card">
					<div class="table-responsive">
						<table class="table table-hover mb-0">
							<thead class="table-light">
								<tr>
									<th class="small">Allergen</th>
									<th class="small">Reaction</th>
									<th class="small">Severity</th>
								</tr>
							</thead>
							<tbody>
								{#each allergies as a (a.id)}
									<tr>
										<td class="small fw-medium text-danger">{a.allergen}</td>
										<td class="small">{a.reaction || '—'}</td>
										<td class="small">
											{#if a.severity}
												<span class="badge {a.severity === 'severe' ? 'bg-danger' : a.severity === 'moderate' ? 'bg-warning text-dark' : 'bg-info'}">
													{a.severity}
												</span>
											{:else}
												—
											{/if}
										</td>
									</tr>
								{/each}
							</tbody>
						</table>
					</div>
				</div>
			{/if}
		{/if}

		<!-- Medications Tab -->
		{#if activeTab === 'medications'}
			{#if medications.length === 0}
				<EmptyState title="No medications" message="No medications on file." />
			{:else}
				<div class="card">
					<div class="table-responsive">
						<table class="table table-hover mb-0">
							<thead class="table-light">
								<tr>
									<th class="small">Medication</th>
									<th class="small">Dosage</th>
									<th class="small">Frequency</th>
									<th class="small">Dates</th>
									<th class="small">Status</th>
								</tr>
							</thead>
							<tbody>
								{#each medications as m (m.id)}
									<tr>
										<td class="small fw-medium">{m.medication_name}</td>
										<td class="small">{m.dosage || '—'}</td>
										<td class="small">{m.frequency || '—'}</td>
										<td class="small">
											{m.start_date || '—'} — {m.end_date || 'ongoing'}
										</td>
										<td class="small">
											<span class="badge {m.status === 'active' ? 'bg-success' : 'bg-secondary'}">
												{m.status || 'Unknown'}
											</span>
										</td>
									</tr>
								{/each}
							</tbody>
						</table>
					</div>
				</div>
			{/if}
		{/if}
	{/if}
</div>
