<script lang="ts">
	import { page } from '$app/stores';
	import { authToken, logout } from '$lib/stores/auth.svelte';
	import Navbar from '$lib/components/Navbar.svelte';
	import Toast from '$lib/components/Toast.svelte';

	let { children } = $props();
	let authenticated = $derived(!!authToken.current);
</script>

<div class="min-h-screen bg-gray-50">
	{#if authenticated}
		<Navbar />
	{/if}

	<main class="container mx-auto px-4 py-6">
		{@render children()}
	</main>

	<Toast />
</div>
