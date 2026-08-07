<script lang="ts">
import { page } from '$app/stores';
import { goto } from '$app/navigation';
import { api } from '$lib/api';
import DiagnosisList from '$lib/components/DiagnosisList.svelte';
	import GrowthChart from '$lib/components/GrowthChart.svelte';

	interface PatientInfo {
		patient_name: string;
		patient_id: string;
		date_of_birth_mdy: string;
		age: number | string;
		address_line_1: string;
		address_line_2: string;
		city: string;
		state: string;
		postal: string;
		csz: string;
		pcp: string;
		facility: string;
		pharmacy: string;
		hasallergy: boolean | string;
	}

	interface Attachment {
		id: number;
		filename: string;
		description: string;
		created_at: string;
	}

	let patientId = $derived($page.params.id || '');
	let patient = $state<PatientInfo | null>(null);
	let attachments = $state<Attachment[]>([]);
	let loading = $state(true);
	let loadingAttachments = $state(false);
	let error = $state('');

	async function loadPatient() {
		loading = true;
		error = '';
		try {
			patient = await api.get<PatientInfo>('/patient/' + patientId + '/info');
		} catch (e: any) {
			error = e.message || 'Failed to load patient';
			patient = null;
		} finally {
			loading = false;
		}
	}

	async function loadAttachments() {
		loadingAttachments = true;
		try {
			attachments = (await api.get<Attachment[]>('/patient/' + patientId + '/attachments')) || [];
		} catch {
			attachments = [];
		} finally {
			loadingAttachments = false;
		}
	}

	function goToProgressNotes() {
		goto('/patients/' + patientId + '/progress-notes');
	}

	function formatDate(dateStr: string): string {
		if (!dateStr) return '';
		try {
			const d = new Date(dateStr);
			if (isNaN(d.getTime())) return dateStr;
			return d.toLocaleDateString('en-US', { year: 'numeric', month: 'short', day: 'numeric' });
		} catch {
			return dateStr;
		}
	}

	function hasAllergy(): boolean {
		if (patient == null) return false;
		if (typeof patient.hasallergy === 'boolean') return patient.hasallergy;
		if (typeof patient.hasallergy === 'string') {
			const v = patient.hasallergy.toLowerCase();
			return v === 'true' || v === 'yes' || v === '1';
		}
		return false;
	}

	// Load on mount and when patientId changes
	$effect(() => {
		if (patientId) {
			loadPatient();
			loadAttachments();
		}
	});
</script>

<div class="max-w-4xl mx-auto">
	<!-- Header with back link -->
	<div class="flex items-center gap-3 mb-6">
		<button
			onclick={() => goto('/patients')}
			class="text-gray-500 hover:text-gray-700 transition-colors p-1 -ml-1"
			aria-label="Back to patient search"
		>
			<svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
			</svg>
		</button>
		<h1 class="text-2xl font-bold text-gray-900">
			{#if loading}
				Loading...
			{:else if patient}
				{patient.patient_name}
			{:else}
				Patient
			{/if}
		</h1>
	</div>

	{#if error}
		<div class="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg text-sm mb-6">
			{error}
		</div>
	{/if}

	{#if loading}
		<div class="flex justify-center py-16">
			<div class="w-6 h-6 border-2 border-blue-500 border-t-transparent rounded-full animate-spin"></div>
		</div>
	{:else if patient}
		<!-- Demographics Card -->
		<div class="bg-white rounded-lg shadow-sm border border-gray-200 mb-6">
			<div class="px-6 py-4 border-b border-gray-100 flex items-center justify-between">
				<h2 class="text-lg font-semibold text-gray-900">Demographics</h2>
				{#if hasAllergy()}
					<span class="px-2.5 py-0.5 bg-red-100 text-red-700 text-xs font-medium rounded-full">Allergies</span>
				{/if}
			</div>
			<div class="px-6 py-4 grid grid-cols-1 sm:grid-cols-2 gap-4">
				<div>
					<span class="block text-xs font-medium text-gray-500 uppercase tracking-wide">Patient ID</span>
					<span class="text-sm text-gray-900 font-mono">{patient.patient_id}</span>
				</div>
				<div>
					<span class="block text-xs font-medium text-gray-500 uppercase tracking-wide">Date of Birth</span>
					<span class="text-sm text-gray-900">{patient.date_of_birth_mdy}</span>
				</div>
				<div>
					<span class="block text-xs font-medium text-gray-500 uppercase tracking-wide">Age</span>
					<span class="text-sm text-gray-900">{patient.age}</span>
				</div>
				<div>
					<span class="block text-xs font-medium text-gray-500 uppercase tracking-wide">PCP</span>
					<span class="text-sm text-gray-900">{patient.pcp || '—'}</span>
				</div>
			</div>
		</div>

		<!-- Address Card -->
		<div class="bg-white rounded-lg shadow-sm border border-gray-200 mb-6">
			<div class="px-6 py-4 border-b border-gray-100">
				<h2 class="text-lg font-semibold text-gray-900">Address</h2>
			</div>
			<div class="px-6 py-4">
				<div class="text-sm text-gray-900 space-y-1">
					<p>{patient.address_line_1}</p>
					{#if patient.address_line_2}
						<p>{patient.address_line_2}</p>
					{/if}
					<p>{patient.csz || `${patient.city}, ${patient.state} ${patient.postal}`}</p>
				</div>
			</div>
		</div>

		<!-- Care Information Card -->
		<div class="bg-white rounded-lg shadow-sm border border-gray-200 mb-6">
			<div class="px-6 py-4 border-b border-gray-100">
				<h2 class="text-lg font-semibold text-gray-900">Care Information</h2>
			</div>
			<div class="px-6 py-4 grid grid-cols-1 sm:grid-cols-2 gap-4">
				<div>
					<span class="block text-xs font-medium text-gray-500 uppercase tracking-wide">Facility</span>
					<span class="text-sm text-gray-900">{patient.facility || '—'}</span>
				</div>
				<div>
					<span class="block text-xs font-medium text-gray-500 uppercase tracking-wide">Pharmacy</span>
					<span class="text-sm text-gray-900">{patient.pharmacy || '—'}</span>
				</div>
			</div>
		</div>

		<!-- Diagnoses -->
		<DiagnosisList patientId={patientId} />

		<GrowthChart {patientId} />

		<!-- EMR Attachments Section -->
		<div class="bg-white rounded-lg shadow-sm border border-gray-200 mb-6">
			<div class="px-6 py-4 border-b border-gray-100">
				<h2 class="text-lg font-semibold text-gray-900">EMR Attachments</h2>
			</div>
			<div class="px-6 py-4">
				{#if loadingAttachments}
					<div class="flex justify-center py-4">
						<div class="w-4 h-4 border-2 border-blue-500 border-t-transparent rounded-full animate-spin"></div>
					</div>
				{:else if attachments.length === 0}
					<p class="text-sm text-gray-500 py-2">No attachments on file</p>
				{:else}
					<div class="divide-y divide-gray-100">
						{#each attachments as att (att.id)}
							<div class="py-3 flex items-center justify-between">
								<div>
									<p class="text-sm font-medium text-gray-900">{att.filename}</p>
									{#if att.description}
										<p class="text-xs text-gray-500 mt-0.5">{att.description}</p>
									{/if}
								</div>
								<span class="text-xs text-gray-400">{formatDate(att.created_at)}</span>
							</div>
						{/each}
					</div>
				{/if}
			</div>
		</div>

		<!-- Quick Actions -->
		<div class="bg-white rounded-lg shadow-sm border border-gray-200 mb-6">
			<div class="px-6 py-4 border-b border-gray-100">
				<h2 class="text-lg font-semibold text-gray-900">Actions</h2>
			</div>
			<div class="px-6 py-4 flex flex-wrap gap-3">
				<button
					onclick={goToProgressNotes}
					class="px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded-lg hover:bg-blue-700 transition-colors"
				>
					Progress Notes
				</button>
				<button
					onclick={() => goto('/patients/' + patientId + '/encounters')}
					class="px-4 py-2 bg-white text-gray-700 text-sm font-medium rounded-lg border border-gray-300 hover:bg-gray-50 transition-colors"
				>
					Encounters
				</button>
				<button
					onclick={() => goto('/patients/' + patientId + '/documents')}
					class="px-4 py-2 bg-white text-gray-700 text-sm font-medium rounded-lg border border-gray-300 hover:bg-gray-50 transition-colors"
				>
					Documents
				</button>
				<button
					onclick={() => goto('/patients/' + patientId + '/allergies')}
					class="px-4 py-2 bg-white text-gray-700 text-sm font-medium rounded-lg border border-gray-300 hover:bg-gray-50 transition-colors"
				>
					Allergies
				</button>
			</div>
		</div>
	{:else if !error}
		<div class="text-center py-16">
			<p class="text-lg text-gray-500">Patient not found</p>
			<button
				onclick={() => goto('/patients')}
				class="mt-4 px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded-lg hover:bg-blue-700 transition-colors"
			>
				Back to Search
			</button>
		</div>
	{/if}
</div>
