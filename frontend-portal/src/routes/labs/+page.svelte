<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib/api';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
	import ErrorBanner from '$lib/components/ErrorBanner.svelte';
	import EmptyState from '$lib/components/EmptyState.svelte';

	interface LabResult {
		id: number;
		test_name: string;
		result_date: string;
		result_value?: string;
		reference_range?: string;
		unit?: string;
		abnormal_flag?: string;
		status?: string;
	}

	let results = $state<LabResult[]>([]);
	let loading = $state(true);
	let error = $state('');

	onMount(() => {
		loadLabs();
	});

	async function loadLabs() {
		error = '';
		loading = true;
		try {
			const data = await api.get<LabResult[]>('/labs');
			results = data || [];
		} catch (e: unknown) {
			error = e instanceof Error ? e.message : 'Failed to load lab results';
		} finally {
			loading = false;
		}
	}
</script>

<svelte:head>
	<title>Lab Results — FreeMED Patient Portal</title>
</svelte:head>

<div>
	<h2 class="fw-bold text-dark mb-4">Lab Results</h2>

	{#if loading}
		<LoadingSpinner message="Loading lab results…" />
	{:else if error}
		<ErrorBanner message={error} onRetry={loadLabs} />
	{:else if results.length === 0}
		<EmptyState
			title="No lab results"
			message="No lab results available at this time."
		/>
	{:else}
		<div class="card">
			<div class="table-responsive">
				<table class="table table-hover mb-0">
					<thead class="table-light">
						<tr>
							<th class="small">Test Name</th>
							<th class="small">Date</th>
							<th class="small">Result</th>
							<th class="small">Reference Range</th>
							<th class="small">Status</th>
						</tr>
					</thead>
					<tbody>
						{#each results as r (r.id)}
							{@const isAbnormal = r.abnormal_flag && r.abnormal_flag !== 'N'}
							<tr class="{isAbnormal ? 'table-warning' : ''}">
								<td class="small fw-medium">{r.test_name}</td>
								<td class="small">{r.result_date}</td>
								<td class="small {isAbnormal ? 'fw-bold text-danger' : ''}">
									{r.result_value || '—'} {r.unit || ''}
									{#if isAbnormal}
										<span class="badge bg-danger ms-1">{r.abnormal_flag}</span>
									{/if}
								</td>
								<td class="small text-muted">{r.reference_range || '—'}</td>
								<td class="small">
									<span class="badge {r.status === 'final' ? 'bg-success' : 'bg-secondary'}">
										{r.status || 'Pending'}
									</span>
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		</div>
	{/if}
</div>
