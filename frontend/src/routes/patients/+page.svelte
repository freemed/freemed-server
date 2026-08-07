<script lang="ts">
	import { goto } from '$app/navigation';
	import { api } from '$lib/api';

	interface PicklistItem {
		id: number;
		value: string;
	}

	interface SearchResult {
		id: number;
		last_name: string;
		first_name: string;
		patient_id: string;
		date_of_birth: string;
	}

	let query = $state('');
	let picklist = $state<PicklistItem[]>([]);
	let results = $state<SearchResult[]>([]);
	let loadingPicklist = $state(false);
	let loadingResults = $state(false);
	let searched = $state(false);
	let error = $state('');

	let debounceTimer: ReturnType<typeof setTimeout>;

	function onQueryInput(value: string) {
		query = value;
		clearTimeout(debounceTimer);

		if (query.length < 2) {
			picklist = [];
			return;
		}

		loadingPicklist = true;
		debounceTimer = setTimeout(async () => {
			try {
				const data = await api.get<PicklistItem[]>('/patients/picklist/' + encodeURIComponent(query));
				picklist = data || [];
			} catch (e: any) {
				picklist = [];
			} finally {
				loadingPicklist = false;
			}
		}, 250);
	}

	function selectFromPicklist(item: PicklistItem) {
		query = item.value;
		picklist = [];
		searchByQuery(item.value);
	}

	async function search() {
		if (query.length < 2) return;
		await searchByQuery(query);
	}

	async function searchByQuery(searchQuery: string) {
		loadingResults = true;
		error = '';
		searched = true;
		picklist = [];
		try {
			const data = await api.get<SearchResult[]>('/patients/picklist/' + encodeURIComponent(searchQuery));
			results = data || [];
		} catch (e: any) {
			error = e.message || 'Search failed';
			results = [];
		} finally {
			loadingResults = false;
		}
	}

	function goToPatient(id: number) {
		goto('/patients/' + id);
	}

	function formatDob(dob: string): string {
		if (!dob) return '';
		try {
			const d = new Date(dob);
			if (isNaN(d.getTime())) return dob;
			return d.toLocaleDateString('en-US', { year: 'numeric', month: 'short', day: 'numeric' });
		} catch {
			return dob;
		}
	}
</script>

<div class="max-w-5xl mx-auto">
	<h1 class="text-2xl font-bold text-gray-900 mb-6">Patient Search</h1>

	<!-- Search area -->
	<div class="relative mb-6">
		<div class="flex gap-2">
			<div class="relative flex-1">
				<input
					type="text"
					bind:value={query}
					oninput={(e) => onQueryInput(e.currentTarget.value)}
					onkeydown={(e) => { if (e.key === 'Enter') search(); }}
					placeholder="Search by name, patient ID, or date of birth..."
					class="w-full px-4 py-2.5 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none transition-shadow"
				/>
				{#if loadingPicklist}
					<div class="absolute right-3 top-1/2 -translate-y-1/2">
						<div class="w-4 h-4 border-2 border-blue-500 border-t-transparent rounded-full animate-spin"></div>
					</div>
				{/if}

				<!-- Picklist dropdown -->
				{#if picklist.length > 0 && !loadingResults}
					<div class="absolute z-10 w-full mt-1 bg-white border border-gray-200 rounded-lg shadow-lg max-h-60 overflow-y-auto">
						{#each picklist as item (item.id)}
							<button
								onclick={() => selectFromPicklist(item)}
								class="w-full text-left px-4 py-2 text-sm hover:bg-blue-50 transition-colors border-b border-gray-100 last:border-b-0"
							>
								{item.value}
							</button>
						{/each}
					</div>
				{/if}
			</div>
			<button
				onclick={search}
				disabled={loadingResults || query.length < 2}
				class="px-6 py-2.5 bg-blue-600 text-white text-sm font-medium rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
			>
				Search
			</button>
		</div>
	</div>

	<!-- Results area -->
	{#if error}
		<div class="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg text-sm mb-6">
			{error}
		</div>
	{/if}

	{#if loadingResults}
		<div class="flex justify-center py-12">
			<div class="w-6 h-6 border-2 border-blue-500 border-t-transparent rounded-full animate-spin"></div>
		</div>
	{:else if searched && results.length === 0 && !error}
		<div class="text-center py-12 text-gray-500">
			<p class="text-lg">No patients found</p>
			<p class="text-sm mt-1">Try adjusting your search terms</p>
		</div>
	{:else if results.length > 0}
		<div class="bg-white rounded-lg shadow-sm border border-gray-200 overflow-x-auto">
			<table class="w-full text-sm">
				<thead class="bg-gray-50 border-b border-gray-200">
					<tr>
						<th class="text-left px-4 py-3 font-medium text-gray-600">Last Name</th>
						<th class="text-left px-4 py-3 font-medium text-gray-600">First Name</th>
						<th class="text-left px-4 py-3 font-medium text-gray-600">Patient ID</th>
						<th class="text-left px-4 py-3 font-medium text-gray-600">Date of Birth</th>
					</tr>
				</thead>
				<tbody>
					{#each results as patient (patient.id)}
						<tr
							onclick={() => goToPatient(patient.id)}
							onkeydown={(e) => { if (e.key === 'Enter') goToPatient(patient.id); }}
							tabindex="0"
							role="button"
							class="border-b border-gray-100 hover:bg-blue-50 cursor-pointer transition-colors focus:outline-none focus:ring-2 focus:ring-inset focus:ring-blue-500"
						>
							<td class="px-4 py-3 text-gray-900 font-medium">{patient.last_name}</td>
							<td class="px-4 py-3 text-gray-700">{patient.first_name}</td>
							<td class="px-4 py-3 text-gray-500 font-mono text-xs">{patient.patient_id}</td>
							<td class="px-4 py-3 text-gray-500">{formatDob(patient.date_of_birth)}</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{/if}
</div>
