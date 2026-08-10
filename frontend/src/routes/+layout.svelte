<script lang="ts">
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { auth, checkAuth } from '$lib/stores/auth.svelte';
	import Navbar from '$lib/components/Navbar.svelte';
	import Toast from '$lib/components/Toast.svelte';
	import { onMount } from 'svelte';

	let { children } = $props();
	let authChecked = $state(false);
	let setupNeeded = $state<boolean | null>(null);

	// pathname() avoids the typed-union narrowness of $page.url.pathname
	function pathname(): string {
		return $page.url.pathname;
	}

	onMount(async () => {
		// Step 1: Check if setup is needed (before any auth checks)
		try {
			const res = await fetch('/api/setup/status');
			if (res.ok) {
				const data = await res.json();
				setupNeeded = data.needs_setup || false;
			}
		} catch {
			setupNeeded = false; // can't reach API, assume not needed
		}

		// Step 2: If setup is needed and not already on /setup, redirect
		if (setupNeeded && pathname() !== '/setup') {
			await goto('/setup');
			return;
		}

		// Step 3: On /setup page, no auth needed
		if (pathname() === '/setup') {
			authChecked = true;
			return;
		}

		// Step 4: Normal auth check
		const ok = await checkAuth().catch(() => false);
		authChecked = true;
		if (!ok && pathname() !== '/login') {
			await goto('/login');
		}
	});
</script>

{#if setupNeeded === true && pathname() === '/setup'}
	<div class="min-h-screen bg-gray-50">
		{@render children()}
	</div>
{:else if !authChecked && pathname() !== '/login' && pathname() !== '/setup'}
	<div class="min-h-screen bg-gray-50 flex items-center justify-center">
		<div class="text-center">
			<svg class="animate-spin h-8 w-8 text-blue-600 mx-auto mb-4" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
				<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
				<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
			</svg>
			<p class="text-gray-500 text-sm">Loading…</p>
		</div>
	</div>
{:else}
	<div class="min-h-screen bg-gray-50">
		{#if auth.authenticated}
			<Navbar />
		{/if}

		<main class="container mx-auto px-4 py-6">
			{@render children()}
		</main>

		<Toast />
	</div>
{/if}
