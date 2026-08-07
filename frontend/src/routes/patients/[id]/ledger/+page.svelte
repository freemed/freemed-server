<script lang="ts">
	import { page } from '$app/stores';
	import { api } from '$lib/api';

	interface LedgerEntry {
		entry_type: string;
		id: number;
		date: string;
		amount: number;
		balance: number;
		cpt_code: string | null;
		description: string | null;
		status: string | null;
	}

	interface LedgerResponse {
		data: LedgerEntry[];
		total: number;
		offset: number;
		limit: number;
	}

	let patientId = $derived($page.params.id || '');

	let entries = $state<LedgerEntry[]>([]);
	let total = $state(0);
	let offset = $state(0);
	let limit = $state(50);
	let loading = $state(true);
	let error = $state('');

	// Date range filters
	let fromDate = $state('');
	let toDate = $state('');

	$effect(() => {
		if (patientId) loadEntries();
	});

	async function loadEntries(resetOffset = true) {
		loading = true;
		error = '';
		const currentOffset = resetOffset ? 0 : offset;
		try {
			const params = new URLSearchParams();
			params.set('offset', String(currentOffset));
			params.set('limit', String(limit));
			if (fromDate) params.set('from', fromDate);
			if (toDate) params.set('to', toDate);

			const data = await api.get<LedgerResponse>(
				`/patient/${patientId}/ledger?${params.toString()}`
			);
			entries = data.data || [];
			total = data.total;
			offset = data.offset;
			limit = data.limit;
			if (resetOffset) offset = 0;
		} catch (e: any) {
			error = e.message || 'Failed to load ledger entries';
		} finally {
			loading = false;
		}
	}

	function applyFilter() {
		loadEntries(true);
	}

	function clearFilter() {
		fromDate = '';
		toDate = '';
		loadEntries(true);
	}

	function prevPage() {
		if (offset > 0) {
			const newOffset = Math.max(0, offset - limit);
			offset = newOffset;
			loadEntries(false);
		}
	}

	function nextPage() {
		if (offset + limit < total) {
			const newOffset = offset + limit;
			offset = newOffset;
			loadEntries(false);
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
			});
		} catch {
			return dateStr;
		}
	}

	function formatCurrency(val: number): string {
		return new Intl.NumberFormat('en-US', {
			style: 'currency',
			currency: 'USD',
			minimumFractionDigits: 2,
			maximumFractionDigits: 2,
		}).format(val);
	}

	function entryTypeBadge(type: string): { label: string; classes: string } {
		switch (type) {
			case 'charge':
				return { label: 'Charge', classes: 'bg-blue-50 text-blue-700' };
			case 'payment':
				return { label: 'Payment', classes: 'bg-green-50 text-green-700' };
			case 'adjustment':
				return { label: 'Adjustment', classes: 'bg-yellow-50 text-yellow-700' };
			default:
				return { label: type, classes: 'bg-gray-50 text-gray-600' };
		}
	}

	let fromPage = $derived(offset + 1);
	let toPage = $derived(Math.min(offset + limit, total));
	let hasPrev = $derived(offset > 0);
	let hasNext = $derived(offset + limit < total);

	function dateToInputValue(dateStr: string): string {
		if (!dateStr) return '';
		try {
			const d = new Date(dateStr);
			if (isNaN(d.getTime())) return '';
			return d.toISOString().slice(0, 10);
		} catch {
			return '';
		}
	}
</script>

<div class="max-w-4xl mx-auto">
	<!-- Header -->
	<div class="flex items-center justify-between mb-6">
		<div>
			<h1 class="text-2xl font-bold text-gray-900">Patient Ledger</h1>
			<p class="text-sm text-gray-500 mt-1">Patient #{patientId}</p>
		</div>
		<a
			href="/patients/{patientId}"
			class="text-sm text-blue-600 hover:text-blue-800 font-medium transition-colors"
		>
			&larr; Back to Patient
		</a>
	</div>

	<!-- Date Range Filter -->
	<div class="bg-white rounded-lg shadow-sm border border-gray-200 p-4 mb-6">
		<div class="flex flex-wrap items-end gap-4">
			<div class="flex flex-col gap-1">
				<label for="from-date" class="text-xs font-medium text-gray-600 uppercase">
					From
				</label>
				<input
					id="from-date"
					type="date"
					bind:value={fromDate}
					class="px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
				/>
			</div>
			<div class="flex flex-col gap-1">
				<label for="to-date" class="text-xs font-medium text-gray-600 uppercase">
					To
				</label>
				<input
					id="to-date"
					type="date"
					bind:value={toDate}
					class="px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
				/>
			</div>
			<button
				onclick={applyFilter}
				class="px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded-lg hover:bg-blue-700 transition-colors"
			>
				Apply Filter
			</button>
			<button
				onclick={clearFilter}
				class="px-4 py-2 border border-gray-300 text-gray-700 text-sm font-medium rounded-lg hover:bg-gray-50 transition-colors"
			>
				Clear
			</button>
		</div>
	</div>

	{#if error}
		<div class="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg text-sm mb-6">
			{error}
		</div>
	{/if}

	<!-- Loading -->
	{#if loading}
		<div class="flex justify-center py-12">
			<div
				class="w-6 h-6 border-2 border-blue-500 border-t-transparent rounded-full animate-spin"
			></div>
		</div>

	<!-- Empty state -->
	{:else if entries.length === 0}
		<div
			class="text-center py-12 text-gray-500 bg-white rounded-lg shadow-sm border border-gray-200"
		>
			<p class="text-lg">No ledger entries</p>
			<p class="text-sm mt-1">
				{#if fromDate || toDate}
					No entries found for the selected date range. Try adjusting your filter.
				{:else}
					No ledger entries exist for this patient yet.
				{/if}
			</p>
		</div>

	<!-- Table -->
	{:else}
		<div class="bg-white rounded-lg shadow-sm border border-gray-200 overflow-x-auto">
			<table class="min-w-full divide-y divide-gray-200">
				<thead class="bg-gray-50">
					<tr>
						<th
							class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
						>
							Date
						</th>
						<th
							class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
						>
							Type
						</th>
						<th
							class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
						>
							Description
						</th>
						<th
							class="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider"
						>
							Amount
						</th>
						<th
							class="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider"
						>
							Balance
						</th>
					</tr>
				</thead>
				<tbody class="bg-white divide-y divide-gray-200">
					{#each entries as entry}
						{@const badge = entryTypeBadge(entry.entry_type)}
						<tr class="hover:bg-gray-50">
							<td class="px-4 py-3 text-sm text-gray-700 whitespace-nowrap">
								{formatDate(entry.date)}
							</td>
							<td class="px-4 py-3 text-sm whitespace-nowrap">
								<span
									class="inline-block px-2 py-0.5 text-xs font-medium rounded {badge.classes}"
								>
									{badge.label}
								</span>
							</td>
							<td
								class="px-4 py-3 text-sm text-gray-700 max-w-[300px] truncate"
								title={entry.description || undefined}
							>
								{entry.description || '—'}
								{#if entry.cpt_code}
									<span class="text-xs text-gray-400 ml-1">({entry.cpt_code})</span>
								{/if}
							</td>
							<td class="px-4 py-3 text-sm text-right whitespace-nowrap">
								<span
									class={entry.entry_type === 'payment'
										? 'text-green-600'
										: 'text-gray-700'}
								>
									{entry.entry_type === 'payment' ? '-' : ''}{formatCurrency(entry.amount)}
								</span>
							</td>
							<td class="px-4 py-3 text-sm text-right whitespace-nowrap font-medium text-gray-700">
								{formatCurrency(entry.balance)}
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>

		<!-- Pagination -->
		<div
			class="flex items-center justify-between mt-4 px-1"
		>
			<p class="text-sm text-gray-500">
				{#if total > 0}
					Showing {fromPage}–{toPage} of {total}
				{:else}
					No entries
				{/if}
			</p>
			<div class="flex items-center gap-2">
				<button
					onclick={prevPage}
					disabled={!hasPrev}
					class="px-3 py-1.5 text-sm font-medium rounded-lg border border-gray-300 transition-colors"
					class:bg-white={hasPrev}
					class:text-gray-700={hasPrev}
					class:hover:bg-gray-50={hasPrev}
					class:bg-gray-100={!hasPrev}
					class:text-gray-400={!hasPrev}
					class:cursor-not-allowed={!hasPrev}
				>
					&larr; Previous
				</button>
				<button
					onclick={nextPage}
					disabled={!hasNext}
					class="px-3 py-1.5 text-sm font-medium rounded-lg border border-gray-300 transition-colors"
					class:bg-white={hasNext}
					class:text-gray-700={hasNext}
					class:hover:bg-gray-50={hasNext}
					class:bg-gray-100={!hasNext}
					class:text-gray-400={!hasNext}
					class:cursor-not-allowed={!hasNext}
				>
					Next &rarr;
				</button>
			</div>
		</div>
	{/if}
</div>
