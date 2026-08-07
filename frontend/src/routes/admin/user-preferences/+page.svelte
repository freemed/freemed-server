<script lang="ts">
	import { api } from '$lib/api';

	interface UserPreference {
		id: number;
		option_key: string;
		default_value: string;
		title: string;
		section: string;
		option_type: string;
		options: string | null;
	}

	interface EditForm {
		default_value: string;
		title: string;
		section: string;
		option_type: string;
		options: string;
	}

	let prefs = $state<UserPreference[]>([]);
	let loading = $state(true);
	let error = $state('');
	let showModal = $state(false);
	let editingPref = $state<UserPreference | null>(null);
	let form = $state<EditForm>({ default_value: '', title: '', section: '', option_type: '', options: '' });
	let formError = $state('');
	let formSaving = $state(false);

	async function loadPrefs() {
		loading = true;
		error = '';
		try {
			prefs = await api.get<UserPreference[]>('/user-preferences');
		} catch (e: any) {
			error = e.message || 'Failed to load user preferences';
		} finally {
			loading = false;
		}
	}

	$effect(() => {
		loadPrefs();
	});

	function openEdit(pref: UserPreference) {
		editingPref = pref;
		form = {
			default_value: pref.default_value,
			title: pref.title,
			section: pref.section,
			option_type: pref.option_type,
			options: pref.options || '',
		};
		formError = '';
		showModal = true;
	}

	function closeModal() {
		showModal = false;
		editingPref = null;
		formSaving = false;
		formError = '';
	}

	async function savePref() {
		if (!editingPref) return;
		formError = '';
		formSaving = true;
		try {
			await api.put(`/user-preferences/${editingPref.option_key}`, {
				default_value: form.default_value,
				title: form.title,
				section: form.section,
				option_type: form.option_type,
				options: form.options || null,
			});
			closeModal();
			await loadPrefs();
		} catch (e: any) {
			formError = e.message || 'Save failed';
		} finally {
			formSaving = false;
		}
	}

	function onModalBackdropClick(e: MouseEvent) {
		if (e.target === e.currentTarget) closeModal();
	}

	function onModalKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') closeModal();
	}
</script>

<div class="max-w-6xl mx-auto">
	<div class="flex items-center justify-between mb-6">
		<h1 class="text-2xl font-bold text-gray-900">User Preferences</h1>
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
	{:else if prefs.length === 0}
		<div class="text-center py-12 text-gray-500">
			<p class="text-lg">No user preferences found</p>
		</div>
	{:else}
		<div class="bg-white rounded-lg shadow-sm border border-gray-200 overflow-hidden">
			<div class="overflow-x-auto">
				<table class="w-full text-sm">
					<thead class="bg-gray-50 border-b border-gray-200">
						<tr>
							<th class="text-left px-4 py-3 font-semibold text-gray-600">Key</th>
							<th class="text-left px-4 py-3 font-semibold text-gray-600">Default Value</th>
							<th class="text-left px-4 py-3 font-semibold text-gray-600">Title</th>
							<th class="text-left px-4 py-3 font-semibold text-gray-600">Section</th>
							<th class="text-left px-4 py-3 font-semibold text-gray-600">Type</th>
							<th class="text-left px-4 py-3 font-semibold text-gray-600">Options</th>
						</tr>
					</thead>
					<tbody class="divide-y divide-gray-100">
						{#each prefs as pref (pref.id)}
							<tr
								class="hover:bg-blue-50 cursor-pointer transition-colors"
								onclick={() => openEdit(pref)}
								onkeydown={(e) => e.key === 'Enter' && openEdit(pref)}
								tabindex="0"
								role="button"
							>
								<td class="px-4 py-3 text-gray-900 font-medium font-mono text-xs">{pref.option_key}</td>
								<td class="px-4 py-3 text-gray-700 max-w-[200px] truncate">{pref.default_value || '—'}</td>
								<td class="px-4 py-3 text-gray-700">{pref.title || '—'}</td>
								<td class="px-4 py-3 text-gray-500">{pref.section || '—'}</td>
								<td class="px-4 py-3 text-gray-500">
									{#if pref.option_type}
										<span class="px-2 py-0.5 bg-gray-100 rounded text-xs font-medium">{pref.option_type}</span>
									{:else}
										—
									{/if}
								</td>
								<td class="px-4 py-3 text-gray-500 max-w-[200px] truncate">{pref.options || '—'}</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		</div>
	{/if}
</div>

<!-- Edit Modal -->
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
		<div
			class="bg-white rounded-xl shadow-xl w-full max-w-lg mx-4 overflow-hidden"
			onclick={(e) => e.stopPropagation()}
			onkeydown={(e) => e.key === 'Escape' && closeModal()}
			role="presentation"
		>
			<div class="flex items-center justify-between px-6 py-4 border-b border-gray-200">
				<h2 class="text-lg font-semibold text-gray-900">
					Edit: <span class="font-mono text-blue-600">{editingPref?.option_key}</span>
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

			<div class="px-6 py-4 space-y-4">
				{#if formError}
					<div class="bg-red-50 border border-red-200 text-red-700 px-3 py-2 rounded-lg text-sm">
						{formError}
					</div>
				{/if}

				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">Title</label>
					<input
						type="text"
						bind:value={form.title}
						class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
						placeholder="Display title"
					/>
				</div>

				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">Default Value</label>
					<input
						type="text"
						bind:value={form.default_value}
						class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
						placeholder="Default value"
					/>
				</div>

				<div class="grid grid-cols-2 gap-4">
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">Section</label>
						<input
							type="text"
							bind:value={form.section}
							class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
							placeholder="e.g. general, billing"
						/>
					</div>
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">Type</label>
						<select
							bind:value={form.option_type}
							class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent bg-white"
						>
							<option value="">—</option>
							<option value="text">Text</option>
							<option value="number">Number</option>
							<option value="boolean">Boolean</option>
							<option value="select">Select</option>
							<option value="textarea">Textarea</option>
							<option value="date">Date</option>
						</select>
					</div>
				</div>

				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">Options</label>
					<textarea
						bind:value={form.options}
						class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
						placeholder="For select type: comma-separated values or JSON"
						rows="2"
					></textarea>
					<p class="text-xs text-gray-400 mt-1">For select type fields: comma-separated list or JSON</p>
				</div>
			</div>

			<div class="px-6 py-4 bg-gray-50 border-t border-gray-200 flex justify-end gap-3">
				<button
					onclick={closeModal}
					class="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors"
				>
					Cancel
				</button>
				<button
					onclick={savePref}
					disabled={formSaving}
					class="px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
				>
					{formSaving ? 'Saving...' : 'Save'}
				</button>
			</div>
		</div>
	</div>
{/if}
