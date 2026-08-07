<script lang="ts">
	import { api } from '$lib/api';

	interface Superbill {
		id: number;
		patient: string;
		date_from: string;
		date_to: string;
		provider: string;
		status: string;
		total_charges: number;
		date_created: string;
		user: string;
	}

	interface SuperbillForm {
		patient: number;
		date_from: string;
		date_to: string;
		provider: number;
		status: string;
		total_charges: number;
	}

	let superbills = $state<Superbill[]>([]);
	let loading = $state(true);
	let error = $state('');
	let showModal = $state(false);
	let submitting = $state(false);
	let formError = $state('');

	let form: SuperbillForm = $state({
		patient: 0,
		date_from: '',
		date_to: '',
		provider: 0,
		status: 'draft',
		total_charges: 0,
	});

	$effect(() => { loadSuperbills(); });

	async function loadSuperbills() {
		loading = true;
		error = '';
		try {
			superbills = await api.get<Superbill[]>('/superbills');
		} catch (e: any) {
			error = e.message || 'Failed to load superbills.';
		} finally { loading = false; }
	}

	function openCreate() {
		formError = '';
		form = { patient: 0, date_from: '', date_to: '', provider: 0, status: 'draft', total_charges: 0 };
		showModal = true;
	}

	function closeModal() {
		showModal = false;
		formError = '';
	}

	async function handleCreate() {
		formError = '';
		if (!form.patient || !form.date_from || !form.date_to || !form.provider) {
			formError = 'Please fill in all required fields.';
			return;
		}
		submitting = true;
		try {
			await api.post('/superbills', form);
			closeModal();
			await loadSuperbills();
		} catch (e: any) {
			formError = e.message || 'Failed to create superbill.';
		} finally { submitting = false; }
	}

	function statusColor(status: string): string {
		switch (status?.toLowerCase()) {
			case 'draft': return 'bg-gray-100 text-gray-700';
			case 'submitted': return 'bg-blue-100 text-blue-800';
			case 'paid': return 'bg-green-100 text-green-800';
			case 'denied': return 'bg-red-100 text-red-800';
			default: return 'bg-gray-100 text-gray-600';
		}
	}

	function formatCurrency(val: number): string {
		return '$' + val.toFixed(2);
	}
</script>

<div class="bg-white rounded-lg shadow-sm border border-gray-200">
	<div class="px-6 py-4 border-b border-gray-100 flex justify-between items-center">
		<h2 class="text-lg font-semibold text-gray-900">Superbills</h2>
		<div class="flex gap-2">
			<button onclick={openCreate} class="px-4 py-2 text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 rounded-md transition-colors">+ New Superbill</button>
			<button onclick={loadSuperbills} class="px-3 py-1 text-sm text-blue-600 hover:bg-blue-50 rounded-md transition-colors">Refresh</button>
		</div>
	</div>

	{#if loading}
		<div class="flex justify-center py-12"><div class="w-6 h-6 border-2 border-blue-500 border-t-transparent rounded-full animate-spin"></div></div>
	{:else if error}
		<p class="p-6 text-red-600 text-sm">{error}</p>
	{:else if superbills.length === 0}
		<p class="p-6 text-gray-500 text-sm">No superbills found.</p>
	{:else}
		<div class="overflow-x-auto">
			<table class="w-full text-sm">
				<thead class="bg-gray-50 text-gray-500 uppercase text-xs">
					<tr>
						<th class="px-6 py-3 text-left">ID</th>
						<th class="px-6 py-3 text-left">Patient</th>
						<th class="px-6 py-3 text-left">Date Range</th>
						<th class="px-6 py-3 text-left">Provider</th>
						<th class="px-6 py-3 text-left">Status</th>
						<th class="px-6 py-3 text-right">Total Charges</th>
						<th class="px-6 py-3 text-left">Created</th>
					</tr>
				</thead>
				<tbody>
					{#each superbills as sb}
						<tr class="border-t border-gray-100 hover:bg-gray-50">
							<td class="px-6 py-3 font-mono text-xs">{sb.id}</td>
							<td class="px-6 py-3">{sb.patient || '—'}</td>
							<td class="px-6 py-3 whitespace-nowrap">{sb.date_from?.slice(0, 10)} – {sb.date_to?.slice(0, 10)}</td>
							<td class="px-6 py-3">{sb.provider || '—'}</td>
							<td class="px-6 py-3">
								<span class="inline-flex px-2 py-0.5 text-xs font-medium rounded-full {statusColor(sb.status)}">
									{sb.status || 'unknown'}
								</span>
							</td>
							<td class="px-6 py-3 text-right font-mono">{formatCurrency(sb.total_charges)}</td>
							<td class="px-6 py-3 whitespace-nowrap">{sb.date_created?.slice(0, 10)}</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{/if}
</div>

<!-- Create Superbill Modal -->
{#if showModal}
	<!-- backdrop -->
	<div class="fixed inset-0 z-40 bg-black/30" onclick={closeModal} onkeydown={(e) => { if (e.key === 'Escape') closeModal(); }}></div>
	<!-- modal -->
	<div class="fixed inset-0 z-50 flex items-center justify-center p-4">
		<div class="bg-white rounded-lg shadow-xl border border-gray-200 w-full max-w-md">
			<div class="px-6 py-4 border-b border-gray-100 flex justify-between items-center">
				<h3 class="text-lg font-semibold text-gray-900">New Superbill</h3>
				<button onclick={closeModal} class="text-gray-400 hover:text-gray-600 text-xl leading-none">&times;</button>
			</div>
			<div class="px-6 py-4 space-y-4">
				{#if formError}
					<p class="text-sm text-red-600 bg-red-50 rounded px-3 py-2">{formError}</p>
				{/if}
				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">Patient ID <span class="text-red-500">*</span></label>
					<input type="number" bind:value={form.patient} class="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-1 focus:ring-blue-500" placeholder="Patient ID" />
				</div>
				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">Date From <span class="text-red-500">*</span></label>
					<input type="date" bind:value={form.date_from} class="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-1 focus:ring-blue-500" />
				</div>
				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">Date To <span class="text-red-500">*</span></label>
					<input type="date" bind:value={form.date_to} class="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-1 focus:ring-blue-500" />
				</div>
				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">Provider ID <span class="text-red-500">*</span></label>
					<input type="number" bind:value={form.provider} class="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-1 focus:ring-blue-500" placeholder="Provider ID" />
				</div>
				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">Status</label>
					<select bind:value={form.status} class="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-1 focus:ring-blue-500">
						<option value="draft">Draft</option>
						<option value="submitted">Submitted</option>
						<option value="paid">Paid</option>
						<option value="denied">Denied</option>
					</select>
				</div>
				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">Total Charges</label>
					<input type="number" step="0.01" bind:value={form.total_charges} class="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-1 focus:ring-blue-500" placeholder="0.00" />
				</div>
			</div>
			<div class="px-6 py-4 border-t border-gray-100 flex justify-end gap-2">
				<button onclick={closeModal} class="px-4 py-2 text-sm font-medium text-gray-700 bg-gray-100 hover:bg-gray-200 rounded-md transition-colors" disabled={submitting}>Cancel</button>
				<button onclick={handleCreate} class="px-4 py-2 text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 rounded-md transition-colors disabled:opacity-50" disabled={submitting}>
					{#if submitting}
						<span class="inline-block w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin mr-1 align-middle"></span>
					{/if}
					Create
				</button>
			</div>
		</div>
	</div>
{/if}
