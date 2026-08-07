<script lang="ts">
	import { api } from '$lib/api';

	interface ClaimLogEntry {
		id: number;
		cltimestamp: string;
		cluser: number;
		clprocedure: number;
		claction: string;
		clcomment: string;
		clformat: string;
		cltarget: string;
	}

	let entries = $state<ClaimLogEntry[]>([]);
	let loading = $state(false);
	let error = $state('');
	let claimId = $state('');
	let searched = $state(false);

	async function search() {
		const id = parseInt(claimId, 10);
		if (!id || id <= 0) {
			error = 'Please enter a valid claim ID.';
			return;
		}
		loading = true;
		error = '';
		searched = true;
		try {
			entries = await api.get<ClaimLogEntry[]>(`/claimlog?claim=${id}`);
		} catch (e: any) {
			error = e.message || 'Failed to load claim log entries.';
			entries = [];
		} finally {
			loading = false;
		}
	}

	function clearSearch() {
		claimId = '';
		entries = [];
		error = '';
		searched = false;
	}

	function formatTimestamp(ts: string): string {
		if (!ts) return '—';
		return new Date(ts).toLocaleString();
	}
</script>

<div class="bg-white rounded-lg shadow-sm border border-gray-200">
	<div class="px-6 py-4 border-b border-gray-100">
		<h2 class="text-lg font-semibold text-gray-900">Claim Log</h2>
	</div>

	<!-- Claim ID filter -->
	<div class="px-6 py-4 border-b border-gray-100 bg-gray-50">
		<div class="flex items-center gap-3">
			<label for="claim-id" class="text-sm font-medium text-gray-700 whitespace-nowrap">Claim ID</label>
			<input
				id="claim-id"
				type="number"
				bind:value={claimId}
				onkeydown={(e) => { if (e.key === 'Enter') search(); }}
				class="w-32 border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-1 focus:ring-blue-500"
				placeholder="e.g. 42"
			/>
			<button
				onclick={search}
				class="px-4 py-2 text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 rounded-md transition-colors"
			>
				Search
			</button>
			{#if searched}
				<button
					onclick={clearSearch}
					class="px-3 py-1 text-sm text-gray-600 hover:bg-gray-200 rounded-md transition-colors"
				>
					Clear
				</button>
			{/if}
		</div>
	</div>

	{#if loading}
		<div class="flex justify-center py-12">
			<div class="w-6 h-6 border-2 border-blue-500 border-t-transparent rounded-full animate-spin"></div>
		</div>
	{:else if error}
		<p class="p-6 text-red-600 text-sm">{error}</p>
	{:else if searched && entries.length === 0}
		<p class="p-6 text-gray-500 text-sm">No claim log entries found for claim #{claimId}.</p>
	{:else if !searched}
		<p class="p-6 text-gray-500 text-sm">Enter a claim ID to view its log entries.</p>
	{:else}
		<div class="overflow-x-auto">
			<table class="w-full text-sm">
				<thead class="bg-gray-50 text-gray-500 uppercase text-xs">
					<tr>
						<th class="px-6 py-3 text-left">Timestamp</th>
						<th class="px-6 py-3 text-left">User</th>
						<th class="px-6 py-3 text-left">Procedure</th>
						<th class="px-6 py-3 text-left">Action</th>
						<th class="px-6 py-3 text-left">Comment</th>
						<th class="px-6 py-3 text-left">Format</th>
						<th class="px-6 py-3 text-left">Target</th>
					</tr>
				</thead>
				<tbody>
					{#each entries as entry}
						<tr class="border-t border-gray-100 hover:bg-gray-50">
							<td class="px-6 py-3 whitespace-nowrap">{formatTimestamp(entry.cltimestamp)}</td>
							<td class="px-6 py-3 font-mono text-xs">{entry.cluser}</td>
							<td class="px-6 py-3 font-mono text-xs">{entry.clprocedure}</td>
							<td class="px-6 py-3">{entry.claction || '—'}</td>
							<td class="px-6 py-3 max-w-xs truncate" title={entry.clcomment}>{entry.clcomment || '—'}</td>
							<td class="px-6 py-3">{entry.clformat || '—'}</td>
							<td class="px-6 py-3">{entry.cltarget || '—'}</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{/if}
</div>
