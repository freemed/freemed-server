<script lang="ts">
	import { api } from '$lib/api';

	interface AgingRow { age_bucket: string; balance: number; }

	let rows = $state<AgingRow[]>([]);
	let loading = $state(true);
	let error = $state('');

	$effect(() => {
		api.get<AgingRow[]>('/aging').then(d => { rows = d; loading = false; }).catch(e => { error = e.message; loading = false; });
	});
</script>

<div class="bg-white rounded-lg shadow-sm border border-gray-200">
	<div class="px-6 py-4 border-b border-gray-100"><h2 class="text-lg font-semibold text-gray-900">Accounts Receivable Aging</h2></div>
	{#if loading}
		<div class="flex justify-center py-12"><div class="w-6 h-6 border-2 border-blue-500 border-t-transparent rounded-full animate-spin"></div></div>
	{:else if error}
		<p class="p-6 text-red-600 text-sm">{error}</p>
	{:else if rows.length === 0}
		<p class="p-6 text-gray-500 text-sm">No aging data available.</p>
	{:else}
		<table class="w-full text-sm">
			<thead class="bg-gray-50 text-gray-500 uppercase text-xs">
				<tr><th class="px-6 py-3 text-left">Age Bucket</th><th class="px-6 py-3 text-right">Balance</th></tr>
			</thead>
			<tbody>
				{#each rows as row}
					<tr class="border-t border-gray-100">
						<td class="px-6 py-3 font-medium">{row.age_bucket} days</td>
						<td class="px-6 py-3 text-right font-mono">${row.balance.toFixed(2)}</td>
					</tr>
				{/each}
			</tbody>
		</table>
	{/if}
</div>
