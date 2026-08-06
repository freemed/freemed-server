<script lang="ts">
	import { page } from '$app/stores';
	import { api } from '$lib/api';

	interface Vital {
		id: number;
		patient: number;
		date_taken: string;
		systolic: number | null;
		diastolic: number | null;
		heart_rate: number | null;
		respiratory_rate: number | null;
		temperature: string | null;
		oxygen_saturation: number | null;
		height_cm: string | null;
		weight_kg: string | null;
		bmi: string | null;
		notes: string | null;
		user: number;
		created_at: string;
	}

	interface VitalsForm {
		date_taken: string;
		systolic: string;
		diastolic: string;
		heart_rate: string;
		respiratory_rate: string;
		temperature: string;
		oxygen_saturation: string;
		height_cm: string;
		weight_kg: string;
		notes: string;
	}

	let vitals = $state<Vital[]>([]);
	let loading = $state(true);
	let error = $state('');
	let patientId = $state('');

	let showForm = $state(false);
	let formSaving = $state(false);
	let formError = $state('');
	let formSuccess = $state('');

	let form: VitalsForm = $state({
		date_taken: new Date().toISOString().slice(0, 16),
		systolic: '',
		diastolic: '',
		heart_rate: '',
		respiratory_rate: '',
		temperature: '',
		oxygen_saturation: '',
		height_cm: '',
		weight_kg: '',
		notes: ''
	});

	$effect(() => {
		const id = $page.params.id;
		if (id) {
			patientId = id;
			loadVitals(id);
		}
	});

	async function loadVitals(id: string) {
		loading = true;
		error = '';
		try {
			const data = await api.get<Vital[]>(`/patient/${id}/vitals`);
			vitals = data || [];
		} catch (e: any) {
			error = e.message || 'Failed to load vitals';
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

	function formatBP(systolic: number | null, diastolic: number | null): string {
		if (systolic == null && diastolic == null) return '—';
		const s = systolic != null ? String(systolic) : '?';
		const d = diastolic != null ? String(diastolic) : '?';
		return `${s}/${d}`;
	}

	function formatValue(val: number | null | string, unit: string = ''): string {
		if (val === null || val === undefined || val === '') return '—';
		return `${String(val)}${unit}`;
	}

	function resetForm() {
		form = {
			date_taken: new Date().toISOString().slice(0, 16),
			systolic: '',
			diastolic: '',
			heart_rate: '',
			respiratory_rate: '',
			temperature: '',
			oxygen_saturation: '',
			height_cm: '',
			weight_kg: '',
			notes: ''
		};
		formError = '';
		formSuccess = '';
	}

	async function submitVitals() {
		formSaving = true;
		formError = '';
		formSuccess = '';
		try {
			const payload: Record<string, any> = {};

			// Format date
			const dt = new Date(form.date_taken);
			payload.date_taken = dt.toISOString();

			// Numeric fields - send null if empty
			const intFields = ['systolic', 'diastolic', 'heart_rate', 'respiratory_rate', 'oxygen_saturation'];
			for (const f of intFields) {
				const v = (form as any)[f];
				payload[f] = v !== '' ? parseInt(v, 10) : null;
			}

			// String fields
			const strFields = ['temperature', 'height_cm', 'weight_kg', 'notes'];
			for (const f of strFields) {
				const v = (form as any)[f];
				payload[f] = v !== '' ? v : null;
			}

			await api.post(`/patient/${patientId}/vitals`, payload);
			formSuccess = 'Vitals recorded successfully';
			showForm = false;
			await loadVitals(patientId);
		} catch (e: any) {
			formError = e.message || 'Failed to save vitals';
		} finally {
			formSaving = false;
		}
	}

	function hasAbnormal(v: Vital): boolean {
		if (v.systolic != null && (v.systolic < 90 || v.systolic > 140)) return true;
		if (v.diastolic != null && (v.diastolic < 60 || v.diastolic > 90)) return true;
		if (v.heart_rate != null && (v.heart_rate < 60 || v.heart_rate > 100)) return true;
		if (v.respiratory_rate != null && (v.respiratory_rate < 12 || v.respiratory_rate > 20)) return true;
		if (v.oxygen_saturation != null && v.oxygen_saturation < 95) return true;
		const temp = parseFloat(v.temperature || '');
		if (!isNaN(temp) && (temp < 36.0 || temp > 37.8)) return true;
		return false;
	}
</script>

<div class="max-w-4xl mx-auto">
	<div class="flex items-center justify-between mb-6">
		<div>
			<h1 class="text-2xl font-bold text-gray-900">Vitals</h1>
			<p class="text-sm text-gray-500 mt-1">Patient #{patientId}</p>
		</div>
		<div class="flex items-center gap-3">
			<a
				href="/patients/{patientId}"
				class="text-sm text-blue-600 hover:text-blue-800 font-medium transition-colors"
			>
				&larr; Back to Patient
			</a>
			<button
				onclick={() => { resetForm(); showForm = !showForm; }}
				class="px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded-lg hover:bg-blue-700 transition-colors"
			>
				{#if showForm}Cancel{:else}+ Record Vitals{/if}
			</button>
		</div>
	</div>

	{#if showForm}
		<div class="bg-white rounded-lg shadow-sm border border-gray-200 p-6 mb-6">
			<h2 class="text-lg font-semibold text-gray-900 mb-4">Record New Vitals</h2>

			{#if formError}
				<div class="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg text-sm mb-4">
					{formError}
				</div>
			{/if}
			{#if formSuccess}
				<div class="bg-green-50 border border-green-200 text-green-700 px-4 py-3 rounded-lg text-sm mb-4">
					{formSuccess}
				</div>
			{/if}

			<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">Date / Time</label>
					<input
						type="datetime-local"
						bind:value={form.date_taken}
						class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
					/>
				</div>

				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">Systolic (mmHg)</label>
					<input
						type="number"
						bind:value={form.systolic}
						placeholder="120"
						class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
					/>
				</div>

				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">Diastolic (mmHg)</label>
					<input
						type="number"
						bind:value={form.diastolic}
						placeholder="80"
						class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
					/>
				</div>

				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">Heart Rate (bpm)</label>
					<input
						type="number"
						bind:value={form.heart_rate}
						placeholder="72"
						class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
					/>
				</div>

				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">Respiratory Rate (/min)</label>
					<input
						type="number"
						bind:value={form.respiratory_rate}
						placeholder="16"
						class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
					/>
				</div>

				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">Temperature (°C)</label>
					<input
						type="number"
						step="0.1"
						bind:value={form.temperature}
						placeholder="37.0"
						class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
					/>
				</div>

				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">O₂ Saturation (%)</label>
					<input
						type="number"
						bind:value={form.oxygen_saturation}
						placeholder="98"
						min="0"
						max="100"
						class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
					/>
				</div>

				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">Height (cm)</label>
					<input
						type="number"
						step="0.1"
						bind:value={form.height_cm}
						placeholder="170.0"
						class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
					/>
				</div>

				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">Weight (kg)</label>
					<input
						type="number"
						step="0.1"
						bind:value={form.weight_kg}
						placeholder="70.0"
						class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
					/>
				</div>

				<div class="md:col-span-2">
					<label class="block text-sm font-medium text-gray-700 mb-1">Notes</label>
					<textarea
						bind:value={form.notes}
						rows="3"
						placeholder="Optional notes..."
						class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
					></textarea>
				</div>
			</div>

			<div class="mt-4 flex justify-end">
				<button
					onclick={submitVitals}
					disabled={formSaving}
					class="px-6 py-2 bg-blue-600 text-white text-sm font-medium rounded-lg hover:bg-blue-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
				>
					{#if formSaving}
						Saving...
					{:else}
						Save Vitals
					{/if}
				</button>
			</div>
		</div>
	{/if}

	{#if loading}
		<div class="flex justify-center py-12">
			<div class="w-6 h-6 border-2 border-blue-500 border-t-transparent rounded-full animate-spin"></div>
		</div>
	{:else if error}
		<div class="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg text-sm mb-6">
			{error}
		</div>
	{:else if vitals.length === 0}
		<div class="text-center py-12 text-gray-500">
			<p class="text-lg">No vitals recorded</p>
			<p class="text-sm mt-1">No vitals have been recorded for this patient yet.</p>
		</div>
	{:else}
		<!-- Vitals Table -->
		<div class="bg-white rounded-lg shadow-sm border border-gray-200 overflow-hidden">
			<div class="overflow-x-auto">
				<table class="min-w-full divide-y divide-gray-200">
					<thead class="bg-gray-50">
						<tr>
							<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Date</th>
							<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">BP</th>
							<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">HR</th>
							<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">RR</th>
							<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Temp</th>
							<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">O₂</th>
							<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Ht</th>
							<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Wt</th>
							<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">BMI</th>
							<th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Notes</th>
						</tr>
					</thead>
					<tbody class="bg-white divide-y divide-gray-200">
						{#each vitals as v, i}
							<tr class="{hasAbnormal(v) ? 'bg-red-50' : ''} {i === 0 ? 'bg-blue-50' : ''}">
								<td class="px-4 py-3 text-sm text-gray-700 whitespace-nowrap">
									{formatDate(v.date_taken)}
									{#if i === 0}
										<span class="ml-2 inline-block px-1.5 py-0.5 text-xs font-medium bg-blue-100 text-blue-700 rounded">Latest</span>
									{/if}
								</td>
								<td class="px-4 py-3 text-sm font-mono whitespace-nowrap">
									<span class={hasAbnormal(v) && (v.systolic != null && (v.systolic < 90 || v.systolic > 140) || v.diastolic != null && (v.diastolic < 60 || v.diastolic > 90)) ? 'text-red-600 font-semibold' : 'text-gray-700'}>
										{formatBP(v.systolic, v.diastolic)}
									</span>
								</td>
								<td class="px-4 py-3 text-sm text-gray-700 whitespace-nowrap">
									<span class={v.heart_rate != null && (v.heart_rate < 60 || v.heart_rate > 100) ? 'text-red-600 font-semibold' : ''}>
										{formatValue(v.heart_rate, ' bpm')}
									</span>
								</td>
								<td class="px-4 py-3 text-sm text-gray-700 whitespace-nowrap">
									<span class={v.respiratory_rate != null && (v.respiratory_rate < 12 || v.respiratory_rate > 20) ? 'text-red-600 font-semibold' : ''}>
										{formatValue(v.respiratory_rate, ' /min')}
									</span>
								</td>
								<td class="px-4 py-3 text-sm text-gray-700 whitespace-nowrap">
									{formatValue(v.temperature, ' °C')}
								</td>
								<td class="px-4 py-3 text-sm text-gray-700 whitespace-nowrap">
									<span class={v.oxygen_saturation != null && v.oxygen_saturation < 95 ? 'text-red-600 font-semibold' : ''}>
										{formatValue(v.oxygen_saturation, '%')}
									</span>
								</td>
								<td class="px-4 py-3 text-sm text-gray-700 whitespace-nowrap">
									{formatValue(v.height_cm, ' cm')}
								</td>
								<td class="px-4 py-3 text-sm text-gray-700 whitespace-nowrap">
									{formatValue(v.weight_kg, ' kg')}
								</td>
								<td class="px-4 py-3 text-sm text-gray-700 whitespace-nowrap">
									{formatValue(v.bmi)}
								</td>
								<td class="px-4 py-3 text-sm text-gray-500 max-w-[200px] truncate">
									{#if v.notes}
										<span title={String(v.notes)}>{String(v.notes).slice(0, 50)}{String(v.notes).length > 50 ? '...' : ''}</span>
									{:else}
										—
									{/if}
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		</div>
	{/if}
</div>
