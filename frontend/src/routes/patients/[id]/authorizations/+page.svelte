<script lang="ts">
	import { page } from '$app/stores';
	import { api } from '$lib/api';

	interface Authorization {
		id: number;
		authnum: string;
		authtype: number;
		authdtbegin: string;
		authdtend: string;
		authvisits: number;
		authvisitsused: number;
		authvisitsremain: number;
		insconame: { String: string; Valid: boolean } | null;
		active: string;
	}

	let authorizations = $state<Authorization[]>([]);
	let loading = $state(true);
	let error = $state('');
	let patientId = $state('');

	$effect(() => {
		const id = $page.params.id;
		if (id) {
			patientId = id;
			loadAuthorizations(id);
		}
	});

	async function loadAuthorizations(id: string) {
		loading = true;
		error = '';
		try {
			const data = await api.get<Authorization[]>(`/patient/${id}/authorizations`);
			authorizations = data || [];
		} catch (e: any) {
			error = e.message || 'Failed to load authorizations';
		} finally {
			loading = false;
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

	function inscoName(auth: Authorization): string {
		if (auth.insconame && auth.insconame.Valid) return auth.insconame.String;
		return '—';
	}

	function authTypeLabel(t: number): string {
		// Common authorization types; extend as needed
		switch (t) {
			case 1: return 'Referral';
			case 2: return 'Procedure';
			case 3: return 'Medication';
			case 4: return 'Admission';
			default: return `Type ${t}`;
		}
	}

	function statusBadgeClass(s: string): string {
		switch (s.toLowerCase()) {
			case 'active':
				return 'bg-green-50 text-green-700';
			case 'expired':
				return 'bg-gray-50 text-gray-500';
			case 'pending':
				return 'bg-yellow-50 text-yellow-700';
			case 'denied':
				return 'bg-red-50 text-red-700';
			default:
				return 'bg-gray-50 text-gray-600';
		}
	}
</script>

<div class="max-w-4xl mx-auto">
	<div class="flex items-center justify-between mb-6">
		<div>
			<h1 class="text-2xl font-bold text-gray-900">Authorizations</h1>
			<p class="text-sm text-gray-500 mt-1">Patient #{patientId}</p>
		</div>
		<div class="flex items-center gap-3">
			<a
				href="/patients/{patientId}"
				class="text-sm text-blue-600 hover:text-blue-800 font-medium transition-colors"
			>
				&larr; Back to Patient
			</a>
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
	{:else if authorizations.length === 0}
		<div class="text-center py-12 text-gray-500 bg-white rounded-lg shadow-sm border border-gray-200">
			<p class="text-lg">No authorizations found</p>
			<p class="text-sm mt-1">No authorizations have been recorded for this patient yet.</p>
		</div>
	{:else}
		<div class="bg-white rounded-lg shadow-sm border border-gray-200 overflow-x-auto">
			<table class="min-w-full divide-y divide-gray-200">
				<thead class="bg-gray-50">
					<tr>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Auth #</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Type</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Begin Date</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">End Date</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Visits Used</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Visits Remaining</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Insurance Company</th>
						<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
					</tr>
				</thead>
				<tbody class="bg-white divide-y divide-gray-200">
					{#each authorizations as auth}
						<tr class="hover:bg-gray-50">
							<td class="px-4 py-3 text-sm font-medium text-gray-900 whitespace-nowrap">
								{auth.authnum || '—'}
							</td>
							<td class="px-4 py-3 text-sm text-gray-600 whitespace-nowrap">
								{authTypeLabel(auth.authtype)}
							</td>
							<td class="px-4 py-3 text-sm text-gray-600 whitespace-nowrap">
								{formatDate(auth.authdtbegin)}
							</td>
							<td class="px-4 py-3 text-sm text-gray-600 whitespace-nowrap">
								{formatDate(auth.authdtend)}
							</td>
							<td class="px-4 py-3 text-sm text-gray-600 whitespace-nowrap">
								{auth.authvisitsused}
							</td>
							<td class="px-4 py-3 text-sm text-gray-600 whitespace-nowrap">
								{auth.authvisitsremain}
							</td>
							<td class="px-4 py-3 text-sm text-gray-600">
								{inscoName(auth)}
							</td>
							<td class="px-4 py-3 text-sm">
								<span class="inline-block px-2 py-0.5 text-xs font-medium rounded {statusBadgeClass(auth.active)}">
									{auth.active}
								</span>
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{/if}
</div>
