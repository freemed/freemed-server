<script lang="ts">
	import { page } from '$app/stores';
	import { api } from '$lib/api';

	interface ClinicalOrder {
		id: number;
		created_at: string;
		updated_at: string;
		deleted_at: string | null;
		patient: number;
		order_type: string;
		description: string;
		status: string;
		date_ordered: string | null;
		ordering_provider: number;
		notes: string | null;
		user: number;
		active: string;
	}

	let orders = $state<ClinicalOrder[]>([]);
	let loading = $state(true);
	let error = $state('');
	let patientId = $state('');

	// Modal state
	let showCreateModal = $state(false);
	let formError = $state('');
	let formSubmitting = $state(false);
	let orderType = $state('');
	let description = $state('');
	let dateOrdered = $state('');
	let status = $state('Ordered');
	let notes = $state('');

	$effect(() => {
		const id = $page.params.id;
		if (id) {
			patientId = id;
			loadOrders(id);
		}
	});

	async function loadOrders(id: string) {
		loading = true;
		error = '';
		try {
			const data = await api.get<ClinicalOrder[]>(`/patient/${id}/clinical-orders`);
			orders = data || [];
		} catch (e: any) {
			error = e.message || 'Failed to load clinical orders';
		} finally {
			loading = false;
		}
	}

	function openCreateModal() {
		orderType = '';
		description = '';
		dateOrdered = '';
		status = 'Ordered';
		notes = '';
		formError = '';
		showCreateModal = true;
	}

	function closeCreateModal() {
		showCreateModal = false;
	}

	async function handleCreate(e: Event) {
		e.preventDefault();

		if (!orderType.trim()) {
			formError = 'Order type is required';
			return;
		}
		if (!description.trim()) {
			formError = 'Description is required';
			return;
		}
		if (!dateOrdered) {
			formError = 'Order date is required';
			return;
		}

		formSubmitting = true;
		formError = '';

		const payload = {
			order_type: orderType.trim(),
			description: description.trim(),
			status: status,
			date_ordered: dateOrdered,
			ordering_provider: 0,
			notes: notes.trim(),
		};

		try {
			await api.post(`/patient/${patientId}/clinical-orders`, payload);
			showCreateModal = false;
			await loadOrders(patientId);
		} catch (e: any) {
			formError = e.message || 'Failed to create clinical order';
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

	function statusBadgeClass(s: string): string {
		switch (s) {
			case 'Ordered':
				return 'bg-blue-50 text-blue-700';
			case 'Pending':
				return 'bg-yellow-50 text-yellow-700';
			case 'Completed':
				return 'bg-green-50 text-green-700';
			case 'Cancelled':
				return 'bg-gray-50 text-gray-500';
			default:
				return 'bg-gray-50 text-gray-600';
		}
	}
</script>

<div class="max-w-4xl mx-auto">
	<div class="flex items-center justify-between mb-6">
		<div>
			<h1 class="text-2xl font-bold text-gray-900">Clinical Orders</h1>
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
				onclick={openCreateModal}
				class="px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded-lg hover:bg-blue-700 transition-colors"
			>
				+ Add Order
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
	{:else if orders.length === 0}
		<div class="text-center py-12 text-gray-500 bg-white rounded-lg shadow-sm border border-gray-200">
			<p class="text-lg">No clinical orders found</p>
			<p class="text-sm mt-1">No orders have been recorded for this patient yet.</p>
		</div>
	{:else}
		<div class="bg-white rounded-lg shadow-sm border border-gray-200 overflow-x-auto">
			<table class="min-w-full divide-y divide-gray-200">
				<thead class="bg-gray-50">
					<tr>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Order Date</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Type</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Description</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
					</tr>
				</thead>
				<tbody class="bg-white divide-y divide-gray-200">
					{#each orders as order}
						<tr class="hover:bg-gray-50">
							<td class="px-4 py-3 text-sm text-gray-600 whitespace-nowrap">
								{formatDate(order.date_ordered)}
							</td>
							<td class="px-4 py-3 text-sm font-medium text-gray-900">
								{order.order_type || '—'}
							</td>
							<td class="px-4 py-3 text-sm text-gray-600">
								{order.description || '—'}
							</td>
							<td class="px-4 py-3 text-sm">
								<span class="inline-block px-2 py-0.5 text-xs font-medium rounded {statusBadgeClass(order.status)}">
									{order.status || '—'}
								</span>
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{/if}
</div>

<!-- Create Clinical Order Modal -->
{#if showCreateModal}
	<!-- svelte-ignore a11y_no_static_element_interactions -->
	<div
		class="fixed inset-0 z-50 flex items-center justify-center"
		onkeydown={(e) => { if (e.key === 'Escape') closeCreateModal(); }}
	>
		<!-- Backdrop -->
		<!-- svelte-ignore a11y_no_static_element_interactions -->
		<div
			class="absolute inset-0 bg-black/50"
			onclick={closeCreateModal}
		></div>

		<!-- Modal -->
		<div class="relative bg-white rounded-lg shadow-xl w-full max-w-lg mx-4 max-h-[90vh] overflow-y-auto">
			<div class="px-6 py-4 border-b border-gray-100 flex items-center justify-between">
				<h2 class="text-lg font-semibold text-gray-900">Add Clinical Order</h2>
				<button
					onclick={closeCreateModal}
					class="text-gray-400 hover:text-gray-600 transition-colors"
					aria-label="Close"
				>
					<svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
					</svg>
				</button>
			</div>

			<form onsubmit={handleCreate}>
				<div class="px-6 py-4 space-y-4">
					{#if formError}
						<div class="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg text-sm">
							{formError}
						</div>
					{/if}

					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">Order Type *</label>
						<input
							type="text"
							bind:value={orderType}
							placeholder="e.g. Lab, Imaging, Referral"
							class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-sm"
						/>
					</div>

					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">Description *</label>
						<textarea
							bind:value={description}
							rows="2"
							placeholder="Description of the order..."
							class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-sm resize-none"
						></textarea>
					</div>

					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">Order Date *</label>
						<input
							type="date"
							bind:value={dateOrdered}
							class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-sm"
						/>
					</div>

					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">Status</label>
						<select
							bind:value={status}
							class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-sm"
						>
							<option value="Ordered">Ordered</option>
							<option value="Pending">Pending</option>
							<option value="Completed">Completed</option>
							<option value="Cancelled">Cancelled</option>
						</select>
					</div>

					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">Notes</label>
						<textarea
							bind:value={notes}
							rows="2"
							placeholder="Additional notes..."
							class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-sm resize-none"
						></textarea>
					</div>
				</div>

				<div class="px-6 py-4 border-t border-gray-100 flex justify-end gap-3">
					<button
						type="button"
						onclick={closeCreateModal}
						class="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors"
					>
						Cancel
					</button>
					<button
						type="submit"
						disabled={formSubmitting}
						class="px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-lg hover:bg-blue-700 disabled:opacity-50 transition-colors"
					>
						{formSubmitting ? 'Saving...' : 'Add Order'}
					</button>
				</div>
			</form>
		</div>
	</div>
{/if}
