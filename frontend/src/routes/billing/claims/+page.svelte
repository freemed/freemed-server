<script lang="ts">
	import { api } from '$lib/api';

	interface Claim {
		id: number;
		cltimestamp: string;
		cpt_code: string;
		username: string;
		status?: string;
		amount?: number;
	}

	let claims = $state<Claim[]>([]);
	let loading = $state(true);
	let error = $state('');
	let activeTab = $state<'recent' | 'pending'>('recent');
	let updating = $state<Record<number, boolean>>({});

	$effect(() => { loadClaims(); });

	async function loadClaims() {
		loading = true;
		error = '';
		try {
			const endpoint = activeTab === 'recent' ? '/claims/recent' : '/claims/pending';
			claims = await api.get<Claim[]>(endpoint);
		} catch (e: any) {
			error = e.message || 'Failed to load claims.';
		} finally { loading = false; }
	}

	function switchTab(tab: 'recent' | 'pending') {
		activeTab = tab;
		loadClaims();
	}

	async function updateStatus(claimId: number, newStatus: string) {
		updating = { ...updating, [claimId]: true };
		try {
			await api.put(`/claims/${claimId}/status`, { status: newStatus });
			claims = claims.map(c => c.id === claimId ? { ...c, status: newStatus } : c);
		} catch (e: any) {
			error = e.message || 'Failed to update status.';
		} finally {
			updating = { ...updating, [claimId]: false };
		}
	}

	function statusColor(status: string | undefined): string {
		switch (status?.toLowerCase()) {
			case 'pending': return 'bg-yellow-100 text-yellow-800';
			case 'submitted': return 'bg-blue-100 text-blue-800';
			case 'paid': return 'bg-green-100 text-green-800';
			case 'denied': return 'bg-red-100 text-red-800';
			default: return 'bg-gray-100 text-gray-600';
		}
	}

	const STATUS_OPTIONS = ['pending', 'submitted', 'paid', 'denied'];
</script>

<div class="bg-white rounded-lg shadow-sm border border-gray-200">
	<div class="px-6 py-4 border-b border-gray-100 flex justify-between items-center">
		<h2 class="text-lg font-semibold text-gray-900">Claims Manager</h2>
		<button onclick={loadClaims} class="px-3 py-1 text-sm text-blue-600 hover:bg-blue-50 rounded-md transition-colors">Refresh</button>
	</div>

	<!-- Tab switcher -->
	<div class="flex border-b border-gray-100 bg-gray-50">
		<button
			onclick={() => switchTab('recent')}
			class="px-6 py-2.5 text-sm font-medium transition-colors border-b-2 {activeTab === 'recent' ? 'border-blue-600 text-blue-600' : 'border-transparent text-gray-500 hover:text-gray-700'}"
		>Recent Claims</button>
		<button
			onclick={() => switchTab('pending')}
			class="px-6 py-2.5 text-sm font-medium transition-colors border-b-2 {activeTab === 'pending' ? 'border-blue-600 text-blue-600' : 'border-transparent text-gray-500 hover:text-gray-700'}"
		>Pending Claims</button>
	</div>

	{#if loading}
		<div class="flex justify-center py-12"><div class="w-6 h-6 border-2 border-blue-500 border-t-transparent rounded-full animate-spin"></div></div>
	{:else if error}
		<p class="p-6 text-red-600 text-sm">{error}</p>
	{:else if claims.length === 0}
		<p class="p-6 text-gray-500 text-sm">No claims found.</p>
	{:else}
		<div class="overflow-x-auto">
			<table class="w-full text-sm">
				<thead class="bg-gray-50 text-gray-500 uppercase text-xs">
					<tr>
						<th class="px-6 py-3 text-left">ID</th>
						<th class="px-6 py-3 text-left">Date</th>
						<th class="px-6 py-3 text-left">CPT</th>
						<th class="px-6 py-3 text-left">User</th>
						<th class="px-6 py-3 text-left">Status</th>
						<th class="px-6 py-3 text-left">Action</th>
					</tr>
				</thead>
				<tbody>
					{#each claims as claim}
						<tr class="border-t border-gray-100 hover:bg-gray-50">
							<td class="px-6 py-3 font-mono text-xs">{claim.id}</td>
							<td class="px-6 py-3 whitespace-nowrap">{claim.cltimestamp?.slice(0, 10)}</td>
							<td class="px-6 py-3 font-mono">{claim.cpt_code || '—'}</td>
							<td class="px-6 py-3">{claim.username || '—'}</td>
							<td class="px-6 py-3">
								<span class="inline-flex px-2 py-0.5 text-xs font-medium rounded-full {statusColor(claim.status)}">
									{claim.status || 'unknown'}
								</span>
							</td>
							<td class="px-6 py-3">
								<select
									value={claim.status || ''}
									onchange={(e) => updateStatus(claim.id, (e.target as HTMLSelectElement).value)}
									disabled={updating[claim.id]}
									class="text-xs border border-gray-300 rounded px-2 py-1 bg-white focus:outline-none focus:ring-1 focus:ring-blue-500 disabled:opacity-50"
								>
									<option value="" disabled>Update...</option>
									{#each STATUS_OPTIONS as s}
										<option value={s} selected={claim.status?.toLowerCase() === s}>{s}</option>
									{/each}
								</select>
								{#if updating[claim.id]}
									<span class="ml-2 inline-block w-3 h-3 border border-blue-400 border-t-transparent rounded-full animate-spin align-middle"></span>
								{/if}
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{/if}
</div>
