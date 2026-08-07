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

	$effect(() => { loadClaims(); });

	async function loadClaims() {
		loading = true;
		try {
			claims = await api.get<Claim[]>('/claims/recent');
		} catch (e: any) {
			error = e.message || 'Failed to load claims.';
		} finally { loading = false; }
	}
</script>

<div class="bg-white rounded-lg shadow-sm border border-gray-200">
	<div class="px-6 py-4 border-b border-gray-100 flex justify-between items-center">
		<h2 class="text-lg font-semibold text-gray-900">Claims Manager</h2>
		<button onclick={loadClaims} class="px-3 py-1 text-sm text-blue-600 hover:bg-blue-50 rounded-md">Refresh</button>
	</div>
	{#if loading}
		<div class="flex justify-center py-12"><div class="w-6 h-6 border-2 border-blue-500 border-t-transparent rounded-full animate-spin"></div></div>
	{:else if error}
		<p class="p-6 text-red-600 text-sm">{error}</p>
	{:else if claims.length === 0}
		<p class="p-6 text-gray-500 text-sm">No claims found.</p>
	{:else}
		<table class="w-full text-sm">
			<thead class="bg-gray-50 text-gray-500 uppercase text-xs">
				<tr><th class="px-6 py-3 text-left">ID</th><th class="px-6 py-3 text-left">Date</th><th class="px-6 py-3 text-left">CPT</th><th class="px-6 py-3 text-left">User</th></tr>
			</thead>
			<tbody>
				{#each claims as claim}
					<tr class="border-t border-gray-100 hover:bg-gray-50">
						<td class="px-6 py-3 font-mono text-xs">{claim.id}</td>
						<td class="px-6 py-3">{claim.cltimestamp?.slice(0, 10)}</td>
						<td class="px-6 py-3 font-mono">{claim.cpt_code || '—'}</td>
						<td class="px-6 py-3">{claim.username || '—'}</td>
					</tr>
				{/each}
			</tbody>
		</table>
	{/if}
</div>
