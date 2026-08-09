<script lang="ts">
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
	import ErrorBanner from '$lib/components/ErrorBanner.svelte';

	interface PatientInfo {
		id: number;
		first_name: string;
		last_name: string;
		patient_id: string;
	}

	interface UpcomingAppointment {
		id: number;
		date_of: string;
		date_of_mdy: string;
		appointment_time: string;
		provider_name: string;
		status: string;
	}

	interface RecentMessage {
		id: number;
		subject: string;
		message_time: string;
		sender: string;
		is_read: number;
	}

	interface BalanceInfo {
		total_outstanding: number;
	}

	let patient = $state<PatientInfo | null>(null);
	let appointments = $state<UpcomingAppointment[]>([]);
	let messages = $state<RecentMessage[]>([]);
	let balance = $state<BalanceInfo | null>(null);
	let loading = $state(true);
	let error = $state('');

	async function portalFetch<T>(path: string): Promise<T> {
		const res = await fetch(`/api/portal${path}`);
		if (res.status === 401) {
			await goto('/login');
			throw new Error('Session expired');
		}
		if (!res.ok) {
			const body = await res.text();
			throw new Error(`API error ${res.status}: ${body}`);
		}
		return res.json();
	}

	async function loadDashboard() {
		loading = true;
		error = '';
		try {
			const [pat, appts, msgs, bal] = await Promise.all([
				portalFetch<PatientInfo>('/me'),
				portalFetch<UpcomingAppointment[]>('/appointments'),
				portalFetch<RecentMessage[]>('/messages'),
				portalFetch<BalanceInfo>('/balance'),
			]);
			patient = pat;
			appointments = (appts || []).filter(a => a.status !== 'cancelled').slice(0, 5);
			messages = (msgs || []).slice(0, 5);
			balance = bal;
		} catch (e: any) {
			error = e.message || 'Failed to load dashboard';
		} finally {
			loading = false;
		}
	}

	$effect(() => {
		loadDashboard();
	});

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

	function formatCurrency(val: number): string {
		return val.toLocaleString('en-US', {
			style: 'currency',
			currency: 'USD',
			minimumFractionDigits: 2,
		});
	}

	function statusColor(status: string): string {
		switch (status?.toLowerCase()) {
			case 'scheduled':
			case 'confirmed':
				return 'bg-green-100 text-green-800';
			case 'requested':
				return 'bg-yellow-100 text-yellow-800';
			case 'cancelled':
				return 'bg-red-100 text-red-800';
			default:
				return 'bg-gray-100 text-gray-600';
		}
	}
</script>

{#if loading}
	<LoadingSpinner message="Loading portal…" />
{:else if error}
	<ErrorBanner message={error} onRetry={loadDashboard} />
{:else}
	<div class="max-w-7xl mx-auto">
		<!-- Welcome header -->
		<div class="mb-6">
			<h1 class="text-2xl font-bold text-gray-900">
				Welcome{patient ? `, ${patient.first_name}` : ''}
			</h1>
			<p class="text-sm text-gray-500 mt-1">
				Patient Portal &mdash; Manage your healthcare online
			</p>
		</div>

		<!-- Quick action cards -->
		<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
			<a href="/portal/schedule" class="bg-white rounded-lg shadow-sm border border-gray-200 p-4 hover:shadow-md transition-shadow">
				<div class="text-2xl mb-2">📅</div>
				<h3 class="font-semibold text-gray-900 text-sm">Schedule Appointment</h3>
				<p class="text-xs text-gray-500 mt-1">Book a new appointment online</p>
			</a>
			<a href="/portal/messages" class="bg-white rounded-lg shadow-sm border border-gray-200 p-4 hover:shadow-md transition-shadow">
				<div class="text-2xl mb-2">✉️</div>
				<h3 class="font-semibold text-gray-900 text-sm">Messages</h3>
				<p class="text-xs text-gray-500 mt-1">Send and receive secure messages</p>
			</a>
			<a href="/portal/billing" class="bg-white rounded-lg shadow-sm border border-gray-200 p-4 hover:shadow-md transition-shadow">
				<div class="text-2xl mb-2">💳</div>
				<h3 class="font-semibold text-gray-900 text-sm">Billing &amp; Statements</h3>
				<p class="text-xs text-gray-500 mt-1">View balance and payment history</p>
			</a>
			<div class="bg-white rounded-lg shadow-sm border border-gray-200 p-4 border-green-200 bg-green-50">
				<div class="text-2xl mb-2">💊</div>
				<h3 class="font-semibold text-gray-900 text-sm">Health Records</h3>
				<p class="text-xs text-gray-500 mt-1">Medications, allergies, vitals, labs</p>
			</div>
		</div>

		<div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
			<!-- Upcoming Appointments -->
			<div class="bg-white rounded-lg shadow-sm border border-gray-200">
				<div class="px-6 py-4 border-b border-gray-100 flex justify-between items-center">
					<h3 class="font-semibold text-gray-900">Upcoming Appointments</h3>
					<a href="/portal/schedule" class="text-xs text-blue-600 hover:text-blue-800">View All &rarr;</a>
				</div>
				<div class="divide-y divide-gray-100">
					{#if appointments.length === 0}
						<p class="p-6 text-sm text-gray-500">No upcoming appointments.</p>
					{:else}
						{#each appointments as appt (appt.id)}
							<div class="px-6 py-3 flex justify-between items-center">
								<div>
									<p class="text-sm font-medium text-gray-900">{appt.date_of_mdy} at {appt.appointment_time}</p>
									<p class="text-xs text-gray-500">{appt.provider_name || 'Provider'}</p>
								</div>
								<span class="inline-flex px-2 py-0.5 text-xs font-medium rounded-full {statusColor(appt.status)}">
									{appt.status}
								</span>
							</div>
						{/each}
					{/if}
				</div>
			</div>

			<!-- Recent Messages -->
			<div class="bg-white rounded-lg shadow-sm border border-gray-200">
				<div class="px-6 py-4 border-b border-gray-100 flex justify-between items-center">
					<h3 class="font-semibold text-gray-900">Recent Messages</h3>
					<a href="/portal/messages" class="text-xs text-blue-600 hover:text-blue-800">View All &rarr;</a>
				</div>
				<div class="divide-y divide-gray-100">
					{#if messages.length === 0}
						<p class="p-6 text-sm text-gray-500">No recent messages.</p>
					{:else}
						{#each messages as msg (msg.id)}
							<div class="px-6 py-3">
								<div class="flex justify-between items-start">
									<div>
										<p class="text-sm font-medium text-gray-900">{msg.subject}</p>
										<p class="text-xs text-gray-500">{msg.sender} &middot; {formatDate(msg.message_time)}</p>
									</div>
									{#if !msg.is_read}
										<span class="w-2 h-2 bg-blue-500 rounded-full mt-1.5 shrink-0"></span>
									{/if}
								</div>
							</div>
						{/each}
					{/if}
				</div>
			</div>
		</div>

		<!-- Balance summary -->
		{#if balance && balance.total_outstanding > 0}
			<div class="mt-6 bg-yellow-50 border border-yellow-200 rounded-lg p-4 flex items-center justify-between">
				<div>
					<p class="text-sm font-medium text-yellow-800">Outstanding Balance</p>
					<p class="text-xs text-yellow-600">Please review your billing statements</p>
				</div>
				<div class="text-right">
					<p class="text-lg font-bold text-yellow-800 font-mono">{formatCurrency(balance.total_outstanding)}</p>
					<a href="/portal/billing" class="text-xs text-yellow-700 hover:text-yellow-900 underline">View statements &rarr;</a>
				</div>
			</div>
		{/if}
	</div>
{/if}
