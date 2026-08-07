<script lang="ts">
	import { page } from '$app/stores';
	import { api } from '$lib/api';

	interface SurgicalHistoryItem {
		id: number;
		created_at: string;
		updated_at: string;
		deleted_at: string | null;
		patient: number;
		operation_date: string | null;
		operation: string;
		user: number;
	}

	let patientId = $derived($page.params.id || '');
	let items = $state<SurgicalHistoryItem[]>([]);
	let loading = $state(true);
	let error = $state('');

	// Modal state
	let showModal = $state(false);
	let formSaving = $state(false);
	let formError = $state('');
	let operationDate = $state('');
	let operationDesc = $state('');

	$effect(() => {
		if (patientId) {
			loadItems();
		}
	});

	async function loadItems() {
		loading = true;
		error = '';
		try {
			const data = await api.get<SurgicalHistoryItem[]>(
				`/patient/${patientId}/surgical-history`
			);
			items = data || [];
		} catch (e: any) {
			error = e.message || 'Failed to load surgical history';
		} finally {
			loading = false;
		}
	}

	function openModal() {
		operationDate = new Date().toISOString().slice(0, 10);
		operationDesc = '';
		formError = '';
		showModal = true;
	}

	function closeModal() {
		showModal = false;
	}

	async function handleSubmit(e: Event) {
		e.preventDefault();

		if (!operationDate.trim()) {
			formError = 'Operation date is required';
			return;
		}
		if (!operationDesc.trim()) {
			formError = 'Operation description is required';
			return;
		}

		formSaving = true;
		formError = '';

		try {
			await api.post(`/patient/${patientId}/surgical-history`, {
				operation_date: operationDate.trim(),
				operation: operationDesc.trim(),
			});
			showModal = false;
			await loadItems();
		} catch (e: any) {
			formError = e.message || 'Failed to save surgical history';
		} finally {
			formSaving = false;
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
</script>

<div class="max-w-4xl mx-auto">
	<!-- Header -->
	<div class="flex items-center justify-between mb-6">
		<div>
			<h1 class="text-2xl font-bold text-gray-900">Surgical History</h1>
			<p class="text-sm text-gray-500 mt-1">Patient #{patientId}</p>
		</div>
		<div class="flex items-center gap-3">
			<a
				href="/patients/{patientId}"
				class="text-sm text-blue-600 hover:text-blue-800 font-medium transition-colors"
			>
				&larr; Back to Patient
			</a>
			<button
				onclick={openModal}
				class="px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded-lg hover:bg-blue-700 transition-colors"
			>
				+ Add Surgery
			</button>
		</div>
	</div>

	<!-- Error -->
	{#if error}
		<div class="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg text-sm mb-6">
			{error}
		</div>
	{/if}

	<!-- Loading / Empty / Table -->
	{#if loading}
		<div class="flex justify-center py-12">
			<div class="w-6 h-6 border-2 border-blue-500 border-t-transparent rounded-full animate-spin"></div>
		</div>
	{:else if items.length === 0}
		<div class="text-center py-12 text-gray-500 bg-white rounded-lg shadow-sm border border-gray-200">
			<p class="text-lg">No surgical history</p>
			<p class="text-sm mt-1">No surgical history entries have been recorded for this patient yet.</p>
		</div>
	{:else}
		<div class="bg-white rounded-lg shadow-sm border border-gray-200 overflow-x-auto">
			<table class="min-w-full divide-y divide-gray-200">
				<thead class="bg-gray-50">
					<tr>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Date</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Operation</th>
					</tr>
				</thead>
				<tbody class="bg-white divide-y divide-gray-200">
					{#each items as item}
						<tr class="hover:bg-gray-50">
							<td class="px-4 py-3 text-sm text-gray-700 whitespace-nowrap">
								{formatDate(item.operation_date)}
							</td>
							<td class="px-4 py-3 text-sm text-gray-700">
								{item.operation}
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{/if}
</div>

<!-- Create Modal (pure CSS overlay — no Bootstrap JS) -->
{#if showModal}
	<!-- Backdrop -->
	<div
		class="fixed inset-0 bg-black/50 z-40 transition-opacity"
		onclick={closeModal}
		onkeydown={(e: KeyboardEvent) => { if (e.key === 'Escape') closeModal(); }}
		role="presentation"
	></div>

	<!-- Modal panel -->
	<div class="fixed inset-0 z-50 flex items-center justify-center p-4">
		<div class="bg-white rounded-xl shadow-xl w-full max-w-md mx-auto" role="dialog" aria-modal="true" aria-labelledby="modal-title">
			<!-- Header -->
			<div class="flex items-center justify-between px-6 py-4 border-b border-gray-200">
				<h2 id="modal-title" class="text-lg font-semibold text-gray-900">Add Surgery</h2>
				<button
					onclick={closeModal}
					class="text-gray-400 hover:text-gray-600 transition-colors"
					aria-label="Close"
				>
					<svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
					</svg>
				</button>
			</div>

			<!-- Form -->
			<form onsubmit={handleSubmit}>
				<div class="px-6 py-4 space-y-4">
					{#if formError}
						<div class="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg text-sm">
							{formError}
						</div>
					{/if}

					<div>
						<label for="op-date" class="block text-sm font-medium text-gray-700 mb-1">Operation Date *</label>
						<input
							id="op-date"
							type="date"
							bind:value={operationDate}
							class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-sm"
							required
						/>
					</div>

					<div>
						<label for="op-desc" class="block text-sm font-medium text-gray-700 mb-1">Operation Description *</label>
						<textarea
							id="op-desc"
							bind:value={operationDesc}
							rows="3"
							placeholder="e.g. Appendectomy, laparoscopic"
							class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-sm resize-none"
							required
						></textarea>
					</div>
				</div>

				<!-- Footer -->
				<div class="flex justify-end gap-3 px-6 py-4 border-t border-gray-200">
					<button
						type="button"
						onclick={closeModal}
						class="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors"
					>
						Cancel
					</button>
					<button
						type="submit"
						disabled={formSaving}
						class="px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-lg hover:bg-blue-700 disabled:opacity-50 transition-colors"
					>
						{formSaving ? 'Saving...' : 'Save'}
					</button>
				</div>
			</form>
		</div>
	</div>
{/if}
