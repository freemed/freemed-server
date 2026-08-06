<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib/api';

	interface DashboardData {
		username: string;
		patientCount: number;
		todayAppointments: number;
		unreadMessages: number;
		activeAuthorizations: number;
	}

	let username = $state('');
	let patientCount = $state<number | null>(null);
	let todayAppointments = $state<number | null>(null);
	let unreadMessages = $state<number | null>(null);
	let activeAuthorizations = $state<number | null>(null);
	let loading = $state(true);
	let error = $state('');

	let welcomeMessage = $derived(
		username ? `Welcome, ${username}` : 'Welcome to FreeMED'
	);
	let patientCountDisplay = $derived(
		patientCount !== null ? patientCount.toLocaleString() : '—'
	);
	let todayAppointmentsDisplay = $derived(
		todayAppointments !== null ? todayAppointments.toLocaleString() : '—'
	);
	let unreadMessagesDisplay = $derived(
		unreadMessages !== null ? unreadMessages.toLocaleString() : '—'
	);
	let activeAuthorizationsDisplay = $derived(
		activeAuthorizations !== null ? activeAuthorizations.toLocaleString() : '—'
	);

	onMount(() => {
		loadDashboard();
	});

	async function loadDashboard() {
		try {
			const data = await api.get<DashboardData>('/dashboard/');
			username = data.username || '';
			patientCount = data.patientCount ?? null;
			todayAppointments = data.todayAppointments ?? null;
			unreadMessages = data.unreadMessages ?? null;
			activeAuthorizations = data.activeAuthorizations ?? null;
		} catch (e: unknown) {
			const msg = e instanceof Error ? e.message : 'Failed to load dashboard';
			error = msg;
		} finally {
			loading = false;
		}
	}
</script>

<div class="max-w-5xl mx-auto">
	<h1 class="text-2xl font-bold text-gray-800 mb-2">{welcomeMessage}</h1>
	<p class="text-gray-500 mb-8">Your practice at a glance</p>

	{#if loading}
		<div class="flex items-center justify-center py-16">
			<div
				class="w-8 h-8 border-4 border-blue-500 border-t-transparent rounded-full animate-spin"
			></div>
			<span class="ml-3 text-gray-500">Loading dashboard…</span>
		</div>
	{:else if error}
		<div class="bg-red-50 border border-red-200 text-red-700 rounded-lg p-4 mb-6">
			{error}
		</div>
	{:else}
		<!-- Stats Row -->
		<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
			<div class="bg-white rounded-xl shadow-sm border border-gray-200 p-5">
				<p class="text-sm text-gray-500 uppercase tracking-wide">Total Patients</p>
				<p class="text-3xl font-bold text-gray-900 mt-1">{patientCountDisplay}</p>
			</div>

			<div class="bg-white rounded-xl shadow-sm border border-gray-200 p-5">
				<p class="text-sm text-gray-500 uppercase tracking-wide">Today's Appointments</p>
				<p class="text-3xl font-bold text-gray-900 mt-1">{todayAppointmentsDisplay}</p>
			</div>

			<div class="bg-white rounded-xl shadow-sm border border-gray-200 p-5">
				<p class="text-sm text-gray-500 uppercase tracking-wide">Unread Messages</p>
				<p class="text-3xl font-bold text-gray-900 mt-1">{unreadMessagesDisplay}</p>
			</div>

			<div class="bg-white rounded-xl shadow-sm border border-gray-200 p-5">
				<p class="text-sm text-gray-500 uppercase tracking-wide">Active Authorizations</p>
				<p class="text-3xl font-bold text-gray-900 mt-1">{activeAuthorizationsDisplay}</p>
			</div>
		</div>

		<!-- Quick Actions -->
		<h2 class="text-lg font-semibold text-gray-700 mb-4">Quick Actions</h2>
		<div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
			<a
				href="/patients"
				class="block bg-white rounded-xl shadow-sm border border-gray-200 p-5 hover:shadow-md hover:border-blue-300 transition-all group"
			>
				<div class="flex items-center gap-3">
					<div
						class="w-10 h-10 rounded-lg bg-blue-100 text-blue-600 flex items-center justify-center text-xl"
					>
						🔍
					</div>
					<div>
						<p class="font-medium text-gray-900 group-hover:text-blue-600 transition-colors">
							Search Patients
						</p>
						<p class="text-sm text-gray-500">Find and view patient records</p>
					</div>
				</div>
			</a>

			<a
				href="/patients/new"
				class="block bg-white rounded-xl shadow-sm border border-gray-200 p-5 hover:shadow-md hover:border-green-300 transition-all group"
			>
				<div class="flex items-center gap-3">
					<div
						class="w-10 h-10 rounded-lg bg-green-100 text-green-600 flex items-center justify-center text-xl"
					>
						➕
					</div>
					<div>
						<p class="font-medium text-gray-900 group-hover:text-green-600 transition-colors">
							New Patient
						</p>
						<p class="text-sm text-gray-500">Add a new patient to the system</p>
					</div>
				</div>
			</a>

			<a
				href="/scheduler"
				class="block bg-white rounded-xl shadow-sm border border-gray-200 p-5 hover:shadow-md hover:border-purple-300 transition-all group"
			>
				<div class="flex items-center gap-3">
					<div
						class="w-10 h-10 rounded-lg bg-purple-100 text-purple-600 flex items-center justify-center text-xl"
					>
						📅
					</div>
					<div>
						<p class="font-medium text-gray-900 group-hover:text-purple-600 transition-colors">
							Scheduler
						</p>
						<p class="text-sm text-gray-500">Manage appointments and visits</p>
					</div>
				</div>
			</a>

			<a
				href="/messages"
				class="block bg-white rounded-xl shadow-sm border border-gray-200 p-5 hover:shadow-md hover:border-orange-300 transition-all group"
			>
				<div class="flex items-center gap-3">
					<div
						class="w-10 h-10 rounded-lg bg-orange-100 text-orange-600 flex items-center justify-center text-xl"
					>
						✉️
					</div>
					<div>
						<p class="font-medium text-gray-900 group-hover:text-orange-600 transition-colors">
							Messages
						</p>
						<p class="text-sm text-gray-500">View and send internal messages</p>
					</div>
				</div>
			</a>
		</div>
	{/if}
</div>
