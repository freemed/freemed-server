<script lang="ts">
	import { api } from '$lib/api';

	interface DashboardData { patientCount: number; todayAppointments: number; unreadMessages: number; activeAuthorizations: number; username: string; }

	let data = $state<DashboardData | null>(null);
	let loading = $state(true);

	$effect(() => {
		api.get<DashboardData>('/dashboard').then(d => { data = d; loading = false; }).catch(() => { loading = false; });
	});
</script>

<div class="grid grid-cols-1 md:grid-cols-3 gap-6">
	<a href="/billing/claims" class="bg-white rounded-lg shadow-sm border border-gray-200 p-6 hover:shadow-md transition-shadow">
		<h3 class="text-lg font-semibold text-gray-900 mb-1">Claims Manager</h3>
		<p class="text-sm text-gray-500">Submit, track, and rebill insurance claims</p>
	</a>
	<a href="/billing/ar" class="bg-white rounded-lg shadow-sm border border-gray-200 p-6 hover:shadow-md transition-shadow">
		<h3 class="text-lg font-semibold text-gray-900 mb-1">Accounts Receivable</h3>
		<p class="text-sm text-gray-500">Aging report and collections tracking</p>
	</a>
	<a href="/billing/remitt" class="bg-white rounded-lg shadow-sm border border-gray-200 p-6 hover:shadow-md transition-shadow">
		<h3 class="text-lg font-semibold text-gray-900 mb-1">Remittance Billing</h3>
		<p class="text-sm text-gray-500">Process remittances and manage billing transport</p>
	</a>
	<a href="/billing/superbills" class="bg-white rounded-lg shadow-sm border border-gray-200 p-6 hover:shadow-md transition-shadow">
		<h3 class="text-lg font-semibold text-gray-900 mb-1">Superbills</h3>
		<p class="text-sm text-gray-500">Generate and process superbills for encounters</p>
	</a>
</div>
