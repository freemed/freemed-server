<script lang="ts">
	import { onMount } from 'svelte';
	import { portalAuth } from '$lib/stores/portal-auth.svelte';
	import { api } from '$lib/api';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
	import ErrorBanner from '$lib/components/ErrorBanner.svelte';

	interface PatientProfile {
		patient_id: string;
		first_name?: string;
		last_name?: string;
		date_of_birth?: string;
		gender?: string;
		address?: string;
		city?: string;
		state?: string;
		zip?: string;
		phone?: string;
		email?: string;
		primary_provider?: string;
	}

	let profile = $state<PatientProfile | null>(null);
	let loading = $state(true);
	let error = $state('');

	onMount(() => {
		loadProfile();
	});

	async function loadProfile() {
		error = '';
		loading = true;
		try {
			const data = await api.get<PatientProfile>('/me');
			profile = data;
		} catch (e: unknown) {
			error = e instanceof Error ? e.message : 'Failed to load profile';
		} finally {
			loading = false;
		}
	}
</script>

<svelte:head>
	<title>Profile — FreeMED Patient Portal</title>
</svelte:head>

<div>
	<h2 class="fw-bold text-dark mb-4">My Profile</h2>

	{#if loading}
		<LoadingSpinner message="Loading profile…" />
	{:else if error}
		<ErrorBanner message={error} onRetry={loadProfile} />
	{:else if profile}
		<div class="row g-4">
			<!-- Personal Information -->
			<div class="col-lg-6">
				<div class="card">
					<div class="card-header bg-white">
						<h5 class="card-title mb-0">Personal Information</h5>
					</div>
					<div class="card-body">
						<div class="mb-3">
							<div class="small text-muted">Name</div>
							<p class="mb-0 fw-medium">{[profile.first_name, profile.last_name].filter(Boolean).join(' ') || '—'}</p>
						</div>
						<div class="mb-3">
							<div class="small text-muted">Patient ID</div>
							<p class="mb-0">{profile.patient_id || '—'}</p>
						</div>
						<div class="mb-3">
							<div class="small text-muted">Date of Birth</div>
							<p class="mb-0">{profile.date_of_birth || '—'}</p>
						</div>
						<div class="mb-3">
							<div class="small text-muted">Gender</div>
							<p class="mb-0">{profile.gender || '—'}</p>
						</div>
					</div>
				</div>
			</div>

			<!-- Contact Information -->
			<div class="col-lg-6">
				<div class="card">
					<div class="card-header bg-white">
						<h5 class="card-title mb-0">Contact Information</h5>
					</div>
					<div class="card-body">
						<div class="mb-3">
							<div class="small text-muted">Phone</div>
							<p class="mb-0">{profile.phone || '—'}</p>
						</div>
						<div class="mb-3">
							<div class="small text-muted">Email</div>
							<p class="mb-0">{profile.email || '—'}</p>
						</div>
						<div class="mb-3">
							<div class="small text-muted">Address</div>
							<p class="mb-0">
								{[profile.address, profile.city, profile.state, profile.zip].filter(Boolean).join(', ') || '—'}
							</p>
						</div>
						<div class="mb-3">
							<div class="small text-muted">Primary Provider</div>
							<p class="mb-0">{profile.primary_provider || '—'}</p>
						</div>
					</div>
				</div>
			</div>
		</div>
	{/if}
</div>
