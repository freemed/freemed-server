<script lang="ts">
	import { page } from '$app/stores';
	import { api } from '$lib/api';

	interface Signature {
		id: number;
		created_at: string;
		updated_at: string;
		deleted_at: string | null;
		patient: number;
		module: string;
		module_field: string;
		oid: number;
		signature_data: string | null;
		format: string;
		collector_location: string;
		collector_model: string;
		collector_jobid: string;
		collector_finished: number;
		user: number;
	}

	let patientId = $derived($page.params.id || '');
	let items = $state<Signature[]>([]);
	let loading = $state(true);
	let error = $state('');

	// Modal state
	let showModal = $state(false);
	let formSaving = $state(false);
	let formError = $state('');

	let sigModule = $state('');
	let sigField = $state('');
	let sigFormat = $state('UNKNOWN');
	let sigOid = $state('0');
	let sigFile: File | null = $state(null);
	let sigFileBase64 = $state('');

	let fileInput: HTMLInputElement | undefined = $state(undefined);

	const FORMATS = ['UNKNOWN', 'JPG', 'PNG', 'TOPAZ'];

	$effect(() => {
		if (patientId) {
			loadItems();
		}
	});

	async function loadItems() {
		loading = true;
		error = '';
		try {
			const data = await api.get<Signature[]>(
				`/patient/${patientId}/signatures`
			);
			items = data || [];
		} catch (e: any) {
			error = e.message || 'Failed to load signatures';
		} finally {
			loading = false;
		}
	}

	function openModal() {
		sigModule = '';
		sigField = '';
		sigFormat = 'UNKNOWN';
		sigOid = '0';
		sigFile = null;
		sigFileBase64 = '';
		formError = '';
		showModal = true;
	}

	function closeModal() {
		showModal = false;
	}

	async function handleFileChange(e: Event) {
		const target = e.target as HTMLInputElement;
		const file = target.files?.[0];
		if (!file) {
			sigFile = null;
			sigFileBase64 = '';
			return;
		}
		sigFile = file;
		// Read file as base64 data URL
		sigFileBase64 = await new Promise<string>((resolve, reject) => {
			const reader = new FileReader();
			reader.onload = () => {
				const result = reader.result as string;
				// Strip data:...;base64, prefix
				const comma = result.indexOf(',');
				resolve(comma >= 0 ? result.slice(comma + 1) : result);
			};
			reader.onerror = () => reject(new Error('Failed to read file'));
			reader.readAsDataURL(file);
		});
	}

	async function handleSubmit(e: Event) {
		e.preventDefault();

		if (!sigModule.trim()) {
			formError = 'Module name is required';
			return;
		}
		if (!sigField.trim()) {
			formError = 'Field name is required';
			return;
		}

		formSaving = true;
		formError = '';

		try {
			await api.post(`/patient/${patientId}/signatures`, {
				module: sigModule.trim(),
				module_field: sigField.trim(),
				format: sigFormat,
				oid: parseInt(sigOid.trim(), 10) || 0,
				signature_data: sigFileBase64 || null,
			});
			showModal = false;
			await loadItems();
		} catch (e: any) {
			formError = e.message || 'Failed to save signature';
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

	function formatBadgeClass(fmt: string): string {
		switch (fmt) {
			case 'JPG':
			case 'JPEG':
				return 'bg-blue-50 text-blue-700';
			case 'PNG':
				return 'bg-green-50 text-green-700';
			case 'TOPAZ':
				return 'bg-purple-50 text-purple-700';
			default:
				return 'bg-gray-50 text-gray-500';
		}
	}
</script>

<div class="max-w-4xl mx-auto">
	<!-- Header -->
	<div class="flex items-center justify-between mb-6">
		<div>
			<h1 class="text-2xl font-bold text-gray-900">Signatures</h1>
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
				+ Add Signature
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
			<p class="text-lg">No signatures</p>
			<p class="text-sm mt-1">No signatures have been recorded for this patient yet.</p>
		</div>
	{:else}
		<div class="bg-white rounded-lg shadow-sm border border-gray-200 overflow-x-auto">
			<table class="min-w-full divide-y divide-gray-200">
				<thead class="bg-gray-50">
					<tr>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Module</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Field</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Format</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Date</th>
					</tr>
				</thead>
				<tbody class="bg-white divide-y divide-gray-200">
					{#each items as item}
						<tr class="hover:bg-gray-50">
							<td class="px-4 py-3 text-sm text-gray-700 whitespace-nowrap">
								{item.module}
							</td>
							<td class="px-4 py-3 text-sm text-gray-700 whitespace-nowrap">
								{item.module_field}
							</td>
							<td class="px-4 py-3 text-sm">
								<span class="inline-block px-2 py-0.5 text-xs font-medium rounded {formatBadgeClass(item.format)}">
									{item.format || 'UNKNOWN'}
								</span>
							</td>
							<td class="px-4 py-3 text-sm text-gray-500 whitespace-nowrap">
								{formatDate(item.created_at)}
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
				<h2 class="text-lg font-semibold text-gray-900">Add Signature</h2>
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
						<label for="sig-module" class="block text-sm font-medium text-gray-700 mb-1">Module Name *</label>
						<input
							id="sig-module"
							type="text"
							bind:value={sigModule}
							placeholder="e.g. encounter, billing"
							class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-sm"
							required
						/>
					</div>

					<div>
						<label for="sig-field" class="block text-sm font-medium text-gray-700 mb-1">Field Name *</label>
						<input
							id="sig-field"
							type="text"
							bind:value={sigField}
							placeholder="e.g. physician_signature"
							class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-sm"
							required
						/>
					</div>

					<div>
						<label for="sig-oid" class="block text-sm font-medium text-gray-700 mb-1">Object ID</label>
						<input
							id="sig-oid"
							type="number"
							bind:value={sigOid}
							placeholder="0"
							class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-sm"
						/>
					</div>

					<div>
						<label for="sig-format" class="block text-sm font-medium text-gray-700 mb-1">Format</label>
						<select
							id="sig-format"
							bind:value={sigFormat}
							class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-sm"
						>
							{#each FORMATS as fmt}
								<option value={fmt}>{fmt}</option>
							{/each}
						</select>
					</div>

					<div>
						<label for="sig-file" class="block text-sm font-medium text-gray-700 mb-1">Signature File</label>
						<input
							id="sig-file"
							type="file"
							accept="image/*,.topaz"
							bind:this={fileInput}
							onchange={handleFileChange}
							class="w-full text-sm text-gray-500 file:mr-4 file:py-2 file:px-4 file:rounded-lg file:border-0 file:text-sm file:font-medium file:bg-blue-50 file:text-blue-700 hover:file:bg-blue-100"
						/>
						{#if sigFile}
							<p class="text-xs text-gray-500 mt-1">{sigFile.name} ({(sigFile.size / 1024).toFixed(1)} KB)</p>
						{/if}
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
