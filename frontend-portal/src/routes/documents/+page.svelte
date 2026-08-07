<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib/api';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
	import ErrorBanner from '$lib/components/ErrorBanner.svelte';
	import EmptyState from '$lib/components/EmptyState.svelte';

	interface Document {
		id: number;
		document_name: string;
		document_type?: string;
		upload_date: string;
		file_size?: number;
	}

	let documents = $state<Document[]>([]);
	let loading = $state(true);
	let error = $state('');

	onMount(() => {
		loadDocuments();
	});

	async function loadDocuments() {
		error = '';
		loading = true;
		try {
			const data = await api.get<Document[]>('/documents');
			documents = data || [];
		} catch (e: unknown) {
			error = e instanceof Error ? e.message : 'Failed to load documents';
		} finally {
			loading = false;
		}
	}

	function formatFileSize(bytes: number | undefined): string {
		if (!bytes) return '—';
		if (bytes < 1024) return `${bytes} B`;
		if (bytes < 1048576) return `${(bytes / 1024).toFixed(1)} KB`;
		return `${(bytes / 1048576).toFixed(1)} MB`;
	}

	function typeIcon(type: string | undefined): string {
		if (!type) return '📄';
		const t = type.toLowerCase();
		if (t.includes('pdf')) return '📕';
		if (t.includes('image') || t.includes('xray') || t.includes('scan')) return '🖼️';
		if (t.includes('lab')) return '🔬';
		return '📄';
	}
</script>

<svelte:head>
	<title>Documents — FreeMED Patient Portal</title>
</svelte:head>

<div>
	<h2 class="fw-bold text-dark mb-4">Documents</h2>

	{#if loading}
		<LoadingSpinner message="Loading documents…" />
	{:else if error}
		<ErrorBanner message={error} onRetry={loadDocuments} />
	{:else if documents.length === 0}
		<EmptyState
			title="No documents"
			message="No documents have been shared with you yet."
		/>
	{:else}
		<div class="row g-3">
			{#each documents as doc (doc.id)}
				<div class="col-12 col-md-6 col-lg-4">
					<div class="card card-hover h-100">
						<div class="card-body">
							<div class="d-flex align-items-start gap-3">
								<div class="fs-3">{typeIcon(doc.document_type)}</div>
								<div class="flex-grow-1 min-w-0">
									<h6 class="card-title mb-1 text-truncate">{doc.document_name}</h6>
									<p class="card-text small text-muted mb-1">
										{doc.document_type || 'Document'} &middot; {formatFileSize(doc.file_size)}
									</p>
									<p class="card-text small text-muted mb-0">
										Uploaded {doc.upload_date}
									</p>
								</div>
							</div>
						</div>
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>
