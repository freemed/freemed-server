<script lang="ts">
	import { api } from '$lib/api';

	interface FormTemplate {
		id: number;
		name: string;
		form_type: string;
		is_default: boolean;
		template_data: string | null;
		description: string | null;
	}

	interface FormTemplateForm {
		name: string;
		form_type: string;
		description: string;
		is_default: boolean;
		template_data: string;
	}

	let templates = $state<FormTemplate[]>([]);
	let loading = $state(true);
	let error = $state('');
	let showModal = $state(false);
	let editingTemplate = $state<FormTemplate | null>(null);
	let form = $state<FormTemplateForm>(emptyForm());
	let formError = $state('');
	let jsonError = $state('');
	let formSaving = $state(false);
	let deleteConfirm = $state<FormTemplate | null>(null);

	function emptyForm(): FormTemplateForm {
		return { name: '', form_type: 'encounter', description: '', is_default: false, template_data: '{}' };
	}

	function validateJson(val: string): boolean {
		if (!val.trim()) return true;
		try {
			JSON.parse(val);
			return true;
		} catch {
			return false;
		}
	}

	async function loadTemplates() {
		loading = true;
		error = '';
		try {
			templates = await api.get<FormTemplate[]>('/form-templates');
		} catch (e: any) {
			error = e.message || 'Failed to load form templates';
		} finally {
			loading = false;
		}
	}

	$effect(() => {
		loadTemplates();
	});

	function openCreate() {
		editingTemplate = null;
		form = emptyForm();
		formError = '';
		jsonError = '';
		showModal = true;
	}

	function openEdit(tpl: FormTemplate) {
		editingTemplate = tpl;
		form = {
			name: tpl.name,
			form_type: tpl.form_type,
			description: tpl.description || '',
			is_default: tpl.is_default,
			template_data: tpl.template_data || '{}',
		};
		formError = '';
		jsonError = '';
		showModal = true;
	}

	function closeModal() {
		showModal = false;
		editingTemplate = null;
		formSaving = false;
		formError = '';
		jsonError = '';
	}

	function formatJson() {
		try {
			const obj = JSON.parse(form.template_data);
			form.template_data = JSON.stringify(obj, null, 2);
			jsonError = '';
		} catch {
			jsonError = 'Invalid JSON';
		}
	}

	async function saveTemplate() {
		formError = '';
		jsonError = '';

		if (!form.name.trim()) {
			formError = 'Name is required';
			return;
		}

		if (!validateJson(form.template_data)) {
			jsonError = 'Template data is not valid JSON';
			return;
		}

		formSaving = true;
		try {
			const payload: Record<string, unknown> = {
				name: form.name,
				form_type: form.form_type || 'encounter',
				description: form.description || null,
				is_default: form.is_default,
				template_data: form.template_data || null,
			};

			if (editingTemplate) {
				await api.put(`/form-templates/${editingTemplate.id}`, payload);
			} else {
				await api.post('/form-templates', payload);
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
		try {
			await api.del(`/form-templates/${deleteConfirm.id}`);
			deleteConfirm = null;
			await loadTemplates();
		} catch (e: any) {
			error = e.message || 'Delete failed';
			deleteConfirm = null;
		}
	}
</script>

<div class="max-w-5xl mx-auto">
	<div class="flex items-center justify-between mb-6">
		<h1 class="text-2xl font-bold text-gray-900">Form Templates</h1>
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
			<p class="text-lg">No form templates found</p>
		</div>
	{:else}
		<div class="bg-white rounded-lg shadow-sm border border-gray-200 overflow-x-auto">
			<table class="w-full text-sm">
				<thead class="bg-gray-50 border-b border-gray-200">
					<tr>
						<th class="text-left px-4 py-3 font-semibold text-gray-600">Name</th>
						<th class="text-left px-4 py-3 font-semibold text-gray-600">Type</th>
						<th class="text-center px-4 py-3 font-semibold text-gray-600">Default</th>
						<th class="text-right px-4 py-3 font-semibold text-gray-600">Actions</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-gray-100">
					{#each templates as tpl (tpl.id)}
						<tr class="hover:bg-gray-50">
							<td class="px-4 py-3 font-medium text-gray-900">{tpl.name}</td>
							<td class="px-4 py-3 text-gray-600">
								<span class="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-gray-100 text-gray-700">
									{tpl.form_type}
								</span>
							</td>
							<td class="px-4 py-3 text-center">
								{#if tpl.is_default}
									<span class="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-700">Default</span>
								{:else}
									<span class="text-gray-400">—</span>
								{/if}
							</td>
							<td class="px-4 py-3 text-right">
								<button
									onclick={() => openEdit(tpl)}
									class="px-3 py-1.5 text-xs font-medium text-blue-600 hover:bg-blue-50 rounded-md transition-colors"
								>
									Edit
								</button>
								<button
									onclick={() => (deleteConfirm = tpl)}
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

<!-- Add/Edit Modal -->
{#if showModal}
	<!-- svelte-ignore a11y_interactive_supports_focus -->
	<div
		class="fixed inset-0 z-50 flex items-center justify-center bg-black/40"
		onclick={closeModal}
		onkeydown={(e) => e.key === 'Escape' && closeModal()}
		role="dialog"
		aria-modal="true"
	>
		<div
			class="bg-white rounded-xl shadow-xl w-full max-w-2xl mx-4 max-h-[90vh] flex flex-col"
			onclick={(e) => e.stopPropagation()}
			onkeydown={(e) => e.key === 'Escape' && closeModal()}
			role="presentation"
		>
			<div class="px-6 py-4 border-b border-gray-200 flex items-center justify-between shrink-0">
				<h2 class="text-lg font-semibold text-gray-900">
					{editingTemplate ? 'Edit Form Template' : 'Add Form Template'}
				</h2>
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

			<div class="px-6 py-4 space-y-4 overflow-y-auto">
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
						placeholder="Form template name"
						required
					/>
				</div>

				<div class="grid grid-cols-2 gap-4">
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">Form Type</label>
						<input
							type="text"
							bind:value={form.form_type}
							class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
							placeholder="e.g. encounter"
						/>
					</div>
					<div class="flex items-end pb-2">
						<label class="flex items-center gap-2 cursor-pointer">
							<input
								type="checkbox"
								bind:checked={form.is_default}
								class="w-4 h-4 text-blue-600 border-gray-300 rounded focus:ring-blue-500"
							/>
							<span class="text-sm font-medium text-gray-700">Set as default</span>
						</label>
					</div>
				</div>

				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">Description</label>
					<textarea
						bind:value={form.description}
						class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
						placeholder="Description"
						rows="2"
					></textarea>
				</div>

				<div>
					<div class="flex items-center justify-between mb-1">
						<label class="block text-sm font-medium text-gray-700">Template Data (JSON)</label>
						<button
							onclick={formatJson}
							class="px-2 py-0.5 text-xs font-medium text-blue-600 hover:bg-blue-50 rounded transition-colors"
						>
							Format JSON
						</button>
					</div>
					<textarea
						bind:value={form.template_data}
						class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm font-mono focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
						placeholder="JSON template data"
						rows="12"
						spellcheck="false"
					></textarea>
					{#if jsonError}
						<p class="text-red-600 text-xs mt-1">{jsonError}</p>
					{/if}
				</div>
			</div>

			<div class="px-6 py-4 border-t border-gray-200 flex justify-end gap-3 shrink-0">
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
	<!-- svelte-ignore a11y_interactive_supports_focus -->
	<div
		class="fixed inset-0 z-50 flex items-center justify-center bg-black/40"
		onclick={() => (deleteConfirm = null)}
		onkeydown={(e) => e.key === 'Escape' && (deleteConfirm = null)}
		role="dialog"
		aria-modal="true"
	>
		<div
			class="bg-white rounded-xl shadow-xl w-full max-w-sm mx-4 p-6"
			onclick={(e) => e.stopPropagation()}
			onkeydown={(e) => e.key === 'Escape' && (deleteConfirm = null)}
			role="presentation"
		>
			<h3 class="text-lg font-semibold text-gray-900 mb-2">Confirm Delete</h3>
			<p class="text-sm text-gray-600 mb-6">
				Are you sure you want to delete template <strong>{deleteConfirm.name}</strong>? This action cannot be undone.
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
					class="px-4 py-2 text-sm font-medium text-white bg-red-600 rounded-lg hover:bg-red-700 transition-colors"
				>
					Delete
				</button>
			</div>
		</div>
	</div>
{/if}
