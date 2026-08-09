<script lang="ts">
	import { api } from '$lib/api';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
	import ErrorBanner from '$lib/components/ErrorBanner.svelte';

	interface EraResult {
		claims_processed: number;
		payments_posted: number;
		adjustments: number;
		errors: number;
		payer: string;
		payment_amount: number;
		trace_number: string;
	}

	let dragging = $state(false);
	let processing = $state(false);
	let uploadError = $state('');
	let result = $state<EraResult | null>(null);

	function onDragOver(e: DragEvent) {
		e.preventDefault();
		dragging = true;
	}

	function onDragLeave() {
		dragging = false;
	}

	function onDrop(e: DragEvent) {
		e.preventDefault();
		dragging = false;
		const files = e.dataTransfer?.files;
		if (files && files.length > 0) {
			uploadFile(files[0]);
		}
	}

	function onFileInput(e: Event) {
		const input = e.target as HTMLInputElement;
		const files = input.files;
		if (files && files.length > 0) {
			uploadFile(files[0]);
		}
	}

	async function uploadFile(file: File) {
		const validExtensions = ['.txt', '.edi', '.835'];
		const ext = '.' + file.name.split('.').pop()?.toLowerCase();
		if (!validExtensions.includes(ext)) {
			uploadError = 'Please upload a .txt, .edi, or .835 ERA file.';
			return;
		}

		uploadError = '';
		result = null;
		processing = true;

		try {
			const formData = new FormData();
			formData.append('file', file);

			const res = await fetch('/api/era/upload', {
				method: 'POST',
				body: formData,
			});

			if (res.status === 401) {
				uploadError = 'Session expired. Please refresh the page.';
				return;
			}

			if (!res.ok) {
				const body = await res.text();
				uploadError = `Upload failed: ${body}`;
				return;
			}

			result = await res.json();
		} catch (e: any) {
			uploadError = e.message || 'Failed to upload ERA file';
		} finally {
			processing = false;
		}
	}

	function formatCurrency(val: number): string {
		return val.toLocaleString('en-US', {
			style: 'currency',
			currency: 'USD',
			minimumFractionDigits: 2,
		});
	}

	function clearResult() {
		result = null;
		uploadError = '';
	}
</script>

<div class="bg-white rounded-lg shadow-sm border border-gray-200">
	<div class="px-6 py-4 border-b border-gray-100">
		<h2 class="text-lg font-semibold text-gray-900">ERA (835) Upload</h2>
		<p class="text-xs text-gray-500 mt-0.5">Upload X12 835 Electronic Remittance Advice files for automatic payment posting</p>
	</div>

	{#if !result}
		<div class="p-6">
			<!-- Drop zone -->
			<label
				class="block w-full cursor-pointer"
				ondragover={onDragOver}
				ondragleave={onDragLeave}
				ondrop={onDrop}
			>
				<div
					class="border-2 border-dashed rounded-lg p-12 text-center transition-colors {dragging ? 'border-blue-400 bg-blue-50' : 'border-gray-300 hover:border-gray-400'}"
				>
					<svg class="w-12 h-12 mx-auto mb-4 {dragging ? 'text-blue-400' : 'text-gray-300'}" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
					</svg>
					{#if processing}
						<div class="flex flex-col items-center">
							<svg class="animate-spin h-8 w-8 text-blue-600 mb-3" fill="none" viewBox="0 0 24 24">
								<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
								<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
							</svg>
							<p class="text-sm text-gray-500">Processing ERA file…</p>
						</div>
					{:else}
						<p class="text-sm text-gray-600 font-medium">
							Drag and drop an 835 ERA file here
						</p>
						<p class="text-xs text-gray-400 mt-1">or click to browse</p>
						<p class="text-xs text-gray-400 mt-2">Accepts .txt, .edi, .835 files</p>
					{/if}
					<input
						type="file"
						accept=".txt,.edi,.835"
						onchange={onFileInput}
						disabled={processing}
						class="hidden"
						{...{ webkitdirectory: false } as any}
					/>
				</div>
			</label>

			{#if uploadError}
				<ErrorBanner
					message={uploadError}
					onRetry={() => { uploadError = ''; }}
				/>
			{/if}
		</div>
	{:else}
		<!-- Results view -->
		<div class="p-6">
			<div class="mb-6 flex justify-between items-start">
				<div>
					<h3 class="text-lg font-semibold text-gray-900">Processing Complete</h3>
					<p class="text-sm text-gray-500 mt-1">
						Payer: {result.payer} &middot; Trace: {result.trace_number}
					</p>
				</div>
				<button
					onclick={clearResult}
					class="px-4 py-2 text-sm font-medium text-blue-600 bg-blue-50 rounded-md hover:bg-blue-100 transition-colors"
				>
					Upload Another
				</button>
			</div>

			<!-- Summary cards -->
			<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
				<div class="bg-blue-50 rounded-lg p-4 text-center">
					<p class="text-3xl font-bold text-blue-700">{result.claims_processed}</p>
					<p class="text-xs text-blue-600 mt-1 font-medium uppercase">Claims Processed</p>
				</div>
				<div class="bg-green-50 rounded-lg p-4 text-center">
					<p class="text-3xl font-bold text-green-700">{result.payments_posted}</p>
					<p class="text-xs text-green-600 mt-1 font-medium uppercase">Payments Posted</p>
				</div>
				<div class="bg-yellow-50 rounded-lg p-4 text-center">
					<p class="text-3xl font-bold text-yellow-700">{result.adjustments}</p>
					<p class="text-xs text-yellow-600 mt-1 font-medium uppercase">Adjustments</p>
				</div>
				<div class="bg-red-50 rounded-lg p-4 text-center">
					<p class="text-3xl font-bold text-red-700">{result.errors}</p>
					<p class="text-xs text-red-600 mt-1 font-medium uppercase">Errors</p>
				</div>
			</div>

			<div class="bg-gray-50 rounded-lg p-4 flex items-center justify-between">
				<span class="text-sm text-gray-600">Total Payment Amount</span>
				<span class="text-lg font-bold text-gray-900 font-mono">{formatCurrency(result.payment_amount)}</span>
			</div>
		</div>
	{/if}
</div>
