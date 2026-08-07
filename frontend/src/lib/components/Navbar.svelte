<script lang="ts">
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { logout } from '$lib/stores/auth.svelte';
	import { api } from '$lib/api';

	interface SearchResult {
		id: number;
		title: string;
		result_type: string;
		label?: string;
	}

	const navItems = [
		{ href: '/', label: 'Home', icon: '🏠' },
		{ href: '/patients', label: 'Patients', icon: '👤' },
		{ href: '/scheduler', label: 'Scheduler', icon: '📅' },
		{ href: '/messages', label: 'Messages', icon: '✉️' },
		{ href: '/billing', label: 'Billing', icon: '💲' },
		{ href: '/admin', label: 'Admin', icon: '🔧' },
	];

	let mobileOpen = $state(false);
	let searchQuery = $state('');
	let searchResults = $state<SearchResult[]>([]);
	let searchOpen = $state(false);
	let searchLoading = $state(false);
	let debounceTimer: ReturnType<typeof setTimeout>;

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

	function onSearchInput(e: Event) {
		const target = e.target as HTMLInputElement;
		searchQuery = target.value;

		clearTimeout(debounceTimer);

		if (searchQuery.length < 2) {
			searchResults = [];
			searchOpen = false;
			return;
		}

		searchLoading = true;
		debounceTimer = setTimeout(async () => {
			try {
				const data = await api.get<SearchResult[]>('/search?q=' + encodeURIComponent(searchQuery));
				searchResults = data || [];
				searchOpen = searchResults.length > 0;
			} catch {
				searchResults = [];
				searchOpen = false;
			} finally {
				searchLoading = false;
			}
		}, 300);
	}

	function onSearchKeydown(e: KeyboardEvent) {
		if (e.key === 'Enter' && searchQuery.length >= 2) {
			clearTimeout(debounceTimer);
			searchLoading = true;
			(async () => {
				try {
					const data = await api.get<SearchResult[]>('/search?q=' + encodeURIComponent(searchQuery));
					searchResults = data || [];
					searchOpen = searchResults.length > 0;
				} catch {
					searchResults = [];
					searchOpen = false;
				} finally {
					searchLoading = false;
				}
			})();
		}
	}

	function navigateResult(result: SearchResult) {
		searchOpen = false;
		searchQuery = '';
		searchResults = [];
		switch (result.result_type) {
			case 'patient':
				goto('/patients/' + result.id);
				break;
			case 'message':
				goto('/messages');
				break;
			case 'appointment':
				goto('/scheduler');
				break;
		}
	}

	function closeSearch() {
		searchOpen = false;
	}

	function handleSearchFocus() {
		if (searchResults.length > 0) {
			searchOpen = true;
		}
	}
</script>

<svelte:window onclick={closeSearch} />

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

			<!-- Global search bar -->
			<div class="hidden md:flex items-center relative">
				<div class="relative">
					<span class="absolute inset-y-0 left-0 flex items-center pl-2 pointer-events-none text-gray-400">
						<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
						</svg>
					</span>
					<input
						type="text"
						placeholder="Search..."
						value={searchQuery}
						oninput={onSearchInput}
						onkeydown={onSearchKeydown}
						onfocus={handleSearchFocus}
						class="w-48 pl-8 pr-3 py-1.5 text-sm border border-gray-300 rounded-md focus:outline-none focus:ring-1 focus:ring-blue-500 focus:border-blue-500 transition-colors"
					/>
					{#if searchLoading}
						<span class="absolute inset-y-0 right-0 flex items-center pr-2 text-gray-400">
							<svg class="animate-spin w-4 h-4" fill="none" viewBox="0 0 24 24">
								<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
								<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
							</svg>
						</span>
					{/if}
				</div>

				<!-- Search results dropdown -->
				{#if searchOpen}
					<div role="presentation" class="absolute top-full mt-1 right-0 w-72 bg-white border border-gray-200 rounded-md shadow-lg z-50 max-h-80 overflow-y-auto" onclick={(e) => e.stopPropagation()} onkeydown={(e) => e.key === 'Escape' && closeSearch()}>
						{#each searchResults as result (result.result_type + '-' + result.id)}
							<button
								onclick={() => navigateResult(result)}
								class="w-full text-left px-3 py-2 text-sm hover:bg-gray-50 border-b border-gray-100 last:border-b-0"
							>
								<div class="flex items-center justify-between">
									<span class="text-gray-800 truncate">{result.title}</span>
									<span class="ml-2 shrink-0 text-xs font-medium px-1.5 py-0.5 rounded {result.result_type === 'patient' ? 'bg-blue-100 text-blue-700' : result.result_type === 'message' ? 'bg-green-100 text-green-700' : 'bg-purple-100 text-purple-700'}">
										{result.result_type}
									</span>
								</div>
								{#if result.label}
									<div class="text-xs text-gray-500 mt-0.5">ID: {result.label}</div>
								{/if}
							</button>
						{:else}
							<div class="px-3 py-2 text-sm text-gray-500">No results found</div>
						{/each}
					</div>
				{/if}
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
				<!-- Mobile search -->
				<div class="px-4 py-2">
					<div class="relative">
						<span class="absolute inset-y-0 left-0 flex items-center pl-2 pointer-events-none text-gray-400">
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
							</svg>
						</span>
						<input
							type="text"
							placeholder="Search..."
							value={searchQuery}
							oninput={onSearchInput}
							onkeydown={onSearchKeydown}
							class="w-full pl-8 pr-3 py-1.5 text-sm border border-gray-300 rounded-md focus:outline-none focus:ring-1 focus:ring-blue-500 focus:border-blue-500"
						/>
						{#if searchLoading}
							<span class="absolute inset-y-0 right-0 flex items-center pr-2 text-gray-400">
								<svg class="animate-spin w-4 h-4" fill="none" viewBox="0 0 24 24">
									<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
									<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
								</svg>
							</span>
						{/if}
					</div>
					<!-- Mobile search results -->
					{#if searchOpen}
						<div class="mt-1 border border-gray-200 rounded-md max-h-60 overflow-y-auto">
							{#each searchResults as result (result.result_type + '-' + result.id)}
								<button
									onclick={() => { mobileOpen = false; navigateResult(result); }}
									class="w-full text-left px-3 py-2 text-sm hover:bg-gray-50 border-b border-gray-100 last:border-b-0"
								>
									<div class="flex items-center justify-between">
										<span class="text-gray-800 truncate">{result.title}</span>
										<span class="ml-2 shrink-0 text-xs font-medium px-1.5 py-0.5 rounded {result.result_type === 'patient' ? 'bg-blue-100 text-blue-700' : result.result_type === 'message' ? 'bg-green-100 text-green-700' : 'bg-purple-100 text-purple-700'}">
											{result.result_type}
										</span>
									</div>
									{#if result.label}
										<div class="text-xs text-gray-500 mt-0.5">ID: {result.label}</div>
									{/if}
								</button>
							{:else}
								<div class="px-3 py-2 text-sm text-gray-500">No results found</div>
							{/each}
						</div>
					{/if}
				</div>

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
