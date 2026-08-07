<script lang="ts">
	import { goto } from '$app/navigation';
	import { login } from '$lib/stores/portal-auth.svelte';

	let patientId = $state('');
	let dateOfBirth = $state('');
	let pin = $state('');
	let serverError = $state('');
	let submitting = $state(false);

	async function handleSubmit(e: Event) {
		e.preventDefault();
		serverError = '';

		if (!patientId.trim() || !dateOfBirth.trim() || !pin.trim()) {
			serverError = 'All fields are required.';
			return;
		}

		submitting = true;
		try {
			const ok = await login(patientId.trim(), dateOfBirth.trim(), pin.trim());
			if (ok) {
				await goto('/dashboard');
			} else {
				serverError = 'Invalid patient ID, date of birth, or PIN. Please try again.';
			}
		} catch {
			serverError = 'An error occurred. Please try again.';
		} finally {
			submitting = false;
		}
	}
</script>

<svelte:head>
	<title>Login — FreeMED Patient Portal</title>
</svelte:head>

<div class="d-flex align-items-center justify-content-center" style="min-height: calc(100vh - 6rem);">
	<div class="col-12 col-sm-8 col-md-6 col-lg-4">
		<div class="card shadow">
			<div class="card-body p-4">
				<div class="text-center mb-4">
					<h3 class="fw-bold text-primary">FreeMED Portal</h3>
					<p class="text-muted small mb-0">Sign in to your patient portal</p>
				</div>

				{#if serverError}
					<div class="alert alert-danger py-2 small" role="alert">
						{serverError}
					</div>
				{/if}

				<form onsubmit={handleSubmit}>
					<div class="mb-3">
						<label for="patientId" class="form-label small fw-medium">Patient ID</label>
						<input
							id="patientId"
							type="text"
							bind:value={patientId}
							class="form-control"
							placeholder="Enter your patient ID"
							disabled={submitting}
							autocomplete="username"
						/>
					</div>

					<div class="mb-3">
						<label for="dateOfBirth" class="form-label small fw-medium">Date of Birth</label>
						<input
							id="dateOfBirth"
							type="date"
							bind:value={dateOfBirth}
							class="form-control"
							disabled={submitting}
						/>
					</div>

					<div class="mb-4">
						<label for="pin" class="form-label small fw-medium">PIN</label>
						<input
							id="pin"
							type="password"
							bind:value={pin}
							class="form-control"
							placeholder="Enter your PIN"
							disabled={submitting}
							autocomplete="current-password"
							maxlength="6"
						/>
					</div>

					<button type="submit" class="btn btn-primary w-100" disabled={submitting}>
						{#if submitting}
							<span class="spinner-border spinner-border-sm me-2" role="status" aria-hidden="true"></span>
							Signing in…
						{:else}
							Sign In
						{/if}
					</button>
				</form>
			</div>
		</div>

		<p class="text-center text-muted small mt-3">
			Need help? Contact your provider's office.
		</p>
	</div>
</div>
