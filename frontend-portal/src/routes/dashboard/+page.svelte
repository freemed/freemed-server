<script lang="ts">
	import { onMount } from 'svelte';
	import { portalAuth } from '$lib/stores/portal-auth.svelte';
	import { api } from '$lib/api';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
	import ErrorBanner from '$lib/components/ErrorBanner.svelte';

	interface AppointmentSummary {
		id: number;
		appointment_date: string;
		appointment_time: string;
		reason: string;
		status: string;
		provider_name?: string;
	}

	let upcomingAppts = $state<AppointmentSummary[]>([]);
	let loading = $state(true);
	let error = $state('');

	onMount(() => {
		loadDashboard();
	});

	async function loadDashboard() {
		error = '';
		loading = true;
		try {
			const data = await api.get<AppointmentSummary[]>('/appointments');
			upcomingAppts = (data || []).slice(0, 5);
		} catch (e: unknown) {
			error = e instanceof Error ? e.message : 'Failed to load dashboard';
		} finally {
			loading = false;
		}
	}
</script>

<svelte:head>
	<title>Dashboard — FreeMED Patient Portal</title>
</svelte:head>

<div>
	<div class="mb-4">
		<h2 class="fw-bold text-dark">Welcome, {portalAuth.displayName}</h2>
		<p class="text-muted">Your health at a glance</p>
	</div>

	{#if loading}
		<LoadingSpinner message="Loading dashboard…" />
	{:else if error}
		<ErrorBanner message={error} onRetry={loadDashboard} />
	{:else}
		<!-- Quick Links -->
		<div class="row g-3 mb-4">
			<div class="col-6 col-md-4 col-lg-2">
				<a href="/appointments" class="card card-hover text-decoration-none h-100">
					<div class="card-body text-center py-3">
						<div class="fs-3 mb-1">📅</div>
						<div class="small fw-medium text-dark">Appointments</div>
					</div>
				</a>
			</div>
			<div class="col-6 col-md-4 col-lg-2">
				<a href="/health" class="card card-hover text-decoration-none h-100">
					<div class="card-body text-center py-3">
						<div class="fs-3 mb-1">❤️</div>
						<div class="small fw-medium text-dark">Health</div>
					</div>
				</a>
			</div>
			<div class="col-6 col-md-4 col-lg-2">
				<a href="/labs" class="card card-hover text-decoration-none h-100">
					<div class="card-body text-center py-3">
						<div class="fs-3 mb-1">🔬</div>
						<div class="small fw-medium text-dark">Labs</div>
					</div>
				</a>
			</div>
			<div class="col-6 col-md-4 col-lg-2">
				<a href="/documents" class="card card-hover text-decoration-none h-100">
					<div class="card-body text-center py-3">
						<div class="fs-3 mb-1">📄</div>
						<div class="small fw-medium text-dark">Documents</div>
					</div>
				</a>
			</div>
			<div class="col-6 col-md-4 col-lg-2">
				<a href="/profile" class="card card-hover text-decoration-none h-100">
					<div class="card-body text-center py-3">
						<div class="fs-3 mb-1">👤</div>
						<div class="small fw-medium text-dark">Profile</div>
					</div>
				</a>
			</div>
		</div>

		<!-- Upcoming Appointments -->
		<div class="card">
			<div class="card-header bg-white">
				<h5 class="card-title mb-0">Upcoming Appointments</h5>
			</div>
			<div class="card-body p-0">
				{#if upcomingAppts.length === 0}
					<div class="p-4 text-center text-muted small">
						No upcoming appointments.
					</div>
				{:else}
					<div class="table-responsive">
						<table class="table table-hover mb-0">
							<thead class="table-light">
								<tr>
									<th class="small">Date</th>
									<th class="small">Time</th>
									<th class="small">Provider</th>
									<th class="small">Reason</th>
									<th class="small">Status</th>
								</tr>
							</thead>
							<tbody>
								{#each upcomingAppts as appt (appt.id)}
									<tr>
										<td class="small">{appt.appointment_date}</td>
										<td class="small">{appt.appointment_time}</td>
										<td class="small">{appt.provider_name || '—'}</td>
										<td class="small">{appt.reason}</td>
										<td class="small">
											<span class="badge {appt.status === 'scheduled' ? 'bg-success' : appt.status === 'cancelled' ? 'bg-danger' : 'bg-secondary'}">
												{appt.status}
											</span>
										</td>
									</tr>
								{/each}
							</tbody>
						</table>
					</div>
				{/if}
			</div>
		</div>
	{/if}
</div>
