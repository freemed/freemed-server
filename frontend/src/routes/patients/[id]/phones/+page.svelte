<script lang="ts">
	import { page } from '$app/stores';
	import { api } from '$lib/api';

	interface Phone {
		id: number;
		patient: number;
		type: string;
		number: string;
		active: string;
		created_at: string;
		updated_at: string;
	}

	interface PhoneForm {
		number: string;
		phone_type: string;
	}

	const PHONE_TYPES = ['home', 'work', 'mobile', 'fax'] as const;

	// Patient ID from route param
	let patientId = $state('');

	// Data state
	let phones = $state<Phone[]>([]);
	let loading = $state(true);
	let error = $state('');

	// Modal state
	let showCreateModal = $state(false);
	let formSaving = $state(false);
	let formError = $state('');
	let form: PhoneForm = $state({ number: '', phone_type: 'home' });

	// Load on mount and when patientId changes
	$effect(() => {
		const id = $page.params.id;
		if (id) {
			patientId = id;
			loadPhones(id);
		}
	});

	async function loadPhones(id: string) {
		loading = true;
		error = '';
		try {
			const data = await api.get<Phone[]>(`/patient/${id}/phones`);
			phones = data || [];
		} catch (e: any) {
			error = e.message || 'Failed to load phones';
		} finally {
			loading = false;
		}
	}

	// --- Create Modal ---
	function openCreateModal() {
		form = { number: '', phone_type: 'home' };
		formError = '';
		showCreateModal = true;
	}

	function closeModal() {
		showCreateModal = false;
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

		// Basic validation
		if (!form.number.trim()) {
			formError = 'Phone number is required.';
			return;
		}
		if (!form.phone_type) {
			formError = 'Phone type is required.';
			return;
		}

		formSaving = true;
		try {
			await api.post(`/patient/${patientId}/phones`, {
				number: form.number.trim(),
				phone_type: form.phone_type
			});

			closeModal();
			await loadPhones(patientId);
		} catch (e: any) {
			formError = e.message || 'Failed to add phone';
		} finally {
			formSaving = false;
		}
	}

	// --- Helpers ---
	function formatDate(dateStr: string): string {
		if (!dateStr) return '';
		try {
			const d = new Date(dateStr);
			if (isNaN(d.getTime())) return dateStr;
			return d.toLocaleDateString('en-US', {
				year: 'numeric',
				month: 'short',
				day: 'numeric'
			});
		} catch {
			return dateStr;
		}
	}

	function typeLabel(type: string): string {
		return type.charAt(0).toUpperCase() + type.slice(1);
	}
</script>

<div class="max-w-4xl mx-auto">
	<!-- Header -->
	<div class="flex items-center justify-between mb-6">
		<div>
			<h1 class="text-2xl font-bold text-gray-900">Phones</h1>
			<p class="text-sm text-gray-500 mt-1">Patient #{patientId}</p>
		</div>
		<a
			href="/patients/{patientId}"
			class="text-sm text-blue-600 hover:text-blue-800 font-medium transition-colors"
		>
			&larr; Back to Patient
		</a>
	</div>

	<!-- Content Card -->
	<div class="bg-white rounded-lg shadow-sm border border-gray-200">
		<div class="p-6">
			<div class="flex items-center justify-between mb-4">
				<h2 class="text-lg font-semibold text-gray-900">Phone Numbers</h2>
				<button
					onclick={openCreateModal}
					class="px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded-lg hover:bg-blue-700 transition-colors"
				>
					+ Add Phone
				</button>
			</div>

			{#if loading}
				<div class="flex justify-center py-12">
					<div
						class="w-6 h-6 border-2 border-blue-500 border-t-transparent rounded-full animate-spin"
					></div>
				</div>
			{:else if error}
				<div class="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg text-sm">
					{error}
				</div>
			{:else if phones.length === 0}
				<div class="text-center py-12 text-gray-500">
					<p class="text-lg">No phone numbers recorded</p>
					<p class="text-sm mt-1">
						Click "Add Phone" to record a phone number for this patient.
					</p>
				</div>
			{:else}
				<div class="overflow-x-auto">
					<table class="w-full text-sm">
						<thead>
							<tr class="border-b border-gray-200">
								<th class="text-left py-2 px-3 font-semibold text-gray-600">Phone Number</th>
								<th class="text-left py-2 px-3 font-semibold text-gray-600">Type</th>
								<th class="text-left py-2 px-3 font-semibold text-gray-600">Recorded</th>
							</tr>
						</thead>
						<tbody class="divide-y divide-gray-100">
							{#each phones as phone (phone.id)}
								<tr class="hover:bg-gray-50 transition-colors">
									<td class="py-2.5 px-3 text-gray-900 font-medium">
										{phone.number}
									</td>
									<td class="py-2.5 px-3 text-gray-700">
										<span
											class="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium"
											class:bg-blue-100={phone.type === 'home'}
											class:text-blue-800={phone.type === 'home'}
											class:bg-green-100={phone.type === 'work'}
											class:text-green-800={phone.type === 'work'}
											class:bg-purple-100={phone.type === 'mobile'}
											class:text-purple-800={phone.type === 'mobile'}
											class:bg-orange-100={phone.type === 'fax'}
											class:text-orange-800={phone.type === 'fax'}
											class:bg-gray-100={!['home', 'work', 'mobile', 'fax'].includes(phone.type)}
											class:text-gray-800={!['home', 'work', 'mobile', 'fax'].includes(phone.type)}
										>
											{typeLabel(phone.type)}
										</span>
									</td>
									<td class="py-2.5 px-3 text-gray-500 text-xs">
										{formatDate(phone.created_at)}
									</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			{/if}
		</div>
	</div>
</div>

<!-- Add Phone Modal -->
{#if showCreateModal}
	<!-- svelte-ignore a11y_no_static_element_interactions -->
	<div
		class="fixed inset-0 z-50 flex items-center justify-center bg-black/40"
		onclick={onModalBackdropClick}
		onkeydown={onModalKeydown}
		role="dialog"
		aria-modal="true"
		aria-label="Add Phone"
		tabindex="-1"
	>
		<div class="bg-white rounded-xl shadow-xl w-full max-w-md mx-4 overflow-hidden">
			<!-- Header -->
			<div class="flex items-center justify-between px-6 py-4 border-b border-gray-200">
				<h2 class="text-lg font-semibold text-gray-900">Add Phone</h2>
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
					<label for="phoneNumber" class="block text-sm font-medium text-gray-700 mb-1">
						Phone Number
					</label>
					<input
						id="phoneNumber"
						type="tel"
						bind:value={form.number}
						placeholder="(555) 123-4567"
						class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none transition-colors"
					/>
				</div>

				<div>
					<label for="phoneType" class="block text-sm font-medium text-gray-700 mb-1">
						Type
					</label>
					<select
						id="phoneType"
						bind:value={form.phone_type}
						class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none transition-colors bg-white"
					>
						{#each PHONE_TYPES as pt}
							<option value={pt}>{typeLabel(pt)}</option>
						{/each}
					</select>
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
