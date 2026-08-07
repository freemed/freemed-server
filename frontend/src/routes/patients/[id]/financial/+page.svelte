<script lang="ts">
	import { page } from '$app/stores';
	import { api } from '$lib/api';

	interface FinancialDemographic {
		id: number;
		created_at: string;
		updated_at: string;
		deleted_at: string | null;
		patient: number;
		income: number;
		id_type: string;
		id_issuer: string;
		id_number: string;
		id_expire: string;
		household_size: number;
		spouse: number;
		children: number;
		other_dependents: number;
		free_text: string | null;
		entry_desc: string;
		entry_ts: string;
		user: number;
		active: string;
	}

	let patientId = $derived($page.params.id || '');

	let records = $state<FinancialDemographic[]>([]);
	let loading = $state(true);
	let error = $state('');

	// Modal state
	let showModal = $state(false);
	let formSaving = $state(false);
	let formError = $state('');

	// Form fields
	let income = $state(0);
	let idType = $state('');
	let idIssuer = $state('');
	let idNumber = $state('');
	let idExpire = $state('');
	let householdSize = $state(1);
	let spouse = $state(0);
	let children = $state(0);
	let otherDependents = $state(0);
	let freeText = $state('');
	let entryDesc = $state('');

	$effect(() => {
		if (patientId) loadRecords();
	});

	async function loadRecords() {
		loading = true;
		error = '';
		try {
			const data = await api.get<FinancialDemographic[]>(`/patient/${patientId}/financial`);
			records = data || [];
		} catch (e: any) {
			error = e.message || 'Failed to load financial records';
		} finally {
			loading = false;
		}
	}

	function openModal() {
		income = 0;
		idType = '';
		idIssuer = '';
		idNumber = '';
		idExpire = '';
		householdSize = 1;
		spouse = 0;
		children = 0;
		otherDependents = 0;
		freeText = '';
		entryDesc = '';
		formError = '';
		showModal = true;
	}

	function closeModal() {
		showModal = false;
	}

	async function handleSubmit(e: SubmitEvent) {
		e.preventDefault();
		formError = '';
		formSaving = true;

		try {
			await api.post(`/patient/${patientId}/financial`, {
				income,
				id_type: idType,
				id_issuer: idIssuer,
				id_number: idNumber,
				id_expire: idExpire,
				household_size: householdSize,
				spouse,
				children,
				other_dependents: otherDependents,
				free_text: freeText || null,
				entry_desc: entryDesc,
			});
			closeModal();
			await loadRecords();
		} catch (e: any) {
			formError = e.message || 'Failed to save financial record';
		} finally {
			formSaving = false;
		}
	}

	function formatCurrency(val: number): string {
		return new Intl.NumberFormat('en-US', {
			style: 'currency',
			currency: 'USD',
			minimumFractionDigits: 0,
			maximumFractionDigits: 0,
		}).format(val);
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
	<!-- Header -->
	<div class="flex items-center justify-between mb-6">
		<div>
			<h1 class="text-2xl font-bold text-gray-900">Financial Demographics</h1>
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
				+ Add Record
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
			<div class="w-6 h-6 border-2 border-blue-500 border-t-transparent rounded-full animate-spin"></div>
		</div>

	<!-- Empty state -->
	{:else if records.length === 0}
		<div class="text-center py-12 text-gray-500 bg-white rounded-lg shadow-sm border border-gray-200">
			<p class="text-lg">No financial records</p>
			<p class="text-sm mt-1">No financial demographics have been recorded for this patient yet.</p>
		</div>

	<!-- Table -->
	{:else}
		<div class="bg-white rounded-lg shadow-sm border border-gray-200 overflow-x-auto">
			<table class="min-w-full divide-y divide-gray-200">
				<thead class="bg-gray-50">
					<tr>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Date</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Income</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Household Size</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">ID Type</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Description</th>
					</tr>
				</thead>
				<tbody class="bg-white divide-y divide-gray-200">
					{#each records as rec}
						<tr class="hover:bg-gray-50">
							<td class="px-4 py-3 text-sm text-gray-700 whitespace-nowrap">
								{formatDate(rec.entry_ts)}
							</td>
							<td class="px-4 py-3 text-sm text-gray-700 whitespace-nowrap">
								{formatCurrency(rec.income)}
							</td>
							<td class="px-4 py-3 text-sm text-gray-700 whitespace-nowrap">
								{rec.household_size}
							</td>
							<td class="px-4 py-3 text-sm text-gray-700 whitespace-nowrap">
								{rec.id_type || '—'}
							</td>
							<td class="px-4 py-3 text-sm text-gray-700 max-w-[300px] truncate">
								{rec.entry_desc || '—'}
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{/if}
</div>

<!-- Create Modal (Pattern A: custom overlay, no Bootstrap JS) -->
{#key showModal}
	{#if showModal}
		<!-- svelte-ignore a11y_no_static_element_interactions -->
		<div
			class="position-fixed top-0 start-0 w-100 h-100"
			style="z-index: 1055; background: rgba(0,0,0,0.5);"
			onclick={closeModal}
			onkeydown={closeModal}
		>
			<div
				class="position-absolute top-50 start-50 translate-middle"
				style="width: 90%; max-width: 700px;"
				onclick={(e: MouseEvent) => e.stopPropagation()}
				onkeydown={(e: KeyboardEvent) => e.stopPropagation()}
			>
				<div class="card shadow">
					<div class="card-header d-flex justify-content-between align-items-center">
						<h5 class="mb-0">Add Financial Record</h5>
						<button type="button" class="btn-close" onclick={closeModal}></button>
					</div>
					<div class="card-body" style="max-height: 70vh; overflow-y: auto;">
						{#if formError}
							<div class="alert alert-danger py-2 mb-3">{formError}</div>
						{/if}

						<form id="financial-form" onsubmit={handleSubmit}>
							<!-- Row 1: Income + ID Type -->
							<div class="row g-3 mb-3">
								<div class="col-md-6">
									<label class="form-label">Annual Income</label>
									<input
										type="number"
										class="form-control"
										bind:value={income}
										min="0"
										step="1"
										placeholder="0"
									/>
								</div>
								<div class="col-md-6">
									<label class="form-label">ID Type</label>
									<input
										type="text"
										class="form-control"
										bind:value={idType}
										placeholder="e.g. Driver's License, Passport"
									/>
								</div>
							</div>

							<!-- Row 2: ID Issuer + ID Number -->
							<div class="row g-3 mb-3">
								<div class="col-md-6">
									<label class="form-label">ID Issuer</label>
									<input
										type="text"
										class="form-control"
										bind:value={idIssuer}
										placeholder="e.g. State of California"
									/>
								</div>
								<div class="col-md-6">
									<label class="form-label">ID Number</label>
									<input
										type="text"
										class="form-control"
										bind:value={idNumber}
										placeholder="e.g. D1234567"
									/>
								</div>
							</div>

							<!-- ID Expiration -->
							<div class="mb-3">
								<label class="form-label">ID Expiration</label>
								<input
									type="date"
									class="form-control"
									bind:value={idExpire}
								/>
							</div>

							<!-- Row 3: Household Size + Spouse -->
							<div class="row g-3 mb-3">
								<div class="col-md-6">
									<label class="form-label">Household Size</label>
									<input
										type="number"
										class="form-control"
										bind:value={householdSize}
										min="1"
										step="1"
									/>
								</div>
								<div class="col-md-6">
									<label class="form-label">Spouse Count</label>
									<input
										type="number"
										class="form-control"
										bind:value={spouse}
										min="0"
										step="1"
									/>
								</div>
							</div>

							<!-- Row 4: Children + Other Dependents -->
							<div class="row g-3 mb-3">
								<div class="col-md-6">
									<label class="form-label">Children Count</label>
									<input
										type="number"
										class="form-control"
										bind:value={children}
										min="0"
										step="1"
									/>
								</div>
								<div class="col-md-6">
									<label class="form-label">Other Dependents</label>
									<input
										type="number"
										class="form-control"
										bind:value={otherDependents}
										min="0"
										step="1"
									/>
								</div>
							</div>

							<!-- Free Text -->
							<div class="mb-3">
								<label class="form-label">Free Text</label>
								<textarea
									class="form-control"
									bind:value={freeText}
									rows="2"
									placeholder="Additional notes..."
								></textarea>
							</div>

							<!-- Entry Description -->
							<div class="mb-3">
								<label class="form-label">Entry Description</label>
								<input
									type="text"
									class="form-control"
									bind:value={entryDesc}
									placeholder="Describe this financial record..."
								/>
							</div>
						</form>
					</div>
					<div class="card-footer text-end">
						<button
							type="button"
							class="btn btn-secondary"
							onclick={closeModal}
						>
							Cancel
						</button>
						<button
							type="submit"
							form="financial-form"
							class="btn btn-primary ms-2"
							disabled={formSaving}
						>
							{formSaving ? 'Saving...' : 'Save Record'}
						</button>
					</div>
				</div>
			</div>
		</div>
	{/if}
{/key}
