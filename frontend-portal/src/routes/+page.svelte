<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { checkAuth } from '$lib/stores/portal-auth.svelte';

	onMount(async () => {
		const ok = await checkAuth();
		if (ok) {
			await goto('/dashboard');
		} else {
			await goto('/login');
		}
	});
</script>

<svelte:head>
	<title>FreeMED Patient Portal</title>
</svelte:head>

<div class="d-flex align-items-center justify-content-center min-vh-100">
	<div class="text-center">
		<div class="spinner-border text-primary mb-3" role="status">
			<span class="visually-hidden">Redirecting…</span>
		</div>
		<p class="text-muted small">Redirecting…</p>
	</div>
</div>
