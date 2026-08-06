<script lang="ts">
	import { api } from '$lib/api';

	interface ConfigGroup {
		c_section: string;
		entries: { key: string; value: string }[];
	}

	let configGroups = $state<ConfigGroup[]>([]);
	let loading = $state(true);
	let error = $state('');

	async function loadSettings() {
		loading = true;
		error = '';
		try {
			const data = await api.get<Record<string, Record<string, string>>>('/config/all');
			// Group by c_section (top-level key in the response)
			configGroups = Object.entries(data).map(([section, entries]) => ({
				c_section: section,
				entries: Object.entries(entries as Record<string, string>).map(([key, value]) => ({
					key,
					value,
				})),
			}));
		} catch (e: any) {
			error = e.message || 'Failed to load settings';
		} finally {
			loading = false;
		}
	}

	$effect(() => {
		loadSettings();
	});
</script>

<div class="max-w-5xl mx-auto">
	<h1 class="text-2xl font-bold text-gray-900 mb-6">Settings</h1>

	{#if loading}
		<div class="flex justify-center py-12">
			<div class="w-6 h-6 border-2 border-blue-500 border-t-transparent rounded-full animate-spin"></div>
		</div>
	{:else if error}
		<div class="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg text-sm mb-6">
			{error}
		</div>
	{:else if configGroups.length === 0}
		<div class="text-center py-12 text-gray-500">
			<p class="text-lg">No settings configured</p>
		</div>
	{:else}
		<div class="space-y-6">
			{#each configGroups as group}
				<div class="bg-white rounded-lg shadow-sm border border-gray-200 overflow-hidden">
					<div class="bg-gray-50 px-4 py-3 border-b border-gray-200">
						<h2 class="text-sm font-semibold text-gray-700 uppercase tracking-wide">
							{group.c_section}
						</h2>
					</div>
					<div class="divide-y divide-gray-100">
						{#each group.entries as entry}
							<div class="px-4 py-3 flex items-center justify-between">
								<span class="text-sm text-gray-600">{entry.key}</span>
								<span class="text-sm font-mono text-gray-900 bg-gray-50 px-2 py-0.5 rounded">
									{entry.value}
								</span>
							</div>
						{/each}
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>
