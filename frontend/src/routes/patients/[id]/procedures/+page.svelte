<script lang="ts">
	import { page } from '$app/stores';
	import { api } from '$lib/api';

	interface Procedure {
		id: number;
		date_of_service: string;
		cpt_id: number;
		cpt_name: string | null;
		charge: number;
		balance: number;
		amount_paid: number;
		status: string | null;
		diagnosis_id: number;
		place_of_service: number;
		provider_id: number;
		units: number;
		diagnosis_set: string;
		balance_original: number;
		date_of_service_end: string | null;
		cpt_modifier_1: number;
		cpt_modifier_2: number;
		cpt_modifier_3: number;
	}

	let procedures = $state<Procedure[]>([]);
	let loading = $state(true);
	let error = $state('');
	let patientId = $state('');

	$effect(() => {
		const id = $page.params.id;
		if (id) {
			patientId = id;
			loadProcedures(id);
		}
	});

	async function loadProcedures(id: string) {
		loading = true;
		error = '';
		try {
			const data = await api.get<Procedure[]>(`/patient/${id}/procedures`);
			procedures = data || [];
		} catch (e: any) {
			error = e.message || 'Failed to load procedures';
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

	function formatCurrency(value: number): string {
		return value.toLocaleString('en-US', { style: 'currency', currency: 'USD' });
	}

	function statusBadgeClass(s: string | null): string {
		if (!s) return 'bg-gray-50 text-gray-600';
		switch (s.toLowerCase()) {
			case 'billed':
				return 'bg-green-50 text-green-700';
			case 'pending':
				return 'bg-yellow-50 text-yellow-700';
			case 'paid':
				return 'bg-blue-50 text-blue-700';
			case 'denied':
				return 'bg-red-50 text-red-700';
			default:
				return 'bg-gray-50 text-gray-600';
		}
	}
</script>

<div class="max-w-5xl mx-auto">
	<div class="flex items-center justify-between mb-6">
		<div>
			<h1 class="text-2xl font-bold text-gray-900">Procedures</h1>
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
	{:else if procedures.length === 0}
		<div class="text-center py-12 text-gray-500 bg-white rounded-lg shadow-sm border border-gray-200">
			<p class="text-lg">No procedures found</p>
			<p class="text-sm mt-1">No procedures have been recorded for this patient yet.</p>
		</div>
	{:else}
		<div class="bg-white rounded-lg shadow-sm border border-gray-200 overflow-x-auto">
			<table class="min-w-full divide-y divide-gray-200">
				<thead class="bg-gray-50">
					<tr>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Date</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">CPT Code</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Description</th>
						<th class="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Charges</th>
						<th class="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Units</th>
						<th class="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Balance</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Provider</th>
					</tr>
				</thead>
				<tbody class="bg-white divide-y divide-gray-200">
					{#each procedures as proc}
						<tr class="hover:bg-gray-50">
							<td class="px-4 py-3 text-sm text-gray-700 whitespace-nowrap">
								{formatDate(proc.date_of_service)}
							</td>
							<td class="px-4 py-3 text-sm text-gray-500 whitespace-nowrap font-mono">
								#{proc.cpt_id}
							</td>
							<td class="px-4 py-3 text-sm text-gray-900">
								{proc.cpt_name || '—'}
							</td>
							<td class="px-4 py-3 text-sm text-gray-700 text-right whitespace-nowrap">
								{formatCurrency(proc.charge)}
							</td>
							<td class="px-4 py-3 text-sm text-gray-700 text-right whitespace-nowrap">
								{proc.units}
							</td>
							<td class="px-4 py-3 text-sm text-gray-700 text-right whitespace-nowrap">
								{formatCurrency(proc.balance)}
							</td>
							<td class="px-4 py-3 text-sm whitespace-nowrap">
								<span class="inline-block px-2 py-0.5 text-xs font-medium rounded {statusBadgeClass(proc.status)}">
									{proc.status || 'unknown'}
								</span>
							</td>
							<td class="px-4 py-3 text-sm text-gray-500 whitespace-nowrap">
								#{proc.provider_id}
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{/if}
</div>
