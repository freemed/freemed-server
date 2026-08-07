<script lang="ts">
	import { api } from '$lib/api';

	interface User {
		id: number;
		username: string;
		userfname: string | null;
		userlname: string | null;
		userdescrip: string | null;
		usertype: string | null;
	}

	interface UserForm {
		username: string;
		password: string;
		first_name: string;
		last_name: string;
		description: string;
		user_type: string;
	}

	let users = $state<User[]>([]);
	let loading = $state(true);
	let error = $state('');
	let showModal = $state(false);
	let editingUser = $state<User | null>(null);
	let form = $state<UserForm>(emptyForm());
	let formError = $state('');
	let formSaving = $state(false);
	let deleteConfirm = $state<User | null>(null);

	function emptyForm(): UserForm {
		return { username: '', password: '', first_name: '', last_name: '', description: '', user_type: '' };
	}

	async function loadUsers() {
		loading = true;
		error = '';
		try {
			users = await api.get<User[]>('/users');
		} catch (e: any) {
			error = e.message || 'Failed to load users';
		} finally {
			loading = false;
		}
	}

	$effect(() => {
		loadUsers();
	});

	function openCreate() {
		editingUser = null;
		form = emptyForm();
		formError = '';
		showModal = true;
	}

	function openEdit(user: User) {
		editingUser = user;
		form = {
			username: user.username,
			password: '',
			first_name: user.userfname || '',
			last_name: user.userlname || '',
			description: user.userdescrip || '',
			user_type: user.usertype || '',
		};
		formError = '';
		showModal = true;
	}

	function closeModal() {
		showModal = false;
		editingUser = null;
		formSaving = false;
		formError = '';
	}

	async function saveUser() {
		formError = '';
		formSaving = true;
		try {
			if (editingUser) {
				await api.put(`/users/${editingUser.id}`, {
					username: form.username,
					first_name: form.first_name,
					last_name: form.last_name,
					description: form.description,
					user_type: form.user_type,
				});
				if (form.password) {
					await api.put(`/users/${editingUser.id}/password`, {
						password: form.password,
					});
				}
			} else {
				await api.post('/users', form);
			}
			closeModal();
			await loadUsers();
		} catch (e: any) {
			formError = e.message || 'Save failed';
		} finally {
			formSaving = false;
		}
	}

	async function confirmDelete() {
		if (!deleteConfirm) return;
		try {
			await api.del(`/users/${deleteConfirm.id}`);
			deleteConfirm = null;
			await loadUsers();
		} catch (e: any) {
			error = e.message || 'Delete failed';
			deleteConfirm = null;
		}
	}
</script>

<div class="max-w-5xl mx-auto">
	<div class="flex items-center justify-between mb-6">
		<h1 class="text-2xl font-bold text-gray-900">User Management</h1>
		<button
			onclick={openCreate}
			class="px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded-lg hover:bg-blue-700 transition-colors"
		>
			+ Add User
		</button>
	</div>

	{#if error}
		<div class="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg text-sm mb-6">
			{error}
		</div>
	{/if}

	{#if loading}
		<div class="flex justify-center py-12">
			<div class="w-6 h-6 border-2 border-blue-500 border-t-transparent rounded-full animate-spin"></div>
		</div>
	{:else if users.length === 0}
		<div class="text-center py-12 text-gray-500">
			<p class="text-lg">No users found</p>
		</div>
	{:else}
		<div class="bg-white rounded-lg shadow-sm border border-gray-200 overflow-x-auto">
			<table class="w-full text-sm">
				<thead class="bg-gray-50 border-b border-gray-200">
					<tr>
						<th class="text-left px-4 py-3 font-semibold text-gray-600">Username</th>
						<th class="text-left px-4 py-3 font-semibold text-gray-600">Name</th>
						<th class="text-left px-4 py-3 font-semibold text-gray-600">Type</th>
						<th class="text-left px-4 py-3 font-semibold text-gray-600">Description</th>
						<th class="text-right px-4 py-3 font-semibold text-gray-600">Actions</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-gray-100">
					{#each users as user (user.id)}
						<tr class="hover:bg-gray-50">
							<td class="px-4 py-3 font-medium text-gray-900">{user.username}</td>
							<td class="px-4 py-3 text-gray-600">
								{[user.userfname, user.userlname].filter(Boolean).join(' ') || '—'}
							</td>
							<td class="px-4 py-3 text-gray-600">{user.usertype || '—'}</td>
							<td class="px-4 py-3 text-gray-600 max-w-xs truncate">{user.userdescrip || '—'}</td>
							<td class="px-4 py-3 text-right">
								<button
									onclick={() => openEdit(user)}
									class="px-3 py-1.5 text-xs font-medium text-blue-600 hover:bg-blue-50 rounded-md transition-colors"
								>
									Edit
								</button>
								<button
									onclick={() => (deleteConfirm = user)}
									class="px-3 py-1.5 text-xs font-medium text-red-600 hover:bg-red-50 rounded-md transition-colors ml-1"
								>
									Delete
								</button>
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{/if}
</div>

<!-- Add/Edit Modal -->
{#if showModal}
	<!-- svelte-ignore a11y_interactive_supports_focus -->
	<div
		class="fixed inset-0 z-50 flex items-center justify-center bg-black/40"
		onclick={closeModal}
		onkeydown={(e) => e.key === 'Escape' && closeModal()}
		role="dialog"
		aria-modal="true"
	>
		<div
			class="bg-white rounded-xl shadow-xl w-full max-w-lg mx-4"
			onclick={(e) => e.stopPropagation()}
			onkeydown={(e) => e.key === 'Escape' && closeModal()}
			role="presentation"
		>
			<div class="px-6 py-4 border-b border-gray-200 flex items-center justify-between">
				<h2 class="text-lg font-semibold text-gray-900">
					{editingUser ? 'Edit User' : 'Add User'}
				</h2>
				<button
					onclick={closeModal}
					class="text-gray-400 hover:text-gray-600 transition-colors"
					aria-label="Close"
				>
					<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
					</svg>
				</button>
			</div>

			<div class="px-6 py-4 space-y-4">
				{#if formError}
					<div class="bg-red-50 border border-red-200 text-red-700 px-3 py-2 rounded-lg text-sm">
						{formError}
					</div>
				{/if}

				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">Username *</label>
					<input
						type="text"
						bind:value={form.username}
						class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
						placeholder="Username"
						required
					/>
				</div>

				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">
						Password {editingUser ? '(leave blank to keep current)' : '*'}
					</label>
					<input
						type="password"
						bind:value={form.password}
						class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
						placeholder={editingUser ? 'New password' : 'Password'}
					/>
				</div>

				<div class="grid grid-cols-2 gap-4">
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">First Name</label>
						<input
							type="text"
							bind:value={form.first_name}
							class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
							placeholder="First name"
						/>
					</div>
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">Last Name</label>
						<input
							type="text"
							bind:value={form.last_name}
							class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
							placeholder="Last name"
						/>
					</div>
				</div>

				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">User Type</label>
					<input
						type="text"
						bind:value={form.user_type}
						class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
						placeholder="e.g. admin, physician, staff"
					/>
				</div>

				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">Description</label>
					<textarea
						bind:value={form.description}
						class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
						placeholder="Description"
						rows="2"
					></textarea>
				</div>
			</div>

			<div class="px-6 py-4 border-t border-gray-200 flex justify-end gap-3">
				<button
					onclick={closeModal}
					class="px-4 py-2 text-sm font-medium text-gray-700 bg-gray-100 rounded-lg hover:bg-gray-200 transition-colors"
				>
					Cancel
				</button>
				<button
					onclick={saveUser}
					disabled={formSaving || !form.username}
					class="px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
				>
					{formSaving ? 'Saving...' : editingUser ? 'Update' : 'Create'}
				</button>
			</div>
		</div>
	</div>
{/if}

<!-- Delete Confirmation Modal -->
{#if deleteConfirm}
	<!-- svelte-ignore a11y_interactive_supports_focus -->
	<div
		class="fixed inset-0 z-50 flex items-center justify-center bg-black/40"
		onclick={() => (deleteConfirm = null)}
		onkeydown={(e) => e.key === 'Escape' && (deleteConfirm = null)}
		role="dialog"
		aria-modal="true"
	>
		<div
			class="bg-white rounded-xl shadow-xl w-full max-w-sm mx-4 p-6"
			onclick={(e) => e.stopPropagation()}
			onkeydown={(e) => e.key === 'Escape' && (deleteConfirm = null)}
			role="presentation"
		>
			<h3 class="text-lg font-semibold text-gray-900 mb-2">Confirm Delete</h3>
			<p class="text-sm text-gray-600 mb-6">
				Are you sure you want to delete user <strong>{deleteConfirm.username}</strong>? This action cannot be undone.
			</p>
			<div class="flex justify-end gap-3">
				<button
					onclick={() => (deleteConfirm = null)}
					class="px-4 py-2 text-sm font-medium text-gray-700 bg-gray-100 rounded-lg hover:bg-gray-200 transition-colors"
				>
					Cancel
				</button>
				<button
					onclick={confirmDelete}
					class="px-4 py-2 text-sm font-medium text-white bg-red-600 rounded-lg hover:bg-red-700 transition-colors"
				>
					Delete
				</button>
			</div>
		</div>
	</div>
{/if}
