<script lang="ts">
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
	import ErrorBanner from '$lib/components/ErrorBanner.svelte';
	import EmptyState from '$lib/components/EmptyState.svelte';

	interface OutstandingCharge {
		id: number;
		procedure: string;
		charge_date: string;
		charge_amount: number;
		paid_amount: number;
		balance: number;
		insurance_paid: number;
	}

	interface Credit {
		id: number;
		amount: number;
		date: string;
		source: string;
	}

	interface Statement {
		id: number;
		created_at: string;
		date_from: string;
		date_to: string;
		provider: number;
		status: string;
		total_charges: number;
		date_created: string;
	}

	interface BalanceResponse {
		total_outstanding: number;
		outstanding_charges: OutstandingCharge[];
		credits: Credit[];
	}

	let balance = $state<BalanceResponse | null>(null);
	let statements = $state<Statement[]>([]);
	let loading = $state(true);
	let error = $state('');

	let creditsTotal = $derived(balance ? (balance.credits || []).reduce((s: number, c: Credit) => s + c.amount, 0) : 0);
	let netBalance = $derived(balance ? balance.total_outstanding - creditsTotal : 0);

	async function portalFetch<T>(path: string): Promise<T> {
		const res = await fetch(`/api/portal${path}`);
		if (!res.ok) {
			const body = await res.text();
			throw new Error(`API error ${res.status}: ${body}`);
		}
		return res.json();
	}

	async function loadBilling() {
		loading = true;
		error = '';
		try {
			const [bal, stmts] = await Promise.all([
				portalFetch<BalanceResponse>('/balance'),
				portalFetch<Statement[]>('/statements'),
			]);
			balance = bal;
			statements = stmts || [];
		} catch (e: any) {
			error = e.message || 'Failed to load billing information';
		} finally {
			loading = false;
		}
	}

	$effect(() => {
		loadBilling();
	});

	function formatCurrency(val: number): string {
		return val.toLocaleString('en-US', {
			style: 'currency',
			currency: 'USD',
			minimumFractionDigits: 2,
		});
	}

	function formatDate(dateStr: string): string {
		if (!dateStr) return '—';
		const d = new Date(dateStr);
		if (isNaN(d.getTime())) return dateStr;
		return d.toLocaleDateString('en-US', {
			year: 'numeric',
			month: 'short',
			day: 'numeric',
		});
	}

	function statusColor(status: string): string {
		switch (status?.toLowerCase()) {
			case 'paid': return 'bg-green-100 text-green-800';
			case 'pending': return 'bg-yellow-100 text-yellow-800';
			case 'submitted': return 'bg-blue-100 text-blue-800';
			default: return 'bg-gray-100 text-gray-600';
		}
	}
</script>

<div class="max-w-7xl mx-auto">
	<div class="mb-6">
		<h1 class="text-2xl font-bold text-gray-900">Billing &amp; Statements</h1>
		<p class="text-sm text-gray-500 mt-1">View your balance, outstanding charges, and statements</p>
	</div>

	{#if loading}
		<LoadingSpinner message="Loading billing info…" />
	{:else if error}
		<ErrorBanner message={error} onRetry={loadBilling} />
	{:else}
		<!-- Balance overview -->
		<div class="grid grid-cols-1 md:grid-cols-3 gap-4 mb-8">
			<div class="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
				<p class="text-xs text-gray-500 uppercase font-medium mb-2">Outstanding Balance</p>
				<p class="text-3xl font-bold font-mono {balance && balance.total_outstanding > 0 ? 'text-red-600' : 'text-green-600'}">
					{balance ? formatCurrency(balance.total_outstanding) : '$0.00'}
				</p>
			</div>
			<div class="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
				<p class="text-xs text-gray-500 uppercase font-medium mb-2">Unapplied Credits</p>
				<p class="text-3xl font-bold font-mono text-blue-600">
					{balance ? formatCurrency((balance.credits || []).reduce((s: number, c: Credit) => s + c.amount, 0)) : '$0.00'}
				</p>
			</div>
			<div class="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
				<p class="text-xs text-gray-500 uppercase font-medium mb-2">Net Balance</p>
				<p class="text-3xl font-bold font-mono {netBalance > 0 ? 'text-red-600' : 'text-green-600'}">
					{netBalance > 0 ? formatCurrency(netBalance) : formatCurrency(0)}
				</p>
			</div>
		</div>

		<!-- Outstanding charges -->
		<div class="bg-white rounded-lg shadow-sm border border-gray-200 mb-8">
			<div class="px-6 py-4 border-b border-gray-100">
				<h3 class="font-semibold text-gray-900">Outstanding Charges</h3>
			</div>
			{#if !balance || (balance.outstanding_charges || []).length === 0}
				<div class="p-6">
					<EmptyState title="No outstanding charges" message="You have no outstanding balances." />
				</div>
			{:else}
				<div class="overflow-x-auto">
					<table class="w-full text-sm">
						<thead class="bg-gray-50 text-gray-500 uppercase text-xs">
							<tr>
								<th class="px-6 py-3 text-left">Date</th>
								<th class="px-6 py-3 text-left">Description</th>
								<th class="px-6 py-3 text-right">Charge</th>
								<th class="px-6 py-3 text-right">Paid</th>
								<th class="px-6 py-3 text-right">Balance</th>
							</tr>
						</thead>
						<tbody>
							{#each balance.outstanding_charges as charge (charge.id)}
								<tr class="border-t border-gray-100 hover:bg-gray-50">
									<td class="px-6 py-3 text-gray-500 whitespace-nowrap">{formatDate(charge.charge_date)}</td>
									<td class="px-6 py-3 text-gray-900">{charge.procedure || '—'}</td>
									<td class="px-6 py-3 text-right font-mono">{formatCurrency(charge.charge_amount)}</td>
									<td class="px-6 py-3 text-right font-mono text-green-600">{formatCurrency(charge.paid_amount)}</td>
									<td class="px-6 py-3 text-right font-mono font-medium {charge.balance > 0 ? 'text-red-600' : ''}">{formatCurrency(charge.balance)}</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			{/if}
		</div>

		<!-- Statements -->
		<div class="bg-white rounded-lg shadow-sm border border-gray-200">
			<div class="px-6 py-4 border-b border-gray-100">
				<h3 class="font-semibold text-gray-900">Recent Statements</h3>
			</div>
			{#if statements.length === 0}
				<div class="p-6">
					<EmptyState title="No statements" message="Statements will appear here once generated." />
				</div>
			{:else}
				<div class="overflow-x-auto">
					<table class="w-full text-sm">
						<thead class="bg-gray-50 text-gray-500 uppercase text-xs">
							<tr>
								<th class="px-6 py-3 text-left">Period</th>
								<th class="px-6 py-3 text-left">Created</th>
								<th class="px-6 py-3 text-right">Total Charges</th>
								<th class="px-6 py-3 text-left">Status</th>
							</tr>
						</thead>
						<tbody>
							{#each statements as stmt (stmt.id)}
								<tr class="border-t border-gray-100 hover:bg-gray-50">
									<td class="px-6 py-3 text-gray-900 whitespace-nowrap">
										{formatDate(stmt.date_from)} &ndash; {formatDate(stmt.date_to)}
									</td>
									<td class="px-6 py-3 text-gray-500 whitespace-nowrap">{formatDate(stmt.date_created)}</td>
									<td class="px-6 py-3 text-right font-mono">{formatCurrency(stmt.total_charges)}</td>
									<td class="px-6 py-3">
										<span class="inline-flex px-2 py-0.5 text-xs font-medium rounded-full {statusColor(stmt.status)}">
											{stmt.status || 'Unknown'}
										</span>
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
