<script lang="ts">
	import { page } from '$app/stores';
	import { api } from '$lib/api';

	interface CoverageInfo {
		id: number;
		patient: number;
		insurance_company: number;
		coverage_type: number;
		policy_number: string;
		group_number: string;
		effective_date: string | null;
		termination_date: string | null;
		primary_coverage: boolean;
	}

	let coverage = $state<CoverageInfo[]>([]);
	let loading = $state(true);
	let error = $state('');
	let patientId = $state('');

	$effect(() => {
		const id = $page.params.id;
		if (id) {
			patientId = id;
			loadCoverage(id);
		}
	});

	async function loadCoverage(id: string) {
		loading = true;
		error = '';
		try {
			const data = await api.get<CoverageInfo | CoverageInfo[]>(`/patient/${id}/coverage-info`);
			// Backend may return single object or array — normalize to array
			coverage = Array.isArray(data) ? data : (data ? [data] : []);
		} catch (e: any) {
			error = e.message || 'Failed to load coverage info';
		} finally {
			loading = false;
		}
	}

	function formatDate(dateStr: string | null): string {
		if (!dateStr) return '—';
		try {
			const d = new Date(dateStr);
			if (isNaN(d.getTime())) return dateStr;
			return d.toLocaleDateString('en-US', {
				year: 'numeric',
				month: 'short',
				day: 'numeric',
			});
		} catch {
			return dateStr;
		}
	}
</script>

<div class="max-w-4xl mx-auto">
	<div class="flex items-center justify-between mb-6">
		<div>
			<h1 class="text-2xl font-bold text-gray-900">Coverage Info</h1>
			<p class="text-sm text-gray-500 mt-1">Patient #{patientId}</p>
		</div>
		<a
			href="/patients/{patientId}"
			class="text-sm text-blue-600 hover:text-blue-800 font-medium transition-colors"
		>
			&larr; Back to Patient
		</a>
	</div>

	{#if error}
		<div class="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg text-sm mb-6">
			{error}
		</div>
	{/if}

	{#if loading}
		<div class="flex justify-center py-12">
			<div class="w-6 h-6 border-2 border-blue-500 border-t-transparent rounded-full animate-spin"></div>
		</div>
	{:else if coverage.length === 0}
		<div class="text-center py-12 text-gray-500 bg-white rounded-lg shadow-sm border border-gray-200">
			<p class="text-lg">No coverage info available</p>
			<p class="text-sm mt-1">No insurance coverage has been recorded for this patient yet.</p>
		</div>
	{:else}
		<div class="bg-white rounded-lg shadow-sm border border-gray-200 overflow-x-auto">
			<table class="min-w-full divide-y divide-gray-200">
				<thead class="bg-gray-50">
					<tr>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Insurance Company</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Coverage Type</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Policy #</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Group #</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Effective Date</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Termination Date</th>
					</tr>
				</thead>
				<tbody class="bg-white divide-y divide-gray-200">
					{#each coverage as cov}
						<tr class="hover:bg-gray-50">
							<td class="px-4 py-3 text-sm text-gray-700 whitespace-nowrap">
								#{cov.insurance_company}
							</td>
							<td class="px-4 py-3 text-sm text-gray-700 whitespace-nowrap">
								#{cov.coverage_type}
							</td>
							<td class="px-4 py-3 text-sm text-gray-700">
								{cov.policy_number || '—'}
							</td>
							<td class="px-4 py-3 text-sm text-gray-700">
								{cov.group_number || '—'}
							</td>
							<td class="px-4 py-3 text-sm text-gray-700 whitespace-nowrap">
								{formatDate(cov.effective_date)}
							</td>
							<td class="px-4 py-3 text-sm text-gray-700 whitespace-nowrap">
								{formatDate(cov.termination_date)}
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{/if}
</div>
