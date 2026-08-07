<script lang="ts">
	import { page } from '$app/stores';
	import { api } from '$lib/api';

	interface Address {
		id: number;
		patient: number;
		line1: string | null;
		line2: string | null;
		city: string | null;
		stpr: string | null;
		postal: string | null;
		zip: string | null;
		active: boolean;
		created_at: string;
		updated_at: string;
	}

	let addresses = $state<Address[]>([]);
	let loading = $state(true);
	let error = $state('');
	let patientId = $state('');

	// Modal state
	let showModal = $state(false);
	let formSaving = $state(false);
	let formError = $state('');

	// Form fields
	let line1 = $state('');
	let line2 = $state('');
	let city = $state('');
	let stpr = $state('');
	let zip = $state('');
	let addressType = $state('');

	$effect(() => {
		const id = $page.params.id;
		if (id) {
			patientId = id;
			loadAddresses(id);
		}
	});

	async function loadAddresses(id: string) {
		loading = true;
		error = '';
		try {
			const data = await api.get<Address[]>(`/patient/${id}/addresses`);
			addresses = data || [];
		} catch (e: any) {
			error = e.message || 'Failed to load addresses';
		} finally {
			loading = false;
		}
	}

	function fmt(s: string | null): string {
		return s || '—';
	}

	function openModal() {
		line1 = '';
		line2 = '';
		city = '';
		stpr = '';
		zip = '';
		addressType = '';
		formError = '';
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

	async function handleSubmit(e: Event) {
		e.preventDefault();

		if (!line1.trim()) {
			formError = 'Address line 1 is required';
			return;
		}
		if (!city.trim()) {
			formError = 'City is required';
			return;
		}
		if (!stpr.trim()) {
			formError = 'State is required';
			return;
		}
		if (!zip.trim()) {
			formError = 'ZIP code is required';
			return;
		}

		formSaving = true;
		formError = '';

		const address = {
			line1: line1.trim(),
			line2: line2.trim(),
			city: city.trim(),
			stpr: stpr.trim(),
			postal: zip.trim(),
		};

		try {
			await api.post(`/patient/${patientId}/addresses/bulk`, {
				addresses: [address],
			});
			closeModal();
			await loadAddresses(patientId);
		} catch (e: any) {
			formError = e.message || 'Failed to save address';
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
</script>

<div class="max-w-4xl mx-auto">
	<div class="flex items-center justify-between mb-6">
		<div>
			<h1 class="text-2xl font-bold text-gray-900">Addresses</h1>
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
				+ Add Address
			</button>
		</div>
	</div>

	{#if error}
		<div class="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg text-sm mb-6">
			{error}
		</div>
	{/if}

	<!-- Address Add Modal -->
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
					<h2 class="text-lg font-semibold text-gray-900">Add Address</h2>
					<button
						onclick={closeModal}
						class="text-gray-400 hover:text-gray-600 p-1 rounded-lg hover:bg-gray-100"
						aria-label="Close"
					>
						<svg
							xmlns="http://www.w3.org/2000/svg"
							class="w-5 h-5"
							fill="none"
							viewBox="0 0 24 24"
							stroke="currentColor"
							stroke-width="2"
						>
							<path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
						</svg>
					</button>
				</div>
				<!-- Body -->
				<div class="px-6 py-4">
					{#if formError}
						<div class="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg text-sm mb-4">
							{formError}
						</div>
					{/if}
					<form onsubmit={handleSubmit}>
						<div class="space-y-4">
							<div>
								<label class="block text-sm font-medium text-gray-700 mb-1">Address Line 1 *</label>
								<input
									type="text"
									bind:value={line1}
									placeholder="e.g. 123 Main St"
									class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-sm"
									required
								/>
							</div>
							<div>
								<label class="block text-sm font-medium text-gray-700 mb-1">Address Line 2</label>
								<input
									type="text"
									bind:value={line2}
									placeholder="e.g. Apt 4B"
									class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-sm"
								/>
							</div>
							<div class="grid grid-cols-2 gap-3">
								<div>
									<label class="block text-sm font-medium text-gray-700 mb-1">City *</label>
									<input
										type="text"
										bind:value={city}
										placeholder="e.g. Hartford"
										class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-sm"
										required
									/>
								</div>
								<div>
									<label class="block text-sm font-medium text-gray-700 mb-1">State *</label>
									<input
										type="text"
										bind:value={stpr}
										placeholder="e.g. CT"
										maxlength="2"
										class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-sm"
										required
									/>
								</div>
							</div>
							<div class="grid grid-cols-2 gap-3">
								<div>
									<label class="block text-sm font-medium text-gray-700 mb-1">ZIP Code *</label>
									<input
										type="text"
										bind:value={zip}
										placeholder="e.g. 06103"
										class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-sm"
										required
									/>
								</div>
								<div>
									<label class="block text-sm font-medium text-gray-700 mb-1">Address Type</label>
									<select
										bind:value={addressType}
										class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-sm bg-white"
									>
										<option value="">Select type...</option>
										<option value="home">Home</option>
										<option value="work">Work</option>
										<option value="billing">Billing</option>
										<option value="mailing">Mailing</option>
										<option value="other">Other</option>
									</select>
								</div>
							</div>
						</div>
						<!-- Footer -->
						<div class="flex justify-end gap-3 mt-6">
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
								{formSaving ? 'Saving...' : 'Save Address'}
							</button>
						</div>
					</form>
				</div>
			</div>
		</div>
	{/if}

	{#if loading}
		<div class="flex justify-center py-12">
			<div class="w-6 h-6 border-2 border-blue-500 border-t-transparent rounded-full animate-spin"></div>
		</div>
	{:else if addresses.length === 0}
		<div class="text-center py-12 text-gray-500 bg-white rounded-lg shadow-sm border border-gray-200">
			<p class="text-lg">No addresses recorded</p>
			<p class="text-sm mt-1">Click "Add Address" to record one.</p>
		</div>
	{:else}
		<div class="bg-white rounded-lg shadow-sm border border-gray-200 overflow-x-auto">
			<table class="min-w-full divide-y divide-gray-200">
				<thead class="bg-gray-50">
					<tr>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Line 1</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Line 2</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">City</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">State</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">ZIP</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Created</th>
					</tr>
				</thead>
				<tbody class="bg-white divide-y divide-gray-200">
					{#each addresses as addr}
						<tr class="hover:bg-gray-50">
							<td class="px-4 py-3 text-sm text-gray-900 font-medium">
								{fmt(addr.line1)}
							</td>
							<td class="px-4 py-3 text-sm text-gray-700">
								{fmt(addr.line2)}
							</td>
							<td class="px-4 py-3 text-sm text-gray-700">
								{fmt(addr.city)}
							</td>
							<td class="px-4 py-3 text-sm text-gray-700">
								{fmt(addr.stpr)}
							</td>
							<td class="px-4 py-3 text-sm text-gray-700">
								{fmt(addr.postal || addr.zip)}
							</td>
							<td class="px-4 py-3 text-sm text-gray-500 whitespace-nowrap">
								{formatDate(addr.created_at)}
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{/if}
</div>
