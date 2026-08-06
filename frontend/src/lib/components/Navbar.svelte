<script lang="ts">
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { logout } from '$lib/stores/auth';

	const navItems = [
		{ href: '/', label: 'Home', icon: '🏠' },
		{ href: '/patients', label: 'Patients', icon: '👤' },
		{ href: '/scheduler', label: 'Scheduler', icon: '📅' },
		{ href: '/messages', label: 'Messages', icon: '✉️' },
		{ href: '/preferences', label: 'Preferences', icon: '⚙️' },
	];

	function isActive(path: string): boolean {
		return $page.url.pathname.startsWith(path);
	}

	async function handleLogout() {
		logout();
		await goto('/login');
	}
</script>

<nav class="bg-white shadow-sm border-b border-gray-200">
	<div class="container mx-auto px-4">
		<div class="flex items-center justify-between h-14">
			<div class="flex items-center gap-1">
				<a href="/" class="text-lg font-bold text-blue-600 mr-4">FreeMED</a>
				{#each navItems as item}
					<button
						onclick={() => goto(item.href)}
						class="px-3 py-2 text-sm rounded-md transition-colors {isActive(item.href) ? 'bg-blue-50 text-blue-700 font-medium' : 'text-gray-600 hover:bg-gray-100'}"
					>
						{item.icon} {item.label}
					</button>
				{/each}
			</div>
			<button
				onclick={handleLogout}
				class="px-3 py-1.5 text-sm text-gray-500 hover:text-red-600 rounded-md hover:bg-red-50 transition-colors"
			>
				Logout
			</button>
		</div>
	</div>
</nav>
