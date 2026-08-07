<script lang="ts">
	import { page } from '$app/stores';
	import { api } from '$lib/api';

	interface Certification {
		id: number;
		created_at: string;
		updated_at: string;
		deleted_at: string | null;
		patient: number;
		cert_type: number;
		cert_form_num: number | null;
		cert_desc: string;
		cert_form_data: string | null;
		user: number;
		active: string;
	}

	let patientId = $derived($page.params.id || '');
	let items = $state<Certification[]>([]);
	let loading = $state(true);
	let error = $state('');

	// Modal state
	let showModal = $state(false);
	let formSaving = $state(false);
	let formError = $state('');
	let certType = $state('');
	let certDesc = $state('');
	let certFormNum = $state('');

	$effect(() => {
		if (patientId) {
			loadItems();
		}
	});

	async function loadItems() {
		loading = true;
		error = '';
		try {
			const data = await api.get<Certification[]>(
				`/patient/${patientId}/certifications`
			);
			items = data || [];
		} catch (e: any) {
			error = e.message || 'Failed to load certifications';
		} finally {
			loading = false;
		}
	}

	function openModal() {
		certType = '';
		certDesc = '';
		certFormNum = '';
		formError = '';
		showModal = true;
	}

	function closeModal() {
		showModal = false;
	}

	async function handleSubmit(e: Event) {
		e.preventDefault();

		if (!certType.trim()) {
			formError = 'Certification type is required';
			return;
		}
		if (!certDesc.trim()) {
			formError = 'Description is required';
			return;
		}

		formSaving = true;
		formError = '';

		const certTypeNum = parseInt(certType.trim(), 10);

		try {
			await api.post(`/patient/${patientId}/certifications`, {
				cert_type: certTypeNum,
				cert_desc: certDesc.trim(),
				cert_form_num: certFormNum.trim()
					? parseInt(certFormNum.trim(), 10)
					: null,
			});
			showModal = false;
			await loadItems();
		} catch (e: any) {
			formError = e.message || 'Failed to save certification';
		} finally {
			formSaving = false;
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

	function certTypeLabel(t: number): string {
		switch (t) {
			case 1:
				return 'Medical';
			case 2:
				return 'Disability';
			case 3:
				return 'Work';
			case 4:
				return 'School';
			case 5:
				return 'DMV';
			default:
				return `Type ${t}`;
		}
	}

	function statusBadgeClass(status: string): string {
		return status === 'active'
			? 'bg-green-50 text-green-700'
			: 'bg-gray-50 text-gray-500';
	}
</script>

<div class="max-w-4xl mx-auto">
	<!-- Header -->
	<div class="flex items-center justify-between mb-6">
		<div>
			<h1 class="text-2xl font-bold text-gray-900">Certifications</h1>
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
				+ Add Certification
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
			<p class="text-lg">No certifications</p>
			<p class="text-sm mt-1">No certifications have been recorded for this patient yet.</p>
		</div>
	{:else}
		<div class="bg-white rounded-lg shadow-sm border border-gray-200 overflow-x-auto">
			<table class="min-w-full divide-y divide-gray-200">
				<thead class="bg-gray-50">
					<tr>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Cert Type</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Description</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Form Number</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
					</tr>
				</thead>
				<tbody class="bg-white divide-y divide-gray-200">
					{#each items as item}
						<tr class="hover:bg-gray-50">
							<td class="px-4 py-3 text-sm text-gray-700 whitespace-nowrap">
								{certTypeLabel(item.cert_type)}
							</td>
							<td class="px-4 py-3 text-sm text-gray-700 max-w-xs truncate">
								{item.cert_desc}
							</td>
							<td class="px-4 py-3 text-sm text-gray-700 whitespace-nowrap">
								{item.cert_form_num ?? '—'}
							</td>
							<td class="px-4 py-3 text-sm">
								<span class="inline-block px-2 py-0.5 text-xs font-medium rounded {statusBadgeClass(item.active)}">
									{item.active}
								</span>
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
		class="fixed inset-0 z-50 flex items-center justify-center"
		onkeydown={(e) => { if (e.key === 'Escape') closeModal(); }}
	>
		<!-- Backdrop -->
		<!-- svelte-ignore a11y_no_static_element_interactions -->
		<div
			class="absolute inset-0 bg-black/50"
			onclick={closeModal}
		></div>

		<!-- Modal panel -->
		<div class="relative bg-white rounded-lg shadow-xl w-full max-w-md mx-4 max-h-[90vh] overflow-y-auto">
			<div class="px-6 py-4 border-b border-gray-100 flex items-center justify-between">
				<h2 class="text-lg font-semibold text-gray-900">Add Certification</h2>
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

			<form onsubmit={handleSubmit}>
				<div class="px-6 py-4 space-y-4">
					{#if formError}
						<div class="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg text-sm">
							{formError}
						</div>
					{/if}

					<div>
						<label for="cert-type" class="block text-sm font-medium text-gray-700 mb-1">Certification Type *</label>
						<select
							id="cert-type"
							bind:value={certType}
							class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-sm"
							required
						>
							<option value="">Select type...</option>
							<option value="1">Medical</option>
							<option value="2">Disability</option>
							<option value="3">Work</option>
							<option value="4">School</option>
							<option value="5">DMV</option>
						</select>
					</div>

					<div>
						<label for="cert-desc" class="block text-sm font-medium text-gray-700 mb-1">Description *</label>
						<textarea
							id="cert-desc"
							bind:value={certDesc}
							rows="3"
							placeholder="Certification description..."
							class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-sm resize-none"
							required
						></textarea>
					</div>

					<div>
						<label for="cert-form-num" class="block text-sm font-medium text-gray-700 mb-1">Form Number</label>
						<input
							id="cert-form-num"
							type="number"
							bind:value={certFormNum}
							placeholder="e.g. 12345"
							class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-sm"
						/>
					</div>
				</div>

				<div class="px-6 py-4 border-t border-gray-100 flex justify-end gap-3">
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
