<script lang="ts">
	import { page } from '$app/stores';
	import { api } from '$lib/api';

	interface Annotation {
		id: number;
		created_at: string;
		updated_at: string;
		deleted_at: string | null;
		atimestamp: string;
		apatient: number;
		amodule: string;
		atable: string;
		aid: number;
		auser: number;
		annotation: string;
	}

	let annotations = $state<Annotation[]>([]);
	let loading = $state(true);
	let error = $state('');
	let patientId = $state('');

	$effect(() => {
		const id = $page.params.id;
		if (id) {
			patientId = id;
			loadAnnotations(id);
		}
	});

	async function loadAnnotations(id: string) {
		loading = true;
		error = '';
		try {
			const data = await api.get<Annotation[]>(`/patient/${id}/annotations`);
			annotations = data || [];
		} catch (e: any) {
			error = e.message || 'Failed to load annotations';
		} finally {
			loading = false;
		}
	}

	function formatDate(dateStr: string): string {
		if (!dateStr) return '—';
		try {
			const d = new Date(dateStr);
			if (isNaN(d.getTime())) return dateStr;
			return d.toLocaleDateString('en-US', {
				year: 'numeric',
				month: 'short',
				day: 'numeric',
				hour: '2-digit',
				minute: '2-digit',
			});
		} catch {
			return dateStr;
		}
	}
</script>

<div class="max-w-4xl mx-auto">
	<div class="flex items-center justify-between mb-6">
		<div>
			<h1 class="text-2xl font-bold text-gray-900">Annotations</h1>
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
	{:else if annotations.length === 0}
		<div class="text-center py-12 text-gray-500 bg-white rounded-lg shadow-sm border border-gray-200">
			<p class="text-lg">No annotations</p>
			<p class="text-sm mt-1">No annotations have been recorded for this patient yet.</p>
		</div>
	{:else}
		<div class="bg-white rounded-lg shadow-sm border border-gray-200 overflow-x-auto">
			<table class="min-w-full divide-y divide-gray-200">
				<thead class="bg-gray-50">
					<tr>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Timestamp</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Module</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Table</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Annotation</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">User</th>
					</tr>
				</thead>
				<tbody class="bg-white divide-y divide-gray-200">
					{#each annotations as note}
						<tr class="hover:bg-gray-50">
							<td class="px-4 py-3 text-sm text-gray-600 whitespace-nowrap">
								{formatDate(note.atimestamp)}
							</td>
							<td class="px-4 py-3 text-sm text-gray-700 whitespace-nowrap">
								{note.amodule || '—'}
							</td>
							<td class="px-4 py-3 text-sm text-gray-700 whitespace-nowrap">
								{note.atable || '—'}
							</td>
							<td class="px-4 py-3 text-sm text-gray-900 max-w-md">
								{note.annotation || '—'}
							</td>
							<td class="px-4 py-3 text-sm text-gray-600 whitespace-nowrap">
								#{note.auser}
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{/if}
</div>
