<script lang="ts">
	import { page } from '$app/stores';
	import { api } from '$lib/api';

	interface Encounter {
		id: number;
		patient: number;
		module: string;
		oid: number;
		stamp: string;
		summary: string;
		locked: boolean;
		annotation: string | null;
		user: number;
		provider: number;
		language: string;
		status: string;
	}

	let encounters = $state<Encounter[]>([]);
	let loading = $state(true);
	let error = $state('');
	let patientId = $state('');
	let expandedId = $state<number | null>(null);

	$effect(() => {
		const id = $page.params.id;
		if (id) {
			patientId = id;
			loadEncounters(id);
		}
	});

	async function loadEncounters(id: string) {
		loading = true;
		error = '';
		try {
			const data = await api.get<Encounter[]>(`/patient/${id}/encounters`);
			encounters = data || [];
		} catch (e: any) {
			error = e.message || 'Failed to load encounters';
		} finally {
			loading = false;
		}
	}

	function formatDate(dateStr: string): string {
		if (!dateStr) return '';
		try {
			const d = new Date(dateStr);
			if (isNaN(d.getTime())) return dateStr;
			return d.toLocaleDateString('en-US', {
				year: 'numeric',
				month: 'long',
				day: 'numeric',
				hour: 'numeric',
				minute: '2-digit',
			});
		} catch {
			return dateStr;
		}
	}

	function toggleExpand(id: number) {
		expandedId = expandedId === id ? null : id;
	}
</script>

<div class="max-w-4xl mx-auto">
	<div class="flex items-center justify-between mb-6">
		<div>
			<h1 class="text-2xl font-bold text-gray-900">Encounters</h1>
			<p class="text-sm text-gray-500 mt-1">Patient #{patientId}</p>
		</div>
		<a
			href="/patients/{patientId}"
			class="text-sm text-blue-600 hover:text-blue-800 font-medium transition-colors"
		>
			&larr; Back to Patient
		</a>
	</div>

	{#if loading}
		<div class="flex justify-center py-12">
			<div class="w-6 h-6 border-2 border-blue-500 border-t-transparent rounded-full animate-spin"></div>
		</div>
	{:else if error}
		<div class="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg text-sm mb-6">
			{error}
		</div>
	{:else if encounters.length === 0}
		<div class="text-center py-12 text-gray-500">
			<p class="text-lg">No encounters found</p>
			<p class="text-sm mt-1">No encounter records exist for this patient yet.</p>
		</div>
	{:else}
		<!-- Timeline -->
		<div class="relative">
			<!-- Vertical line -->
			<div class="absolute left-6 top-0 bottom-0 w-0.5 bg-gray-200"></div>

			<div class="space-y-6">
				{#each encounters as encounter, i}
					<div class="relative pl-16">
						<!-- Timeline dot -->
						<button
							class="absolute left-4 top-5 w-4 h-4 rounded-full border-2 {i === 0 ? 'bg-blue-500 border-blue-500' : 'bg-white border-gray-300'} z-10 cursor-pointer hover:scale-110 transition-transform"
							onclick={() => toggleExpand(encounter.id)}
							aria-label={expandedId === encounter.id ? 'Collapse encounter details' : 'Expand encounter details'}
							title={expandedId === encounter.id ? 'Collapse' : 'Expand'}
						></button>

						<!-- Encounter card -->
						<div class="bg-white rounded-lg shadow-sm border border-gray-200 p-4">
							<div class="flex items-center justify-between mb-2">
								<div class="flex items-center gap-2">
									{#if encounter.locked}
										<span class="inline-block px-2 py-0.5 text-xs font-medium bg-orange-50 text-orange-700 rounded">Locked</span>
									{:else}
										<span class="inline-block px-2 py-0.5 text-xs font-medium bg-green-50 text-green-700 rounded">Open</span>
									{/if}
									<span class="text-sm text-gray-500">
										{formatDate(encounter.stamp)}
									</span>
								</div>
								<button
									class="text-xs text-blue-600 hover:text-blue-800 transition-colors"
									onclick={() => toggleExpand(encounter.id)}
								>
									{expandedId === encounter.id ? 'Collapse' : 'Details'}
								</button>
							</div>

							<p class="text-sm text-gray-700 font-medium mb-1">
								{encounter.summary || 'No summary'}
							</p>

							{#if expandedId === encounter.id}
								<div class="mt-3 pt-3 border-t border-gray-100 space-y-2 text-sm">
									<div class="grid grid-cols-2 gap-2">
										<div>
											<span class="font-medium text-gray-600">Encounter ID:</span>
											<span class="text-gray-700 ml-1">{encounter.id}</span>
										</div>
										<div>
											<span class="font-medium text-gray-600">Record ID:</span>
											<span class="text-gray-700 ml-1">{encounter.oid}</span>
										</div>
										<div>
											<span class="font-medium text-gray-600">Language:</span>
											<span class="text-gray-700 ml-1">{encounter.language || 'N/A'}</span>
										</div>
										<div>
											<span class="font-medium text-gray-600">Status:</span>
											<span class="text-gray-700 ml-1">{encounter.status || 'N/A'}</span>
										</div>
										<div>
											<span class="font-medium text-gray-600">Provider:</span>
											<span class="text-gray-700 ml-1">{encounter.provider || 'N/A'}</span>
										</div>
									</div>
									{#if encounter.annotation}
										<div>
											<span class="font-medium text-gray-600">Annotation:</span>
											<span class="text-gray-700 ml-1">{encounter.annotation}</span>
										</div>
									{/if}
								</div>
							{/if}
						</div>
					</div>
				{/each}
			</div>
		</div>
	{/if}
</div>
