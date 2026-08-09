<script lang="ts">
	import { page } from '$app/stores';
	import { api } from '$lib/api';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
	import ErrorBanner from '$lib/components/ErrorBanner.svelte';

	interface SocialHistory {
		id: number;
		patient: number;
		smoking_status: string;
		smoking_detail: string;
		alcohol_use: string;
		alcohol_detail: string;
		drug_use: string;
		drug_detail: string;
		exercise_frequency: string;
		occupation: string;
		living_situation: string;
		food_insecurity: boolean;
		transportation_access: boolean;
		notes: string | null;
		recorded_date: string;
		user: number;
		active: string;
		created_at: string;
		updated_at: string;
	}

	let patientId = $state('');
	let loading = $state(true);
	let error = $state('');
	let latest = $state<SocialHistory | null>(null);
	let history = $state<SocialHistory[]>([]);

	// Form state
	let smokingStatus = $state('');
	let smokingDetail = $state('');
	let alcoholUse = $state('');
	let alcoholDetail = $state('');
	let drugUse = $state('');
	let drugDetail = $state('');
	let exerciseFrequency = $state('');
	let occupation = $state('');
	let livingSituation = $state('');
	let foodInsecurity = $state(false);
	let transportationAccess = $state(false);
	let notes = $state('');
	let recordedDate = $state('');

	let formSubmitting = $state(false);
	let formError = $state('');

	$effect(() => {
		const id = $page.params.id;
		if (id) {
			patientId = id;
			loadData(id);
		}
	});

	async function loadData(id: string) {
		loading = true;
		error = '';
		try {
			const [latestData, historyData] = await Promise.all([
				api.get<SocialHistory | null>(`/patient/${id}/social-history/latest`),
				api.get<SocialHistory[]>(`/patient/${id}/social-history`),
			]);
			latest = latestData;
			history = historyData || [];

			// Pre-populate form with latest data
			if (latestData) {
				smokingStatus = latestData.smoking_status;
				smokingDetail = latestData.smoking_detail;
				alcoholUse = latestData.alcohol_use;
				alcoholDetail = latestData.alcohol_detail;
				drugUse = latestData.drug_use;
				drugDetail = latestData.drug_detail;
				exerciseFrequency = latestData.exercise_frequency;
				occupation = latestData.occupation;
				livingSituation = latestData.living_situation;
				foodInsecurity = latestData.food_insecurity;
				transportationAccess = latestData.transportation_access;
				notes = latestData.notes || '';
			}
			// Always default recorded_date to today
			if (!recordedDate) {
				recordedDate = new Date().toISOString().split('T')[0];
			}
		} catch (e: any) {
			error = e.message || 'Failed to load social history';
		} finally {
			loading = false;
		}
	}

	function resetForm() {
		smokingStatus = '';
		smokingDetail = '';
		alcoholUse = '';
		alcoholDetail = '';
		drugUse = '';
		drugDetail = '';
		exerciseFrequency = '';
		occupation = '';
		livingSituation = '';
		foodInsecurity = false;
		transportationAccess = false;
		notes = '';
		recordedDate = new Date().toISOString().split('T')[0];
		formError = '';
	}

	async function handleSubmit(e: Event) {
		e.preventDefault();
		formError = '';

		if (!recordedDate) {
			formError = 'Recorded date is required';
			return;
		}

		formSubmitting = true;
		try {
			await api.post(`/patient/${patientId}/social-history`, {
				smoking_status: smokingStatus,
				smoking_detail: smokingDetail.trim(),
				alcohol_use: alcoholUse,
				alcohol_detail: alcoholDetail.trim(),
				drug_use: drugUse,
				drug_detail: drugDetail.trim(),
				exercise_frequency: exerciseFrequency,
				occupation: occupation.trim(),
				living_situation: livingSituation,
				food_insecurity: foodInsecurity,
				transportation_access: transportationAccess,
				notes: notes.trim(),
				recorded_date: recordedDate,
			});
			resetForm();
			await loadData(patientId);
		} catch (e: any) {
			formError = e.message || 'Failed to save social history';
		} finally {
			formSubmitting = false;
		}
	}

	async function handleRemove(entryId: number) {
		if (!confirm('Are you sure you want to remove this entry?')) return;
		try {
			await api.del(`/patient/${patientId}/social-history/${entryId}`);
			await loadData(patientId);
		} catch (e: any) {
			alert(e.message || 'Failed to remove entry');
		}
	}

	function formatDate(dateStr: string): string {
		if (!dateStr) return '—';
		try {
			const d = new Date(dateStr);
			if (isNaN(d.getTime())) return dateStr;
			return d.toLocaleDateString('en-US', { year: 'numeric', month: 'short', day: 'numeric' });
		} catch {
			return dateStr;
		}
	}

	function formatLabel(val: string): string {
		if (!val) return '—';
		return val.replace(/-/g, ' ').replace(/\b\w/g, (c) => c.toUpperCase());
	}
</script>

<div class="max-w-4xl mx-auto">
	<div class="flex items-center justify-between mb-6">
		<div>
			<h1 class="text-2xl font-bold text-gray-900">Social History</h1>
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
		<ErrorBanner message={error} onRetry={() => loadData(patientId)} />
	{/if}

	{#if loading}
		<LoadingSpinner message="Loading social history…" />
	{:else}
		<!-- Record New Entry Form -->
		<div class="bg-white rounded-lg shadow-sm border border-gray-200 p-6 mb-6">
			<h2 class="text-lg font-semibold text-gray-800 mb-4">
				{#if latest}
					Record New Social History Entry
				{:else}
					Record Social History
				{/if}
			</h2>

			{#if formError}
				<div class="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg text-sm mb-4">
					{formError}
				</div>
			{/if}

			<form onsubmit={handleSubmit}>
				<div class="grid grid-cols-1 md:grid-cols-2 gap-4 mb-4">
					<!-- Smoking Status -->
					<div>
						<label for="smoking-status" class="block text-sm font-medium text-gray-700 mb-1">Smoking Status</label>
						<select
							id="smoking-status"
							bind:value={smokingStatus}
							class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:ring-1 focus:ring-blue-500"
						>
							<option value="">— Select —</option>
							<option value="current-smoker">Current Smoker</option>
							<option value="former-smoker">Former Smoker</option>
							<option value="never-smoker">Never Smoker</option>
							<option value="unknown">Unknown</option>
						</select>
					</div>

					<!-- Smoking Detail -->
					<div>
						<label for="smoking-detail" class="block text-sm font-medium text-gray-700 mb-1">Smoking Detail</label>
						<input
							id="smoking-detail"
							type="text"
							bind:value={smokingDetail}
							placeholder="e.g. 1 pack/day for 10 years"
							class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:ring-1 focus:ring-blue-500"
						/>
					</div>

					<!-- Alcohol Use -->
					<div>
						<label for="alcohol-use" class="block text-sm font-medium text-gray-700 mb-1">Alcohol Use</label>
						<select
							id="alcohol-use"
							bind:value={alcoholUse}
							class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:ring-1 focus:ring-blue-500"
						>
							<option value="">— Select —</option>
							<option value="none">None</option>
							<option value="light">Light</option>
							<option value="moderate">Moderate</option>
							<option value="heavy">Heavy</option>
							<option value="former">Former</option>
						</select>
					</div>

					<!-- Alcohol Detail -->
					<div>
						<label for="alcohol-detail" class="block text-sm font-medium text-gray-700 mb-1">Alcohol Detail</label>
						<input
							id="alcohol-detail"
							type="text"
							bind:value={alcoholDetail}
							placeholder="e.g. 2 drinks/weekend"
							class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:ring-1 focus:ring-blue-500"
						/>
					</div>

					<!-- Drug Use -->
					<div>
						<label for="drug-use" class="block text-sm font-medium text-gray-700 mb-1">Drug Use</label>
						<select
							id="drug-use"
							bind:value={drugUse}
							class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:ring-1 focus:ring-blue-500"
						>
							<option value="">— Select —</option>
							<option value="none">None</option>
							<option value="former">Former</option>
							<option value="current">Current</option>
						</select>
					</div>

					<!-- Drug Detail -->
					<div>
						<label for="drug-detail" class="block text-sm font-medium text-gray-700 mb-1">Drug Detail</label>
						<input
							id="drug-detail"
							type="text"
							bind:value={drugDetail}
							placeholder="e.g. occasional marijuana use"
							class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:ring-1 focus:ring-blue-500"
						/>
					</div>

					<!-- Exercise Frequency -->
					<div>
						<label for="exercise-frequency" class="block text-sm font-medium text-gray-700 mb-1">Exercise Frequency</label>
						<select
							id="exercise-frequency"
							bind:value={exerciseFrequency}
							class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:ring-1 focus:ring-blue-500"
						>
							<option value="">— Select —</option>
							<option value="never">Never</option>
							<option value="rarely">Rarely</option>
							<option value="occasional">Occasional</option>
							<option value="regular">Regular</option>
							<option value="daily">Daily</option>
						</select>
					</div>

					<!-- Occupation -->
					<div>
						<label for="occupation" class="block text-sm font-medium text-gray-700 mb-1">Occupation</label>
						<input
							id="occupation"
							type="text"
							bind:value={occupation}
							placeholder="e.g. Construction Worker"
							class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:ring-1 focus:ring-blue-500"
						/>
					</div>

					<!-- Living Situation -->
					<div>
						<label for="living-situation" class="block text-sm font-medium text-gray-700 mb-1">Living Situation</label>
						<select
							id="living-situation"
							bind:value={livingSituation}
							class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:ring-1 focus:ring-blue-500"
						>
							<option value="">— Select —</option>
							<option value="alone">Alone</option>
							<option value="with-family">With Family</option>
							<option value="assisted-living">Assisted Living</option>
							<option value="homeless">Homeless</option>
							<option value="other">Other</option>
						</select>
					</div>

					<!-- Recorded Date -->
					<div>
						<label for="recorded-date" class="block text-sm font-medium text-gray-700 mb-1">Recorded Date</label>
						<input
							id="recorded-date"
							type="date"
							bind:value={recordedDate}
							required
							class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:ring-1 focus:ring-blue-500"
						/>
					</div>
				</div>

				<!-- Checkboxes row -->
				<div class="flex flex-wrap gap-6 mb-4">
					<label class="flex items-center gap-2 text-sm text-gray-700">
						<input type="checkbox" bind:checked={foodInsecurity} class="rounded border-gray-300 text-blue-600 focus:ring-blue-500" />
						Food Insecurity
					</label>
					<label class="flex items-center gap-2 text-sm text-gray-700">
						<input type="checkbox" bind:checked={transportationAccess} class="rounded border-gray-300 text-blue-600 focus:ring-blue-500" />
						Transportation Access
					</label>
				</div>

				<!-- Notes -->
				<div class="mb-4">
					<label for="notes" class="block text-sm font-medium text-gray-700 mb-1">Notes</label>
					<textarea
						id="notes"
						bind:value={notes}
						rows="3"
						placeholder="Additional notes about social history…"
						class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:ring-1 focus:ring-blue-500"
					></textarea>
				</div>

				<div class="flex items-center gap-3">
					<button
						type="submit"
						disabled={formSubmitting}
						class="px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
					>
						{formSubmitting ? 'Saving…' : 'Save Social History'}
					</button>
					<button
						type="button"
						onclick={resetForm}
						class="px-4 py-2 border border-gray-300 text-gray-700 text-sm font-medium rounded-lg hover:bg-gray-50 transition-colors"
					>
						Reset
					</button>
				</div>
			</form>
		</div>

		<!-- Latest Entry Summary -->
		{#if latest}
			<div class="bg-blue-50 rounded-lg border border-blue-200 p-6 mb-6">
				<h2 class="text-lg font-semibold text-blue-900 mb-3">Current Social History</h2>
				<div class="grid grid-cols-1 md:grid-cols-2 gap-3 text-sm">
					<div>
						<span class="text-blue-700 font-medium">Smoking:</span>
						<span class="text-blue-900 ml-1">{formatLabel(latest.smoking_status)}</span>
						{#if latest.smoking_detail}
							<span class="text-blue-600 ml-1">({latest.smoking_detail})</span>
						{/if}
					</div>
					<div>
						<span class="text-blue-700 font-medium">Alcohol:</span>
						<span class="text-blue-900 ml-1">{formatLabel(latest.alcohol_use)}</span>
						{#if latest.alcohol_detail}
							<span class="text-blue-600 ml-1">({latest.alcohol_detail})</span>
						{/if}
					</div>
					<div>
						<span class="text-blue-700 font-medium">Drugs:</span>
						<span class="text-blue-900 ml-1">{formatLabel(latest.drug_use)}</span>
						{#if latest.drug_detail}
							<span class="text-blue-600 ml-1">({latest.drug_detail})</span>
						{/if}
					</div>
					<div>
						<span class="text-blue-700 font-medium">Exercise:</span>
						<span class="text-blue-900 ml-1">{formatLabel(latest.exercise_frequency)}</span>
					</div>
					<div>
						<span class="text-blue-700 font-medium">Occupation:</span>
						<span class="text-blue-900 ml-1">{latest.occupation || '—'}</span>
					</div>
					<div>
						<span class="text-blue-700 font-medium">Living:</span>
						<span class="text-blue-900 ml-1">{formatLabel(latest.living_situation)}</span>
					</div>
					<div>
						<span class="text-blue-700 font-medium">Food Insecurity:</span>
						<span class="text-blue-900 ml-1">{latest.food_insecurity ? 'Yes' : 'No'}</span>
					</div>
					<div>
						<span class="text-blue-700 font-medium">Transportation:</span>
						<span class="text-blue-900 ml-1">{latest.transportation_access ? 'Yes' : 'No'}</span>
					</div>
				</div>
				{#if latest.notes}
					<div class="mt-3 pt-3 border-t border-blue-200">
						<span class="text-blue-700 font-medium text-sm">Notes:</span>
						<p class="text-blue-900 text-sm mt-1">{latest.notes}</p>
					</div>
				{/if}
				<div class="mt-2 text-xs text-blue-500">
					Recorded: {formatDate(latest.recorded_date)}
				</div>
			</div>
		{/if}

		<!-- History Timeline -->
		<div class="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
			<h2 class="text-lg font-semibold text-gray-800 mb-4">History</h2>

			{#if history.length === 0}
				<p class="text-center text-gray-500 py-6 text-sm">No social history entries recorded yet.</p>
			{:else}
				<div class="space-y-4">
					{#each history as entry}
						<div class="border border-gray-200 rounded-lg p-4 hover:bg-gray-50 transition-colors">
							<div class="flex items-start justify-between mb-2">
								<span class="text-sm font-medium text-gray-700">{formatDate(entry.recorded_date)}</span>
								<button
									onclick={() => handleRemove(entry.id)}
									class="text-xs text-red-500 hover:text-red-700 font-medium transition-colors"
								>
									Remove
								</button>
							</div>
							<div class="grid grid-cols-1 md:grid-cols-2 gap-2 text-sm text-gray-600">
								<div>
									<span class="font-medium">Smoking:</span> {formatLabel(entry.smoking_status)}
									{#if entry.smoking_detail}
										<span class="text-gray-500">({entry.smoking_detail})</span>
									{/if}
								</div>
								<div>
									<span class="font-medium">Alcohol:</span> {formatLabel(entry.alcohol_use)}
									{#if entry.alcohol_detail}
										<span class="text-gray-500">({entry.alcohol_detail})</span>
									{/if}
								</div>
								<div>
									<span class="font-medium">Drugs:</span> {formatLabel(entry.drug_use)}
									{#if entry.drug_detail}
										<span class="text-gray-500">({entry.drug_detail})</span>
									{/if}
								</div>
								<div>
									<span class="font-medium">Exercise:</span> {formatLabel(entry.exercise_frequency)}
								</div>
								<div>
									<span class="font-medium">Occupation:</span> {entry.occupation || '—'}
								</div>
								<div>
									<span class="font-medium">Living:</span> {formatLabel(entry.living_situation)}
								</div>
								<div>
									<span class="font-medium">Food Insecurity:</span> {entry.food_insecurity ? 'Yes' : 'No'}
								</div>
								<div>
									<span class="font-medium">Transportation:</span> {entry.transportation_access ? 'Yes' : 'No'}
								</div>
							</div>
							{#if entry.notes}
								<div class="mt-2 pt-2 border-t border-gray-100">
									<p class="text-sm text-gray-500">{entry.notes}</p>
								</div>
							{/if}
						</div>
					{/each}
				</div>
			{/if}
		</div>
	{/if}
</div>
