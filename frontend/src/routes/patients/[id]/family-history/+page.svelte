<script lang="ts">
	import { page } from '$app/stores';
	import { api } from '$lib/api';

	interface FamilyHistoryRecord {
		id: number;
		patient: number;
		relationship: string;
		condition_name: string;
		icd10_code: string;
		onset_age: number;
		deceased: boolean;
		notes: string | null;
		user: number;
		active: string;
		created_at: string;
		updated_at: string;
		deleted_at: string | null;
	}

	interface FamilyHistoryForm {
		relationship: string;
		condition_name: string;
		icd10_code: string;
		onset_age: number;
		deceased: boolean;
		notes: string;
	}

	const RELATIONSHIPS = [
		'Mother',
		'Father',
		'Sister',
		'Brother',
		'Son',
		'Daughter',
		'Maternal Grandmother',
		'Maternal Grandfather',
		'Paternal Grandmother',
		'Paternal Grandfather',
		'Aunt',
		'Uncle',
		'Cousin',
		'Other'
	];

	// Patient ID from route param
	let patientId = $state('');

	// Data state
	let records = $state<FamilyHistoryRecord[]>([]);
	let loading = $state(true);
	let error = $state('');

	// Modal state
	let showCreateModal = $state(false);
	let showEditModal = $state(false);
	let showDeleteConfirm = $state(false);
	let editingRecord = $state<FamilyHistoryRecord | null>(null);
	let deletingRecord = $state<FamilyHistoryRecord | null>(null);
	let formSaving = $state(false);
	let formError = $state('');
	let form: FamilyHistoryForm = $state({
		relationship: '',
		condition_name: '',
		icd10_code: '',
		onset_age: 0,
		deceased: false,
		notes: ''
	});

	// Load on mount
	$effect(() => {
		const id = $page.params.id;
		if (id) {
			patientId = id;
			loadRecords(id);
		}
	});

	async function loadRecords(id: string) {
		loading = true;
		error = '';
		try {
			const data = await api.get<FamilyHistoryRecord[]>(`/patient/${id}/family-history`);
			records = data || [];
		} catch (e: any) {
			error = e.message || 'Failed to load family history';
		} finally {
			loading = false;
		}
	}

	// --- Create Modal ---
	function openCreateModal() {
		form = {
			relationship: '',
			condition_name: '',
			icd10_code: '',
			onset_age: 0,
			deceased: false,
			notes: ''
		};
		formError = '';
		showCreateModal = true;
	}

	// --- Edit Modal ---
	function openEditModal(record: FamilyHistoryRecord) {
		editingRecord = record;
		form = {
			relationship: record.relationship,
			condition_name: record.condition_name,
			icd10_code: record.icd10_code,
			onset_age: record.onset_age,
			deceased: record.deceased,
			notes: record.notes || ''
		};
		formError = '';
		showEditModal = true;
	}

	function closeModal() {
		showCreateModal = false;
		showEditModal = false;
		editingRecord = null;
		formError = '';
	}

	function onModalBackdropClick(e: MouseEvent) {
		if (e.target === e.currentTarget) {
			closeModal();
		}
	}

	function onModalKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') {
			closeModal();
		}
	}

	async function handleCreate() {
		formError = '';

		if (!form.relationship) {
			formError = 'Relationship is required.';
			return;
		}
		if (!form.condition_name.trim()) {
			formError = 'Condition name is required.';
			return;
		}

		formSaving = true;
		try {
			await api.post(`/patient/${patientId}/family-history`, {
				relationship: form.relationship,
				condition_name: form.condition_name.trim(),
				icd10_code: form.icd10_code.trim(),
				onset_age: form.onset_age,
				deceased: form.deceased,
				notes: form.notes.trim()
			});
			closeModal();
			await loadRecords(patientId);
		} catch (e: any) {
			formError = e.message || 'Failed to create record';
		} finally {
			formSaving = false;
		}
	}

	async function handleUpdate() {
		formError = '';
		if (!editingRecord) return;

		if (!form.relationship) {
			formError = 'Relationship is required.';
			return;
		}
		if (!form.condition_name.trim()) {
			formError = 'Condition name is required.';
			return;
		}

		formSaving = true;
		try {
			await api.put(`/patient/${patientId}/family-history/${editingRecord.id}`, {
				relationship: form.relationship,
				condition_name: form.condition_name.trim(),
				icd10_code: form.icd10_code.trim(),
				onset_age: form.onset_age,
				deceased: form.deceased,
				notes: form.notes.trim()
			});
			closeModal();
			await loadRecords(patientId);
		} catch (e: any) {
			formError = e.message || 'Failed to update record';
		} finally {
			formSaving = false;
		}
	}

	// --- Delete ---
	function openDeleteConfirm(record: FamilyHistoryRecord) {
		deletingRecord = record;
		showDeleteConfirm = true;
	}

	function closeDeleteConfirm() {
		showDeleteConfirm = false;
		deletingRecord = null;
	}

	async function handleDelete() {
		if (!deletingRecord) return;
		formSaving = true;
		try {
			await api.del(`/patient/${patientId}/family-history/${deletingRecord.id}`);
			closeDeleteConfirm();
			await loadRecords(patientId);
		} catch (e: any) {
			error = e.message || 'Failed to delete record';
		} finally {
			formSaving = false;
		}
	}

	// --- Helpers ---
	function formatDate(dateStr: string | null): string {
		if (!dateStr) return '—';
		try {
			const d = new Date(dateStr);
			if (isNaN(d.getTime())) return dateStr;
			return d.toLocaleDateString('en-US', {
				year: 'numeric',
				month: 'short',
				day: 'numeric'
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
			<h1 class="text-2xl font-bold text-gray-900">Family Health History</h1>
			<p class="text-sm text-gray-500 mt-1">Patient #{patientId}</p>
		</div>
		<div class="flex items-center gap-3">
			<button
				onclick={openCreateModal}
				class="px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded-lg hover:bg-blue-700 transition-colors"
			>
				+ Add Family History
			</button>
			<a
				href="/patients/{patientId}"
				class="text-sm text-blue-600 hover:text-blue-800 font-medium transition-colors"
			>
				&larr; Back to Patient
			</a>
		</div>
	</div>

	<!-- Content -->
	<div class="bg-white rounded-lg shadow-sm border border-gray-200">
		<div class="p-6">
			{#if loading}
				<div class="flex justify-center py-12">
					<div
						class="w-6 h-6 border-2 border-blue-500 border-t-transparent rounded-full animate-spin"
					></div>
				</div>
			{:else if error}
				<div class="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg text-sm">
					{error}
				</div>
			{:else if records.length === 0}
				<div class="text-center py-12 text-gray-500">
					<p class="text-lg">No family health history recorded</p>
					<p class="text-sm mt-1">
						Click "Add Family History" to record a condition for a family member.
					</p>
				</div>
			{:else}
				<div class="overflow-x-auto">
					<table class="w-full text-sm">
						<thead>
							<tr class="border-b border-gray-200">
								<th class="text-left py-2 px-3 font-semibold text-gray-600">Relationship</th>
								<th class="text-left py-2 px-3 font-semibold text-gray-600">Condition</th>
								<th class="text-left py-2 px-3 font-semibold text-gray-600">ICD-10</th>
								<th class="text-left py-2 px-3 font-semibold text-gray-600">Onset Age</th>
								<th class="text-left py-2 px-3 font-semibold text-gray-600">Deceased</th>
								<th class="text-left py-2 px-3 font-semibold text-gray-600">Recorded</th>
								<th class="text-right py-2 px-3 font-semibold text-gray-600">Actions</th>
							</tr>
						</thead>
						<tbody class="divide-y divide-gray-100">
							{#each records as record (record.id)}
								<tr class="hover:bg-gray-50 transition-colors">
									<td class="py-2.5 px-3 text-gray-900 font-medium">{record.relationship}</td>
									<td class="py-2.5 px-3 text-gray-700">
										{record.condition_name}
										{#if record.notes}
											<p class="text-xs text-gray-500 mt-0.5">{record.notes}</p>
										{/if}
									</td>
									<td class="py-2.5 px-3 text-gray-600 font-mono text-xs">{record.icd10_code || '—'}</td>
									<td class="py-2.5 px-3 text-gray-600">{record.onset_age > 0 ? record.onset_age : '—'}</td>
									<td class="py-2.5 px-3 text-gray-600">{record.deceased ? 'Yes' : 'No'}</td>
									<td class="py-2.5 px-3 text-gray-500 text-xs">{formatDate(record.created_at)}</td>
									<td class="py-2.5 px-3 text-right">
										<button
											onclick={() => openEditModal(record)}
											class="text-blue-600 hover:text-blue-800 text-xs font-medium mr-3 transition-colors"
										>
											Edit
										</button>
										<button
											onclick={() => openDeleteConfirm(record)}
											class="text-red-600 hover:text-red-800 text-xs font-medium transition-colors"
										>
											Delete
										</button>
									</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			{/if}
		</div>
	</div>
</div>

<!-- Create/Edit Modal -->
{#if showCreateModal || showEditModal}
	<!-- svelte-ignore a11y_no_static_element_interactions -->
	<div
		class="fixed inset-0 z-50 flex items-center justify-center bg-black/40"
		onclick={onModalBackdropClick}
		onkeydown={onModalKeydown}
		role="dialog"
		aria-modal="true"
		aria-label={showCreateModal ? 'Add Family History' : 'Edit Family History'}
		tabindex="-1"
	>
		<div class="bg-white rounded-xl shadow-xl w-full max-w-lg mx-4 overflow-hidden">
			<!-- Header -->
			<div class="flex items-center justify-between px-6 py-4 border-b border-gray-200">
				<h2 class="text-lg font-semibold text-gray-900">
					{showCreateModal ? 'Add Family History' : 'Edit Family History'}
				</h2>
				<button
					onclick={closeModal}
					class="text-gray-400 hover:text-gray-600 transition-colors p-1 rounded-lg hover:bg-gray-100"
					aria-label="Close"
				>
					<svg
						xmlns="http://www.w3.org/2000/svg"
						class="h-5 w-5"
						viewBox="0 0 20 20"
						fill="currentColor"
					>
						<path
							fill-rule="evenodd"
							d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z"
							clip-rule="evenodd"
						/>
					</svg>
				</button>
			</div>

			<!-- Body -->
			<div class="px-6 py-4 space-y-4">
				{#if formError}
					<div class="bg-red-50 border border-red-200 text-red-700 px-3 py-2 rounded-lg text-sm">
						{formError}
					</div>
				{/if}

				<div>
					<label for="fhRelationship" class="block text-sm font-medium text-gray-700 mb-1">
						Relationship
					</label>
					<select
						id="fhRelationship"
						bind:value={form.relationship}
						class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none transition-colors"
					>
						<option value="">-- Select Relationship --</option>
						{#each RELATIONSHIPS as rel}
							<option value={rel}>{rel}</option>
						{/each}
					</select>
				</div>

				<div>
					<label for="fhCondition" class="block text-sm font-medium text-gray-700 mb-1">
						Condition Name
					</label>
					<input
						id="fhCondition"
						type="text"
						bind:value={form.condition_name}
						placeholder="e.g. Hypertension, Diabetes, Cancer"
						class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none transition-colors"
					/>
				</div>

				<div>
					<label for="fhIcd10" class="block text-sm font-medium text-gray-700 mb-1">
						ICD-10 Code
					</label>
					<input
						id="fhIcd10"
						type="text"
						bind:value={form.icd10_code}
						placeholder="e.g. I10, E11.9"
						class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none transition-colors"
					/>
				</div>

				<div class="grid grid-cols-2 gap-4">
					<div>
						<label for="fhOnsetAge" class="block text-sm font-medium text-gray-700 mb-1">
							Onset Age
						</label>
						<input
							id="fhOnsetAge"
							type="number"
							min="0"
							bind:value={form.onset_age}
							class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none transition-colors"
						/>
					</div>

					<div class="flex items-center pt-6">
						<label class="flex items-center gap-2 cursor-pointer">
							<input
								type="checkbox"
								bind:checked={form.deceased}
								class="w-4 h-4 text-blue-600 border-gray-300 rounded focus:ring-blue-500"
							/>
							<span class="text-sm text-gray-700">Deceased</span>
						</label>
					</div>
				</div>

				<div>
					<label for="fhNotes" class="block text-sm font-medium text-gray-700 mb-1">
						Notes
					</label>
					<textarea
						id="fhNotes"
						bind:value={form.notes}
						rows="3"
						placeholder="Additional details..."
						class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none transition-colors resize-none"
					></textarea>
				</div>
			</div>

			<!-- Footer -->
			<div class="px-6 py-4 bg-gray-50 border-t border-gray-200 flex justify-end gap-3">
				<button
					onclick={closeModal}
					class="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors"
				>
					Cancel
				</button>
				<button
					onclick={showCreateModal ? handleCreate : handleUpdate}
					disabled={formSaving}
					class="px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors flex items-center gap-2"
				>
					{#if formSaving}
						<div
							class="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin"
						></div>
						Saving...
					{:else}
						Save
					{/if}
				</button>
			</div>
		</div>
	</div>
{/if}

<!-- Delete Confirmation Modal -->
{#if showDeleteConfirm && deletingRecord}
	<!-- svelte-ignore a11y_no_static_element_interactions -->
	<div
		class="fixed inset-0 z-50 flex items-center justify-center bg-black/40"
		onclick={() => closeDeleteConfirm()}
		onkeydown={(e) => { if (e.key === 'Escape') closeDeleteConfirm(); }}
		role="dialog"
		aria-modal="true"
		aria-label="Confirm Delete"
		tabindex="-1"
	>
		<div class="bg-white rounded-xl shadow-xl w-full max-w-md mx-4 overflow-hidden">
			<div class="px-6 py-4 border-b border-gray-200">
				<h2 class="text-lg font-semibold text-gray-900">Confirm Delete</h2>
			</div>
			<div class="px-6 py-4">
				<p class="text-sm text-gray-600">
					Are you sure you want to remove <strong>{deletingRecord.relationship}</strong> — <em>{deletingRecord.condition_name}</em> from the family history?
				</p>
			</div>
			<div class="px-6 py-4 bg-gray-50 border-t border-gray-200 flex justify-end gap-3">
				<button
					onclick={closeDeleteConfirm}
					class="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors"
				>
					Cancel
				</button>
				<button
					onclick={handleDelete}
					disabled={formSaving}
					class="px-4 py-2 text-sm font-medium text-white bg-red-600 rounded-lg hover:bg-red-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors flex items-center gap-2"
				>
					{#if formSaving}
						<div
							class="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin"
						></div>
						Deleting...
					{:else}
						Delete
					{/if}
				</button>
			</div>
		</div>
	</div>
{/if}
