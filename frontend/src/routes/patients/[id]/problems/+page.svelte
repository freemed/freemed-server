<script lang="ts">
	import { page } from '$app/stores';
	import { api } from '$lib/api';

	interface Problem {
		id: number;
		patient: number;
		date: string;
		problem: string;
		created_at: string;
	}

	interface ProblemForm {
		date: string;
		problem: string;
	}

	// Patient ID from route param
	let patientId = $state('');

	// Tab state
	type TabKey = 'current' | 'chronic';
	let activeTab = $state<TabKey>('current');

	// Current problems state
	let currentProblems = $state<Problem[]>([]);
	let currentLoading = $state(true);
	let currentError = $state('');

	// Chronic problems state
	let chronicProblems = $state<Problem[]>([]);
	let chronicLoading = $state(true);
	let chronicError = $state('');

	// Modal state
	let showCreateModal = $state(false);
	let createModalTab = $state<TabKey>('current');
	let formSaving = $state(false);
	let formError = $state('');
	let form: ProblemForm = $state({ date: '', problem: '' });

	// Load on mount and when patientId changes
	$effect(() => {
		const id = $page.params.id;
		if (id) {
			patientId = id;
			loadCurrentProblems(id);
			loadChronicProblems(id);
		}
	});

	// --- Current Problems ---
	async function loadCurrentProblems(id: string) {
		currentLoading = true;
		currentError = '';
		try {
			const data = await api.get<Problem[]>(`/patient/${id}/current-problems`);
			currentProblems = data || [];
		} catch (e: any) {
			currentError = e.message || 'Failed to load current problems';
		} finally {
			currentLoading = false;
		}
	}

	// --- Chronic Problems ---
	async function loadChronicProblems(id: string) {
		chronicLoading = true;
		chronicError = '';
		try {
			const data = await api.get<Problem[]>(`/patient/${id}/chronic-problems`);
			chronicProblems = data || [];
		} catch (e: any) {
			chronicError = e.message || 'Failed to load chronic problems';
		} finally {
			chronicLoading = false;
		}
	}

	// --- Create Modal ---
	function openCreateModal(tab: TabKey) {
		createModalTab = tab;
		form = { date: new Date().toISOString().slice(0, 10), problem: '' };
		formError = '';
		showCreateModal = true;
	}

	function closeModal() {
		showCreateModal = false;
		formError = '';
	}

	function onModalBackdropClick(e: MouseEvent) {
		if (e.target === e.currentTarget) {
			closeModal();
		}
	}

	function onModalKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') {
			closeModal();
		}
	}

	async function handleCreate() {
		formError = '';

		// Basic validation
		if (!form.date.trim()) {
			formError = 'Date is required.';
			return;
		}
		if (!form.problem.trim()) {
			formError = 'Problem description is required.';
			return;
		}

		formSaving = true;
		try {
			const endpoint =
				createModalTab === 'current'
					? `/patient/${patientId}/current-problems`
					: `/patient/${patientId}/chronic-problems`;

			await api.post(endpoint, {
				date: form.date.trim(),
				problem: form.problem.trim()
			});

			closeModal();

			// Reload the appropriate list
			if (createModalTab === 'current') {
				await loadCurrentProblems(patientId);
			} else {
				await loadChronicProblems(patientId);
			}
		} catch (e: any) {
			formError = e.message || 'Failed to create problem';
		} finally {
			formSaving = false;
		}
	}

	// --- Helpers ---
	function formatDate(dateStr: string): string {
		if (!dateStr) return '';
		try {
			const d = new Date(dateStr);
			if (isNaN(d.getTime())) return dateStr;
			return d.toLocaleDateString('en-US', {
				year: 'numeric',
				month: 'short',
				day: 'numeric'
			});
		} catch {
			return dateStr;
		}
	}

	function tabLabel(tab: TabKey): string {
		return tab === 'current' ? 'Current Problems' : 'Chronic Problems';
	}
</script>

<div class="max-w-4xl mx-auto">
	<!-- Header -->
	<div class="flex items-center justify-between mb-6">
		<div>
			<h1 class="text-2xl font-bold text-gray-900">Problems</h1>
			<p class="text-sm text-gray-500 mt-1">Patient #{patientId}</p>
		</div>
		<a
			href="/patients/{patientId}"
			class="text-sm text-blue-600 hover:text-blue-800 font-medium transition-colors"
		>
			&larr; Back to Patient
		</a>
	</div>

	<!-- Tabs -->
	<div class="bg-white rounded-lg shadow-sm border border-gray-200 mb-6">
		<div class="border-b border-gray-200">
			<div class="flex -mb-px" role="tablist">
				{#each (['current', 'chronic'] as const) as tab}
					<button
						class="px-6 py-3 text-sm font-medium border-b-2 transition-colors whitespace-nowrap"
						class:border-blue-500={activeTab === tab}
						class:text-blue-600={activeTab === tab}
						class:border-transparent={activeTab !== tab}
						class:text-gray-500={activeTab !== tab}
						class:hover:text-gray-700={activeTab !== tab}
						class:hover:border-gray-300={activeTab !== tab}
						onclick={() => (activeTab = tab)}
						role="tab"
						aria-selected={activeTab === tab}
					>
						{tabLabel(tab)}
					</button>
				{/each}
				</div>
		</div>

		<!-- Current Problems Tab Panel -->
		{#if activeTab === 'current'}
			<div class="p-6" role="tabpanel">
				<div class="flex items-center justify-between mb-4">
					<h2 class="text-lg font-semibold text-gray-900">Current Problems</h2>
					<button
						onclick={() => openCreateModal('current')}
						class="px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded-lg hover:bg-blue-700 transition-colors"
					>
						+ Add Current Problem
					</button>
				</div>

				{#if currentLoading}
					<div class="flex justify-center py-12">
						<div
							class="w-6 h-6 border-2 border-blue-500 border-t-transparent rounded-full animate-spin"
						></div>
					</div>
				{:else if currentError}
					<div class="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg text-sm">
						{currentError}
					</div>
				{:else if currentProblems.length === 0}
					<div class="text-center py-12 text-gray-500">
						<p class="text-lg">No current problems recorded</p>
						<p class="text-sm mt-1">
							Click "Add Current Problem" to record a problem for this patient.
						</p>
					</div>
				{:else}
					<div class="overflow-x-auto">
						<table class="w-full text-sm">
							<thead>
								<tr class="border-b border-gray-200">
									<th class="text-left py-2 px-3 font-semibold text-gray-600">Date</th>
									<th class="text-left py-2 px-3 font-semibold text-gray-600">Problem</th>
									<th class="text-left py-2 px-3 font-semibold text-gray-600">Recorded</th>
								</tr>
							</thead>
							<tbody class="divide-y divide-gray-100">
								{#each currentProblems as problem (problem.id)}
									<tr class="hover:bg-gray-50 transition-colors">
										<td class="py-2.5 px-3 text-gray-900 font-medium">
											{formatDate(problem.date)}
										</td>
										<td class="py-2.5 px-3 text-gray-700">{problem.problem}</td>
										<td class="py-2.5 px-3 text-gray-500 text-xs">
											{formatDate(problem.created_at)}
										</td>
									</tr>
								{/each}
							</tbody>
						</table>
					</div>
				{/if}
			</div>
		{/if}

		<!-- Chronic Problems Tab Panel -->
		{#if activeTab === 'chronic'}
			<div class="p-6" role="tabpanel">
				<div class="flex items-center justify-between mb-4">
					<h2 class="text-lg font-semibold text-gray-900">Chronic Problems</h2>
					<button
						onclick={() => openCreateModal('chronic')}
						class="px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded-lg hover:bg-blue-700 transition-colors"
					>
						+ Add Chronic Problem
					</button>
				</div>

				{#if chronicLoading}
					<div class="flex justify-center py-12">
						<div
							class="w-6 h-6 border-2 border-blue-500 border-t-transparent rounded-full animate-spin"
						></div>
					</div>
				{:else if chronicError}
					<div class="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg text-sm">
						{chronicError}
					</div>
				{:else if chronicProblems.length === 0}
					<div class="text-center py-12 text-gray-500">
						<p class="text-lg">No chronic problems recorded</p>
						<p class="text-sm mt-1">
							Click "Add Chronic Problem" to record a chronic problem for this patient.
						</p>
					</div>
				{:else}
					<div class="overflow-x-auto">
						<table class="w-full text-sm">
							<thead>
								<tr class="border-b border-gray-200">
									<th class="text-left py-2 px-3 font-semibold text-gray-600">Date</th>
									<th class="text-left py-2 px-3 font-semibold text-gray-600">Problem</th>
									<th class="text-left py-2 px-3 font-semibold text-gray-600">Recorded</th>
								</tr>
							</thead>
							<tbody class="divide-y divide-gray-100">
								{#each chronicProblems as problem (problem.id)}
									<tr class="hover:bg-gray-50 transition-colors">
										<td class="py-2.5 px-3 text-gray-900 font-medium">
											{formatDate(problem.date)}
										</td>
										<td class="py-2.5 px-3 text-gray-700">{problem.problem}</td>
										<td class="py-2.5 px-3 text-gray-500 text-xs">
											{formatDate(problem.created_at)}
										</td>
									</tr>
								{/each}
							</tbody>
						</table>
					</div>
				{/if}
			</div>
		{/if}
	</div>
</div>

<!-- Create Problem Modal -->
{#if showCreateModal}
	<!-- svelte-ignore a11y_no_static_element_interactions -->
	<div
		class="fixed inset-0 z-50 flex items-center justify-center bg-black/40"
		onclick={onModalBackdropClick}
		onkeydown={onModalKeydown}
		role="dialog"
		aria-modal="true"
		aria-label="Add {tabLabel(createModalTab)}"
		tabindex="-1"
	>
		<div class="bg-white rounded-xl shadow-xl w-full max-w-md mx-4 overflow-hidden">
			<!-- Header -->
			<div class="flex items-center justify-between px-6 py-4 border-b border-gray-200">
				<h2 class="text-lg font-semibold text-gray-900">
					Add {tabLabel(createModalTab)}
				</h2>
				<button
					onclick={closeModal}
					class="text-gray-400 hover:text-gray-600 transition-colors p-1 rounded-lg hover:bg-gray-100"
					aria-label="Close"
				>
					<svg
						xmlns="http://www.w3.org/2000/svg"
						class="h-5 w-5"
						viewBox="0 0 20 20"
						fill="currentColor"
					>
						<path
							fill-rule="evenodd"
							d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z"
							clip-rule="evenodd"
						/>
					</svg>
				</button>
			</div>

			<!-- Body -->
			<div class="px-6 py-4 space-y-4">
				{#if formError}
					<div class="bg-red-50 border border-red-200 text-red-700 px-3 py-2 rounded-lg text-sm">
						{formError}
					</div>
				{/if}

				<div>
					<label for="problemDate" class="block text-sm font-medium text-gray-700 mb-1">
						Date
					</label>
					<input
						id="problemDate"
						type="date"
						bind:value={form.date}
						class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none transition-colors"
					/>
				</div>

				<div>
					<label for="problemDesc" class="block text-sm font-medium text-gray-700 mb-1">
						Problem Description
					</label>
					<input
						id="problemDesc"
						type="text"
						bind:value={form.problem}
						placeholder="Enter problem description"
						class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none transition-colors"
					/>
				</div>
			</div>

			<!-- Footer -->
			<div class="px-6 py-4 bg-gray-50 border-t border-gray-200 flex justify-end gap-3">
				<button
					onclick={closeModal}
					class="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors"
				>
					Cancel
				</button>
				<button
					onclick={handleCreate}
					disabled={formSaving}
					class="px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors flex items-center gap-2"
				>
					{#if formSaving}
						<div
							class="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin"
						></div>
						Saving...
					{:else}
						Save
					{/if}
				</button>
			</div>
		</div>
	</div>
{/if}
