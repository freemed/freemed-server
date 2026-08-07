<script lang="ts">
	import { page } from '$app/stores';
	import { api } from '$lib/api';

	interface Claim {
		id: number;
		cltimestamp: string;
		claction: string;
		cpt_code: number | null;
		diagnosis_code: number | null;
		charges: number | null;
		username: string | null;
	}

	let claims = $state<Claim[]>([]);
	let loading = $state(true);
	let error = $state('');
	let patientId = $state('');

	$effect(() => {
		const id = $page.params.id;
		if (id) {
			patientId = id;
			loadClaims(id);
		}
	});

	async function loadClaims(id: string) {
		loading = true;
		error = '';
		try {
			const data = await api.get<Claim[]>(`/patient/${id}/claims`);
			claims = data || [];
		} catch (e: any) {
			error = e.message || 'Failed to load claims';
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

	function formatCurrency(value: number | null): string {
		if (value == null) return '—';
		return value.toLocaleString('en-US', { style: 'currency', currency: 'USD' });
	}

	function statusBadgeClass(s: string): string {
		switch (s.toLowerCase()) {
			case 'pending':
				return 'bg-yellow-50 text-yellow-700';
			case 'submitted':
				return 'bg-blue-50 text-blue-700';
			case 'paid':
				return 'bg-green-50 text-green-700';
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
			<h1 class="text-2xl font-bold text-gray-900">Claims</h1>
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
	{:else if claims.length === 0}
		<div class="text-center py-12 text-gray-500 bg-white rounded-lg shadow-sm border border-gray-200">
			<p class="text-lg">No claims found</p>
			<p class="text-sm mt-1">No claims have been submitted for this patient yet.</p>
		</div>
	{:else}
		<div class="bg-white rounded-lg shadow-sm border border-gray-200 overflow-x-auto">
			<table class="min-w-full divide-y divide-gray-200">
				<thead class="bg-gray-50">
					<tr>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Claim Date</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">CPT</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Diagnosis</th>
						<th class="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Charges</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
					</tr>
				</thead>
				<tbody class="bg-white divide-y divide-gray-200">
					{#each claims as claim (claim.id)}
						<tr class="hover:bg-gray-50">
							<td class="px-4 py-3 text-sm text-gray-700 whitespace-nowrap">
								{formatDate(claim.cltimestamp)}
							</td>
							<td class="px-4 py-3 text-sm text-gray-500 whitespace-nowrap font-mono">
								{claim.cpt_code != null ? '#' + claim.cpt_code : '—'}
							</td>
							<td class="px-4 py-3 text-sm text-gray-500 whitespace-nowrap font-mono">
								{claim.diagnosis_code != null ? '#' + claim.diagnosis_code : '—'}
							</td>
							<td class="px-4 py-3 text-sm text-gray-700 text-right whitespace-nowrap">
								{formatCurrency(claim.charges)}
							</td>
							<td class="px-4 py-3 text-sm whitespace-nowrap">
								<span class="inline-block px-2 py-0.5 text-xs font-medium rounded {statusBadgeClass(claim.claction)}">
									{claim.claction || 'unknown'}
								</span>
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{/if}
</div>
