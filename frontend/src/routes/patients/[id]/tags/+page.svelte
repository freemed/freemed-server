<script lang="ts">
	import { page } from '$app/stores';
	import { api } from '$lib/api';

	interface TagItem {
		id: number;
		tag: string;
		patient: number;
		user: number;
		datecreate: string;
		dateexpire: string | null;
		created_at: string;
		updated_at: string;
		deleted_at: string | null;
	}

	let patientId = $state('');

	// Table state
	let tags = $state<TagItem[]>([]);
	let loading = $state(true);
	let error = $state('');

	// Modal state
	let showModal = $state(false);
	let formSaving = $state(false);
	let formError = $state('');
	let tagName = $state('');
	let expiryDate = $state('');

	$effect(() => {
		const id = $page.params.id;
		if (id) {
			patientId = id;
			loadTags(id);
		}
	});

	async function loadTags(id: string) {
		loading = true;
		error = '';
		try {
			const data = await api.get<TagItem[]>(`/patient/${id}/tags`);
			tags = data || [];
		} catch (e: any) {
			error = e.message || 'Failed to load tags';
		} finally {
			loading = false;
		}
	}

	function openModal() {
		tagName = '';
		expiryDate = '';
		formError = '';
		showModal = true;
	}

	function closeModal() {
		showModal = false;
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

		if (!tagName.trim()) {
			formError = 'Tag name is required.';
			return;
		}

		formSaving = true;
		try {
			const payload: Record<string, string> = { tag: tagName.trim() };
			if (expiryDate.trim()) {
				payload.dateexpire = expiryDate;
			}

			await api.post(`/patient/${patientId}/tags`, payload);
			closeModal();
			await loadTags(patientId);
		} catch (e: any) {
			formError = e.message || 'Failed to create tag';
		} finally {
			formSaving = false;
		}
	}

	async function handleExpire(tagId: number) {
		try {
			await api.del(`/patient/${patientId}/tags/${tagId}`);
			await loadTags(patientId);
		} catch (e: any) {
			// Reload anyway — the tag may have been expired
			loadTags(patientId);
		}
	}

	function formatDate(dateStr: string | null): string {
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
			<h1 class="text-2xl font-bold text-gray-900">Tags</h1>
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
				+ Add Tag
			</button>
		</div>
	</div>

	{#if error}
		<div class="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg text-sm mb-6">
			{error}
		</div>
	{/if}

	<div class="bg-white rounded-lg shadow-sm border border-gray-200">
		{#if loading}
			<div class="flex justify-center py-12">
				<div
					class="w-6 h-6 border-2 border-blue-500 border-t-transparent rounded-full animate-spin"
				></div>
			</div>
		{:else if error}
			<div class="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg text-sm m-6">
				{error}
			</div>
		{:else if tags.length === 0}
			<div class="text-center py-12 text-gray-500">
				<p class="text-lg">No tags</p>
				<p class="text-sm mt-1">
					Click "Add Tag" to assign a tag to this patient.
				</p>
			</div>
		{:else}
			<div class="overflow-x-auto">
				<table class="w-full text-sm">
					<thead>
						<tr class="border-b border-gray-200">
							<th class="text-left py-2 px-3 font-semibold text-gray-600">Tag</th>
							<th class="text-left py-2 px-3 font-semibold text-gray-600">Created</th>
							<th class="text-left py-2 px-3 font-semibold text-gray-600">Expires</th>
							<th class="text-right py-2 px-3 font-semibold text-gray-600">Actions</th>
						</tr>
					</thead>
					<tbody class="divide-y divide-gray-100">
						{#each tags as tagItem (tagItem.id)}
							<tr class="hover:bg-gray-50 transition-colors">
								<td class="py-2.5 px-3 text-gray-900 font-medium">
									{tagItem.tag}
								</td>
								<td class="py-2.5 px-3 text-gray-500 text-xs">
									{formatDate(tagItem.datecreate)}
								</td>
								<td class="py-2.5 px-3 text-gray-500 text-xs">
									{formatDate(tagItem.dateexpire)}
								</td>
								<td class="py-2.5 px-3 text-right">
									<button
										onclick={() => handleExpire(tagItem.id)}
										class="text-xs text-red-600 hover:text-red-800 font-medium transition-colors px-2 py-1 rounded hover:bg-red-50"
									>
										Expire
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

<!-- Add Tag Modal -->
{#if showModal}
	<!-- svelte-ignore a11y_no_static_element_interactions -->
	<div
		class="fixed inset-0 z-50 flex items-center justify-center bg-black/40"
		onclick={onModalBackdropClick}
		onkeydown={onModalKeydown}
		role="dialog"
		aria-modal="true"
		aria-label="Add Tag"
		tabindex="-1"
	>
		<div class="bg-white rounded-xl shadow-xl w-full max-w-md mx-4 overflow-hidden">
			<!-- Header -->
			<div class="flex items-center justify-between px-6 py-4 border-b border-gray-200">
				<h2 class="text-lg font-semibold text-gray-900">Add Tag</h2>
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
					<label for="tagName" class="block text-sm font-medium text-gray-700 mb-1">
						Tag Name
					</label>
					<input
						id="tagName"
						type="text"
						bind:value={tagName}
						placeholder="Enter tag name"
						class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none transition-colors"
					/>
				</div>

				<div>
					<label for="tagExpiry" class="block text-sm font-medium text-gray-700 mb-1">
						Expiry Date <span class="text-gray-400 font-normal">(optional)</span>
					</label>
					<input
						id="tagExpiry"
						type="date"
						bind:value={expiryDate}
						class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none transition-colors"
					/>
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
					onclick={handleCreate}
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
