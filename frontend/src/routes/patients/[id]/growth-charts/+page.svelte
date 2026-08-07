<script lang="ts">
	import { page } from '$app/stores';
	import { api } from '$lib/api';

	interface GrowthChart {
		id: number;
		record_date: string;
		age_months: string | null;
		height_cm: string | null;
		weight_kg: string | null;
		head_circumference_cm: string | null;
		bmi: string | null;
		notes: string;
	}

	let growthCharts = $state<GrowthChart[]>([]);
	let loading = $state(true);
	let error = $state('');
	let patientId = $state('');

	$effect(() => {
		const id = $page.params.id;
		if (id) {
			patientId = id;
			loadGrowthCharts(id);
		}
	});

	async function loadGrowthCharts(id: string) {
		loading = true;
		error = '';
		try {
			const data = await api.get<GrowthChart[]>(`/patient/${id}/growth-charts`);
			growthCharts = data || [];
		} catch (e: any) {
			error = e.message || 'Failed to load growth charts';
		} finally {
			loading = false;
		}
	}

	function formatDate(dateStr: string): string {
		if (!dateStr) return '';
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

	function formatValue(val: string | null, unit: string = ''): string {
		if (val === null || val === undefined || val === '') return '—';
		return `${val}${unit}`;
	}
</script>

<div class="max-w-4xl mx-auto">
	<div class="flex items-center justify-between mb-6">
		<div>
			<h1 class="text-2xl font-bold text-gray-900">Growth Charts</h1>
			<p class="text-sm text-gray-500 mt-1">Patient #{patientId}</p>
		</div>
		<div class="flex items-center gap-3">
			<a
				href="/patients/{patientId}"
				class="text-sm text-blue-600 hover:text-blue-800 font-medium transition-colors"
			>
				&larr; Back to Patient
			</a>
		</div>
	</div>

	{#if loading}
		<div class="flex justify-center py-12">
			<div class="w-6 h-6 border-2 border-blue-500 border-t-transparent rounded-full animate-spin"></div>
		</div>
	{:else if error}
		<div class="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg text-sm mb-6">
			{error}
		</div>
	{:else if growthCharts.length === 0}
		<div class="text-center py-12 text-gray-500">
			<p class="text-lg">No growth charts recorded</p>
			<p class="text-sm mt-1">No growth chart data has been recorded for this patient yet.</p>
		</div>
	{:else}
		<div class="bg-white rounded-lg shadow-sm border border-gray-200 overflow-hidden">
			<div class="overflow-x-auto">
				<table class="min-w-full divide-y divide-gray-200">
					<thead class="bg-gray-50">
						<tr>
							<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Date</th>
							<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Age (mo)</th>
							<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Height</th>
							<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Weight</th>
							<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">BMI</th>
							<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Head Circ.</th>
							<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Notes</th>
						</tr>
					</thead>
					<tbody class="bg-white divide-y divide-gray-200">
						{#each growthCharts as gc, i}
							<tr class="hover:bg-gray-50 transition-colors">
								<td class="px-4 py-3 text-sm text-gray-700 whitespace-nowrap">
									{formatDate(gc.record_date)}
									{#if i === 0}
										<span class="ml-2 inline-block px-1.5 py-0.5 text-xs font-medium bg-blue-100 text-blue-700 rounded">Latest</span>
									{/if}
								</td>
								<td class="px-4 py-3 text-sm text-gray-700 whitespace-nowrap">
									{formatValue(gc.age_months)}
								</td>
								<td class="px-4 py-3 text-sm text-gray-700 whitespace-nowrap">
									{formatValue(gc.height_cm, ' cm')}
								</td>
								<td class="px-4 py-3 text-sm text-gray-700 whitespace-nowrap">
									{formatValue(gc.weight_kg, ' kg')}
								</td>
								<td class="px-4 py-3 text-sm text-gray-700 whitespace-nowrap">
									{formatValue(gc.bmi)}
								</td>
								<td class="px-4 py-3 text-sm text-gray-700 whitespace-nowrap">
									{formatValue(gc.head_circumference_cm, ' cm')}
								</td>
								<td class="px-4 py-3 text-sm text-gray-500 max-w-[200px] truncate">
									{#if gc.notes}
										<span title={gc.notes}>{gc.notes.slice(0, 50)}{gc.notes.length > 50 ? '...' : ''}</span>
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
</div>
