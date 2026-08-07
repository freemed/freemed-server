<script lang="ts">
	import { api } from '$lib/api';

	// --- Status ---
	interface RemittStatus {
		configured: boolean;
		url: string;
	}

	// --- Months ---
	interface MonthEntry {
		id: number;
		month: string;
		count: number;
	}

	// --- Patients to Bill ---
	interface PatientToBill {
		id: number;
		ptlname: string;
		ptfname: string;
	}

	// --- Procedures ---
	interface ProcedureToBill {
		id: number;
		procpatient: number;
		proccpt: number;
		proccharges: number;
		procunits: number;
		procdt: string;
		procstatus: string | null;
		cptnameext: string | null;
		procbalorig: number;
		procbalcurrent: number;
		procdiagset: string;
	}

	// --- Claim Info ---
	interface ClaimInfo {
		id: number;
		created_at: string;
		updated_at: string;
		billkeydate: string;
		billkey: string | null;
		bkprocs: string;
	}

	// --- Rebill ---
	interface RebillEntry {
		id: number;
		billkeydate: string;
		billkey: string | null;
		bkprocs: string;
	}

	// --- Reactive state ---
	let status = $state<RemittStatus | null>(null);
	let months = $state<MonthEntry[]>([]);
	let patientsToBill = $state<PatientToBill[]>([]);
	let claimInfo = $state<ClaimInfo[]>([]);
	let rebillList = $state<RebillEntry[]>([]);

	let loading = $state(true);
	let error = $state('');
	let statusLoading = $state(true);

	// Patient procedures expansion
	let expandedPatientId = $state<number | null>(null);
	let patientProcedures = $state<ProcedureToBill[]>([]);
	let proceduresLoading = $state(false);
	let proceduresError = $state('');

	// Mark billed checkboxes
	let selectedClaimIds = $state<Set<number>>(new Set());
	let markingBilled = $state(false);
	let markBilledResult = $state('');

	// Section visibility
	let showClaimInfo = $state(false);

	// --- Load all data ---
	$effect(() => {
		loadAll();
	});

	async function loadAll() {
		loading = true;
		error = '';
		try {
			const [s, m, p, c, r] = await Promise.all([
				api.get<RemittStatus>('/remitt/status'),
				api.get<MonthEntry[]>('/remitt/months'),
				api.get<PatientToBill[]>('/remitt/patients-to-bill'),
				api.get<ClaimInfo[]>('/remitt/claim-info'),
				api.get<RebillEntry[]>('/remitt/rebill-list'),
			]);
			status = s;
			months = m || [];
			patientsToBill = p || [];
			claimInfo = c || [];
			rebillList = r || [];
			statusLoading = false;
		} catch (e: any) {
			error = e.message || 'Failed to load remittance data';
			statusLoading = false;
		} finally {
			loading = false;
		}
	}

	async function loadProcedures(patientId: number) {
		proceduresLoading = true;
		proceduresError = '';
		try {
			patientProcedures = await api.get<ProcedureToBill[]>(
				`/remitt/patient/${patientId}/procedures-to-bill`
			);
		} catch (e: any) {
			proceduresError = e.message || 'Failed to load procedures';
			patientProcedures = [];
		} finally {
			proceduresLoading = false;
		}
	}

	function togglePatientProcedures(patientId: number) {
		if (expandedPatientId === patientId) {
			expandedPatientId = null;
			patientProcedures = [];
		} else {
			expandedPatientId = patientId;
			loadProcedures(patientId);
		}
	}

	function toggleClaim(id: number) {
		const next = new Set(selectedClaimIds);
		if (next.has(id)) {
			next.delete(id);
		} else {
			next.add(id);
		}
		selectedClaimIds = next;
	}

	function toggleAllClaims() {
		if (selectedClaimIds.size === claimInfo.length) {
			selectedClaimIds = new Set();
		} else {
			selectedClaimIds = new Set(claimInfo.map((c) => c.id));
		}
	}

	async function markBilled() {
		if (selectedClaimIds.size === 0) return;
		markingBilled = true;
		markBilledResult = '';
		try {
			const result = await api.post<{ status: string; count: number }>(
				'/remitt/mark-billed',
				{ ids: [...selectedClaimIds] }
			);
			markBilledResult = `Marked ${result.count} claim(s) as billed.`;
			selectedClaimIds = new Set();
			// Refresh claim info
			claimInfo = await api.get<ClaimInfo[]>('/remitt/claim-info');
		} catch (e: any) {
			markBilledResult = e.message || 'Failed to mark as billed';
		} finally {
			markingBilled = false;
		}
	}

	function formatDate(dateStr: string | null): string {
		if (!dateStr) return '—';
		try {
			const d = new Date(dateStr);
			if (isNaN(d.getTime())) return dateStr;
			return d.toLocaleDateString('en-US', {
				year: 'numeric',
				month: 'short',
				day: 'numeric',
			});
		} catch {
			return dateStr;
		}
	}

	function formatCurrency(val: number): string {
		return '$' + val.toFixed(2);
	}

	function monthLabel(ym: string): string {
		const [y, m] = ym.split('-');
		const months = [
			'Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun',
			'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec',
		];
		return `${months[parseInt(m, 10) - 1] || m} ${y}`;
	}
</script>

<div class="max-w-5xl mx-auto space-y-6">
	<h1 class="text-2xl font-bold text-gray-900">Remittance Billing</h1>

	<!-- ===== Status Banner ===== -->
	{#if !statusLoading}
		<div
			class="px-4 py-3 rounded-lg text-sm border {status?.configured
				? 'bg-green-50 border-green-200 text-green-800'
				: 'bg-yellow-50 border-yellow-200 text-yellow-800'}"
		>
			{#if status?.configured}
				<strong>Remitt configured:</strong> {status.url}
			{:else}
				<strong>Remitt not configured.</strong> Set the remitt_url configuration to enable transport.
			{/if}
		</div>
	{/if}

	{#if loading}
		<div class="flex justify-center py-12">
			<div class="w-6 h-6 border-2 border-blue-500 border-t-transparent rounded-full animate-spin"></div>
		</div>
	{:else if error}
		<div class="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg text-sm">
			{error}
		</div>
	{:else}
		<!-- ===== Process Buttons ===== -->
		<div class="bg-white rounded-lg shadow-sm border border-gray-200">
			<div class="px-6 py-4 border-b border-gray-100">
				<h2 class="text-lg font-semibold text-gray-900">Actions</h2>
			</div>
			<div class="p-6 flex flex-wrap gap-4">
				<button
					disabled
					class="px-4 py-2 text-sm font-medium rounded-md bg-gray-100 text-gray-400 cursor-not-allowed"
					title="Claim processing transport not yet implemented"
				>
					Process Claims
				</button>
				<button
					disabled
					class="px-4 py-2 text-sm font-medium rounded-md bg-gray-100 text-gray-400 cursor-not-allowed"
					title="Statement generation transport not yet implemented"
				>
					Process Statement
				</button>
				<button
					onclick={loadAll}
					class="px-4 py-2 text-sm font-medium text-blue-600 hover:bg-blue-50 rounded-md border border-blue-200 transition-colors"
				>
					Refresh All
				</button>
			</div>
		</div>

		<!-- ===== Billing Months ===== -->
		<div class="bg-white rounded-lg shadow-sm border border-gray-200">
			<div class="px-6 py-4 border-b border-gray-100">
				<h2 class="text-lg font-semibold text-gray-900">Billing Months</h2>
			</div>
			{#if months.length === 0}
				<p class="p-6 text-gray-500 text-sm">No billing months found.</p>
			{:else}
				<div class="p-6 flex flex-wrap gap-3">
					{#each months as m}
						<div class="inline-flex items-center gap-2 px-4 py-2 bg-gray-50 border border-gray-200 rounded-lg text-sm">
							<span class="font-medium text-gray-700">{monthLabel(m.month)}</span>
							<span
								class="inline-flex items-center justify-center w-6 h-6 text-xs font-bold rounded-full {m.count > 0
									? 'bg-blue-100 text-blue-700'
									: 'bg-gray-200 text-gray-500'}"
							>
								{m.count}
							</span>
						</div>
					{/each}
				</div>
			{/if}
		</div>

		<!-- ===== Patients to Bill ===== -->
		<div class="bg-white rounded-lg shadow-sm border border-gray-200">
			<div class="px-6 py-4 border-b border-gray-100">
				<h2 class="text-lg font-semibold text-gray-900">
					Patients to Bill
					{#if patientsToBill.length > 0}
						<span class="ml-2 text-sm font-normal text-gray-500">({patientsToBill.length})</span>
					{/if}
				</h2>
			</div>
			{#if patientsToBill.length === 0}
				<p class="p-6 text-gray-500 text-sm">No patients with unbilled procedures.</p>
			{:else}
			<div class="overflow-x-auto">
				<table class="w-full text-sm">
					<thead class="bg-gray-50 text-gray-500 uppercase text-xs">
						<tr>
							<th class="px-6 py-3 text-left">Patient</th>
							<th class="px-6 py-3 text-right w-24">Actions</th>
						</tr>
					</thead>
					<tbody>
						{#each patientsToBill as pt}
							<tr class="border-t border-gray-100 hover:bg-gray-50">
								<td class="px-6 py-3">
									<span class="font-medium text-gray-900">{pt.ptlname}, {pt.ptfname}</span>
									<span class="ml-2 text-xs text-gray-400">#{pt.id}</span>
								</td>
								<td class="px-6 py-3 text-right">
									<button
										onclick={() => togglePatientProcedures(pt.id)}
										class="text-xs text-blue-600 hover:text-blue-800 font-medium transition-colors"
									>
										{expandedPatientId === pt.id ? 'Collapse' : 'Procedures'}
									</button>
								</td>
							</tr>
							{#if expandedPatientId === pt.id}
								<tr>
									<td colspan="2" class="px-6 py-3 bg-gray-50">
										{#if proceduresLoading}
											<div class="flex justify-center py-4">
												<div class="w-5 h-5 border-2 border-blue-500 border-t-transparent rounded-full animate-spin"></div>
											</div>
										{:else if proceduresError}
											<p class="text-red-600 text-xs">{proceduresError}</p>
										{:else if patientProcedures.length === 0}
											<p class="text-gray-500 text-xs">No procedures found.</p>
										{:else}
										<div class="overflow-x-auto">
											<table class="w-full text-xs border border-gray-200 rounded bg-white">
												<thead class="bg-gray-100 text-gray-500 uppercase">
													<tr>
														<th class="px-3 py-2 text-left">Date</th>
														<th class="px-3 py-2 text-left">CPT</th>
														<th class="px-3 py-2 text-left">Description</th>
														<th class="px-3 py-2 text-right">Charges</th>
														<th class="px-3 py-2 text-right">Balance</th>
														<th class="px-3 py-2 text-left">Status</th>
													</tr>
												</thead>
												<tbody>
													{#each patientProcedures as proc}
														<tr class="border-t border-gray-100">
															<td class="px-3 py-2 font-mono whitespace-nowrap">{formatDate(proc.procdt)}</td>
															<td class="px-3 py-2 font-mono">{proc.proccpt || '—'}</td>
															<td class="px-3 py-2 text-gray-600">{proc.cptnameext || '—'}</td>
															<td class="px-3 py-2 text-right font-mono">{formatCurrency(proc.proccharges)}</td>
															<td class="px-3 py-2 text-right font-mono">{formatCurrency(proc.procbalcurrent)}</td>
															<td class="px-3 py-2">
																{#if proc.procstatus}
																	<span class="inline-block px-2 py-0.5 text-xs font-medium rounded bg-yellow-50 text-yellow-700">
																		{proc.procstatus}
																	</span>
																{:else}
																	—
																{/if}
															</td>
														</tr>
													{/each}
												</tbody>
											</table>
										</div>
										{/if}
									</td>
								</tr>
							{/if}
						{/each}
					</tbody>
				</table>
			</div>
			{/if}
		</div>

		<!-- ===== Claim Info (Mark as Billed) ===== -->
		<div class="bg-white rounded-lg shadow-sm border border-gray-200">
			<div class="px-6 py-4 border-b border-gray-100 flex justify-between items-center">
				<button
					onclick={() => (showClaimInfo = !showClaimInfo)}
					class="text-lg font-semibold text-gray-900 hover:text-blue-600 transition-colors flex items-center gap-2"
				>
					<span class="text-sm text-gray-400 transition-transform" class:rotate-90={showClaimInfo}>&#9654;</span>
					Claim Info
					{#if claimInfo.length > 0}
						<span class="text-sm font-normal text-gray-500">({claimInfo.length})</span>
					{/if}
				</button>
				{#if showClaimInfo && selectedClaimIds.size > 0}
					<button
						onclick={markBilled}
						disabled={markingBilled}
						class="px-4 py-1.5 text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 rounded-md disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
					>
						{markingBilled ? 'Marking...' : `Mark ${selectedClaimIds.size} as Billed`}
					</button>
				{/if}
			</div>
			{#if markBilledResult}
				<div class="px-6 py-2 bg-blue-50 border-b border-blue-100 text-blue-700 text-sm">
					{markBilledResult}
				</div>
			{/if}
			{#if showClaimInfo}
				{#if claimInfo.length === 0}
					<p class="p-6 text-gray-500 text-sm">No claim info entries found.</p>
				{:else}
			<div class="overflow-x-auto">
					<table class="w-full text-sm">
						<thead class="bg-gray-50 text-gray-500 uppercase text-xs">
							<tr>
								<th class="px-6 py-3 w-10">
									<input
										type="checkbox"
										checked={selectedClaimIds.size === claimInfo.length && claimInfo.length > 0}
										onchange={toggleAllClaims}
										class="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
									/>
								</th>
								<th class="px-6 py-3 text-left">Bill Key Date</th>
								<th class="px-6 py-3 text-left">Procedures</th>
								<th class="px-6 py-3 text-left">Updated</th>
							</tr>
						</thead>
						<tbody>
							{#each claimInfo as claim}
								<tr class="border-t border-gray-100 hover:bg-gray-50">
									<td class="px-6 py-3">
										<input
											type="checkbox"
											checked={selectedClaimIds.has(claim.id)}
											onchange={() => toggleClaim(claim.id)}
											class="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
										/>
									</td>
									<td class="px-6 py-3 font-mono text-xs whitespace-nowrap">
										{formatDate(claim.billkeydate)}
									</td>
									<td class="px-6 py-3 text-xs text-gray-600 max-w-xs truncate">
										{claim.bkprocs || '—'}
									</td>
									<td class="px-6 py-3 text-xs text-gray-400">
										{formatDate(claim.updated_at)}
									</td>
								</tr>
							{/each}
						</tbody>
					</table>
			</div>
				{/if}
			{/if}
		</div>

		<!-- ===== Rebill List ===== -->
		<div class="bg-white rounded-lg shadow-sm border border-gray-200">
			<div class="px-6 py-4 border-b border-gray-100">
				<h2 class="text-lg font-semibold text-gray-900">
					Rebill Candidates
					{#if rebillList.length > 0}
						<span class="ml-2 text-sm font-normal text-gray-500">({rebillList.length})</span>
					{/if}
				</h2>
			</div>
			{#if rebillList.length === 0}
				<p class="p-6 text-gray-500 text-sm">No rebill candidates.</p>
			{:else}
			<div class="overflow-x-auto">
				<table class="w-full text-sm">
					<thead class="bg-gray-50 text-gray-500 uppercase text-xs">
						<tr>
							<th class="px-6 py-3 text-left">Bill Key Date</th>
							<th class="px-6 py-3 text-left">Procedures</th>
						</tr>
					</thead>
					<tbody>
						{#each rebillList as entry}
							<tr class="border-t border-gray-100 hover:bg-gray-50">
								<td class="px-6 py-3 font-mono text-xs whitespace-nowrap">
									{formatDate(entry.billkeydate)}
								</td>
								<td class="px-6 py-3 text-xs text-gray-600 max-w-xs truncate">
									{entry.bkprocs || '—'}
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
