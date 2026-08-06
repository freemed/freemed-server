<script lang="ts">
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { logout } from '$lib/stores/auth.svelte';

	const navItems = [
		{ href: '/', label: 'Home', icon: '🏠' },
		{ href: '/patients', label: 'Patients', icon: '👤' },
		{ href: '/scheduler', label: 'Scheduler', icon: '📅' },
		{ href: '/messages', label: 'Messages', icon: '✉️' },
		{ href: '/preferences', label: 'Preferences', icon: '⚙️' },
	];

	let mobileOpen = $state(false);

	function isActive(path: string): boolean {
		return $page.url.pathname.startsWith(path);
	}

	function navigate(path: string) {
		mobileOpen = false;
		goto(path);
	}

	async function handleLogout() {
		mobileOpen = false;
		logout();
		await goto('/login');
	}
</script>

<nav class="bg-white shadow-sm border-b border-gray-200">
	<div class="container mx-auto px-4">
		<div class="flex items-center justify-between h-14">
			<a href="/" class="text-lg font-bold text-blue-600 shrink-0">FreeMED</a>

			<!-- Desktop nav -->
			<div class="hidden md:flex items-center gap-1">
				{#each navItems as item}
					<button
						onclick={() => navigate(item.href)}
						class="px-3 py-2 text-sm rounded-md transition-colors {isActive(item.href) ? 'bg-blue-50 text-blue-700 font-medium' : 'text-gray-600 hover:bg-gray-100'}"
					>
						{item.icon} {item.label}
					</button>
				{/each}
			</div>

			<div class="hidden md:block">
				<button
					onclick={handleLogout}
					class="px-3 py-1.5 text-sm text-gray-500 hover:text-red-600 rounded-md hover:bg-red-50 transition-colors"
				>
					Logout
				</button>
			</div>

			<!-- Mobile hamburger -->
			<button
				onclick={() => mobileOpen = !mobileOpen}
				class="md:hidden p-2 text-gray-500 hover:text-gray-700 rounded-md"
				aria-label="Toggle menu"
			>
				<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					{#if mobileOpen}
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
					{:else}
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
					{/if}
				</svg>
			</button>
		</div>

		<!-- Mobile menu -->
		{#if mobileOpen}
			<div class="md:hidden pb-3 border-t border-gray-100">
				{#each navItems as item}
					<button
						onclick={() => navigate(item.href)}
						class="block w-full text-left px-4 py-2 text-sm rounded-md transition-colors {isActive(item.href) ? 'bg-blue-50 text-blue-700 font-medium' : 'text-gray-600 hover:bg-gray-100'}"
					>
						{item.icon} {item.label}
					</button>
				{/each}
				<button
					onclick={handleLogout}
					class="block w-full text-left px-4 py-2 text-sm text-red-600 hover:bg-red-50 rounded-md transition-colors mt-1"
				>
					🚪 Logout
				</button>
			</div>
		{/if}
	</div>
</nav>
