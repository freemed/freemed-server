<script lang="ts">
	import { page } from '$app/stores';
	import { api } from '$lib/api';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
	import ErrorBanner from '$lib/components/ErrorBanner.svelte';

	interface EligibilityBenefit {
		service_type_code: string;
		service_type_desc: string;
		coverage_level: string;
		plan_coverage: string;
		time_period: string;
		in_plan_network: string;
		amount: number;
		quantity: number;
	}

	interface EligibilityResult {
		status: string;
		payer_name: string;
		payer_id: string;
		trace_number: string;
		response_date: string;
		active_coverage: boolean;
		subscriber: {
			last_name: string;
			first_name: string;
			member_id: string;
			eligibility_date: string;
			active_coverage: boolean;
			plan_coverage: string;
		};
		benefits: EligibilityBenefit[];
		x12_270?: string;
	}

	let patientId = $state('');
	let loading = $state(false);
	let error = $state('');
	let result = $state<EligibilityResult | null>(null);
	let generated270 = $state('');
	let parseError = $state('');
	let uploadLoading = $state(false);

	$effect(() => {
		const id = $page.params.id;
		if (id) patientId = id;
	});

	async function checkEligibility() {
		loading = true;
		error = '';
		result = null;
		try {
			const data = await api.post<EligibilityResult>(`/eligibility/generate/${patientId}`, {});
			generated270 = data.x12_270 || '';
			result = data;
		} catch (e: any) {
			error = e.message || 'An error occurred';
		} finally {
			loading = false;
		}
	}

	async function upload271(e: Event) {
		const input = e.target as HTMLInputElement;
		const file = input.files?.[0];
		if (!file) return;

		uploadLoading = true;
		parseError = '';
		const formData = new FormData();
		formData.append('file', file);

		try {
			const res = await fetch('/api/eligibility/parse', {
				method: 'POST',
				body: formData,
				credentials: 'include'
			});
			if (!res.ok) {
				const body = await res.json();
				throw new Error(body.message || 'Failed to parse 271 response');
			}
			const parsed: EligibilityResult = await res.json();
			result = parsed;
		} catch (e: any) {
			parseError = e.message || 'Failed to parse file';
		} finally {
			uploadLoading = false;
		}
	}

	function coverageLevelLabel(code: string): string {
		const labels: Record<string, string> = {
			'1': 'Active Coverage',
			'A': 'Co-Insurance',
			'B': 'Co-Payment',
			'C': 'Deductible',
			'D': 'Co-Insurance (Patient)',
			'E': 'Patient Co-Payment',
			'F': 'Limitations',
			'G': 'Out of Pocket',
			'H': 'Stop Loss',
			'I': 'Non-Covered',
		};
		return labels[code] || `Code ${code}`;
	}

	function formatDate(dateStr: string): string {
		if (!dateStr) return '—';
		if (dateStr.length === 8) {
			const y = dateStr.slice(0, 4);
			const m = dateStr.slice(4, 6);
			const d = dateStr.slice(6, 8);
			return `${m}/${d}/${y}`;
		}
		return dateStr;
	}

	function formatCurrency(amount: number): string {
		if (!amount || amount === 0) return '—';
		return amount.toLocaleString('en-US', { style: 'currency', currency: 'USD', minimumFractionDigits: 2 });
	}
</script>

<div class="max-w-4xl mx-auto p-6">
	<h1 class="text-2xl font-bold text-gray-900 mb-6">Insurance Eligibility Verification</h1>

	<!-- Check Eligibility Button -->
	<div class="bg-white rounded-lg shadow-sm border border-gray-200 p-6 mb-6">
		<h2 class="text-lg font-semibold text-gray-800 mb-3">Eligibility Inquiry (X12 270)</h2>
		<p class="text-sm text-gray-500 mb-4">
			Generate an X12 270 eligibility inquiry for this patient's primary insurance.
			Submit the generated inquiry to the payer's clearinghouse and parse the 271 response.
		</p>

		<button
			on:click={checkEligibility}
			disabled={loading}
			class="px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded-md hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
		>
			{#if loading}
				Generating...
			{:else}
				Check Eligibility
			{/if}
		</button>

		{#if error}
			<ErrorBanner message={error} />
		{/if}
	</div>

	<!-- Generated 270 Display -->
	{#if generated270}
		<div class="bg-white rounded-lg shadow-sm border border-gray-200 p-6 mb-6">
			<h2 class="text-lg font-semibold text-gray-800 mb-3">Generated X12 270 Inquiry</h2>
			<pre class="bg-gray-50 border border-gray-200 rounded-md p-4 text-xs text-gray-700 overflow-x-auto max-h-64">{generated270}</pre>
			<p class="text-xs text-gray-400 mt-2">Copy this inquiry and submit to the payer's clearinghouse. Upload the 271 response below.</p>
		</div>
	{/if}

	<!-- Upload 271 Response -->
	<div class="bg-white rounded-lg shadow-sm border border-gray-200 p-6 mb-6">
		<h2 class="text-lg font-semibold text-gray-800 mb-3">Upload 271 Response</h2>
		<p class="text-sm text-gray-500 mb-3">
			Upload the payer's X12 271 eligibility response file for parsing.
		</p>
		<label class="inline-block px-4 py-2 bg-gray-100 text-sm font-medium text-gray-700 rounded-md border border-gray-300 cursor-pointer hover:bg-gray-200">
			<input
				type="file"
				accept=".txt,.edi,.271"
				on:change={upload271}
				class="hidden"
				disabled={uploadLoading}
			/>
			{#if uploadLoading}
				Parsing...
			{:else}
				Upload 271 File
			{/if}
		</label>
		{#if parseError}
			<ErrorBanner message={parseError} />
		{/if}
	</div>

	<!-- Results -->
	{#if loading}
		<LoadingSpinner message="Checking eligibility..." />
	{:else if result?.benefits && result.benefits.length > 0}
		<div class="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
			<h2 class="text-lg font-semibold text-gray-800 mb-4">
				Eligibility Results
				{#if result.payer_name}
					<span class="text-sm font-normal text-gray-500 ml-2">— {result.payer_name}</span>
				{/if}
			</h2>

			<!-- Subscriber Info -->
			{#if result.subscriber}
				<div class="grid grid-cols-2 gap-4 mb-6 p-4 bg-blue-50 rounded-md">
					<div>
						<span class="text-xs text-gray-500">Subscriber</span>
						<p class="text-sm font-medium text-gray-900">{result.subscriber.last_name}, {result.subscriber.first_name}</p>
					</div>
					<div>
						<span class="text-xs text-gray-500">Member ID</span>
						<p class="text-sm font-medium text-gray-900">{result.subscriber.member_id || '—'}</p>
					</div>
					<div>
						<span class="text-xs text-gray-500">Eligibility Date</span>
						<p class="text-sm font-medium text-gray-900">{formatDate(result.subscriber.eligibility_date)}</p>
					</div>
					<div>
						<span class="text-xs text-gray-500">Coverage Status</span>
						<p class="text-sm font-medium {result.active_coverage ? 'text-green-600' : 'text-red-600'}">
							{result.active_coverage ? 'Active' : 'Inactive'}
						</p>
					</div>
				</div>
			{/if}

			<!-- Benefits Table -->
			<div class="overflow-x-auto">
				<table class="min-w-full text-sm">
					<thead>
						<tr class="border-b border-gray-200">
							<th class="text-left py-2 px-3 text-xs font-medium text-gray-500 uppercase">Service</th>
							<th class="text-left py-2 px-3 text-xs font-medium text-gray-500 uppercase">Coverage</th>
							<th class="text-left py-2 px-3 text-xs font-medium text-gray-500 uppercase">Plan</th>
							<th class="text-right py-2 px-3 text-xs font-medium text-gray-500 uppercase">Amount</th>
							<th class="text-left py-2 px-3 text-xs font-medium text-gray-500 uppercase">Period</th>
							<th class="text-left py-2 px-3 text-xs font-medium text-gray-500 uppercase">Network</th>
						</tr>
					</thead>
					<tbody>
						{#each result.benefits as benefit}
							<tr class="border-b border-gray-100 hover:bg-gray-50">
								<td class="py-2 px-3 text-gray-900">{benefit.service_type_desc}</td>
								<td class="py-2 px-3">
									<span class="inline-block px-2 py-0.5 text-xs rounded-full {benefit.coverage_level === '1' ? 'bg-green-100 text-green-800' : 'bg-gray-100 text-gray-800'}">
										{coverageLevelLabel(benefit.coverage_level)}
									</span>
								</td>
								<td class="py-2 px-3 text-gray-700">{benefit.plan_coverage || '—'}</td>
								<td class="py-2 px-3 text-right font-mono text-gray-900">{formatCurrency(benefit.amount)}</td>
								<td class="py-2 px-3 text-gray-600">{formatDate(benefit.time_period) || '—'}</td>
								<td class="py-2 px-3 text-gray-600">{benefit.in_plan_network || '—'}</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		</div>
	{:else if result}
		<div class="bg-white rounded-lg shadow-sm border border-gray-200 p-6 text-center text-gray-500">
			No benefit details found in the response.
		</div>
	{/if}
</div>
