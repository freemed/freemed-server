<script lang="ts">
	import { page } from '$app/stores';
	import { api } from '$lib/api';

	interface DrugSample {
		id: number;
		drugcode: string;
		lot: string;
		samplecount: number;
		samplecountremain: number;
		logdate: string;
		received: string;
		drugndc: string;
		drugclass: string;
		packagecount: number;
		location: string;
		drugco: string;
		drugrep: string;
		invoice: string;
		expiration: string;
		assignedto: string;
		loguser: number;
		created_at: string;
	}

	let drugSamples = $state<DrugSample[]>([]);
	let loading = $state(true);
	let error = $state('');
	let patientId = $state('');

	// Form state
	let showForm = $state(false);
	let formSaving = $state(false);
	let formError = $state('');

	// Form fields
	let drug = $state('');
	let lotNumber = $state('');
	let quantity = $state('');
	let date = $state('');

	$effect(() => {
		const id = $page.params.id;
		if (id) {
			patientId = id;
			loadDrugSamples(id);
		}
	});

	async function loadDrugSamples(id: string) {
		loading = true;
		error = '';
		try {
			const data = await api.get<DrugSample[]>(`/patient/${id}/drug-samples`);
			drugSamples = data || [];
		} catch (e: any) {
			error = e.message || 'Failed to load drug samples';
		} finally {
			loading = false;
		}
	}

	function openForm() {
		drug = '';
		lotNumber = '';
		quantity = '';
		date = new Date().toISOString().slice(0, 10);
		formError = '';
		showForm = true;
	}

	function cancelForm() {
		showForm = false;
	}

	async function handleSubmit(e: Event) {
		e.preventDefault();

		if (!drug.trim()) {
			formError = 'Drug name is required';
			return;
		}
		if (!lotNumber.trim()) {
			formError = 'Lot number is required';
			return;
		}
		if (!quantity || parseInt(quantity) <= 0) {
			formError = 'Quantity must be a positive number';
			return;
		}
		if (!date.trim()) {
			formError = 'Date is required';
			return;
		}

		formSaving = true;
		formError = '';

		const payload = {
			drug_code: drug.trim(),
			lot: lotNumber.trim(),
			sample_count: parseInt(quantity, 10),
			received: date,
		};

		try {
			await api.post(`/patient/${patientId}/drug-samples`, payload);
			showForm = false;
			await loadDrugSamples(patientId);
		} catch (e: any) {
			formError = e.message || 'Failed to save drug sample';
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
</script>

<div class="max-w-4xl mx-auto">
	<div class="flex items-center justify-between mb-6">
		<div>
			<h1 class="text-2xl font-bold text-gray-900">Drug Samples</h1>
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
				onclick={openForm}
				class="px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded-lg hover:bg-blue-700 transition-colors"
			>
				+ Add Drug Sample
			</button>
		</div>
	</div>

	{#if error}
		<div class="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg text-sm mb-6">
			{error}
		</div>
	{/if}

	<!-- Add Form -->
	{#if showForm}
		<div class="bg-white rounded-lg shadow-sm border border-gray-200 p-6 mb-6">
			<h2 class="text-lg font-semibold text-gray-800 mb-4">Record Drug Sample</h2>
			{#if formError}
				<div class="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg text-sm mb-4">
					{formError}
				</div>
			{/if}
			<form onsubmit={handleSubmit}>
				<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">Drug *</label>
						<input
							type="text"
							bind:value={drug}
							placeholder="e.g. Lipitor"
							class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-sm"
							required
						/>
					</div>
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">Lot Number *</label>
						<input
							type="text"
							bind:value={lotNumber}
							placeholder="e.g. AB12345"
							class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-sm"
							required
						/>
					</div>
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">Quantity *</label>
						<input
							type="number"
							bind:value={quantity}
							placeholder="e.g. 10"
							min="1"
							class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-sm"
							required
						/>
					</div>
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">Date *</label>
						<input
							type="date"
							bind:value={date}
							class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-sm"
							required
						/>
					</div>
				</div>
				<div class="flex justify-end gap-3 mt-6">
					<button
						type="button"
						onclick={cancelForm}
						class="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors"
					>
						Cancel
					</button>
					<button
						type="submit"
						disabled={formSaving}
						class="px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-lg hover:bg-blue-700 disabled:opacity-50 transition-colors"
					>
						{formSaving ? 'Saving...' : 'Save Drug Sample'}
					</button>
				</div>
			</form>
		</div>
	{/if}

	{#if loading}
		<div class="flex justify-center py-12">
			<div class="w-6 h-6 border-2 border-blue-500 border-t-transparent rounded-full animate-spin"></div>
		</div>
	{:else if drugSamples.length === 0}
		<div class="text-center py-12 text-gray-500 bg-white rounded-lg shadow-sm border border-gray-200">
			<p class="text-lg">No drug samples recorded</p>
			<p class="text-sm mt-1">No drug samples have been recorded for this patient yet.</p>
		</div>
	{:else}
		<div class="bg-white rounded-lg shadow-sm border border-gray-200 overflow-x-auto">
			<table class="min-w-full divide-y divide-gray-200">
				<thead class="bg-gray-50">
					<tr>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Drug Name</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Lot</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Quantity</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Date</th>
					</tr>
				</thead>
				<tbody class="bg-white divide-y divide-gray-200">
					{#each drugSamples as ds}
						<tr class="hover:bg-gray-50">
							<td class="px-4 py-3 text-sm text-gray-900 font-medium">
								{ds.drugcode || '—'}
							</td>
							<td class="px-4 py-3 text-sm text-gray-700">
								{ds.lot || '—'}
							</td>
							<td class="px-4 py-3 text-sm text-gray-700">
								{ds.samplecount}
							</td>
							<td class="px-4 py-3 text-sm text-gray-700 whitespace-nowrap">
								{formatDate(ds.logdate)}
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{/if}
</div>
