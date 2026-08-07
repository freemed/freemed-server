<script lang="ts">
	import { page } from '$app/stores';
	import { api } from '$lib/api';

	interface Immunization {
		id: number;
		dateof: string;
		patient: number;
		provider: number;
		admin_provider: number;
		immunization: number;
		route: number;
		body_site: number;
		manufacturer: string | null;
		lot_number: string | null;
		previous_doses: number;
		recovered: boolean;
		notes: string | null;
		active: string;
		created_at: string;
	}

	let immunizations = $state<Immunization[]>([]);
	let loading = $state(true);
	let error = $state('');
	let patientId = $state('');

	// Form state
	let showForm = $state(false);
	let formSaving = $state(false);
	let formError = $state('');

	// Form fields
	let dateof = $state('');
	let vaccineId = $state('');
	let lotNumber = $state('');
	let manufacturer = $state('');
	let providerId = $state('');
	let notes = $state('');

	$effect(() => {
		const id = $page.params.id;
		if (id) {
			patientId = id;
			loadImmunizations(id);
		}
	});

	async function loadImmunizations(id: string) {
		loading = true;
		error = '';
		try {
			const data = await api.get<Immunization[]>(`/patient/${id}/immunizations`);
			immunizations = data || [];
		} catch (e: any) {
			error = e.message || 'Failed to load immunizations';
		} finally {
			loading = false;
		}
	}

	function openForm() {
		dateof = new Date().toISOString().slice(0, 10);
		vaccineId = '';
		lotNumber = '';
		manufacturer = '';
		providerId = '';
		notes = '';
		formError = '';
		showForm = true;
	}

	function cancelForm() {
		showForm = false;
	}

	async function handleSubmit(e: Event) {
		e.preventDefault();
		if (!dateof.trim()) {
			formError = 'Date administered is required';
			return;
		}

		formSaving = true;
		formError = '';

		const payload: Record<string, any> = {
			dateof: dateof,
		};

		if (vaccineId.trim()) payload.immunization = parseInt(vaccineId, 10);
		if (lotNumber.trim()) payload.lot_number = lotNumber.trim();
		if (manufacturer.trim()) payload.manufacturer = manufacturer.trim();
		if (providerId.trim()) payload.provider = parseInt(providerId, 10);
		if (notes.trim()) payload.notes = notes.trim();

		try {
			await api.post(`/patient/${patientId}/immunizations`, payload);
			showForm = false;
			await loadImmunizations(patientId);
		} catch (e: any) {
			formError = e.message || 'Failed to save immunization';
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
			<h1 class="text-2xl font-bold text-gray-900">Immunizations</h1>
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
				+ Record Immunization
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
			<h2 class="text-lg font-semibold text-gray-800 mb-4">Record Immunization</h2>
			{#if formError}
				<div class="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg text-sm mb-4">
					{formError}
				</div>
			{/if}
			<form onsubmit={handleSubmit}>
				<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">Date Administered *</label>
						<input
							type="date"
							bind:value={dateof}
							class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-sm"
							required
						/>
					</div>
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">Vaccine ID</label>
						<input
							type="number"
							bind:value={vaccineId}
							placeholder="e.g. 1"
							class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-sm"
						/>
					</div>
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">Lot Number</label>
						<input
							type="text"
							bind:value={lotNumber}
							placeholder="e.g. AB12345"
							class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-sm"
						/>
					</div>
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">Manufacturer</label>
						<input
							type="text"
							bind:value={manufacturer}
							placeholder="e.g. Pfizer"
							class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-sm"
						/>
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
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">Notes</label>
						<input
							type="text"
							bind:value={notes}
							placeholder="Optional notes..."
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
						disabled={formSaving}
						class="px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-lg hover:bg-blue-700 disabled:opacity-50 transition-colors"
					>
						{formSaving ? 'Saving...' : 'Save Immunization'}
					</button>
				</div>
			</form>
		</div>
	{/if}

	{#if loading}
		<div class="flex justify-center py-12">
			<div class="w-6 h-6 border-2 border-blue-500 border-t-transparent rounded-full animate-spin"></div>
		</div>
	{:else if immunizations.length === 0}
		<div class="text-center py-12 text-gray-500 bg-white rounded-lg shadow-sm border border-gray-200">
			<p class="text-lg">No immunizations recorded</p>
			<p class="text-sm mt-1">No immunizations have been recorded for this patient yet.</p>
		</div>
	{:else}
		<div class="bg-white rounded-lg shadow-sm border border-gray-200 overflow-x-auto">
			<table class="min-w-full divide-y divide-gray-200">
				<thead class="bg-gray-50">
					<tr>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Date</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Vaccine</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Lot #</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Manufacturer</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Provider</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Notes</th>
					</tr>
				</thead>
				<tbody class="bg-white divide-y divide-gray-200">
					{#each immunizations as imm}
						<tr class="hover:bg-gray-50">
							<td class="px-4 py-3 text-sm text-gray-700 whitespace-nowrap">
								{formatDate(imm.dateof)}
							</td>
							<td class="px-4 py-3 text-sm text-gray-700">
								#{imm.immunization}
							</td>
							<td class="px-4 py-3 text-sm text-gray-700">
								{imm.lot_number || '—'}
							</td>
							<td class="px-4 py-3 text-sm text-gray-700">
								{imm.manufacturer || '—'}
							</td>
							<td class="px-4 py-3 text-sm text-gray-700">
								#{imm.provider}
							</td>
							<td class="px-4 py-3 text-sm text-gray-500 max-w-[200px] truncate">
								{#if imm.notes}
									<span title={String(imm.notes)}>{String(imm.notes).slice(0, 50)}{String(imm.notes).length > 50 ? '...' : ''}</span>
								{:else}
									—
								{/if}
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{/if}
</div>
