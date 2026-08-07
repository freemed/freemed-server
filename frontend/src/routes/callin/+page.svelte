<script lang="ts">
	import { api } from '$lib/api';

	interface CallinRecord {
		id: number;
		cilname: string;
		cifname: string;
		cicomplaint: string;
		created_at: string;
		updated_at: string;
	}

	let records = $state<CallinRecord[]>([]);
	let loading = $state(true);
	let error = $state('');

	let expandedId = $state<number | null>(null);
	let detailLoading = $state(false);
	let detailRecord = $state<CallinRecord | null>(null);

	async function loadRecords() {
		loading = true;
		error = '';
		try {
			records = await api.get<CallinRecord[]>('/callin');
		} catch (e: any) {
			error = e.message || 'Failed to load call-in records';
		} finally {
			loading = false;
		}
	}

	$effect(() => {
		loadRecords();
	});

	async function toggleDetail(id: number) {
		if (expandedId === id) {
			expandedId = null;
			detailRecord = null;
			return;
		}
		expandedId = id;
		detailLoading = true;
		detailRecord = null;
		try {
			detailRecord = await api.get<CallinRecord>(`/callin/${id}`);
		} catch (e: any) {
			detailRecord = null;
		} finally {
			detailLoading = false;
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
				hour: 'numeric',
				minute: '2-digit',
			});
		} catch {
			return dateStr;
		}
	}
</script>

<div class="max-w-5xl mx-auto">
	<h1 class="text-2xl font-bold text-gray-900 mb-6">Call-In Patient Log</h1>

	{#if error}
		<div class="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg text-sm mb-6">
			{error}
		</div>
	{/if}

	{#if loading}
		<div class="flex justify-center py-12">
			<div class="w-6 h-6 border-2 border-blue-500 border-t-transparent rounded-full animate-spin"></div>
		</div>
	{:else if records.length === 0}
		<div class="text-center py-12 text-gray-500">
			<p class="text-lg">No call-in records found</p>
		</div>
	{:else}
		<div class="bg-white rounded-lg shadow-sm border border-gray-200 overflow-x-auto">
			<table class="w-full text-sm">
				<thead class="bg-gray-50 border-b border-gray-200">
					<tr>
						<th class="text-left px-4 py-3 font-semibold text-gray-600">Name</th>
						<th class="text-left px-4 py-3 font-semibold text-gray-600">Complaint</th>
						<th class="text-left px-4 py-3 font-semibold text-gray-600">Date</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-gray-100">
					{#each records as record (record.id)}
						<tr
							onclick={() => toggleDetail(record.id)}
							onkeydown={(e) => { if (e.key === 'Enter') toggleDetail(record.id); }}
							tabindex="0"
							role="button"
							class="cursor-pointer hover:bg-blue-50 transition-colors focus:outline-none focus:ring-2 focus:ring-inset focus:ring-blue-500"
						>
							<td class="px-4 py-3 font-medium text-gray-900">
								{[record.cifname, record.cilname].filter(Boolean).join(' ') || '—'}
							</td>
							<td class="px-4 py-3 text-gray-600 max-w-xs truncate">
								{record.cicomplaint || '—'}
							</td>
							<td class="px-4 py-3 text-gray-500">
								{formatDate(record.created_at)}
							</td>
						</tr>

						{#if expandedId === record.id}
							<tr class="bg-blue-50/50">
								<td colspan="3" class="px-4 py-4">
									{#if detailLoading}
										<div class="flex justify-center py-4">
											<div class="w-5 h-5 border-2 border-blue-500 border-t-transparent rounded-full animate-spin"></div>
										</div>
									{:else if detailRecord}
										<dl class="grid grid-cols-1 sm:grid-cols-2 gap-3">
											<div>
												<dt class="text-xs font-medium text-gray-500 uppercase tracking-wide">Full Name</dt>
												<dd class="text-sm text-gray-900 mt-0.5">
													{[detailRecord.cifname, detailRecord.cilname].filter(Boolean).join(' ') || '—'}
												</dd>
											</div>
											<div>
												<dt class="text-xs font-medium text-gray-500 uppercase tracking-wide">Complaint</dt>
												<dd class="text-sm text-gray-900 mt-0.5">
													{detailRecord.cicomplaint || '—'}
												</dd>
											</div>
											<div>
												<dt class="text-xs font-medium text-gray-500 uppercase tracking-wide">Created</dt>
												<dd class="text-sm text-gray-900 mt-0.5">
													{formatDate(detailRecord.created_at)}
												</dd>
											</div>
											<div>
												<dt class="text-xs font-medium text-gray-500 uppercase tracking-wide">Last Updated</dt>
												<dd class="text-sm text-gray-900 mt-0.5">
													{formatDate(detailRecord.updated_at)}
												</dd>
											</div>
										</dl>
									{:else}
										<p class="text-sm text-gray-500">Failed to load details.</p>
									{/if}
								</td>
							</tr>
						{/if}
					{/each}
				</tbody>
			</table>
		</div>
	{/if}
</div>
