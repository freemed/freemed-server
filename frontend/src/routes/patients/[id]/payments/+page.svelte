<script lang="ts">
	import { page } from '$app/stores';
	import { api } from '$lib/api';

	interface Payment {
		id: number;
		date: string | null;
		amount: number;
		type: number;
		description: string;
		procedure_id: number;
		source: number;
		reference_number: string | null;
	}

	let payments = $state<Payment[]>([]);
	let loading = $state(true);
	let error = $state('');
	let patientId = $state('');

	$effect(() => {
		const id = $page.params.id;
		if (id) {
			patientId = id;
			loadPayments(id);
		}
	});

	async function loadPayments(id: string) {
		loading = true;
		error = '';
		try {
			const data = await api.get<Payment[]>(`/patient/${id}/payments`);
			payments = data || [];
		} catch (e: any) {
			error = e.message || 'Failed to load payments';
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

	function formatCurrency(amount: number): string {
		return amount.toLocaleString('en-US', {
			style: 'currency',
			currency: 'USD',
			minimumFractionDigits: 2,
		});
	}

	function formatType(type: number): string {
		const types: Record<number, string> = {
			0: 'Payment',
			1: 'Copay',
			2: 'Deductible',
			3: 'Coinsurance',
			4: 'Write-off',
			5: 'Withhold',
			6: 'Refund',
		};
		return types[type] ?? `Type ${type}`;
	}
</script>

<div class="max-w-4xl mx-auto">
	<div class="flex items-center justify-between mb-6">
		<div>
			<h1 class="text-2xl font-bold text-gray-900">Payments</h1>
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

	{#if error}
		<div class="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg text-sm mb-6">
			{error}
		</div>
	{/if}

	{#if loading}
		<div class="flex justify-center py-12">
			<div class="w-6 h-6 border-2 border-blue-500 border-t-transparent rounded-full animate-spin"></div>
		</div>
	{:else if payments.length === 0}
		<div class="text-center py-12 text-gray-500 bg-white rounded-lg shadow-sm border border-gray-200">
			<p class="text-lg">No payments recorded</p>
			<p class="text-sm mt-1">No payments have been recorded for this patient yet.</p>
		</div>
	{:else}
		<div class="bg-white rounded-lg shadow-sm border border-gray-200 overflow-x-auto">
			<table class="min-w-full divide-y divide-gray-200">
				<thead class="bg-gray-50">
					<tr>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Date</th>
						<th class="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Amount</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Type</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Description</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Source</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Reference #</th>
					</tr>
				</thead>
				<tbody class="bg-white divide-y divide-gray-200">
					{#each payments as payment (payment.id)}
						<tr class="hover:bg-gray-50">
							<td class="px-4 py-3 text-sm text-gray-900 whitespace-nowrap">
								{formatDate(payment.date)}
							</td>
							<td class="px-4 py-3 text-sm text-gray-900 text-right font-mono whitespace-nowrap">
								{formatCurrency(payment.amount)}
							</td>
							<td class="px-4 py-3 text-sm text-gray-700">
								{formatType(payment.type)}
							</td>
							<td class="px-4 py-3 text-sm text-gray-700 max-w-xs truncate">
								{payment.description || '—'}
							</td>
							<td class="px-4 py-3 text-sm text-gray-700">
								{payment.source || '—'}
							</td>
							<td class="px-4 py-3 text-sm text-gray-700 font-mono">
								{payment.reference_number || '—'}
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{/if}
</div>
