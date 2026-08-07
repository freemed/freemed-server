<script lang="ts">
	import { page } from '$app/stores';
	import { api } from '$lib/api';

	interface EpisodeOfCare {
		id: number;
		start_date: string;
		end_date: string | null;
		description: string;
		status: string;
		provider: number;
		patient: number;
		notes: string;
		active: string;
		created_at: string;
	}

	let eocs = $state<EpisodeOfCare[]>([]);
	let loading = $state(true);
	let error = $state('');
	let patientId = $state('');

	// Modal state
	let showModal = $state(false);
	let formSubmitting = $state(false);
	let formError = $state('');

	// Form fields
	let startDate = $state('');
	let endDate = $state('');
	let description = $state('');
	let status = $state('open');
	let providerId = $state('');

	$effect(() => {
		const id = $page.params.id;
		if (id) {
			patientId = id;
			loadEOCs(id);
		}
	});

	async function loadEOCs(id: string) {
		loading = true;
		error = '';
		try {
			const data = await api.get<EpisodeOfCare[]>(`/patient/${id}/episodes-of-care`);
			eocs = data || [];
		} catch (e: any) {
			error = e.message || 'Failed to load episodes of care';
		} finally {
			loading = false;
		}
	}

	function openCreateModal() {
		startDate = '';
		endDate = '';
		description = '';
		status = 'open';
		providerId = '';
		formError = '';
		showModal = true;
	}

	function closeModal() {
		showModal = false;
	}

	function handleOverlayClick(e: MouseEvent) {
		if (e.target === e.currentTarget) {
			closeModal();
		}
	}

	function handleEscape(e: KeyboardEvent) {
		if (e.key === 'Escape') {
			closeModal();
		}
	}

	async function handleSubmit(e: Event) {
		e.preventDefault();
		if (!startDate.trim()) {
			formError = 'Start date is required';
			return;
		}

		formSubmitting = true;
		formError = '';

		const payload: Record<string, any> = {
			start_date: startDate.trim(),
			description: description.trim(),
			status: status,
			provider_id: parseInt(providerId, 10) || 0,
		};

		if (endDate.trim()) payload.end_date = endDate.trim();

		try {
			await api.post(`/patient/${patientId}/episodes-of-care`, payload);
			showModal = false;
			await loadEOCs(patientId);
		} catch (e: any) {
			formError = e.message || 'Failed to create episode of care';
		} finally {
			formSubmitting = false;
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

	function statusBadge(s: string): { label: string; classes: string } {
		switch (s) {
			case 'active':
				return { label: 'Active', classes: 'bg-blue-50 text-blue-700' };
			case 'open':
				return { label: 'Open', classes: 'bg-yellow-50 text-yellow-700' };
			case 'closed':
				return { label: 'Closed', classes: 'bg-gray-100 text-gray-600' };
			default:
				return { label: s, classes: 'bg-gray-50 text-gray-500' };
		}
	}
</script>

<svelte:window onkeydown={handleEscape} />

<div class="max-w-4xl mx-auto">
	<div class="flex items-center justify-between mb-6">
		<div>
			<h1 class="text-2xl font-bold text-gray-900">Episodes of Care</h1>
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
				onclick={openCreateModal}
				class="px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded-lg hover:bg-blue-700 transition-colors"
			>
				+ New Episode of Care
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
	{:else if eocs.length === 0}
		<div class="text-center py-12 text-gray-500 bg-white rounded-lg shadow-sm border border-gray-200">
			<p class="text-lg">No episodes of care</p>
			<p class="text-sm mt-1">No episodes of care have been recorded for this patient yet.</p>
		</div>
	{:else}
		<div class="bg-white rounded-lg shadow-sm border border-gray-200 overflow-x-auto">
			<table class="min-w-full divide-y divide-gray-200">
				<thead class="bg-gray-50">
					<tr>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Start Date</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">End Date</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Description</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Provider</th>
					</tr>
				</thead>
				<tbody class="bg-white divide-y divide-gray-200">
					{#each eocs as eoc}
						{@const badge = statusBadge(eoc.status)}
						<tr class="hover:bg-gray-50">
							<td class="px-4 py-3 text-sm text-gray-700 whitespace-nowrap">
								{formatDate(eoc.start_date)}
							</td>
							<td class="px-4 py-3 text-sm text-gray-700 whitespace-nowrap">
								{formatDate(eoc.end_date)}
							</td>
							<td class="px-4 py-3 text-sm text-gray-700 max-w-[300px] truncate">
								{eoc.description || '—'}
							</td>
							<td class="px-4 py-3 text-sm whitespace-nowrap">
								<span class="inline-block px-2 py-0.5 text-xs font-medium rounded {badge.classes}">
									{badge.label}
								</span>
							</td>
							<td class="px-4 py-3 text-sm text-gray-700 whitespace-nowrap">
								{eoc.provider ? `#${eoc.provider}` : '—'}
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{/if}
</div>

<!-- Create Modal -->
{#if showModal}
	<!-- svelte-ignore a11y_no_static_element_interactions -->
	<div
		class="fixed inset-0 z-50 flex items-center justify-center bg-black/50"
		onclick={handleOverlayClick}
		onkeydown={(e) => { if (e.key === 'Escape') closeModal(); }}
		role="dialog"
		aria-modal="true"
		tabindex="-1"
	>
		<div class="bg-white rounded-lg shadow-xl w-full max-w-lg mx-4 p-6">
			<div class="flex items-center justify-between mb-4">
				<h2 class="text-lg font-semibold text-gray-800">New Episode of Care</h2>
				<button
					onclick={closeModal}
					class="text-gray-400 hover:text-gray-600 transition-colors"
					aria-label="Close"
				>
					<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
					</svg>
				</button>
			</div>

			{#if formError}
				<div class="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg text-sm mb-4">
					{formError}
				</div>
			{/if}

			<form onsubmit={handleSubmit}>
				<div class="space-y-4">
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">Start Date *</label>
						<input
							type="date"
							bind:value={startDate}
							class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-sm"
							required
						/>
					</div>
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">End Date</label>
						<input
							type="date"
							bind:value={endDate}
							class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-sm"
						/>
					</div>
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">Description</label>
						<input
							type="text"
							bind:value={description}
							placeholder="e.g. Physical therapy for knee injury"
							class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-sm"
						/>
					</div>
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">Status</label>
						<select
							bind:value={status}
							class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-sm bg-white"
						>
							<option value="open">Open</option>
							<option value="active">Active</option>
							<option value="closed">Closed</option>
						</select>
					</div>
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">Provider ID</label>
						<input
							type="number"
							bind:value={providerId}
							placeholder="e.g. 1"
							class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-sm"
						/>
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
						disabled={formSubmitting}
						class="px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-lg hover:bg-blue-700 disabled:opacity-50 transition-colors"
					>
						{formSubmitting ? 'Creating...' : 'Create Episode'}
					</button>
				</div>
			</form>
		</div>
	</div>
{/if}
