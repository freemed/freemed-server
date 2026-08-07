<script lang="ts">
	import { page } from '$app/stores';
	import { api } from '$lib/api';

	interface PhotoID {
		id: number;
		created_at: string;
		description: string;
		page_count: number;
		photo_mime: string;
		photo: string | null;
	}

	let photoIDs = $state<PhotoID[]>([]);
	let loading = $state(true);
	let error = $state('');
	let patientId = $state('');

	// Modal state
	let showModal = $state(false);
	let modalSaving = $state(false);
	let modalError = $state('');

	// Form fields
	let formDescription = $state('');
	let formPageCount = $state(1);
	let selectedFile = $state<File | null>(null);
	let filePreview = $state('');

	$effect(() => {
		const id = $page.params.id;
		if (id) {
			patientId = id;
			loadPhotoIDs(id);
		}
	});

	async function loadPhotoIDs(id: string) {
		loading = true;
		error = '';
		try {
			const data = await api.get<PhotoID[]>(`/patient/${id}/photo-id`);
			photoIDs = data || [];
		} catch (e: any) {
			error = e.message || 'Failed to load photo IDs';
		} finally {
			loading = false;
		}
	}

	function openModal() {
		formDescription = '';
		formPageCount = 1;
		selectedFile = null;
		filePreview = '';
		modalError = '';
		showModal = true;
	}

	function closeModal() {
		showModal = false;
	}

	function onModalBackdropClick(e: MouseEvent) {
		if (e.target === e.currentTarget) closeModal();
	}

	function onModalKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') closeModal();
	}

	function handleFileSelect(e: Event) {
		const input = e.target as HTMLInputElement;
		const file = input.files?.[0];
		if (file) {
			selectedFile = file;
			// Auto-detect page count from filename hint or default to 1
			// Generate a small preview for display
			const reader = new FileReader();
			reader.onload = () => {
				filePreview = reader.result as string;
			};
			reader.readAsDataURL(file);
		}
	}

	async function handleSubmit(e: Event) {
		e.preventDefault();

		if (!formDescription.trim()) {
			modalError = 'Description is required';
			return;
		}
		if (!selectedFile) {
			modalError = 'Please select a file to upload';
			return;
		}
		if (formPageCount < 1) {
			modalError = 'Page count must be at least 1';
			return;
		}

		modalSaving = true;
		modalError = '';

		try {
			// Read file as base64 (strip data: prefix)
			const base64 = await readFileAsBase64(selectedFile);
			const payload = {
				description: formDescription.trim(),
				page_count: formPageCount,
				photo: base64,
				photo_mime: selectedFile.type || 'application/octet-stream',
			};

			await api.post(`/patient/${patientId}/photo-id`, payload);
			closeModal();
			await loadPhotoIDs(patientId);
		} catch (e: any) {
			modalError = e.message || 'Failed to upload photo ID';
		} finally {
			modalSaving = false;
		}
	}

	function readFileAsBase64(file: File): Promise<string> {
		return new Promise((resolve, reject) => {
			const reader = new FileReader();
			reader.onload = () => {
				const result = reader.result as string;
				// Strip the data:...;base64, prefix
				const base64 = result.split(',')[1] || result;
				resolve(base64);
			};
			reader.onerror = () => reject(new Error('Failed to read file'));
			reader.readAsDataURL(file);
		});
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

	function formatFileSize(file: File): string {
		const kb = file.size / 1024;
		if (kb < 1024) return `${kb.toFixed(1)} KB`;
		return `${(kb / 1024).toFixed(1)} MB`;
	}
</script>

<div class="max-w-4xl mx-auto">
	<div class="flex items-center justify-between mb-6">
		<div>
			<h1 class="text-2xl font-bold text-gray-900">Photo ID</h1>
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
				+ Upload Photo ID
			</button>
		</div>
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
	{:else if photoIDs.length === 0}
		<div class="text-center py-12 text-gray-500 bg-white rounded-lg shadow-sm border border-gray-200">
			<p class="text-lg">No photo IDs uploaded</p>
			<p class="text-sm mt-1">No photo identification has been recorded for this patient yet.</p>
		</div>
	{:else}
		<div class="bg-white rounded-lg shadow-sm border border-gray-200 overflow-x-auto">
			<table class="min-w-full divide-y divide-gray-200">
				<thead class="bg-gray-50">
					<tr>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Description</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Pages</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Date</th>
					</tr>
				</thead>
				<tbody class="bg-white divide-y divide-gray-200">
					{#each photoIDs as photo}
						<tr class="hover:bg-gray-50">
							<td class="px-4 py-3 text-sm text-gray-900 font-medium">
								{photo.description || '—'}
							</td>
							<td class="px-4 py-3 text-sm text-gray-700">
								{photo.page_count}
							</td>
							<td class="px-4 py-3 text-sm text-gray-700 whitespace-nowrap">
								{formatDate(photo.created_at)}
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{/if}

	<!-- Upload Modal -->
	{#if showModal}
		<!-- svelte-ignore a11y_no_static_element_interactions -->
		<div
			class="fixed inset-0 z-50 flex items-center justify-center bg-black/40"
			onclick={onModalBackdropClick}
			onkeydown={onModalKeydown}
			role="dialog"
			aria-modal="true"
			tabindex="-1"
		>
			<div class="bg-white rounded-xl shadow-xl w-full max-w-md mx-4 overflow-hidden">
				<!-- Header -->
				<div class="flex items-center justify-between px-6 py-4 border-b border-gray-200">
					<h2 class="text-lg font-semibold text-gray-900">Upload Photo ID</h2>
					<button
						onclick={closeModal}
						class="text-gray-400 hover:text-gray-600 p-1 rounded-lg hover:bg-gray-100"
						aria-label="Close"
					>
						<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
							<path fill-rule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clip-rule="evenodd" />
						</svg>
					</button>
				</div>

				<!-- Body -->
				<div class="px-6 py-4">
					{#if modalError}
						<div class="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg text-sm mb-4">
							{modalError}
						</div>
					{/if}
					<form onsubmit={handleSubmit}>
						<div class="space-y-4">
							<!-- Description -->
							<div>
								<label for="description" class="block text-sm font-medium text-gray-700 mb-1">Description *</label>
								<input
									id="description"
									type="text"
									bind:value={formDescription}
									placeholder="e.g. Driver's License, Passport"
									class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-sm"
									required
								/>
							</div>

							<!-- Page Count -->
							<div>
								<label for="pageCount" class="block text-sm font-medium text-gray-700 mb-1">Page Count</label>
								<input
									id="pageCount"
									type="number"
									bind:value={formPageCount}
									min="1"
									class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-sm"
								/>
							</div>

							<!-- File Upload -->
							<div>
								<label for="file" class="block text-sm font-medium text-gray-700 mb-1">Photo ID File *</label>
								<div
									class="relative border-2 border-dashed border-gray-300 rounded-lg p-6 text-center hover:border-blue-400 transition-colors cursor-pointer"
									class:border-blue-400={!!selectedFile}
									class:bg-blue-50={!!selectedFile}
								>
									{#if selectedFile}
										<div class="space-y-1">
											<svg xmlns="http://www.w3.org/2000/svg" class="h-8 w-8 mx-auto text-blue-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
											</svg>
											<p class="text-sm font-medium text-blue-700">{selectedFile.name}</p>
											<p class="text-xs text-blue-500">{formatFileSize(selectedFile)}</p>
											{#if filePreview}
												<img src={filePreview} alt="Preview" class="max-h-32 mx-auto mt-2 rounded border border-gray-200" />
											{/if}
										</div>
									{:else}
										<div class="space-y-2">
											<svg xmlns="http://www.w3.org/2000/svg" class="h-10 w-10 mx-auto text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
											</svg>
											<p class="text-sm text-gray-500">Drag and drop or click to upload</p>
											<p class="text-xs text-gray-400">PNG, JPG, GIF up to 10MB</p>
										</div>
									{/if}
									<input
										id="file"
										type="file"
										accept="image/*"
										onchange={handleFileSelect}
										class="absolute inset-0 w-full h-full opacity-0 cursor-pointer"
									/>
								</div>
							</div>
						</div>
					</form>
				</div>

				<!-- Footer -->
				<div class="px-6 py-4 bg-gray-50 border-t border-gray-200 flex justify-end gap-3">
					<button
						type="button"
						onclick={closeModal}
						class="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors"
					>
						Cancel
					</button>
					<button
						type="button"
						onclick={handleSubmit}
						disabled={modalSaving}
						class="px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-lg hover:bg-blue-700 disabled:opacity-50 transition-colors"
					>
						{modalSaving ? 'Uploading...' : 'Upload Photo ID'}
					</button>
				</div>
			</div>
		</div>
	{/if}
</div>
