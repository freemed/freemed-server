<script lang="ts">
	import { api } from '$lib/api';

	interface Diagnosis {
		dx_type: string;
		icd9code: string;
		icd9descrip: string;
	}

	interface Props {
		patientId: string;
	}

	let { patientId }: Props = $props();

	let diagnoses = $state<Diagnosis[]>([]);
	let loading = $state(true);
	let error = $state('');

	async function loadDiagnoses() {
		loading = true;
		error = '';
		try {
			diagnoses = (await api.get<Diagnosis[]>('/patient/' + patientId + '/diagnoses')) || [];
		} catch (e: any) {
			error = e.message || 'Failed to load diagnoses';
			diagnoses = [];
		} finally {
			loading = false;
		}
	}

	$effect(() => {
		if (patientId) {
			loadDiagnoses();
		}
	});
</script>

<div class="bg-white rounded-lg shadow-sm border border-gray-200 mb-6">
	<div class="px-6 py-4 border-b border-gray-100">
		<h2 class="text-lg font-semibold text-gray-900">Diagnoses</h2>
	</div>
	<div class="px-6 py-4">
		{#if loading}
			<div class="flex justify-center py-4">
				<div class="w-4 h-4 border-2 border-blue-500 border-t-transparent rounded-full animate-spin"></div>
			</div>
		{:else if error}
			<p class="text-sm text-red-600">{error}</p>
		{:else if diagnoses.length === 0}
			<p class="text-sm text-gray-500 py-2">No diagnoses on file</p>
		{:else}
			<div class="divide-y divide-gray-100">
				{#each diagnoses as dx, i (i)}
					<div class="py-3 flex items-start gap-4">
						<span
							class="mt-0.5 px-2 py-0.5 text-xs font-medium rounded-full shrink-0
							{dx.dx_type === 'primary' ? 'bg-blue-100 text-blue-700' : 'bg-gray-100 text-gray-600'}"
						>
							{dx.dx_type}
						</span>
						<div>
							<span class="text-sm font-mono font-medium text-gray-900">{dx.icd9code}</span>
							<p class="text-sm text-gray-600 mt-0.5">{dx.icd9descrip}</p>
						</div>
					</div>
				{/each}
			</div>
		{/if}
	</div>
</div>
