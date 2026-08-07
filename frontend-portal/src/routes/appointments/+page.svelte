<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib/api';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
	import ErrorBanner from '$lib/components/ErrorBanner.svelte';
	import EmptyState from '$lib/components/EmptyState.svelte';

	interface Appointment {
		id: number;
		appointment_date: string;
		appointment_time: string;
		reason: string;
		status: string;
		provider_name?: string;
	}

	let appointments = $state<Appointment[]>([]);
	let loading = $state(true);
	let error = $state('');

	// Request form state
	let showRequestForm = $state(false);
	let reqDate = $state('');
	let reqHour = $state('9');
	let reqMinute = $state('0');
	let reqProviderId = $state('');
	let reqReason = $state('');
	let reqSubmitting = $state(false);
	let reqError = $state('');
	let reqSuccess = $state(false);

	onMount(() => {
		loadAppointments();
	});

	async function loadAppointments() {
		error = '';
		loading = true;
		try {
			const data = await api.get<Appointment[]>('/appointments');
			appointments = data || [];
		} catch (e: unknown) {
			error = e instanceof Error ? e.message : 'Failed to load appointments';
		} finally {
			loading = false;
		}
	}

	async function requestAppointment() {
		reqError = '';
		reqSuccess = false;

		if (!reqDate || !reqReason.trim()) {
			reqError = 'Date and reason are required.';
			return;
		}

		reqSubmitting = true;
		try {
			await api.post('/appointments/request', {
				date: reqDate,
				hour: parseInt(reqHour),
				minute: parseInt(reqMinute),
				provider_id: reqProviderId || undefined,
				reason: reqReason.trim(),
			});
			reqSuccess = true;
			showRequestForm = false;
			reqDate = '';
			reqHour = '9';
			reqMinute = '0';
			reqProviderId = '';
			reqReason = '';
			await loadAppointments();
		} catch (e: unknown) {
			reqError = e instanceof Error ? e.message : 'Failed to request appointment';
		} finally {
			reqSubmitting = false;
		}
	}
</script>

<svelte:head>
	<title>Appointments — FreeMED Patient Portal</title>
</svelte:head>

<div>
	<div class="d-flex justify-content-between align-items-center mb-4">
		<h2 class="fw-bold text-dark mb-0">Appointments</h2>
		<button class="btn btn-primary" onclick={() => showRequestForm = !showRequestForm}>
			{#if showRequestForm}
				Cancel
			{:else}
				+ Request Appointment
			{/if}
		</button>
	</div>

	{#if reqSuccess}
		<div class="alert alert-success alert-dismissible fade show" role="alert">
			Appointment request submitted successfully!
			<button type="button" class="btn-close" onclick={() => reqSuccess = false} aria-label="Close"></button>
		</div>
	{/if}

	{#if showRequestForm}
		<div class="card mb-4">
			<div class="card-header bg-white">
				<h5 class="card-title mb-0">Request Appointment</h5>
			</div>
			<div class="card-body">
				{#if reqError}
					<div class="alert alert-danger py-2 small" role="alert">{reqError}</div>
				{/if}
				<div class="row g-3">
					<div class="col-md-4">
						<label for="reqDate" class="form-label small fw-medium">Date</label>
						<input id="reqDate" type="date" class="form-control" bind:value={reqDate} disabled={reqSubmitting} />
					</div>
					<div class="col-md-2">
						<label for="reqHour" class="form-label small fw-medium">Hour</label>
						<select id="reqHour" class="form-select" bind:value={reqHour} disabled={reqSubmitting}>
							{#each Array(12) as _, i}
								{@const h = i + 8}
								<option value={h}>{h}:00</option>
							{/each}
						</select>
					</div>
					<div class="col-md-2">
						<label for="reqMinute" class="form-label small fw-medium">Minute</label>
						<select id="reqMinute" class="form-select" bind:value={reqMinute} disabled={reqSubmitting}>
							<option value="0">:00</option>
							<option value="15">:15</option>
							<option value="30">:30</option>
							<option value="45">:45</option>
						</select>
					</div>
					<div class="col-md-4">
						<label for="reqProviderId" class="form-label small fw-medium">Provider ID (optional)</label>
						<input id="reqProviderId" type="text" class="form-control" bind:value={reqProviderId} placeholder="Leave blank for any provider" disabled={reqSubmitting} />
					</div>
					<div class="col-12">
						<label for="reqReason" class="form-label small fw-medium">Reason for Visit</label>
						<input id="reqReason" type="text" class="form-control" bind:value={reqReason} placeholder="e.g., Annual checkup, Follow-up" disabled={reqSubmitting} />
					</div>
					<div class="col-12">
						<button class="btn btn-primary" onclick={requestAppointment} disabled={reqSubmitting}>
							{#if reqSubmitting}
								<span class="spinner-border spinner-border-sm me-2" role="status" aria-hidden="true"></span>
								Submitting…
							{:else}
								Submit Request
							{/if}
						</button>
					</div>
				</div>
			</div>
		</div>
	{/if}

	{#if loading}
		<LoadingSpinner message="Loading appointments…" />
	{:else if error}
		<ErrorBanner message={error} onRetry={loadAppointments} />
	{:else if appointments.length === 0}
		<EmptyState title="No appointments" message="You have no appointments on file. Request one above!" />
	{:else}
		<div class="card">
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
						{#each appointments as appt (appt.id)}
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
		</div>
	{/if}
</div>
