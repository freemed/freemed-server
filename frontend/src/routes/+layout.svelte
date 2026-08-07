<script lang="ts">
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { auth, checkAuth } from '$lib/stores/auth.svelte';
	import Navbar from '$lib/components/Navbar.svelte';
	import Toast from '$lib/components/Toast.svelte';
	import { onMount } from 'svelte';

	let { children } = $props();
	let authChecked = $state(false);
	let currentPath = $derived($page.url.pathname);

	onMount(async () => {
		const ok = await checkAuth().catch(() => false);
		authChecked = true;
		if (!ok && currentPath !== '/login') {
			await goto('/login');
		}
	});
</script>

{#if !authChecked && currentPath !== '/login'}
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
