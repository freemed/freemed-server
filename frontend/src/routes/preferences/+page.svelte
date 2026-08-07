<script lang="ts">
	import { api } from '$lib/api';

	interface ConfigItem {
		id: number;
		c_option: string;
		c_value: string;
		c_title: string;
		c_section: string;
		c_type: string;
		c_options: string;
	}

	interface SectionGroup {
		section: string;
		items: ConfigItem[];
	}

	let items = $state<ConfigItem[]>([]);
	let loading = $state(true);
	let error = $state('');

	const sections = $derived.by<SectionGroup[]>(() => {
		const map = new Map<string, ConfigItem[]>();
		for (const item of items) {
			const section = item.c_section || 'General';
			if (!map.has(section)) map.set(section, []);
			map.get(section)!.push(item);
		}
		return Array.from(map.entries())
			.map(([section, sectionItems]) => ({ section, items: sectionItems }))
			.sort((a, b) => a.section.localeCompare(b.section));
	});

	async function loadPreferences() {
		loading = true;
		error = '';
		try {
			const data = await api.get<ConfigItem[]>('/config/all');
			items = data || [];
		} catch (e: any) {
			error = e.message || 'Failed to load preferences';
			items = [];
		} finally {
			loading = false;
		}
	}

	function displayValue(item: ConfigItem): string {
		if (item.c_type === 'boolean' || item.c_type === 'bool') {
			return item.c_value === '1' || item.c_value === 'true' ? 'Yes' : 'No';
		}
		if (item.c_value === null || item.c_value === undefined || item.c_value === '') {
			return '<em>—</em>';
		}
		return item.c_value;
	}

	loadPreferences();
</script>

<div class="max-w-4xl mx-auto">
	<h1 class="text-2xl font-bold text-gray-900 mb-6">Preferences</h1>

	{#if error}
		<div class="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg text-sm mb-6">
			{error}
		</div>
	{/if}

	{#if loading}
		<div class="flex justify-center py-12">
			<div class="w-6 h-6 border-2 border-blue-500 border-t-transparent rounded-full animate-spin"></div>
		</div>
	{:else if sections.length === 0}
		<div class="text-center py-12 text-gray-500">
			<p class="text-lg">No preferences found</p>
		</div>
	{:else}
		<div class="space-y-6">
			{#each sections as group (group.section)}
				<div class="bg-white rounded-lg border border-gray-200 overflow-x-auto">
					<div class="bg-gray-50 px-4 py-3 border-b border-gray-200">
						<h2 class="text-sm font-semibold text-gray-700 uppercase tracking-wider">
							{group.section}
						</h2>
					</div>
					<table class="w-full text-sm">
						<thead class="sr-only">
							<tr>
								<th>Option</th>
								<th>Value</th>
							</tr>
						</thead>
						<tbody>
							{#each group.items as item (item.id)}
								<tr class="border-b border-gray-100 last:border-b-0">
									<td class="px-4 py-2.5 text-gray-700 font-medium w-1/3">
										{item.c_title || item.c_option}
									</td>
									<td class="px-4 py-2.5 text-gray-600">
										{@html displayValue(item)}
									</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			{/each}
		</div>
	{/if}
</div>
