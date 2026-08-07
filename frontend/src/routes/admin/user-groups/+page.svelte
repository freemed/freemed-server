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

	interface Group {
		id: number;
		groupname: string;
		groupdescrip: string | null;
	}

	interface UserGroup {
		group_id: number;
		groupname: string;
		groupdescrip: string;
	}

	// --- Users ---
	let allUsers = $state<User[]>([]);
	let usersLoading = $state(true);
	let usersError = $state('');
	let searchQuery = $state('');
	let selectedUser = $state<User | null>(null);

	// --- Groups ---
	let allGroups = $state<Group[]>([]);
	let assignedGroups = $state<UserGroup[]>([]);
	let groupsLoading = $state(false);
	let groupsError = $state('');

	// Add group modal
	let showAddModal = $state(false);
	let addGroupId = $state<number | null>(null);
	let addError = $state('');
	let addSaving = $state(false);

	// Delete confirm
	let deleteConfirm = $state<number | null>(null);
	let deleteError = $state('');

	let filteredUsers = $derived.by(() => {
		const q = searchQuery.trim().toLowerCase();
		if (!q) return allUsers;
		return allUsers.filter(
			(u) =>
				u.username.toLowerCase().includes(q) ||
				(u.userfname || '').toLowerCase().includes(q) ||
				(u.userlname || '').toLowerCase().includes(q) ||
				(u.userdescrip || '').toLowerCase().includes(q)
		);
	});

	async function loadUsers() {
		usersLoading = true;
		usersError = '';
		try {
			allUsers = await api.get<User[]>('/users');
		} catch (e: any) {
			usersError = e.message || 'Failed to load users';
		} finally {
			usersLoading = false;
		}
	}

	async function loadGroups() {
		try {
			allGroups = await api.get<Group[]>('/acl/groups');
		} catch (_e) {
			// ignore
		}
	}

	async function loadUserGroups() {
		if (!selectedUser) return;
		groupsLoading = true;
		groupsError = '';
		try {
			assignedGroups = await api.get<UserGroup[]>(`/user-groups?user=${selectedUser.id}`);
		} catch (e: any) {
			groupsError = e.message || 'Failed to load groups';
		} finally {
			groupsLoading = false;
		}
	}

	function selectUser(user: User) {
		selectedUser = user;
	}

	function availableGroups(): Group[] {
		const ids = new Set(assignedGroups.map((g) => g.group_id));
		return allGroups.filter((g) => !ids.has(g.id));
	}

	function openAddModal() {
		addGroupId = null;
		addError = '';
		showAddModal = true;
	}

	function closeAddModal() {
		showAddModal = false;
		addSaving = false;
		addError = '';
	}

	async function addUserGroup() {
		if (!selectedUser || !addGroupId) return;
		addError = '';
		addSaving = true;
		try {
			await api.post('/user-groups', {
				user_id: selectedUser.id,
				group_id: addGroupId,
			});
			closeAddModal();
			await loadUserGroups();
		} catch (e: any) {
			addError = e.message || 'Failed to add group';
		} finally {
			addSaving = false;
		}
	}

	async function confirmDelete() {
		if (!selectedUser || !deleteConfirm) return;
		deleteError = '';
		try {
			await api.del(`/user-groups?user=${selectedUser.id}&group=${deleteConfirm}`);
			deleteConfirm = null;
			await loadUserGroups();
		} catch (e: any) {
			deleteError = e.message || 'Failed to remove group';
			deleteConfirm = null;
		}
	}

	$effect(() => {
		loadUsers();
		loadGroups();
	});

	$effect(() => {
		if (selectedUser) {
			loadUserGroups();
		}
	});
</script>

<div class="max-w-6xl mx-auto">
	<h1 class="text-2xl font-bold text-gray-900 mb-6">User Group Management</h1>

	{#if usersError}
		<div class="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg text-sm mb-6">
			{usersError}
		</div>
	{/if}

	<div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
		<!-- Left Panel: User Selector -->
		<div class="bg-white rounded-lg shadow-sm border border-gray-200">
			<div class="px-4 py-3 border-b border-gray-200">
				<h2 class="text-sm font-semibold text-gray-700 uppercase tracking-wide">Users</h2>
			</div>

			<!-- Search -->
			<div class="px-3 py-2">
				<input
					type="text"
					bind:value={searchQuery}
					placeholder="Search by name or username..."
					class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
				/>
			</div>

			<!-- User List -->
			<div class="max-h-96 overflow-y-auto">
				{#if usersLoading}
					<div class="flex justify-center py-8">
						<div class="w-5 h-5 border-2 border-blue-500 border-t-transparent rounded-full animate-spin"></div>
					</div>
				{:else if filteredUsers.length === 0}
					<div class="text-center py-8 text-gray-500 text-sm">
						{searchQuery ? 'No matching users found' : 'No users found'}
					</div>
				{:else}
					<ul class="divide-y divide-gray-100">
						{#each filteredUsers as user (user.id)}
							<li>
								<button
									onclick={() => selectUser(user)}
									class="w-full text-left px-4 py-3 text-sm hover:bg-gray-50 transition-colors {selectedUser?.id === user.id
										? 'bg-blue-50 border-l-2 border-blue-500'
										: 'border-l-2 border-transparent'}"
								>
									<div class="font-medium text-gray-900">{user.username}</div>
									<div class="text-gray-500 text-xs mt-0.5">
										{[user.userfname, user.userlname].filter(Boolean).join(' ') || '—'}
										{#if user.usertype}
											<span class="ml-2 px-1.5 py-0.5 bg-gray-100 rounded text-xs">{user.usertype}</span>
										{/if}
									</div>
								</button>
							</li>
						{/each}
					</ul>
				{/if}
			</div>
		</div>

		<!-- Right Panel: Group Assignments -->
		<div class="lg:col-span-2 bg-white rounded-lg shadow-sm border border-gray-200">
			<div class="px-4 py-3 border-b border-gray-200 flex items-center justify-between">
				<h2 class="text-sm font-semibold text-gray-700 uppercase tracking-wide">
					Group Assignments
					{#if selectedUser}
						<span class="font-normal text-gray-500 normal-case ml-2">
							— {selectedUser.username}
						</span>
					{/if}
				</h2>
				{#if selectedUser}
					<button
						onclick={openAddModal}
						disabled={availableGroups().length === 0}
						class="px-3 py-1.5 bg-blue-600 text-white text-xs font-medium rounded-md hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
					>
						+ Add Group
					</button>
				{/if}
			</div>

			{#if !selectedUser}
				<div class="flex items-center justify-center py-16 text-gray-400">
					<div class="text-center">
						<svg class="w-12 h-12 mx-auto mb-3 text-gray-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M15 15l-2 5L9 9l11 4-5 2zm0 0l5 5M7.188 2.239l.777 2.897M5.136 7.965l-2.898-.777M13.95 4.05l-2.122 2.122m-5.657 5.656l-2.12 2.122" />
						</svg>
						<p class="text-sm">Select a user from the left panel to manage their groups</p>
					</div>
				</div>
			{:else if groupsError}
				<div class="bg-red-50 border border-red-200 text-red-700 px-4 py-3 mx-4 mt-4 rounded-lg text-sm">
					{groupsError}
				</div>
			{:else if deleteError}
				<div class="bg-red-50 border border-red-200 text-red-700 px-4 py-3 mx-4 mt-4 rounded-lg text-sm">
					{deleteError}
				</div>
			{:else if groupsLoading}
				<div class="flex justify-center py-12">
					<div class="w-6 h-6 border-2 border-blue-500 border-t-transparent rounded-full animate-spin"></div>
				</div>
			{:else if assignedGroups.length === 0}
				<div class="text-center py-12 text-gray-500">
					<p class="text-sm">No groups assigned to this user</p>
				</div>
			{:else}
				<div class="overflow-x-auto">
					<table class="w-full text-sm">
						<thead class="bg-gray-50 border-b border-gray-200">
							<tr>
								<th class="text-left px-4 py-3 font-semibold text-gray-600">Group</th>
								<th class="text-left px-4 py-3 font-semibold text-gray-600">Description</th>
								<th class="text-right px-4 py-3 font-semibold text-gray-600">Actions</th>
							</tr>
						</thead>
						<tbody class="divide-y divide-gray-100">
							{#each assignedGroups as ug (ug.group_id)}
								<tr class="hover:bg-gray-50">
									<td class="px-4 py-3 font-medium text-gray-900">{ug.groupname}</td>
									<td class="px-4 py-3 text-gray-600 max-w-md truncate">{ug.groupdescrip || '—'}</td>
									<td class="px-4 py-3 text-right">
										<button
											onclick={() => (deleteConfirm = ug.group_id)}
											class="px-3 py-1.5 text-xs font-medium text-red-600 hover:bg-red-50 rounded-md transition-colors"
										>
											Remove
										</button>
									</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			{/if}
		</div>
	</div>
</div>

<!-- Add Group Modal -->
{#if showAddModal}
	<!-- svelte-ignore a11y_interactive_supports_focus -->
	<div
		class="fixed inset-0 z-50 flex items-center justify-center bg-black/40"
		onclick={closeAddModal}
		onkeydown={(e) => e.key === 'Escape' && closeAddModal()}
		role="dialog"
		aria-modal="true"
	>
		<div
			class="bg-white rounded-xl shadow-xl w-full max-w-sm mx-4"
			onclick={(e) => e.stopPropagation()}
			onkeydown={(e) => e.key === 'Escape' && closeAddModal()}
			role="presentation"
		>
			<div class="px-6 py-4 border-b border-gray-200 flex items-center justify-between">
				<h2 class="text-lg font-semibold text-gray-900">Add Group</h2>
				<button
					onclick={closeAddModal}
					class="text-gray-400 hover:text-gray-600 transition-colors"
					aria-label="Close"
				>
					<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
					</svg>
				</button>
			</div>

			<div class="px-6 py-4">
				{#if addError}
					<div class="bg-red-50 border border-red-200 text-red-700 px-3 py-2 rounded-lg text-sm mb-4">
						{addError}
					</div>
				{/if}

				<label class="block text-sm font-medium text-gray-700 mb-1">Select Group</label>
				<select
					bind:value={addGroupId}
					class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
				>
					<option value={null}>-- Choose a group --</option>
					{#each availableGroups() as group (group.id)}
						<option value={group.id}>
							{group.groupname}
							{#if group.groupdescrip} — {group.groupdescrip}{/if}
						</option>
					{/each}
				</select>

				{#if availableGroups().length === 0 && !groupsLoading}
					<p class="text-xs text-gray-500 mt-2">This user is already in all available groups.</p>
				{/if}
			</div>

			<div class="px-6 py-4 border-t border-gray-200 flex justify-end gap-3">
				<button
					onclick={closeAddModal}
					class="px-4 py-2 text-sm font-medium text-gray-700 bg-gray-100 rounded-lg hover:bg-gray-200 transition-colors"
				>
					Cancel
				</button>
				<button
					onclick={addUserGroup}
					disabled={addSaving || !addGroupId}
					class="px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
				>
					{addSaving ? 'Adding...' : 'Add'}
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
			<h3 class="text-lg font-semibold text-gray-900 mb-2">Confirm Remove</h3>
			<p class="text-sm text-gray-600 mb-6">
				Are you sure you want to remove this group assignment from <strong>{selectedUser?.username}</strong>?
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
					Remove
				</button>
			</div>
		</div>
	</div>
{/if}
