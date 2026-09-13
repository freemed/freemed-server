<script lang="ts">
	import { api } from '$lib/api';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
	import ErrorBanner from '$lib/components/ErrorBanner.svelte';
	import EmptyState from '$lib/components/EmptyState.svelte';
	import { z } from 'zod';

	interface FhirClient {
		id: number;
		client_id: string;
		client_name: string;
		redirect_uris: string;
		grant_types: string;
		scopes: string;
		is_confidential: boolean;
		active: boolean;
		created_at: string;
	}

	// The 201 from POST /smart/clients. For a confidential client the server
	// returns the generated client_secret exactly once and only ever stores its
	// bcrypt hash, so this response is the only chance to show it.
	interface CreatedClient {
		id: number;
		client_id: string;
		client_secret?: string;
		client_secret_note?: string;
	}

	let clients = $state<FhirClient[]>([]);
	let loading = $state(true);
	let error = $state('');

	let showModal = $state(false);
	let formSaving = $state(false);
	let formError = $state('');
	let deleting = $state<Record<number, boolean>>({});
	let createdClient = $state<CreatedClient | null>(null);
	let copied = $state('');

	const formSchema = z.object({
		client_name: z.string().min(1, 'Client name is required'),
		redirect_uris: z.string().min(1, 'Redirect URIs are required'),
		scopes: z.string().optional(),
		is_confidential: z.boolean(),
	});

	let form = $state({
		client_name: '',
		redirect_uris: '',
		scopes: '',
		is_confidential: false,
	});
	let fieldErrors = $state<Record<string, string>>({});

	function resetForm() {
		form = { client_name: '', redirect_uris: '', scopes: '', is_confidential: false };
		fieldErrors = {};
		formError = '';
	}

	function openCreate() {
		resetForm();
		showModal = true;
	}

	function closeModal() {
		showModal = false;
		resetForm();
	}

	async function loadClients() {
		loading = true;
		error = '';
		try {
			clients = await api.get<FhirClient[]>('/smart/clients');
		} catch (e: any) {
			error = e.message || 'Failed to load SMART clients';
		} finally {
			loading = false;
		}
	}

	$effect(() => {
		loadClients();
	});

	async function saveClient(e: SubmitEvent) {
		e.preventDefault();
		formError = '';
		fieldErrors = {};

		const result = formSchema.safeParse(form);
		if (!result.success) {
			for (const issue of result.error.issues) {
				fieldErrors = { ...fieldErrors, [issue.path[0] as string]: issue.message };
			}
			return;
		}

		formSaving = true;
		try {
			const created = await api.post<CreatedClient>('/smart/clients', result.data);
			closeModal();
			// Hold the 201 body: for a confidential client it is the one and
			// only place the plaintext client_secret is ever available.
			createdClient = created;
			await loadClients();
		} catch (e: any) {
			formError = e.message || 'Failed to register client';
		} finally {
			formSaving = false;
		}
	}

	function closeCreated() {
		createdClient = null;
		copied = '';
	}

	async function copyValue(value: string, what: string) {
		try {
			await navigator.clipboard.writeText(value);
			copied = what;
			setTimeout(() => {
				if (copied === what) copied = '';
			}, 2000);
		} catch {
			copied = '';
		}
	}

	async function deleteClient(client: FhirClient) {
		if (!confirm(`Delete "${client.client_name}"? This cannot be undone.`)) return;

		deleting = { ...deleting, [client.id]: true };
		try {
			await api.del(`/smart/clients/${client.id}`);
			clients = clients.filter((c) => c.id !== client.id);
		} catch (e: any) {
			error = e.message || 'Failed to delete client';
		} finally {
			deleting = { ...deleting, [client.id]: false };
		}
	}

	function formatDate(dateStr: string): string {
		if (!dateStr) return '—';
		const d = new Date(dateStr);
		if (isNaN(d.getTime())) return dateStr;
		return d.toLocaleDateString('en-US', {
			year: 'numeric',
			month: 'short',
			day: 'numeric',
		});
	}
</script>

<div class="bg-white rounded-lg shadow-sm border border-gray-200">
	<div class="px-6 py-4 border-b border-gray-100 flex justify-between items-center">
		<div>
			<h2 class="text-lg font-semibold text-gray-900">SMART on FHIR App Registration</h2>
			<p class="text-xs text-gray-500 mt-0.5">Manage registered FHIR client applications</p>
		</div>
		<div class="flex gap-2">
			<button onclick={loadClients} class="px-3 py-1.5 text-sm text-gray-600 hover:bg-gray-100 rounded-md transition-colors">Refresh</button>
			<button onclick={openCreate} class="px-4 py-1.5 text-sm font-medium text-white bg-blue-600 rounded-md hover:bg-blue-700 transition-colors">
				+ Register New App
			</button>
		</div>
	</div>

	{#if loading}
		<LoadingSpinner message="Loading registered apps…" />
	{:else if error}
		<ErrorBanner message={error} onRetry={loadClients} />
	{:else if clients.length === 0}
		<EmptyState
			title="No registered apps"
			message="Register a FHIR client app to enable SMART on FHIR integration."
			actionLabel="Register New App"
			onAction={openCreate}
		/>
	{:else}
		<div class="overflow-x-auto">
			<table class="w-full text-sm">
				<thead class="bg-gray-50 text-gray-500 uppercase text-xs">
					<tr>
						<th class="px-6 py-3 text-left">Client Name</th>
						<th class="px-6 py-3 text-left">Client ID</th>
						<th class="px-6 py-3 text-left">Redirect URIs</th>
						<th class="px-6 py-3 text-left">Scopes</th>
						<th class="px-6 py-3 text-left">Type</th>
						<th class="px-6 py-3 text-left">Status</th>
						<th class="px-6 py-3 text-left">Created</th>
						<th class="px-6 py-3 text-right">Actions</th>
					</tr>
				</thead>
				<tbody>
					{#each clients as client (client.id)}
						<tr class="border-t border-gray-100 hover:bg-gray-50">
							<td class="px-6 py-3 font-medium text-gray-900">{client.client_name}</td>
							<td class="px-6 py-3 font-mono text-xs text-gray-500 max-w-[180px] truncate" title={client.client_id}>
								{client.client_id}
							</td>
							<td class="px-6 py-3 text-gray-600 max-w-[200px] truncate" title={client.redirect_uris}>
								{client.redirect_uris}
							</td>
							<td class="px-6 py-3 text-gray-600 max-w-[150px] truncate" title={client.scopes}>
								{client.scopes}
							</td>
							<td class="px-6 py-3">
								<span class="inline-flex px-2 py-0.5 text-xs font-medium rounded-full {client.is_confidential ? 'bg-purple-100 text-purple-800' : 'bg-gray-100 text-gray-600'}">
									{client.is_confidential ? 'Confidential' : 'Public'}
								</span>
							</td>
							<td class="px-6 py-3">
								<span class="inline-flex px-2 py-0.5 text-xs font-medium rounded-full {client.active ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'}">
									{client.active ? 'Active' : 'Inactive'}
								</span>
							</td>
							<td class="px-6 py-3 text-gray-500 whitespace-nowrap">{formatDate(client.created_at)}</td>
							<td class="px-6 py-3 text-right">
								<button
									onclick={() => deleteClient(client)}
									disabled={deleting[client.id]}
									class="px-3 py-1 text-xs font-medium text-red-600 hover:bg-red-50 rounded-md transition-colors disabled:opacity-50"
								>
									{deleting[client.id] ? 'Deleting…' : 'Delete'}
								</button>
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{/if}
</div>

<!-- Register Modal -->
{#if showModal}
	<div class="fixed inset-0 z-50 flex items-center justify-center bg-black/40" role="dialog" aria-modal="true">
		<div class="bg-white rounded-lg shadow-xl w-full max-w-lg mx-4" onclick={(e: MouseEvent) => e.stopPropagation()} onkeydown={() => {}}>
			<div class="px-6 py-4 border-b border-gray-100 flex justify-between items-center">
				<h3 class="text-lg font-semibold text-gray-900">Register New FHIR App</h3>
				<button
					onclick={closeModal}
					class="p-1 text-gray-400 hover:text-gray-600 rounded-md hover:bg-gray-100"
					aria-label="Close"
				>
					<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
					</svg>
				</button>
			</div>

			<form onsubmit={saveClient} class="p-6 space-y-4">
				{#if formError}
					<div class="bg-red-50 border border-red-200 rounded-md p-3 text-sm text-red-700">{formError}</div>
				{/if}

				<div>
					<label for="client_name" class="block text-sm font-medium text-gray-700 mb-1">Client Name *</label>
					<input
						id="client_name"
						type="text"
						bind:value={form.client_name}
						class="w-full px-3 py-2 text-sm border rounded-md transition-colors focus:outline-none focus:ring-1 focus:ring-blue-500 {fieldErrors.client_name ? 'border-red-300' : 'border-gray-300'}"
						placeholder="e.g., My Health App"
					/>
					{#if fieldErrors.client_name}
						<p class="text-red-600 text-xs mt-1">{fieldErrors.client_name}</p>
					{/if}
				</div>

				<div>
					<label for="redirect_uris" class="block text-sm font-medium text-gray-700 mb-1">Redirect URIs *</label>
					<input
						id="redirect_uris"
						type="text"
						bind:value={form.redirect_uris}
						class="w-full px-3 py-2 text-sm border rounded-md transition-colors focus:outline-none focus:ring-1 focus:ring-blue-500 {fieldErrors.redirect_uris ? 'border-red-300' : 'border-gray-300'}"
						placeholder="https://app.example.com/callback"
					/>
					{#if fieldErrors.redirect_uris}
						<p class="text-red-600 text-xs mt-1">{fieldErrors.redirect_uris}</p>
					{/if}
				</div>

				<div>
					<label for="scopes" class="block text-sm font-medium text-gray-700 mb-1">Scopes</label>
					<input
						id="scopes"
						type="text"
						bind:value={form.scopes}
						class="w-full px-3 py-2 text-sm border border-gray-300 rounded-md transition-colors focus:outline-none focus:ring-1 focus:ring-blue-500"
						placeholder="launch patient/*.read openid fhirUser"
					/>
					<p class="text-xs text-gray-400 mt-1">Default: launch patient/*.read openid fhirUser</p>
				</div>

				<div class="flex items-center gap-3">
					<label class="flex items-center gap-2 cursor-pointer">
						<input
							type="checkbox"
							bind:checked={form.is_confidential}
							class="w-4 h-4 text-blue-600 border-gray-300 rounded focus:ring-blue-500"
						/>
						<span class="text-sm text-gray-700">Confidential Client</span>
					</label>
					<span class="text-xs text-gray-400">(requires client secret for token exchange)</span>
				</div>

				<div class="flex justify-end gap-3 pt-2">
					<button
						type="button"
						onclick={closeModal}
						class="px-4 py-2 text-sm text-gray-600 hover:bg-gray-100 rounded-md transition-colors"
					>
						Cancel
					</button>
					<button
						type="submit"
						disabled={formSaving}
						class="px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-md hover:bg-blue-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
					>
						{formSaving ? 'Registering…' : 'Register App'}
					</button>
				</div>
			</form>
		</div>
	</div>
{/if}

<!-- One-time client credentials. A confidential client's secret is generated
     server-side, returned once in the 201 body and never stored in plaintext, so
     this is the only place it can be handed over. It is deliberately not kept in
     component state after the dialog is dismissed. -->
{#if createdClient}
	<div
		class="fixed inset-0 z-50 flex items-center justify-center bg-black/40"
		role="dialog"
		aria-modal="true"
		aria-labelledby="created-client-title"
	>
		<div class="bg-white rounded-lg shadow-xl w-full max-w-lg mx-4">
			<div class="px-6 py-4 border-b border-gray-100">
				<h3 id="created-client-title" class="text-lg font-semibold text-gray-900">
					Application registered
				</h3>
				<p class="text-xs text-gray-500 mt-0.5">
					Give these credentials to the application developer.
				</p>
			</div>

			<div class="p-6 space-y-4">
				{#if createdClient.client_secret}
					<div class="bg-amber-50 border border-amber-200 rounded-md p-3 text-sm text-amber-800">
						{createdClient.client_secret_note ??
							'This secret is shown once and cannot be retrieved. Store it now.'}
					</div>
				{:else}
					<div class="bg-gray-50 border border-gray-200 rounded-md p-3 text-sm text-gray-600">
						This is a public client, so no client secret was generated. Public clients must not be
						given one; the secret exists only for confidential clients.
					</div>
				{/if}

				<div>
					<div class="flex items-center justify-between mb-1">
						<span class="text-sm font-medium text-gray-700">Client ID</span>
						<button
							type="button"
							onclick={() => copyValue(createdClient!.client_id, 'client_id')}
							class="px-2 py-1 text-xs font-medium text-blue-600 hover:bg-blue-50 rounded-md transition-colors"
						>
							{copied === 'client_id' ? 'Copied' : 'Copy'}
						</button>
					</div>
					<code
						class="block w-full px-3 py-2 text-xs font-mono bg-gray-50 border border-gray-200 rounded-md break-all"
					>
						{createdClient.client_id}
					</code>
				</div>

				{#if createdClient.client_secret}
					<div>
						<div class="flex items-center justify-between mb-1">
							<span class="text-sm font-medium text-gray-700">Client Secret</span>
							<button
								type="button"
								onclick={() => copyValue(createdClient!.client_secret ?? '', 'client_secret')}
								class="px-2 py-1 text-xs font-medium text-blue-600 hover:bg-blue-50 rounded-md transition-colors"
							>
								{copied === 'client_secret' ? 'Copied' : 'Copy'}
							</button>
						</div>
						<code
							class="block w-full px-3 py-2 text-xs font-mono bg-amber-50 border border-amber-200 rounded-md break-all"
						>
							{createdClient.client_secret}
						</code>
					</div>
				{/if}

				<div class="flex justify-end pt-2">
					<button
						type="button"
						onclick={closeCreated}
						class="px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-md hover:bg-blue-700 transition-colors"
					>
						Done
					</button>
				</div>
			</div>
		</div>
	</div>
{/if}
