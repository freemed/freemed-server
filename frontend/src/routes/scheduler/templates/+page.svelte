<script lang="ts">
	import { api } from '$lib/api';
	import { toast } from '$lib/stores/toast.svelte';

	// --- Types ---

	interface Template {
		id: number;
		atname: string;
		atduration: number;
		atequipment: string | null;
		atcolor: string;
		created_at: string;
		updated_at: string;
	}

	interface TemplateForm {
		name: string;
		duration: number;
		equipment: string;
		color: string;
	}

	// --- State ---

	let templates = $state<Template[]>([]);
	let loading = $state(true);
	let error = $state('');

	let showModal = $state(false);
	let editingTemplate = $state<Template | null>(null);
	let form = $state<TemplateForm>(emptyForm());
	let formError = $state('');
	let formSaving = $state(false);

	let deleteConfirm = $state<Template | null>(null);
	let deleting = $state(false);

	// --- Helpers ---

	function emptyForm(): TemplateForm {
		return { name: '', duration: 30, equipment: '', color: '#3b82f6' };
	}

	// --- Data loading ---

	async function loadTemplates() {
		loading = true;
		error = '';
		try {
			templates = await api.get<Template[]>('/scheduler/templates');
		} catch (e: any) {
			error = e.message || 'Failed to load templates';
		} finally {
			loading = false;
		}
	}

	$effect(() => {
		loadTemplates();
	});

	// --- Modal handlers ---

	function openCreate() {
		editingTemplate = null;
		form = emptyForm();
		formError = '';
		showModal = true;
	}

	function openEdit(t: Template) {
		editingTemplate = t;
		form = {
			name: t.atname,
			duration: t.atduration,
			equipment: t.atequipment || '',
			color: t.atcolor || '#3b82f6',
		};
		formError = '';
		showModal = true;
	}

	function closeModal() {
		showModal = false;
		editingTemplate = null;
		formSaving = false;
		formError = '';
	}

	async function saveTemplate() {
		formError = '';
		if (!form.name.trim()) {
			formError = 'Name is required.';
			return;
		}
		if (form.duration <= 0) {
			formError = 'Duration must be greater than 0.';
			return;
		}
		formSaving = true;
		try {
			const payload = {
				name: form.name.trim(),
				duration: form.duration,
				equipment: form.equipment.trim() || null,
				color: form.color.trim() || '#3b82f6',
			};
			if (editingTemplate) {
				await api.put(`/scheduler/templates/${editingTemplate.id}`, payload);
				toast.success('Template updated.');
			} else {
				await api.post('/scheduler/templates', payload);
				toast.success('Template created.');
			}
			closeModal();
			await loadTemplates();
		} catch (e: any) {
			formError = e.message || 'Save failed';
		} finally {
			formSaving = false;
		}
	}

	async function confirmDelete() {
		if (!deleteConfirm) return;
		deleting = true;
		try {
			await api.del(`/scheduler/templates/${deleteConfirm.id}`);
			toast.success('Template deleted.');
			deleteConfirm = null;
			await loadTemplates();
		} catch (e: any) {
			toast.error(e.message || 'Delete failed');
			deleteConfirm = null;
		} finally {
			deleting = false;
		}
	}

	// --- Form helpers ---

	function handleModalBackdrop(e: MouseEvent) {
		if (e.target === e.currentTarget) closeModal();
	}

	function handleModalKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') closeModal();
	}
</script>

<div class="max-w-5xl mx-auto">
	<div class="flex items-center justify-between mb-6">
		<div>
			<h1 class="text-2xl font-bold text-gray-900">Appointment Templates</h1>
			<p class="text-sm text-gray-500 mt-1">
				Manage recurring appointment templates with name, duration, equipment, and color.
			</p>
		</div>
		<button
			onclick={openCreate}
			class="px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded-lg hover:bg-blue-700 transition-colors"
		>
			+ Add Template
		</button>
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
	{:else if templates.length === 0}
		<div class="text-center py-12 text-gray-500">
			<p class="text-lg">No templates found</p>
			<p class="text-sm mt-1">Create a template to get started.</p>
		</div>
	{:else}
		<div class="bg-white rounded-lg shadow-sm border border-gray-200 overflow-x-auto">
			<table class="w-full text-sm">
				<thead class="bg-gray-50 border-b border-gray-200">
					<tr>
						<th class="text-left px-4 py-3 font-semibold text-gray-600">Color</th>
						<th class="text-left px-4 py-3 font-semibold text-gray-600">Name</th>
						<th class="text-left px-4 py-3 font-semibold text-gray-600">Duration</th>
						<th class="text-left px-4 py-3 font-semibold text-gray-600">Equipment</th>
						<th class="text-right px-4 py-3 font-semibold text-gray-600">Actions</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-gray-100">
					{#each templates as t (t.id)}
						<tr class="hover:bg-gray-50">
							<td class="px-4 py-3">
								<span
									class="inline-block w-5 h-5 rounded-full border border-gray-300"
									style="background-color: {t.atcolor || '#3b82f6'}"
								></span>
							</td>
							<td class="px-4 py-3 font-medium text-gray-900">{t.atname}</td>
							<td class="px-4 py-3 text-gray-600">{t.atduration} min</td>
							<td class="px-4 py-3 text-gray-600 max-w-xs truncate">{t.atequipment || '—'}</td>
							<td class="px-4 py-3 text-right">
								<button
									onclick={() => openEdit(t)}
									class="px-3 py-1.5 text-xs font-medium text-blue-600 hover:bg-blue-50 rounded-md transition-colors"
								>
									Edit
								</button>
								<button
									onclick={() => (deleteConfirm = t)}
									class="px-3 py-1.5 text-xs font-medium text-red-600 hover:bg-red-50 rounded-md transition-colors ml-1"
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

<!-- Create/Edit Modal -->
{#if showModal}
	<!-- svelte-ignore a11y_no_static_element_interactions -->
	<div
		class="fixed inset-0 z-50 flex items-center justify-center bg-black/40"
		onclick={handleModalBackdrop}
		onkeydown={handleModalKeydown}
		role="dialog"
		aria-modal="true"
		aria-label={editingTemplate ? 'Edit template' : 'Create template'}
		tabindex="-1"
	>
		<div
			class="bg-white rounded-xl shadow-xl w-full max-w-lg mx-4 overflow-hidden"
			onclick={(e) => e.stopPropagation()}
			onkeydown={handleModalKeydown}
			role="presentation"
		>
			<!-- Header -->
			<div class="flex items-center justify-between px-6 py-4 border-b border-gray-200">
				<h2 class="text-lg font-semibold text-gray-900">
					{editingTemplate ? 'Edit Template' : 'Add Template'}
				</h2>
				<button
					onclick={closeModal}
					class="text-gray-400 hover:text-gray-600 transition-colors p-1 rounded-lg hover:bg-gray-100"
					aria-label="Close"
				>
					<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
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
					<label class="block text-sm font-medium text-gray-700 mb-1">Name *</label>
					<input
						type="text"
						bind:value={form.name}
						class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
						placeholder="e.g. Annual Physical"
						required
					/>
				</div>

				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">Duration (minutes) *</label>
					<input
						type="number"
						bind:value={form.duration}
						class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
						min="5"
						step="5"
						placeholder="30"
						required
					/>
					<p class="text-xs text-gray-400 mt-1">Typical durations: 15, 30, 45, 60 minutes</p>
				</div>

				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">Equipment</label>
					<input
						type="text"
						bind:value={form.equipment}
						class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
						placeholder="e.g. Exam Room 1, Ultrasound"
					/>
				</div>

				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">Color</label>
					<div class="flex items-center gap-2">
						<input
							type="color"
							bind:value={form.color}
							class="w-10 h-10 rounded border border-gray-300 cursor-pointer"
						/>
						<input
							type="text"
							bind:value={form.color}
							class="flex-1 px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
							placeholder="#3b82f6"
						/>
					</div>
				</div>
			</div>

			<!-- Footer -->
			<div class="px-6 py-4 border-t border-gray-200 flex justify-end gap-3">
				<button
					onclick={closeModal}
					class="px-4 py-2 text-sm font-medium text-gray-700 bg-gray-100 rounded-lg hover:bg-gray-200 transition-colors"
				>
					Cancel
				</button>
				<button
					onclick={saveTemplate}
					disabled={formSaving || !form.name.trim()}
					class="px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
				>
					{formSaving ? 'Saving...' : editingTemplate ? 'Update' : 'Create'}
				</button>
			</div>
		</div>
	</div>
{/if}

<!-- Delete Confirmation Modal -->
{#if deleteConfirm}
	<!-- svelte-ignore a11y_no_static_element_interactions -->
	<div
		class="fixed inset-0 z-50 flex items-center justify-center bg-black/40"
		onclick={() => (deleteConfirm = null)}
		onkeydown={(e) => e.key === 'Escape' && (deleteConfirm = null)}
		role="dialog"
		aria-modal="true"
		aria-label="Confirm delete"
		tabindex="-1"
	>
		<div
			class="bg-white rounded-xl shadow-xl w-full max-w-sm mx-4 p-6"
			onclick={(e) => e.stopPropagation()}
			role="presentation"
		>
			<h3 class="text-lg font-semibold text-gray-900 mb-2">Confirm Delete</h3>
			<p class="text-sm text-gray-600 mb-6">
				Are you sure you want to delete template <strong>{deleteConfirm.atname}</strong>? This action cannot be undone.
			</p>
			<div class="flex justify-end gap-3">
				<button
					onclick={() => (deleteConfirm = null)}
					class="px-4 py-2 text-sm font-medium text-gray-700 bg-gray-100 rounded-lg hover:bg-gray-200 transition-colors"
				>
					Cancel
				</button>
				<button
					onclick={confirmDelete}
					disabled={deleting}
					class="px-4 py-2 text-sm font-medium text-white bg-red-600 rounded-lg hover:bg-red-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
				>
					{deleting ? 'Deleting...' : 'Delete'}
				</button>
			</div>
		</div>
	</div>
{/if}
