<script lang="ts">
	import { page } from '$app/stores';
	import { api } from '$lib/api';

	interface Correspondence {
		id: number;
		date: string;
		type: string;
		summary: string;
		body: string;
		status: string;
		patient: number;
		created_at: string;
	}

	let items = $state<Correspondence[]>([]);
	let loading = $state(true);
	let error = $state('');
	let patientId = $state('');

	// Modal state
	let showModal = $state(false);
	let modalSaving = $state(false);
	let modalError = $state('');
	let modalFieldErrors = $state<Record<string, string>>({});

	// Form fields
	let formType = $state('');
	let formSummary = $state('');
	let formBody = $state('');

	// Correspondence types
	const correspondenceTypes = [
		'Letter',
		'Email',
		'Fax',
		'Referral Letter',
		'Consult Note',
		'Other',
	];

	$effect(() => {
		const id = $page.params.id;
		if (id) {
			patientId = id;
			loadCorrespondence(id);
		}
	});

	async function loadCorrespondence(id: string) {
		loading = true;
		error = '';
		try {
			const data = await api.get<Correspondence[]>(`/patient/${id}/correspondence`);
			items = data || [];
		} catch (e: any) {
			error = e.message || 'Failed to load correspondence';
		} finally {
			loading = false;
		}
	}

	function openModal() {
		formType = '';
		formSummary = '';
		formBody = '';
		modalError = '';
		modalFieldErrors = {};
		showModal = true;
	}

	function closeModal() {
		showModal = false;
	}

	function validate(): boolean {
		const errs: Record<string, string> = {};
		if (!formType.trim()) errs.type = 'Type is required';
		if (!formSummary.trim()) errs.summary = 'Summary is required';
		if (!formBody.trim()) errs.body = 'Body is required';
		modalFieldErrors = errs;
		return Object.keys(errs).length === 0;
	}

	async function handleSubmit(e: Event) {
		e.preventDefault();
		if (!validate()) return;

		modalSaving = true;
		modalError = '';

		try {
			await api.post(`/patient/${patientId}/correspondence`, {
				type: formType.trim(),
				summary: formSummary.trim(),
				body: formBody.trim(),
			});
			showModal = false;
			await loadCorrespondence(patientId);
		} catch (e: any) {
			modalError = e.message || 'Failed to save correspondence';
		} finally {
			modalSaving = false;
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

	function statusClass(status: string): string {
		switch ((status || '').toLowerCase()) {
			case 'sent':
				return 'bg-green-100 text-green-800';
			case 'draft':
				return 'bg-yellow-100 text-yellow-800';
			case 'pending':
				return 'bg-blue-100 text-blue-800';
			case 'failed':
				return 'bg-red-100 text-red-800';
			default:
				return 'bg-gray-100 text-gray-800';
		}
	}
</script>

<div class="max-w-4xl mx-auto">
	<div class="flex items-center justify-between mb-6">
		<div>
			<h1 class="text-2xl font-bold text-gray-900">Correspondence</h1>
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
				+ Compose
			</button>
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
	{:else if items.length === 0}
		<div class="text-center py-12 text-gray-500 bg-white rounded-lg shadow-sm border border-gray-200">
			<p class="text-lg">No correspondence</p>
			<p class="text-sm mt-1">No correspondence has been recorded for this patient yet.</p>
		</div>
	{:else}
		<div class="bg-white rounded-lg shadow-sm border border-gray-200 overflow-x-auto">
			<table class="min-w-full divide-y divide-gray-200">
				<thead class="bg-gray-50">
					<tr>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Date</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Type</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Summary</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
					</tr>
				</thead>
				<tbody class="bg-white divide-y divide-gray-200">
					{#each items as item}
						<tr class="hover:bg-gray-50">
							<td class="px-4 py-3 text-sm text-gray-700 whitespace-nowrap">
								{formatDate(item.date)}
							</td>
							<td class="px-4 py-3 text-sm text-gray-700">
								{item.type || '—'}
							</td>
							<td class="px-4 py-3 text-sm text-gray-700 max-w-[300px] truncate" title={String(item.summary || '')}>
								{item.summary || '—'}
							</td>
							<td class="px-4 py-3 text-sm">
								<span class="inline-flex px-2 py-0.5 text-xs font-medium rounded-full {statusClass(item.status)}">
									{item.status || 'Unknown'}
								</span>
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{/if}
</div>

<!-- Compose Modal -->
{#if showModal}
	<!-- svelte-ignore a11y_no_static_element_interactions -->
	<div
		class="fixed inset-0 z-50 flex items-center justify-center bg-black/40"
		onkeydown={(e: KeyboardEvent) => { if (e.key === 'Escape') closeModal(); }}
	>
		<!-- svelte-ignore a11y_click_events_have_key_events -->
		<div
			class="bg-white rounded-xl shadow-xl w-full max-w-lg mx-4 max-h-[90vh] overflow-y-auto"
			role="dialog"
			aria-modal="true"
			aria-label="Compose correspondence"
		>
			<div class="flex items-center justify-between px-6 py-4 border-b border-gray-200">
				<h2 class="text-lg font-semibold text-gray-800">Compose Correspondence</h2>
				<button
					onclick={closeModal}
					class="p-1 text-gray-400 hover:text-gray-600 hover:bg-gray-100 rounded-lg transition-colors"
					aria-label="Close modal"
				>
					<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
					</svg>
				</button>
			</div>

			<div class="px-6 py-4">
				{#if modalError}
					<div class="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg text-sm mb-4">
						{modalError}
					</div>
				{/if}

				<form onsubmit={handleSubmit}>
					<div class="space-y-4">
						<!-- Type -->
						<div>
							<label for="formType" class="block text-sm font-medium text-gray-700 mb-1">
								Type <span class="text-red-500">*</span>
							</label>
							<select
								id="formType"
								bind:value={formType}
								class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none transition-shadow"
								class:border-red-300={!!modalFieldErrors.type}
								class:border-gray-300={!modalFieldErrors.type}
							>
								<option value="">-- Select type --</option>
								{#each correspondenceTypes as t}
									<option value={t}>{t}</option>
								{/each}
							</select>
							{#if modalFieldErrors.type}
								<p class="text-red-500 text-xs mt-1">{modalFieldErrors.type}</p>
							{/if}
						</div>

						<!-- Summary -->
						<div>
							<label for="formSummary" class="block text-sm font-medium text-gray-700 mb-1">
								Summary <span class="text-red-500">*</span>
							</label>
							<input
								id="formSummary"
								type="text"
								bind:value={formSummary}
								placeholder="Brief summary of this correspondence..."
								class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none transition-shadow"
								class:border-red-300={!!modalFieldErrors.summary}
								class:border-gray-300={!modalFieldErrors.summary}
							/>
							{#if modalFieldErrors.summary}
								<p class="text-red-500 text-xs mt-1">{modalFieldErrors.summary}</p>
							{/if}
						</div>

						<!-- Body -->
						<div>
							<label for="formBody" class="block text-sm font-medium text-gray-700 mb-1">
								Body <span class="text-red-500">*</span>
							</label>
							<textarea
								id="formBody"
								bind:value={formBody}
								rows="8"
								placeholder="Type the full correspondence body..."
								class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none transition-shadow resize-y"
								class:border-red-300={!!modalFieldErrors.body}
								class:border-gray-300={!modalFieldErrors.body}
							></textarea>
							{#if modalFieldErrors.body}
								<p class="text-red-500 text-xs mt-1">{modalFieldErrors.body}</p>
							{/if}
						</div>
					</div>

					<div class="flex justify-end gap-3 mt-6">
						<button
							type="button"
							onclick={closeModal}
							class="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors"
						>
							Cancel
						</button>
						<button
							type="submit"
							disabled={modalSaving}
							class="px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
						>
							{#if modalSaving}
								Saving...
							{:else}
								Save
							{/if}
						</button>
					</div>
				</form>
			</div>
		</div>
	</div>
{/if}
