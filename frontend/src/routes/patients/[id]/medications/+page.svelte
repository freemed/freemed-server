<script lang="ts">
	import { page } from '$app/stores';
	import { api } from '$lib/api';

	interface Medication {
		id: number;
		patient: number;
		drug_name: string;
		dosage: string;
		frequency: string;
		start_date: string;
		end_date: string | null;
		prescribing_provider: number;
		active: string;
		created_at: string;
	}

	let medications = $state<Medication[]>([]);
	let loading = $state(true);
	let error = $state('');
	let patientId = $state('');

	// Form state
	let showForm = $state(false);
	let editingId = $state<number | null>(null);
	let formError = $state('');
	let formSubmitting = $state(false);
	let drugName = $state('');
	let dosage = $state('');
	let frequency = $state('');
	let startDate = $state('');
	let endDate = $state('');
	let prescribingProvider = $state(0);

	// Track which med's end-date field to show in edit mode
	let editingMed = $state<Medication | null>(null);

	$effect(() => {
		const id = $page.params.id;
		if (id) {
			patientId = id;
			loadMeds(id);
		}
	});

	async function loadMeds(id: string) {
		loading = true;
		error = '';
		try {
			const data = await api.get<Medication[]>(`/patient/${id}/medications?all=1`);
			medications = data || [];
		} catch (e: any) {
			error = e.message || 'Failed to load medications';
		} finally {
			loading = false;
		}
	}

	function openAddForm() {
		editingId = null;
		editingMed = null;
		drugName = '';
		dosage = '';
		frequency = '';
		startDate = '';
		endDate = '';
		prescribingProvider = 0;
		formError = '';
		showForm = true;
	}

	function openEditForm(med: Medication) {
		editingId = med.id;
		editingMed = med;
		drugName = med.drug_name;
		dosage = med.dosage;
		frequency = med.frequency;
		startDate = med.start_date ? med.start_date.slice(0, 10) : '';
		endDate = med.end_date ? med.end_date.slice(0, 10) : '';
		prescribingProvider = med.prescribing_provider;
		formError = '';
		showForm = true;
	}

	function cancelForm() {
		showForm = false;
		editingId = null;
		editingMed = null;
	}

	async function handleSubmit(e: Event) {
		e.preventDefault();
		if (!drugName.trim()) {
			formError = 'Drug name is required';
			return;
		}

		formSubmitting = true;
		formError = '';

		const payload = {
			drug_name: drugName.trim(),
			dosage: dosage.trim(),
			frequency: frequency.trim(),
			start_date: startDate || null,
			end_date: endDate || null,
			prescribing_provider: prescribingProvider,
		};

		try {
			if (editingId) {
				await api.put(`/medications/${editingId}`, payload);
			} else {
				await api.post(`/patient/${patientId}/medications`, payload);
			}
			showForm = false;
			editingId = null;
			editingMed = null;
			await loadMeds(patientId);
		} catch (e: any) {
			formError = e.message || 'Failed to save medication';
		} finally {
			formSubmitting = false;
		}
	}

	async function discontinueMed(med: Medication) {
		if (!confirm(`Discontinue ${med.drug_name}?`)) return;
		try {
			await api.del(`/medications/${med.id}`);
			await loadMeds(patientId);
		} catch (e: any) {
			error = e.message || 'Failed to discontinue medication';
		}
	}

	function formatDate(dateStr: string | null): string {
		if (!dateStr) return '';
		try {
			const d = new Date(dateStr);
			if (isNaN(d.getTime())) return dateStr;
			return d.toLocaleDateString('en-US', { year: 'numeric', month: 'short', day: 'numeric' });
		} catch {
			return dateStr;
		}
	}
</script>

<div class="max-w-4xl mx-auto">
	<div class="flex items-center justify-between mb-6">
		<div>
			<h1 class="text-2xl font-bold text-gray-900">Medications</h1>
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
				onclick={openAddForm}
				class="px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded-lg hover:bg-blue-700 transition-colors"
			>
				+ Add Medication
			</button>
		</div>
	</div>

	{#if error}
		<div class="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg text-sm mb-6">
			{error}
		</div>
	{/if}

	<!-- Add/Edit Form -->
	{#if showForm}
		<div class="bg-white rounded-lg shadow-sm border border-gray-200 p-6 mb-6">
			<h2 class="text-lg font-semibold text-gray-800 mb-4">
				{editingId ? 'Edit Medication' : 'Add Medication'}
			</h2>
			{#if formError}
				<div class="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg text-sm mb-4">
					{formError}
				</div>
			{/if}
			<form onsubmit={handleSubmit}>
				<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
					<div class="md:col-span-2">
						<label class="block text-sm font-medium text-gray-700 mb-1">Drug Name *</label>
						<input
							type="text"
							bind:value={drugName}
							class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-sm"
							required
						/>
					</div>
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">Dosage</label>
						<input
							type="text"
							bind:value={dosage}
							placeholder="e.g. 10mg"
							class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-sm"
						/>
					</div>
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">Frequency</label>
						<input
							type="text"
							bind:value={frequency}
							placeholder="e.g. Once daily"
							class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-sm"
						/>
					</div>
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">Start Date</label>
						<input
							type="date"
							bind:value={startDate}
							class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-sm"
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
						<label class="block text-sm font-medium text-gray-700 mb-1">Prescribing Provider ID</label>
						<input
							type="number"
							bind:value={prescribingProvider}
							class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-sm"
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
						disabled={formSubmitting}
						class="px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-lg hover:bg-blue-700 disabled:opacity-50 transition-colors"
					>
						{formSubmitting ? 'Saving...' : editingId ? 'Update' : 'Add Medication'}
					</button>
				</div>
			</form>
		</div>
	{/if}

	{#if loading}
		<div class="flex justify-center py-12">
			<div class="w-6 h-6 border-2 border-blue-500 border-t-transparent rounded-full animate-spin"></div>
		</div>
	{:else if medications.length === 0}
		<div class="text-center py-12 text-gray-500">
			<p class="text-lg">No medications found</p>
			<p class="text-sm mt-1">No medications have been prescribed for this patient yet.</p>
		</div>
	{:else}
		<div class="bg-white rounded-lg shadow-sm border border-gray-200 overflow-x-auto">
			<table class="min-w-full divide-y divide-gray-200">
				<thead class="bg-gray-50">
					<tr>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Drug</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Dosage</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Frequency</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Start</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">End</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
						<th class="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Actions</th>
					</tr>
				</thead>
				<tbody class="bg-white divide-y divide-gray-200">
					{#each medications as med}
						<tr class="hover:bg-gray-50">
							<td class="px-4 py-3 text-sm font-medium text-gray-900">{med.drug_name}</td>
							<td class="px-4 py-3 text-sm text-gray-600">{med.dosage || '—'}</td>
							<td class="px-4 py-3 text-sm text-gray-600">{med.frequency || '—'}</td>
							<td class="px-4 py-3 text-sm text-gray-600">{formatDate(med.start_date) || '—'}</td>
							<td class="px-4 py-3 text-sm text-gray-600">{formatDate(med.end_date) || '—'}</td>
							<td class="px-4 py-3 text-sm">
								{#if med.active === 'active'}
									<span class="inline-block px-2 py-0.5 text-xs font-medium bg-green-50 text-green-700 rounded">Active</span>
								{:else}
									<span class="inline-block px-2 py-0.5 text-xs font-medium bg-gray-50 text-gray-500 rounded">Discontinued</span>
								{/if}
							</td>
							<td class="px-4 py-3 text-sm text-right whitespace-nowrap">
								{#if med.active === 'active'}
									<button
										onclick={() => openEditForm(med)}
										class="text-blue-600 hover:text-blue-800 font-medium mr-3"
									>
										Edit
									</button>
									<button
										onclick={() => discontinueMed(med)}
										class="text-red-600 hover:text-red-800 font-medium"
									>
										Discontinue
									</button>
								{:else}
									<span class="text-gray-400">—</span>
								{/if}
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{/if}
</div>
