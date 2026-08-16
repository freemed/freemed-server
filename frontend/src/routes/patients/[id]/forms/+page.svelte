<script lang="ts">
	import { page } from '$app/stores';
	import { api } from '$lib/api';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
	import ErrorBanner from '$lib/components/ErrorBanner.svelte';
	import EmptyState from '$lib/components/EmptyState.svelte';

	interface Control {
		variable: string;
		name: string;
		type: string;
		default: string;
		options: string;
		limits: string;
		uuid: string;
	}

	interface FormResult {
		id: number;
		fr_timestamp: string;
		fr_template: string;
		fr_formname: string;
		fr_patient: number;
		user: number;
		active: string;
		created_at: string;
		updated_at: string;
		deleted_at: string | null;
	}

	interface FormDetail {
		form: FormResult;
		values: Record<string, string>;
		controls: Control[];
	}

	interface Template {
		id: number;
		name: string;
		description: string | null;
		form_type: string;
		template_data: string | null;
		is_default: boolean;
	}

	let patientId = $derived($page.params.id || '');

	// List state
	let forms = $state<FormResult[]>([]);
	let loading = $state(true);
	let error = $state('');

	// Templates (for create)
	let templates = $state<Template[]>([]);
	let templatesError = $state('');

	// Modal state
	let showModal = $state(false);
	let modalMode = $state<'create' | 'edit'>('create');
	let editingId = $state<number | null>(null);
	let selectedTemplateId = $state<number | null>(null);
	let controls = $state<Control[]>([]);
	let controlsLoading = $state(false);
	let controlsError = $state('');
	let formValues = $state<Record<string, string | string[]>>({});
	let saving = $state(false);
	let formError = $state('');

	$effect(() => {
		if (patientId) {
			loadForms();
			loadTemplates();
		}
	});

	async function loadForms() {
		loading = true;
		error = '';
		try {
			const data = await api.get<FormResult[]>(`/patient/${patientId}/forms`);
			forms = data || [];
		} catch (e: any) {
			error = e.message || 'Failed to load forms';
		} finally {
			loading = false;
		}
	}

	async function loadTemplates() {
		templatesError = '';
		try {
			const data = await api.get<Template[]>('/form-templates/');
			templates = data || [];
		} catch (e: any) {
			templatesError = e.message || 'Failed to load form templates';
		}
	}

	function openCreate() {
		modalMode = 'create';
		editingId = null;
		selectedTemplateId = null;
		controls = [];
		formValues = {};
		controlsError = '';
		formError = '';
		showModal = true;
	}

	async function openEdit(form: FormResult) {
		modalMode = 'edit';
		editingId = form.id;
		selectedTemplateId = null;
		controls = [];
		formValues = {};
		controlsError = '';
		formError = '';
		controlsLoading = true;
		showModal = true;
		try {
			const data = await api.get<FormDetail>(`/patient/${patientId}/forms/${form.id}`);
			controls = data.controls || [];
			initFormValues(controls, data.values || {});
		} catch (e: any) {
			controlsError = e.message || 'Failed to load form';
		} finally {
			controlsLoading = false;
		}
	}

	async function onTemplateSelect() {
		controls = [];
		formValues = {};
		controlsError = '';
		if (!selectedTemplateId) return;
		controlsLoading = true;
		try {
			const data = await api.get<Control[]>(`/form-templates/${selectedTemplateId}/controls`);
			controls = data || [];
			initFormValues(controls, {});
		} catch (e: any) {
			controlsError = e.message || 'Failed to load template controls';
		} finally {
			controlsLoading = false;
		}
	}

	function initFormValues(ctrls: Control[], existing: Record<string, string>) {
		const v: Record<string, string | string[]> = {};
		for (const c of ctrls) {
			const raw = existing[c.name] !== undefined ? existing[c.name] : c.default;
			if (c.type === 'multiple') {
				v[c.variable] = raw ? raw.split(',') : [];
			} else if (c.type === 'boolean') {
				v[c.variable] = raw === '1' || raw === 'true' ? '1' : '0';
			} else {
				v[c.variable] = raw || '';
			}
		}
		formValues = v;
	}

	function stringValue(variable: string): string {
		const v = formValues[variable];
		return typeof v === 'string' ? v : '';
	}

	function optionsFor(c: Control): string[] {
		return c.options ? c.options.split('|') : [];
	}

	function multiSelected(variable: string, opt: string): boolean {
		const v = formValues[variable];
		return Array.isArray(v) && v.includes(opt);
	}

	function onMultiChange(variable: string, e: Event) {
		const sel = e.currentTarget as HTMLSelectElement;
		formValues[variable] = Array.from(sel.selectedOptions).map((o) => o.value);
	}

	function closeModal() {
		showModal = false;
		formError = '';
		controlsError = '';
	}

	function onModalBackdropClick(e: MouseEvent) {
		if (e.target === e.currentTarget) closeModal();
	}

	function onModalKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') closeModal();
	}

	async function handleSave() {
		formError = '';
		saving = true;
		try {
			const payload: Record<string, string> = {};
			for (const c of controls) {
				const v = formValues[c.variable];
				payload[c.variable] = Array.isArray(v) ? v.join(',') : (v ?? '');
			}
			if (modalMode === 'create') {
				await api.post(`/patient/${patientId}/forms`, {
					template_id: selectedTemplateId,
					values: payload
				});
			} else {
				await api.put(`/patient/${patientId}/forms/${editingId}`, { values: payload });
			}
			closeModal();
			await loadForms();
		} catch (e: any) {
			formError = e.message || 'Failed to save form';
		} finally {
			saving = false;
		}
	}

	async function handleDelete(form: FormResult) {
		if (!confirm(`Delete form "${form.fr_formname}"?`)) return;
		try {
			await api.del(`/patient/${patientId}/forms/${form.id}`);
			await loadForms();
		} catch (e: any) {
			alert(e.message || 'Failed to delete form');
		}
	}

	function formatDate(dateStr: string | null): string {
		if (!dateStr) return '—';
		const d = new Date(dateStr);
		if (isNaN(d.getTime())) return dateStr;
		return d.toLocaleDateString('en-US', { year: 'numeric', month: 'short', day: 'numeric' });
	}

	function controlLabel(c: Control): string {
		return c.name || c.variable;
	}
</script>

<div class="max-w-4xl mx-auto">
	<!-- Header -->
	<div class="flex items-center justify-between mb-6">
		<div>
			<h1 class="text-2xl font-bold text-gray-900">Forms</h1>
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
				onclick={openCreate}
				class="px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded-lg hover:bg-blue-700 transition-colors"
			>
				+ New Form
			</button>
		</div>
	</div>

	<!-- List -->
	{#if loading}
		<LoadingSpinner message="Loading forms..." />
	{:else if error}
		<ErrorBanner message={error} onRetry={loadForms} />
	{:else if forms.length === 0}
		<EmptyState
			title="No forms recorded"
			message="Click New Form to fill out a clinical form for this patient."
			actionLabel="New Form"
			onAction={openCreate}
		/>
	{:else}
		<div class="bg-white rounded-lg shadow-sm border border-gray-200 overflow-hidden">
			<table class="w-full text-sm">
				<thead>
					<tr class="border-b border-gray-200 bg-gray-50">
						<th class="text-left py-3 px-4 font-semibold text-gray-600">Date</th>
						<th class="text-left py-3 px-4 font-semibold text-gray-600">Form</th>
						<th class="text-left py-3 px-4 font-semibold text-gray-600">Template</th>
						<th class="text-right py-3 px-4 font-semibold text-gray-600">Actions</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-gray-100">
					{#each forms as form (form.id)}
						<tr class="hover:bg-gray-50 transition-colors">
							<td class="py-2.5 px-4 text-gray-900 font-medium">
								{formatDate(form.fr_timestamp)}
							</td>
							<td class="py-2.5 px-4 text-gray-700">{form.fr_formname}</td>
							<td class="py-2.5 px-4 text-gray-500">{form.fr_template}</td>
							<td class="py-2.5 px-4 text-right space-x-2">
								<button
									onclick={() => openEdit(form)}
									class="px-3 py-1 text-xs font-medium text-blue-600 bg-blue-50 rounded-md hover:bg-blue-100 transition-colors"
								>
									Edit
								</button>
								<button
									onclick={() => handleDelete(form)}
									class="px-3 py-1 text-xs font-medium text-red-600 bg-red-50 rounded-md hover:bg-red-100 transition-colors"
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

<!-- Form Modal -->
{#if showModal}
	<!-- svelte-ignore a11y_no_static_element_interactions -->
	<div
		class="fixed inset-0 z-50 flex items-center justify-center bg-black/40 overflow-y-auto"
		onclick={onModalBackdropClick}
		onkeydown={onModalKeydown}
		role="dialog"
		aria-modal="true"
		aria-label={modalMode === 'create' ? 'New Form' : 'Edit Form'}
		tabindex="-1"
	>
		<div class="bg-white rounded-xl shadow-xl w-full max-w-lg mx-4 my-8 overflow-hidden">
			<!-- Header -->
			<div class="flex items-center justify-between px-6 py-4 border-b border-gray-200">
				<h2 class="text-lg font-semibold text-gray-900">
					{modalMode === 'create' ? 'New Form' : 'Edit Form'}
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
			<div class="px-6 py-4 space-y-4 max-h-[60vh] overflow-y-auto">
				{#if formError}
					<div class="bg-red-50 border border-red-200 text-red-700 px-3 py-2 rounded-lg text-sm">
						{formError}
					</div>
				{/if}

				{#if modalMode === 'create'}
					<div>
						<label for="templateSelect" class="block text-sm font-medium text-gray-700 mb-1">
							Template
						</label>
						{#if templatesError}
							<div class="bg-red-50 border border-red-200 text-red-700 px-3 py-2 rounded-lg text-sm">
								{templatesError}
							</div>
						{:else}
							<select
								id="templateSelect"
								class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none transition-colors"
								onchange={(e) => {
									selectedTemplateId = Number((e.currentTarget as HTMLSelectElement).value) || null;
									onTemplateSelect();
								}}
							>
								<option value="">— Select a template —</option>
								{#each templates as t (t.id)}
									<option value={t.id}>{t.name}</option>
								{/each}
							</select>
						{/if}
					</div>
				{/if}

				{#if controlsLoading}
					<LoadingSpinner message="Loading controls..." />
				{:else if controlsError}
					<ErrorBanner message={controlsError} />
				{:else if controls.length > 0}
					{#each controls as c (c.variable)}
						<div>
							<label class="block text-sm font-medium text-gray-700 mb-1">
								{controlLabel(c)}
							</label>

							{#if c.type === 'boolean'}
								<div class="flex items-center gap-2">
									<input
										type="checkbox"
										checked={stringValue(c.variable) === '1' || stringValue(c.variable) === 'true'}
										onchange={(e) => {
											formValues[c.variable] = (e.currentTarget as HTMLInputElement)
												.checked
												? '1'
												: '0';
										}}
										class="h-4 w-4 text-blue-600 rounded border-gray-300 focus:ring-blue-500"
									/>
									<span class="text-sm text-gray-600">Yes</span>
								</div>
							{:else if c.type === 'date'}
								<input
									type="date"
									value={stringValue(c.variable)}
									oninput={(e) => {
										formValues[c.variable] = (e.currentTarget as HTMLInputElement).value;
									}}
									class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none transition-colors"
								/>
							{:else if c.type === 'select'}
								<select
									value={stringValue(c.variable)}
									onchange={(e) => {
										formValues[c.variable] = (e.currentTarget as HTMLSelectElement).value;
									}}
									class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none transition-colors"
								>
									<option value="">— Select —</option>
									{#each optionsFor(c) as opt}
										<option value={opt}>{opt}</option>
									{/each}
								</select>
							{:else if c.type === 'multiple'}
								<select
									multiple
									size={Math.min(optionsFor(c).length || 4, 6)}
									onchange={(e) => onMultiChange(c.variable, e)}
									class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none transition-colors"
								>
									{#each optionsFor(c) as opt}
										<option value={opt} selected={multiSelected(c.variable, opt)}>{opt}</option>
									{/each}
								</select>
							{:else}
								<input
									type="text"
									value={stringValue(c.variable)}
									oninput={(e) => {
										formValues[c.variable] = (e.currentTarget as HTMLInputElement).value;
									}}
									class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none transition-colors"
								/>
							{/if}
						</div>
					{/each}
				{:else if modalMode === 'edit'}
					<p class="text-sm text-gray-500">This form has no controls.</p>
				{/if}
			</div>

			<!-- Footer -->
			<div class="px-6 py-4 bg-gray-50 border-t border-gray-200 flex justify-end gap-3">
				<button
					onclick={closeModal}
					class="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors"
				>
					Cancel
				</button>
				{#if controls.length > 0}
					<button
						onclick={handleSave}
						disabled={saving}
						class="px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors flex items-center gap-2"
					>
						{#if saving}
							<div
								class="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin"
							></div>
							Saving...
						{:else}
							Save
						{/if}
					</button>
				{/if}
			</div>
		</div>
	</div>
{/if}
