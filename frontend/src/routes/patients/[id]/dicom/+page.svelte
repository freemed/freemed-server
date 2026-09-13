<script lang="ts">
	import { page } from '$app/stores';
	import { api } from '$lib/api';
	import DicomViewer from '$lib/components/DicomViewer.svelte';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
	import ErrorBanner from '$lib/components/ErrorBanner.svelte';
	import EmptyState from '$lib/components/EmptyState.svelte';

	interface DicomItem {
		id: number;
		created_at: string;
		study_description: string;
		filename: string;
		study_date: string;
		institution_name: string;
		institution_address: string;
		study_uid: string;
		series_uid: string;
		sop_uid: string;
		referring_provider: string;
		modality: string;
		storage_status: string;
	}

	let patientId = $derived($page.params.id || '');

	let items = $state<DicomItem[]>([]);
	let loading = $state(true);
	let error = $state('');
	let uploading = $state(false);
	let selectedFile = $state<File | null>(null);
	let selected = $state<DicomItem | null>(null);

	$effect(() => {
		if (patientId) loadItems(patientId);
	});

	async function loadItems(id: string) {
		loading = true;
		error = '';
		try {
			const data = await api.get<DicomItem[]>(`/patient/${id}/dicom`);
			items = data || [];
		} catch (e: any) {
			error = e.message || 'Failed to load DICOM studies';
		} finally {
			loading = false;
		}
	}

	function handleFileSelect(e: Event) {
		const input = e.target as HTMLInputElement;
		selectedFile = input.files?.[0] ?? null;
	}

	async function uploadDicom() {
		if (!selectedFile || !patientId) {
			error = 'Please select a DICOM file to upload';
			return;
		}
		uploading = true;
		error = '';
		try {
			const formData = new FormData();
			formData.append('file', selectedFile);
			const res = await fetch(`/api/patient/${patientId}/dicom`, {
				method: 'POST',
				body: formData,
				credentials: 'include'
			});
			if (!res.ok) {
				const body = await res.json().catch(() => ({}));
				throw new Error(body.message || `Upload failed (${res.status})`);
			}
			selectedFile = null;
			await loadItems(patientId);
		} catch (e: any) {
			error = e.message || 'Failed to upload DICOM';
		} finally {
			uploading = false;
		}
	}

	async function deleteDicom(id: number) {
		if (!patientId) return;
		if (!confirm('Remove this DICOM object?')) return;
		try {
			await api.del(`/patient/${patientId}/dicom/${id}`);
			await loadItems(patientId);
		} catch (e: any) {
			error = e.message || 'Failed to remove DICOM object';
		}
	}

	/** WADO-RS instance retrieval needs all three UIDs; the list endpoint returns
	 *  empty strings when metadata could not be parsed at upload time. */
	function hasUids(item: DicomItem): boolean {
		return !!(item.study_uid && item.series_uid && item.sop_uid);
	}

	/** Raw-bytes fallback that only needs the local row id. */
	function rawPath(item: DicomItem): string {
		return `/api/patient/${patientId}/dicom/${item.id}`;
	}

	function viewDicom(item: DicomItem) {
		if (!hasUids(item)) return;
		selected = item;
	}

	function closeViewer() {
		selected = null;
	}

	function formatDate(dateStr: string): string {
		if (!dateStr) return '—';
		const d = new Date(dateStr);
		if (isNaN(d.getTime())) return dateStr;
		return d.toLocaleDateString('en-US', {
			year: 'numeric',
			month: 'short',
			day: 'numeric'
		});
	}

	function formatFileSize(file: File): string {
		const kb = file.size / 1024;
		if (kb < 1024) return `${kb.toFixed(1)} KB`;
		return `${(kb / 1024).toFixed(1)} MB`;
	}
</script>

<div class="max-w-5xl mx-auto">
	<div class="flex items-center justify-between mb-6">
		<div>
			<h1 class="text-2xl font-bold text-gray-900">DICOM Imaging</h1>
			<p class="text-sm text-gray-500 mt-1">Patient #{patientId}</p>
		</div>
		<a
			href="/patients/{patientId}"
			class="text-sm text-blue-600 hover:text-blue-800 font-medium transition-colors"
		>
			&larr; Back to Patient
		</a>
	</div>

	{#if error}
		<div class="mb-6">
			<ErrorBanner message={error} onRetry={patientId ? () => loadItems(patientId) : undefined} />
		</div>
	{/if}

	<!-- Upload -->
	<div class="bg-white rounded-lg shadow-sm border border-gray-200 p-4 mb-6">
		<label for="dicom-file" class="block text-sm font-medium text-gray-700 mb-2">
			Upload DICOM study
		</label>
		<div class="flex items-center gap-3">
			<div
				class="relative flex-1 border-2 border-dashed border-gray-300 rounded-lg p-4 text-center hover:border-blue-400 transition-colors cursor-pointer"
				class:border-blue-400={!!selectedFile}
				class:bg-blue-50={!!selectedFile}
			>
				{#if selectedFile}
					<div class="space-y-1">
						<p class="text-sm font-medium text-blue-700">{selectedFile.name}</p>
						<p class="text-xs text-blue-500">{formatFileSize(selectedFile)}</p>
					</div>
				{:else}
					<p class="text-sm text-gray-500">Click to select a DICOM file</p>
				{/if}
				<input
					id="dicom-file"
					type="file"
					accept=".dcm,application/dicom"
					onchange={handleFileSelect}
					class="absolute inset-0 w-full h-full opacity-0 cursor-pointer"
				/>
			</div>
			<button
				type="button"
				onclick={uploadDicom}
				disabled={uploading || !selectedFile}
				class="px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded-lg hover:bg-blue-700 disabled:opacity-50 transition-colors whitespace-nowrap"
			>
				{uploading ? 'Uploading...' : 'Upload'}
			</button>
		</div>
	</div>

	{#if loading}
		<div class="bg-white rounded-lg shadow-sm border border-gray-200">
			<LoadingSpinner message="Loading DICOM studies…" />
		</div>
	{:else if items.length === 0}
		<div class="bg-white rounded-lg shadow-sm border border-gray-200">
			<EmptyState
				title="No DICOM studies"
				message="No imaging has been stored for this patient yet."
			/>
		</div>
	{:else}
		<div class="bg-white rounded-lg shadow-sm border border-gray-200 overflow-x-auto">
			<table class="min-w-full divide-y divide-gray-200">
				<thead class="bg-gray-50">
					<tr>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Modality</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Description</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Study Date</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Institution</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Study UID</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Actions</th>
					</tr>
				</thead>
				<tbody class="bg-white divide-y divide-gray-200">
					{#each items as item (item.id)}
						<tr class="hover:bg-gray-50">
							<td class="px-4 py-3 text-sm text-gray-900 font-semibold">
								{item.modality || '—'}
							</td>
							<td class="px-4 py-3 text-sm text-gray-700">
								{item.study_description || item.filename || '—'}
							</td>
							<td class="px-4 py-3 text-sm text-gray-700 whitespace-nowrap">
								{formatDate(item.study_date)}
							</td>
							<td class="px-4 py-3 text-sm text-gray-700">
								{item.institution_name || '—'}
							</td>
							<td class="px-4 py-3 text-sm text-gray-500 font-mono text-xs max-w-[180px] truncate">
								{item.study_uid || '—'}
							</td>
							<td class="px-4 py-3 text-sm whitespace-nowrap">
								{#if hasUids(item)}
									<button
										type="button"
										onclick={() => viewDicom(item)}
										class="text-blue-600 hover:text-blue-800 font-medium mr-3"
									>
										View
									</button>
								{:else}
									<button
										type="button"
										disabled
										title="No Study / Series / SOP Instance UID is stored for this object, so it cannot be addressed through the WADO-RS instance endpoint."
										class="text-gray-400 font-medium mr-3 cursor-not-allowed"
									>
										View
									</button>
								{/if}
								<a
									href={rawPath(item)}
									class="text-gray-600 hover:text-gray-800 font-medium mr-3"
								>
									Download
								</a>
								<button
									type="button"
									onclick={() => deleteDicom(item.id)}
									class="text-red-600 hover:text-red-800 font-medium"
								>
									Remove
								</button>
								{#if !hasUids(item)}
									<p class="text-xs text-gray-400 mt-1 max-w-[240px] whitespace-normal">
										Preview unavailable: this object has no Study / Series / SOP Instance UID
										(its header could not be parsed at upload), so it cannot be fetched from
										WADO-RS. The raw file can still be downloaded.
									</p>
								{/if}
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{/if}
</div>

{#if selected}
	{#key selected.id}
		<DicomViewer
			patientId={patientId}
			studyUid={selected.study_uid}
			seriesUid={selected.series_uid}
			sopUid={selected.sop_uid}
			label={selected.study_description || selected.filename || `Instance ${selected.id}`}
			rawPath={rawPath(selected)}
			onClose={closeViewer}
		/>
	{/key}
{/if}
