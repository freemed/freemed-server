<script lang="ts">
	import '../app.css';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { portalAuth, checkAuth, logout } from '$lib/stores/portal-auth.svelte';
	import { onMount } from 'svelte';

	let { children } = $props();
	let authChecked = $state(false);
	let currentPath = $derived($page.url.pathname);

	const navItems = [
		{ href: '/dashboard', label: 'Dashboard', icon: '📊' },
		{ href: '/appointments', label: 'Appointments', icon: '📅' },
		{ href: '/health', label: 'Health', icon: '❤️' },
		{ href: '/labs', label: 'Labs', icon: '🔬' },
		{ href: '/documents', label: 'Documents', icon: '📄' },
		{ href: '/profile', label: 'Profile', icon: '👤' },
	];

	onMount(async () => {
		const ok = await checkAuth().catch(() => false);
		authChecked = true;
		if (!ok && currentPath !== '/login') {
			await goto('/login');
		}
	});

	function isActive(path: string): boolean {
		return $page.url.pathname.startsWith(path);
	}

	async function handleLogout() {
		logout();
		await goto('/login');
	}
</script>

{#if !authChecked && currentPath !== '/login'}
	<div class="d-flex align-items-center justify-content-center min-vh-100">
		<div class="text-center">
			<div class="spinner-border text-primary mb-3" role="status">
				<span class="visually-hidden">Loading…</span>
			</div>
			<p class="text-muted small">Loading…</p>
		</div>
	</div>
{:else}
	<div class="min-vh-100 bg-light">
		{#if portalAuth.authenticated}
			<nav class="navbar navbar-expand-lg navbar-light bg-white navbar-portal">
				<div class="container">
					<a class="navbar-brand fw-bold text-primary" href="/dashboard">FreeMED Portal</a>

					<button class="navbar-toggler" type="button" data-bs-toggle="collapse" data-bs-target="#portalNavbar"
						aria-controls="portalNavbar" aria-expanded="false" aria-label="Toggle navigation">
						<span class="navbar-toggler-icon"></span>
					</button>

					<div class="collapse navbar-collapse" id="portalNavbar">
						<ul class="navbar-nav me-auto">
							{#each navItems as item}
								<li class="nav-item">
									<a
										href={item.href}
										class="nav-link {isActive(item.href) ? 'active fw-semibold' : ''}"
									>
										{item.icon} {item.label}
									</a>
								</li>
							{/each}
						</ul>
						<div class="d-flex align-items-center gap-2">
							<span class="navbar-text small text-muted me-2">
								{portalAuth.displayName}
							</span>
							<button class="btn btn-outline-danger btn-sm" onclick={handleLogout}>
								Logout
							</button>
						</div>
					</div>
				</div>
			</nav>
		{/if}

		<main class="container py-4">
			{@render children()}
		</main>
	</div>
{/if}
