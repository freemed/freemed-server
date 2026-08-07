<script lang="ts">
	import { page } from '$app/stores';
	import { api } from '$lib/api';

	interface Letter {
		id: number;
		created_at: string;
		updated_at: string;
		deleted_at: string | null;
		patient: number;
		letter_type: string;
		recipient: string;
		subject: string;
		body: { String: string; Valid: boolean } | null;
		date_sent: string | null;
		user: number;
		active: string;
	}

	let letters = $state<Letter[]>([]);
	let loading = $state(true);
	let error = $state('');
	let patientId = $state('');

	// Modal state
	let showComposeModal = $state(false);
	let formError = $state('');
	let formSubmitting = $state(false);
	let composeSubject = $state('');
	let composeBody = $state('');
	let composeRecipient = $state('');

	$effect(() => {
		const id = $page.params.id;
		if (id) {
			patientId = id;
			loadLetters(id);
		}
	});

	async function loadLetters(id: string) {
		loading = true;
		error = '';
		try {
			const data = await api.get<Letter[]>(`/patient/${id}/letters`);
			letters = data || [];
		} catch (e: any) {
			error = e.message || 'Failed to load letters';
		} finally {
			loading = false;
		}
	}

	function openComposeModal() {
		composeSubject = '';
		composeBody = '';
		composeRecipient = '';
		formError = '';
		showComposeModal = true;
	}

	function closeComposeModal() {
		showComposeModal = false;
	}

	async function handleCompose(e: Event) {
		e.preventDefault();

		if (!composeSubject.trim()) {
			formError = 'Subject is required';
			return;
		}
		if (!composeRecipient.trim()) {
			formError = 'Recipient is required';
			return;
		}

		formSubmitting = true;
		formError = '';

		const payload = {
			letter_type: '',
			recipient: composeRecipient.trim(),
			subject: composeSubject.trim(),
			body: composeBody.trim(),
			date_sent: new Date().toISOString().split('T')[0],
		};

		try {
			await api.post(`/patient/${patientId}/letters`, payload);
			showComposeModal = false;
			await loadLetters(patientId);
		} catch (e: any) {
			formError = e.message || 'Failed to create letter';
		} finally {
			formSubmitting = false;
		}
	}

	function formatDate(dateStr: string | null): string {
		if (!dateStr) return '—';
		try {
			const d = new Date(dateStr);
			if (isNaN(d.getTime())) return dateStr;
			return d.toLocaleDateString('en-US', { year: 'numeric', month: 'short', day: 'numeric' });
		} catch {
			return dateStr;
		}
	}

	function bodyPreview(body: Letter['body']): string {
		if (!body || !body.Valid) return '—';
		const text = body.String;
		if (text.length <= 80) return text;
		return text.substring(0, 80) + '…';
	}
</script>

<div class="max-w-4xl mx-auto">
	<div class="flex items-center justify-between mb-6">
		<div>
			<h1 class="text-2xl font-bold text-gray-900">Letters</h1>
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
				onclick={openComposeModal}
				class="px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded-lg hover:bg-blue-700 transition-colors"
			>
				+ Compose Letter
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
	{:else if letters.length === 0}
		<div class="text-center py-12 text-gray-500 bg-white rounded-lg shadow-sm border border-gray-200">
			<p class="text-lg">No letters found</p>
			<p class="text-sm mt-1">No letters have been composed for this patient yet.</p>
		</div>
	{:else}
		<div class="bg-white rounded-lg shadow-sm border border-gray-200 overflow-x-auto">
			<table class="min-w-full divide-y divide-gray-200">
				<thead class="bg-gray-50">
					<tr>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Date</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Subject</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Recipient</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Preview</th>
					</tr>
				</thead>
				<tbody class="bg-white divide-y divide-gray-200">
					{#each letters as letter}
						<tr class="hover:bg-gray-50">
							<td class="px-4 py-3 text-sm text-gray-600 whitespace-nowrap">
								{formatDate(letter.date_sent)}
							</td>
							<td class="px-4 py-3 text-sm font-medium text-gray-900">
								{letter.subject || '—'}
							</td>
							<td class="px-4 py-3 text-sm text-gray-600 whitespace-nowrap">
								{letter.recipient || '—'}
							</td>
							<td class="px-4 py-3 text-sm text-gray-600 max-w-xs truncate">
								{bodyPreview(letter.body)}
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{/if}
</div>

<!-- Compose Letter Modal -->
{#if showComposeModal}
	<!-- svelte-ignore a11y_no_static_element_interactions -->
	<div
		class="fixed inset-0 z-50 flex items-center justify-center"
		onkeydown={(e) => { if (e.key === 'Escape') closeComposeModal(); }}
	>
		<!-- Backdrop -->
		<!-- svelte-ignore a11y_no_static_element_interactions -->
		<div
			class="absolute inset-0 bg-black/50"
			onclick={closeComposeModal}
		></div>

		<!-- Modal -->
		<div class="relative bg-white rounded-lg shadow-xl w-full max-w-lg mx-4 max-h-[90vh] overflow-y-auto">
			<div class="px-6 py-4 border-b border-gray-100 flex items-center justify-between">
				<h2 class="text-lg font-semibold text-gray-900">Compose Letter</h2>
				<button
					onclick={closeComposeModal}
					class="text-gray-400 hover:text-gray-600 transition-colors"
					aria-label="Close"
				>
					<svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
					</svg>
				</button>
			</div>

			<form onsubmit={handleCompose}>
				<div class="px-6 py-4 space-y-4">
					{#if formError}
						<div class="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg text-sm">
							{formError}
						</div>
					{/if}

					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">Subject *</label>
						<input
							type="text"
							bind:value={composeSubject}
							placeholder="e.g. Referral Letter, Follow-up Instructions"
							class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-sm"
							required
						/>
					</div>

					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">Recipient *</label>
						<input
							type="text"
							bind:value={composeRecipient}
							placeholder="e.g. Dr. Smith, Cardiology Clinic"
							class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-sm"
							required
						/>
					</div>

					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">Body</label>
						<textarea
							bind:value={composeBody}
							rows="8"
							placeholder="Compose the letter content…"
							class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-sm resize-y"
						></textarea>
					</div>
				</div>

				<div class="px-6 py-4 border-t border-gray-100 flex justify-end gap-3">
					<button
						type="button"
						onclick={closeComposeModal}
						class="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors"
					>
						Cancel
					</button>
					<button
						type="submit"
						disabled={formSubmitting}
						class="px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-lg hover:bg-blue-700 disabled:opacity-50 transition-colors"
					>
						{formSubmitting ? 'Saving...' : 'Compose Letter'}
					</button>
				</div>
			</form>
		</div>
	</div>
{/if}
