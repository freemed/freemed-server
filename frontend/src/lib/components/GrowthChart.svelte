<script lang="ts">
	interface VitalsPoint {
		date: string;
		height_cm: number | null;
		weight_kg: number | null;
		bmi: number | null;
	}

	let { patientId }: { patientId: string } = $props();
	let vitals = $state<VitalsPoint[]>([]);
	let loading = $state(true);
	let error = $state('');
	let chartType = $state<'weight' | 'height' | 'bmi'>('weight');

	$effect(() => {
		loadVitals();
	});

	async function loadVitals() {
		loading = true;
		error = '';
		try {
			const { api } = await import('$lib/api');
			const data = await api.get<VitalsPoint[]>(`/patient/${patientId}/vitals`);
			vitals = data || [];
		} catch (e: any) {
			error = e.message || 'Failed to load vitals.';
		} finally {
			loading = false;
		}
	}

	function getChartData(): { label: string; value: number }[] {
		return vitals
			.filter(v => {
				const val = chartType === 'weight' ? v.weight_kg :
					chartType === 'height' ? v.height_cm : v.bmi;
				return val != null && val > 0;
			})
			.map(v => ({
				label: v.date.slice(0, 10),
				value: chartType === 'weight' ? (v.weight_kg || 0) :
					chartType === 'height' ? (v.height_cm || 0) : (v.bmi || 0)
			}));
	}

	function formatLabel(label: string): string {
		const d = new Date(label);
		return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
	}

	function chartMax(): number {
		const data = getChartData();
		if (data.length === 0) return 100;
		const max = Math.max(...data.map(d => d.value));
		return Math.ceil(max * 1.2);
	}
</script>

<div class="bg-white rounded-lg shadow-sm border border-gray-200 p-4 mb-6">
	<h2 class="text-lg font-semibold text-gray-900 mb-3">Growth Charts</h2>

	{#if loading}
		<div class="flex justify-center py-8">
			<div class="w-6 h-6 border-2 border-blue-500 border-t-transparent rounded-full animate-spin"></div>
		</div>
	{:else if error}
		<p class="text-red-600 text-sm">{error}</p>
	{:else if vitals.length === 0}
		<p class="text-gray-500 text-sm py-8 text-center">No vitals data recorded yet.</p>
	{:else}
		<div class="flex gap-2 mb-4">
			{#each ['weight', 'height', 'bmi'] as type}
				<button
					onclick={() => (chartType = type as typeof chartType)}
					class="px-3 py-1 text-xs font-medium rounded-full transition-colors {chartType === type ? 'bg-blue-600 text-white' : 'bg-gray-100 text-gray-600 hover:bg-gray-200'}"
				>
					{type === 'weight' ? 'Weight' : type === 'height' ? 'Height' : 'BMI'}
				</button>
			{/each}
		</div>

		{@const data = getChartData()}
		{@const maxVal = chartMax()}
		{@const height = 160}
		{@const width = data.length > 1 ? (data.length - 1) * 60 + 20 : 200}

		<div style="overflow-x: auto;">
			<svg viewBox="0 0 {Math.max(width, 300)} {height + 30}" class="w-full" style="min-width: 300px;">
				<!-- Y axis labels -->
				<text x="5" y="15" class="text-xs fill-gray-400">{maxVal}</text>
				<text x="5" y="{height + 10}" class="text-xs fill-gray-400">0</text>

				<!-- Bars -->
				{#each data as point, i}
					{@const barW = Math.min(40, Math.max(8, (Math.max(width, 300) - 40) / data.length - 8))}
					{@const x = 30 + i * (barW + 8)}
					{@const barH = maxVal > 0 ? (point.value / maxVal) * height : 0}
					{@const y = height - barH}
					<rect {x} {y} width={barW} height={barH} rx="3" class="fill-blue-500" />
					<text x={x + barW / 2} y={height + 16} text-anchor="middle" class="text-xs fill-gray-400">{formatLabel(point.label)}</text>
					<text x={x + barW / 2} y={y - 4} text-anchor="middle" class="text-xs fill-gray-600">{point.value}</text>
				{/each}
			</svg>
		</div>
	{/if}
</div>
