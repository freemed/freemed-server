<script lang="ts">
	import { page } from '$app/stores';
	import { api } from '$lib/api';

	interface Lab {
		id: number;
		patient: number;
		lab_name: string;
		lab_date: string;
		result: string;
		unit: string;
		reference_range: string;
		status: string;
		notes: string | null;
		active: string;
		created_at: string;
	}

	let labs = $state<Lab[]>([]);
	let loading = $state(true);
	let error = $state('');
	let patientId = $state('');

	$effect(() => {
		const id = $page.params.id;
		if (id) {
			patientId = id;
			loadLabs(id);
		}
	});

	async function loadLabs(id: string) {
		loading = true;
		error = '';
		try {
			const data = await api.get<Lab[]>(`/patient/${id}/labs`);
			labs = data || [];
		} catch (e: any) {
			error = e.message || 'Failed to load labs';
		} finally {
			loading = false;
		}
	}

	function formatDate(dateStr: string): string {
		if (!dateStr) return '—';
		try {
			const d = new Date(dateStr);
			if (isNaN(d.getTime())) return dateStr;
			return d.toLocaleDateString('en-US', { year: 'numeric', month: 'short', day: 'numeric' });
		} catch {
			return dateStr;
		}
	}
</script>

<div class="max-w-4xl mx-auto">
	<div class="flex items-center justify-between mb-6">
		<div>
			<h1 class="text-2xl font-bold text-gray-900">Lab Results</h1>
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
	{:else if labs.length === 0}
		<div class="text-center py-12 text-gray-500 bg-white rounded-lg shadow-sm border border-gray-200">
			<p class="text-lg">No lab results found</p>
			<p class="text-sm mt-1">Lab results from external lab systems will appear here.</p>
		</div>
	{:else}
		<div class="bg-white rounded-lg shadow-sm border border-gray-200 overflow-x-auto">
			<table class="min-w-full divide-y divide-gray-200">
				<thead class="bg-gray-50">
					<tr>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Lab Name</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Date Ordered</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Result Value</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Reference Range</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
					</tr>
				</thead>
				<tbody class="bg-white divide-y divide-gray-200">
					{#each labs as lab}
						<tr class="hover:bg-gray-50">
							<td class="px-4 py-3 text-sm font-medium text-gray-900">{lab.lab_name}</td>
							<td class="px-4 py-3 text-sm text-gray-600">{formatDate(lab.lab_date)}</td>
							<td class="px-4 py-3 text-sm text-gray-600">
								{lab.result || '—'}
								{#if lab.unit}
									<span class="text-gray-400 ml-1">{lab.unit}</span>
								{/if}
							</td>
							<td class="px-4 py-3 text-sm text-gray-600">{lab.reference_range || '—'}</td>
							<td class="px-4 py-3 text-sm">
								{#if lab.status === 'final'}
									<span class="inline-block px-2 py-0.5 text-xs font-medium bg-green-50 text-green-700 rounded">Final</span>
								{:else if lab.status === 'preliminary'}
									<span class="inline-block px-2 py-0.5 text-xs font-medium bg-yellow-50 text-yellow-700 rounded">Preliminary</span>
								{:else if lab.status === 'corrected'}
									<span class="inline-block px-2 py-0.5 text-xs font-medium bg-blue-50 text-blue-700 rounded">Corrected</span>
								{:else}
									<span class="inline-block px-2 py-0.5 text-xs font-medium bg-gray-50 text-gray-500 rounded">{lab.status || 'Unknown'}</span>
								{/if}
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{/if}
</div>
