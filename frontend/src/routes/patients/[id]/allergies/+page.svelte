<script lang="ts">
	import { page } from '$app/stores';
	import { api } from '$lib/api';

	interface Allergy {
		id: number;
		created_at: string;
		updated_at: string;
		deleted_at: string | null;
		patient: number;
		active: string;
	}

	let allergies = $state<Allergy[]>([]);
	let loading = $state(true);
	let error = $state('');
	let patientId = $state('');
	let adding = $state(false);

	$effect(() => {
		const id = $page.params.id;
		if (id) {
			patientId = id;
			loadAllergies(id);
		}
	});

	async function loadAllergies(id: string) {
		loading = true;
		error = '';
		try {
			const data = await api.get<Allergy[]>(`/patient/${id}/allergies`);
			allergies = data || [];
		} catch (e: any) {
			error = e.message || 'Failed to load allergies';
		} finally {
			loading = false;
		}
	}

	async function addAllergy() {
		adding = true;
		error = '';
		try {
			const result = await api.post<{ id: number }>(`/patient/${patientId}/allergies`, {});
			await loadAllergies(patientId);
		} catch (e: any) {
			error = e.message || 'Failed to add allergy';
		} finally {
			adding = false;
		}
	}

	async function deactivateAllergy(allergyId: number) {
		error = '';
		try {
			await api.del(`/allergies/${allergyId}`);
			// Remove from local list
			allergies = allergies.filter((a) => a.id !== allergyId);
		} catch (e: any) {
			error = e.message || 'Failed to deactivate allergy';
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
</script>

<div class="max-w-4xl mx-auto">
	<div class="flex items-center justify-between mb-6">
		<div>
			<h1 class="text-2xl font-bold text-gray-900">Allergies</h1>
			<p class="text-sm text-gray-500 mt-1">Patient #{patientId}</p>
		</div>
		<a
			href="/patients/{patientId}"
			class="text-sm text-blue-600 hover:text-blue-800 font-medium transition-colors"
		>
			&larr; Back to Patient
		</a>
	</div>

	{#if error}
		<div class="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg text-sm mb-6">
			{error}
		</div>
	{/if}

	<!-- Add Allergy Form -->
	<div class="bg-white rounded-lg shadow-sm border border-gray-200 p-4 mb-6">
		<div class="flex items-center justify-between">
			<h2 class="text-lg font-semibold text-gray-900">Manage Allergies</h2>
			<button
				onclick={addAllergy}
				disabled={adding}
				class="px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
			>
				{#if adding}
					Adding...
				{:else}
					+ Add Allergy
				{/if}
			</button>
		</div>
	</div>

	{#if loading}
		<div class="flex justify-center py-12">
			<div class="w-6 h-6 border-2 border-blue-500 border-t-transparent rounded-full animate-spin"></div>
		</div>
	{:else if allergies.length === 0}
		<div class="text-center py-12 text-gray-500 bg-white rounded-lg shadow-sm border border-gray-200">
			<p class="text-lg">No allergies recorded</p>
			<p class="text-sm mt-1">Click "Add Allergy" to record a new allergy for this patient.</p>
		</div>
	{:else}
		<div class="bg-white rounded-lg shadow-sm border border-gray-200">
			<div class="divide-y divide-gray-100">
				{#each allergies as allergy (allergy.id)}
					<div class="px-6 py-4 flex items-center justify-between">
						<div>
							<p class="text-sm font-medium text-gray-900">
								Allergy #{allergy.id}
							</p>
							<p class="text-xs text-gray-500 mt-0.5">
								Recorded {formatDate(allergy.created_at)}
							</p>
						</div>
						<div class="flex items-center gap-3">
							<span class="inline-block px-2 py-0.5 text-xs font-medium bg-green-50 text-green-700 rounded">
								Active
							</span>
							<button
								onclick={() => deactivateAllergy(allergy.id)}
								class="text-sm text-red-600 hover:text-red-800 font-medium transition-colors"
							>
								Deactivate
							</button>
						</div>
					</div>
				{/each}
			</div>
		</div>
	{/if}
</div>
