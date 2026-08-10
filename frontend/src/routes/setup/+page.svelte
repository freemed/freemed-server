<script lang="ts">
	import { goto } from '$app/navigation';
	import { z } from 'zod';

	// ---- Validation schemas ----
	const adminSchema = z.object({
		username: z.string().min(3, 'Username must be at least 3 characters'),
		password: z.string().min(8, 'Password must be at least 8 characters'),
		first_name: z.string().optional(),
		last_name: z.string().optional(),
		email: z.string().email('Invalid email').optional().or(z.literal('')),
		title: z.string().optional(),
	});

	type AdminForm = z.infer<typeof adminSchema>;

	// ---- State ----
	let step = $state(1);
	let loading = $state(false);
	let error = $state('');
	let success = $state(false);

	let admin: AdminForm = $state({
		username: '',
		password: '',
		first_name: '',
		last_name: '',
		email: '',
		title: '',
	});

	let fieldErrors = $state<Record<string, string>>({});
	let alreadySetup = $state(false);

	function validateStep1(): boolean {
		fieldErrors = {};
		const result = adminSchema.safeParse(admin);
		if (!result.success) {
			for (const issue of result.error.issues) {
				fieldErrors = { ...fieldErrors, [issue.path[0] as string]: issue.message };
			}
			return false;
		}
		return true;
	}

	function nextStep() {
		if (step === 1) {
			if (!validateStep1()) return;
		}
		step++;
	}

	function prevStep() {
		error = '';
		step--;
	}

	async function handleSubmit() {
		error = '';
		loading = true;

		try {
			const res = await fetch('/api/setup/initialize', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(admin),
			});

			if (!res.ok) {
				const body = await res.json();
				if (res.status === 409) {
					alreadySetup = true;
					error = body.message || 'System has already been initialized.';
				} else {
					error = body.message || 'Failed to initialize the system. Please try again.';
				}
				loading = false;
				return;
			}

			success = true;
			loading = false;

			// Redirect to login after a brief delay
			setTimeout(() => {
				goto('/login');
			}, 2500);
		} catch (e: unknown) {
			const msg = e instanceof Error ? e.message : 'Network error. Please try again.';
			error = msg;
			loading = false;
		}
	}
</script>

<svelte:head>
	<title>Setup — FreeMED EMR</title>
</svelte:head>

<div class="flex items-center justify-center min-h-screen bg-gradient-to-br from-blue-50 to-gray-100 px-4">
	<div class="w-full max-w-lg">
		<!-- Header -->
		<div class="text-center mb-8">
			<h1 class="text-3xl font-bold text-gray-900">FreeMED EMR</h1>
			<p class="text-gray-500 mt-2">Installation Wizard</p>
		</div>

		<!-- Steps indicator -->
		<div class="flex items-center justify-center gap-2 mb-8">
			<div
				class="w-8 h-8 rounded-full flex items-center justify-center text-sm font-medium
					{step === 1 ? 'bg-blue-600 text-white' : step > 1 ? 'bg-green-500 text-white' : 'bg-gray-200 text-gray-500'}"
			>
				{step > 1 ? '✓' : '1'}
			</div>
			<div class="w-12 h-px bg-gray-300"></div>
			<div
				class="w-8 h-8 rounded-full flex items-center justify-center text-sm font-medium
					{step === 2 ? 'bg-blue-600 text-white' : 'bg-gray-200 text-gray-500'}"
			>
				2
			</div>
		</div>

		<!-- Card -->
		<div class="bg-white rounded-xl shadow-lg p-8">
			{#if success}
				<div class="text-center py-6">
					<div class="w-16 h-16 bg-green-100 rounded-full flex items-center justify-center mx-auto mb-4">
						<svg class="w-8 h-8 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
						</svg>
					</div>
					<h2 class="text-xl font-semibold text-gray-900 mb-2">Setup Complete!</h2>
					<p class="text-gray-500">Your installation has been configured successfully. Redirecting to login…</p>
				</div>

			{:else if alreadySetup}
				<div class="text-center py-6">
					<div class="w-16 h-16 bg-yellow-100 rounded-full flex items-center justify-center mx-auto mb-4">
						<svg class="w-8 h-8 text-yellow-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L4.082 16.5c-.77.833.192 2.5 1.732 2.5z" />
						</svg>
					</div>
					<h2 class="text-xl font-semibold text-gray-900 mb-2">Already Configured</h2>
					<p class="text-gray-500">{error}</p>
					<button
						onclick={() => goto('/login')}
						class="mt-4 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors text-sm font-medium"
					>
						Go to Login
					</button>
				</div>

			{:else if step === 1}
				<h2 class="text-lg font-semibold text-gray-800 mb-6">Create Administrator Account</h2>

				{#if error}
					<div
						class="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg mb-6 text-sm"
						role="alert"
					>
						{error}
					</div>
				{/if}

				<form onsubmit={(e) => { e.preventDefault(); nextStep(); }} class="space-y-4">
					<div>
						<label for="username" class="block text-sm font-medium text-gray-700 mb-1">
							Username <span class="text-red-500">*</span>
						</label>
						<input
							id="username"
							type="text"
							bind:value={admin.username}
							autocomplete="username"
							class="w-full px-3 py-2 border rounded-lg shadow-sm text-sm
								{fieldErrors.username ? 'border-red-300 focus:ring-red-500 focus:border-red-500' : 'border-gray-300 focus:ring-blue-500 focus:border-blue-500'}
								focus:outline-none focus:ring-2 disabled:bg-gray-100 disabled:cursor-not-allowed"
							placeholder="e.g. admin"
						/>
						{#if fieldErrors.username}
							<p class="text-red-600 text-xs mt-1">{fieldErrors.username}</p>
						{/if}
					</div>

					<div>
						<label for="password" class="block text-sm font-medium text-gray-700 mb-1">
							Password <span class="text-red-500">*</span>
						</label>
						<input
							id="password"
							type="password"
							bind:value={admin.password}
							autocomplete="new-password"
							class="w-full px-3 py-2 border rounded-lg shadow-sm text-sm
								{fieldErrors.password ? 'border-red-300 focus:ring-red-500 focus:border-red-500' : 'border-gray-300 focus:ring-blue-500 focus:border-blue-500'}
								focus:outline-none focus:ring-2 disabled:bg-gray-100 disabled:cursor-not-allowed"
							placeholder="Minimum 8 characters"
						/>
						{#if fieldErrors.password}
							<p class="text-red-600 text-xs mt-1">{fieldErrors.password}</p>
						{/if}
					</div>

					<div class="grid grid-cols-2 gap-4">
						<div>
							<label for="first_name" class="block text-sm font-medium text-gray-700 mb-1">
								First Name
							</label>
							<input
								id="first_name"
								type="text"
								bind:value={admin.first_name}
								class="w-full px-3 py-2 border border-gray-300 rounded-lg shadow-sm text-sm
									focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
								placeholder="e.g. John"
							/>
						</div>
						<div>
							<label for="last_name" class="block text-sm font-medium text-gray-700 mb-1">
								Last Name
							</label>
							<input
								id="last_name"
								type="text"
								bind:value={admin.last_name}
								class="w-full px-3 py-2 border border-gray-300 rounded-lg shadow-sm text-sm
									focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
								placeholder="e.g. Smith"
							/>
						</div>
					</div>

					<div>
						<label for="email" class="block text-sm font-medium text-gray-700 mb-1">
							Email
						</label>
						<input
							id="email"
							type="email"
							bind:value={admin.email}
							class="w-full px-3 py-2 border rounded-lg shadow-sm text-sm
								{fieldErrors.email ? 'border-red-300 focus:ring-red-500 focus:border-red-500' : 'border-gray-300 focus:ring-blue-500 focus:border-blue-500'}
								focus:outline-none focus:ring-2"
							placeholder="admin@example.com"
						/>
						{#if fieldErrors.email}
							<p class="text-red-600 text-xs mt-1">{fieldErrors.email}</p>
						{/if}
					</div>

					<div>
						<label for="title" class="block text-sm font-medium text-gray-700 mb-1">
							Title
						</label>
						<input
							id="title"
							type="text"
							bind:value={admin.title}
							class="w-full px-3 py-2 border border-gray-300 rounded-lg shadow-sm text-sm
								focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
							placeholder="e.g. System Administrator"
						/>
					</div>

					<div class="flex justify-end pt-4">
						<button
							type="submit"
							class="px-6 py-2.5 bg-blue-600 text-white font-medium text-sm rounded-lg
								hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500
								transition-colors"
						>
							Continue
						</button>
					</div>
				</form>

			{:else if step === 2}
				<h2 class="text-lg font-semibold text-gray-800 mb-6">Confirm Setup</h2>

				{#if error}
					<div
						class="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg mb-6 text-sm"
						role="alert"
					>
						{error}
					</div>
				{/if}

				<div class="bg-gray-50 rounded-lg p-4 mb-6 space-y-3 text-sm">
					<div class="flex justify-between">
						<span class="text-gray-500">Username</span>
						<span class="text-gray-900 font-medium">{admin.username}</span>
					</div>
					<div class="flex justify-between">
						<span class="text-gray-500">Password</span>
						<span class="text-gray-900 font-medium">{'●'.repeat(admin.password.length)}</span>
					</div>
					{#if admin.first_name || admin.last_name}
						<div class="flex justify-between">
							<span class="text-gray-500">Name</span>
							<span class="text-gray-900 font-medium">{admin.first_name} {admin.last_name}</span>
						</div>
					{/if}
					{#if admin.email}
						<div class="flex justify-between">
							<span class="text-gray-500">Email</span>
							<span class="text-gray-900 font-medium">{admin.email}</span>
						</div>
					{/if}
					{#if admin.title}
						<div class="flex justify-between">
							<span class="text-gray-500">Title</span>
							<span class="text-gray-900 font-medium">{admin.title}</span>
						</div>
					{/if}
					<div class="flex justify-between">
						<span class="text-gray-500">Role</span>
						<span class="text-gray-900 font-medium">Administrator</span>
					</div>
				</div>

				<p class="text-sm text-gray-500 mb-6">
					This will create the initial administrator account. You will be able to add more users and configure the system after logging in.
				</p>

				<div class="flex justify-between">
					<button
						onclick={prevStep}
						disabled={loading}
						class="px-6 py-2.5 border border-gray-300 text-gray-700 font-medium text-sm rounded-lg
							hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500
							transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
					>
						Back
					</button>
					<button
						onclick={handleSubmit}
						disabled={loading}
						class="px-6 py-2.5 bg-blue-600 text-white font-medium text-sm rounded-lg
							hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500
							transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
					>
						{#if loading}
							<svg class="animate-spin h-4 w-4 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
								<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
								<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
							</svg>
							Initializing…
						{:else}
							Complete Setup
						{/if}
					</button>
				</div>
			{/if}
		</div>

		<!-- Footer -->
		<p class="text-center text-xs text-gray-400 mt-6">
			FreeMED EMR — Open Source Electronic Medical Records
		</p>
	</div>
</div>
