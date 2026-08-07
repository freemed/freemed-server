<script lang="ts">
	import { api } from '$lib/api';

	interface Group {
		id: number;
		groupname: string;
		groupdescrip: string | null;
	}

	interface Permission {
		id: number;
		permission_name: string;
		permission_desc: string | null;
	}

	// --- Tab state ---
	type Tab = 'groups' | 'permissions' | 'group-permissions' | 'user-groups';
	let activeTab = $state<Tab>('groups');

	// --- Groups ---
	let groups = $state<Group[]>([]);
	let groupsLoading = $state(true);
	let groupsError = $state('');

	// Group modal
	let showGroupModal = $state(false);
	let editingGroup = $state<Group | null>(null);
	let groupForm = $state({ groupname: '', groupdescrip: '' });
	let groupFormError = $state('');
	let groupFormSaving = $state(false);

	// Delete confirm
	let deleteGroupConfirm = $state<Group | null>(null);

	function emptyGroupForm() {
		return { groupname: '', groupdescrip: '' };
	}

	async function loadGroups() {
		groupsLoading = true;
		groupsError = '';
		try {
			groups = await api.get<Group[]>('/acl/groups');
		} catch (e: any) {
			groupsError = e.message || 'Failed to load groups';
		} finally {
			groupsLoading = false;
		}
	}

	$effect(() => {
		loadGroups();
	});

	function openCreateGroup() {
		editingGroup = null;
		groupForm = emptyGroupForm();
		groupFormError = '';
		showGroupModal = true;
	}

	function openEditGroup(group: Group) {
		editingGroup = group;
		groupForm = { groupname: group.groupname, groupdescrip: group.groupdescrip || '' };
		groupFormError = '';
		showGroupModal = true;
	}

	function closeGroupModal() {
		showGroupModal = false;
		editingGroup = null;
		groupFormSaving = false;
		groupFormError = '';
	}

	async function saveGroup() {
		groupFormError = '';
		groupFormSaving = true;
		try {
			if (editingGroup) {
				await api.put(`/acl/groups/${editingGroup.id}`, groupForm);
			} else {
				await api.post('/acl/groups', groupForm);
			}
			closeGroupModal();
			await loadGroups();
		} catch (e: any) {
			groupFormError = e.message || 'Save failed';
		} finally {
			groupFormSaving = false;
		}
	}

	async function confirmDeleteGroup() {
		if (!deleteGroupConfirm) return;
		try {
			await api.del(`/acl/groups/${deleteGroupConfirm.id}`);
			deleteGroupConfirm = null;
			await loadGroups();
		} catch (e: any) {
			groupsError = e.message || 'Delete failed';
			deleteGroupConfirm = null;
		}
	}

	// --- Permissions ---
	let permissions = $state<Permission[]>([]);
	let permissionsLoading = $state(true);
	let permissionsError = $state('');

	// Permission modal
	let showPermModal = $state(false);
	let permForm = $state({ permission_name: '', permission_desc: '' });
	let permFormError = $state('');
	let permFormSaving = $state(false);
	let deletePermConfirm = $state<Permission | null>(null);

	function emptyPermForm() {
		return { permission_name: '', permission_desc: '' };
	}

	async function loadPermissions() {
		permissionsLoading = true;
		permissionsError = '';
		try {
			permissions = await api.get<Permission[]>('/acl/permissions');
		} catch (e: any) {
			permissionsError = e.message || 'Failed to load permissions';
		} finally {
			permissionsLoading = false;
		}
	}

	$effect(() => {
		loadPermissions();
	});

	function openCreatePermission() {
		permForm = emptyPermForm();
		permFormError = '';
		showPermModal = true;
	}

	function closePermModal() {
		showPermModal = false;
		permFormSaving = false;
		permFormError = '';
	}

	async function savePermission() {
		permFormError = '';
		permFormSaving = true;
		try {
			await api.post('/acl/permissions', permForm);
			closePermModal();
			await loadPermissions();
		} catch (e: any) {
			permFormError = e.message || 'Save failed';
		} finally {
			permFormSaving = false;
		}
	}

	async function confirmDeletePermission() {
		if (!deletePermConfirm) return;
		try {
			await api.del(`/acl/permissions/${deletePermConfirm.id}`);
			deletePermConfirm = null;
			await loadPermissions();
		} catch (e: any) {
			permissionsError = e.message || 'Delete failed';
			deletePermConfirm = null;
		}
	}

	// --- Group Permissions ---
	let gpGroups = $state<Group[]>([]);
	let gpSelectedGroupId = $state<number | null>(null);
	let gpPermissionIds = $state<number[]>([]);
	let gpLoading = $state(false);
	let gpError = $state('');
	let gpAllPermissions = $state<Permission[]>([]);

	// Add permission modal
	let showGpAddModal = $state(false);
	let gpAddPermId = $state<number | null>(null);
	let gpAddError = $state('');
	let gpAddSaving = $state(false);

	// Delete confirm
	let deleteGpConfirm = $state<number | null>(null);
	let gpDeleteError = $state('');

	async function loadGpData() {
		if (!gpSelectedGroupId) return;
		gpError = '';
		gpLoading = true;
		try {
			gpAllPermissions = await api.get<Permission[]>('/acl/permissions');
			gpPermissionIds = await api.get<number[]>(`/acl/groups/${gpSelectedGroupId}/permissions`);
		} catch (e: any) {
			gpError = e.message || 'Failed to load data';
		} finally {
			gpLoading = false;
		}
	}

	async function loadGpGroups() {
		try {
			gpGroups = await api.get<Group[]>('/acl/groups');
		} catch (_e) {
			// ignore
		}
	}

	$effect(() => {
		loadGpGroups();
	});

	$effect(() => {
		if (gpSelectedGroupId && activeTab === 'group-permissions') {
			loadGpData();
		}
	});

	function openGpAddModal() {
		gpAddPermId = null;
		gpAddError = '';
		showGpAddModal = true;
	}

	function closeGpAddModal() {
		showGpAddModal = false;
		gpAddSaving = false;
		gpAddError = '';
	}

	async function addGroupPermission() {
		if (!gpSelectedGroupId || !gpAddPermId) return;
		gpAddError = '';
		gpAddSaving = true;
		try {
			await api.post(`/acl/groups/${gpSelectedGroupId}/permissions`, { permission_id: gpAddPermId });
			closeGpAddModal();
			await loadGpData();
		} catch (e: any) {
			gpAddError = e.message || 'Failed to add permission';
		} finally {
			gpAddSaving = false;
		}
	}

	async function confirmDeleteGp() {
		if (!gpSelectedGroupId || !deleteGpConfirm) return;
		gpDeleteError = '';
		try {
			await api.del(`/acl/groups/${gpSelectedGroupId}/permissions/${deleteGpConfirm}`);
			deleteGpConfirm = null;
			await loadGpData();
		} catch (e: any) {
			gpDeleteError = e.message || 'Delete failed';
			deleteGpConfirm = null;
		}
	}

	function getPermName(id: number): string {
		const p = gpAllPermissions.find((p) => p.id === id);
		return p ? p.permission_name : `#${id}`;
	}

	function availableGpPerms(): Permission[] {
		return gpAllPermissions.filter((p) => !gpPermissionIds.includes(p.id));
	}

	// --- User Groups ---
	let ugUserIdInput = $state('');
	let ugSearchedUserId = $state<number | null>(null);
	let ugGroupIds = $state<number[]>([]);
	let ugLoading = $state(false);
	let ugError = $state('');
	let ugAllGroups = $state<Group[]>([]);

	// Add group modal
	let showUgAddModal = $state(false);
	let ugAddGroupId = $state<number | null>(null);
	let ugAddError = $state('');
	let ugAddSaving = $state(false);

	// Delete confirm
	let deleteUgConfirm = $state<number | null>(null);
	let ugDeleteError = $state('');

	async function loadUgData() {
		if (!ugSearchedUserId) return;
		ugError = '';
		ugLoading = true;
		try {
			ugAllGroups = await api.get<Group[]>('/acl/groups');
			ugGroupIds = await api.get<number[]>(`/acl/users/${ugSearchedUserId}/groups`);
		} catch (e: any) {
			ugError = e.message || 'Failed to load data';
		} finally {
			ugLoading = false;
		}
	}

	$effect(() => {
		if (ugSearchedUserId && activeTab === 'user-groups') {
			loadUgData();
		}
	});

	function searchUserGroups() {
		const id = parseInt(ugUserIdInput.trim(), 10);
		if (isNaN(id) || id <= 0) {
			ugError = 'Please enter a valid numeric user ID';
			return;
		}
		ugSearchedUserId = id;
	}

	function openUgAddModal() {
		ugAddGroupId = null;
		ugAddError = '';
		showUgAddModal = true;
	}

	function closeUgAddModal() {
		showUgAddModal = false;
		ugAddSaving = false;
		ugAddError = '';
	}

	async function addUserGroup() {
		if (!ugSearchedUserId || !ugAddGroupId) return;
		ugAddError = '';
		ugAddSaving = true;
		try {
			await api.post(`/acl/users/${ugSearchedUserId}/groups`, { group_id: ugAddGroupId });
			closeUgAddModal();
			await loadUgData();
		} catch (e: any) {
			ugAddError = e.message || 'Failed to add group';
		} finally {
			ugAddSaving = false;
		}
	}

	async function confirmDeleteUg() {
		if (!ugSearchedUserId || !deleteUgConfirm) return;
		ugDeleteError = '';
		try {
			await api.del(`/acl/users/${ugSearchedUserId}/groups/${deleteUgConfirm}`);
			deleteUgConfirm = null;
			await loadUgData();
		} catch (e: any) {
			ugDeleteError = e.message || 'Delete failed';
			deleteUgConfirm = null;
		}
	}

	function getGroupName(id: number): string {
		const g = ugAllGroups.find((g) => g.id === id);
		return g ? g.groupname : `#${id}`;
	}

	function availableUgGroups(): Group[] {
		return ugAllGroups.filter((g) => !ugGroupIds.includes(g.id));
	}

	// --- Tab switching ---
	function switchTab(tab: Tab) {
		activeTab = tab;
		// Trigger reloads
		if (tab === 'groups') loadGroups();
		else if (tab === 'permissions') loadPermissions();
		else if (tab === 'group-permissions') {
			loadGpGroups();
			loadGpData();
		} else if (tab === 'user-groups') {
			loadUgData();
		}
	}
</script>

<div class="max-w-5xl mx-auto">
	<h1 class="text-2xl font-bold text-gray-900 mb-6">ACL Management</h1>

	<!-- Tab Navigation -->
	<nav class="flex border-b border-gray-200 mb-6">
		<button
			onclick={() => switchTab('groups')}
			class="px-4 py-2.5 text-sm font-medium border-b-2 transition-colors {activeTab === 'groups'
				? 'border-blue-600 text-blue-600'
				: 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'}"
		>
			Groups
		</button>
		<button
			onclick={() => switchTab('permissions')}
			class="px-4 py-2.5 text-sm font-medium border-b-2 transition-colors {activeTab === 'permissions'
				? 'border-blue-600 text-blue-600'
				: 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'}"
		>
			Permissions
		</button>
		<button
			onclick={() => switchTab('group-permissions')}
			class="px-4 py-2.5 text-sm font-medium border-b-2 transition-colors {activeTab === 'group-permissions'
				? 'border-blue-600 text-blue-600'
				: 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'}"
		>
			Group Permissions
		</button>
		<button
			onclick={() => switchTab('user-groups')}
			class="px-4 py-2.5 text-sm font-medium border-b-2 transition-colors {activeTab === 'user-groups'
				? 'border-blue-600 text-blue-600'
				: 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'}"
		>
			User Groups
		</button>
	</nav>

	<!-- ==================== GROUPS TAB ==================== -->
	{#if activeTab === 'groups'}
		<div class="flex items-center justify-between mb-6">
			<h2 class="text-lg font-semibold text-gray-900">Groups</h2>
			<button
				onclick={openCreateGroup}
				class="px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded-lg hover:bg-blue-700 transition-colors"
			>
				+ Add Group
			</button>
		</div>

		{#if groupsError}
			<div class="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg text-sm mb-6">
				{groupsError}
			</div>
		{/if}

		{#if groupsLoading}
			<div class="flex justify-center py-12">
				<div class="w-6 h-6 border-2 border-blue-500 border-t-transparent rounded-full animate-spin"></div>
			</div>
		{:else if groups.length === 0}
			<div class="text-center py-12 text-gray-500">
				<p class="text-lg">No groups found</p>
			</div>
		{:else}
			<div class="bg-white rounded-lg shadow-sm border border-gray-200 overflow-x-auto">
				<table class="w-full text-sm">
					<thead class="bg-gray-50 border-b border-gray-200">
						<tr>
							<th class="text-left px-4 py-3 font-semibold text-gray-600">Group Name</th>
							<th class="text-left px-4 py-3 font-semibold text-gray-600">Description</th>
							<th class="text-right px-4 py-3 font-semibold text-gray-600">Actions</th>
						</tr>
					</thead>
					<tbody class="divide-y divide-gray-100">
						{#each groups as group (group.id)}
							<tr class="hover:bg-gray-50">
								<td class="px-4 py-3 font-medium text-gray-900">{group.groupname}</td>
								<td class="px-4 py-3 text-gray-600 max-w-md truncate">{group.groupdescrip || '—'}</td>
								<td class="px-4 py-3 text-right">
									<button
										onclick={() => openEditGroup(group)}
										class="px-3 py-1.5 text-xs font-medium text-blue-600 hover:bg-blue-50 rounded-md transition-colors"
									>
										Edit
									</button>
									<button
										onclick={() => (deleteGroupConfirm = group)}
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
	{/if}

	<!-- ==================== PERMISSIONS TAB ==================== -->
	{#if activeTab === 'permissions'}
		<div class="flex items-center justify-between mb-6">
			<h2 class="text-lg font-semibold text-gray-900">Permissions</h2>
			<button
				onclick={openCreatePermission}
				class="px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded-lg hover:bg-blue-700 transition-colors"
			>
				+ Add Permission
			</button>
		</div>

		{#if permissionsError}
			<div class="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg text-sm mb-6">
				{permissionsError}
			</div>
		{/if}

		{#if permissionsLoading}
			<div class="flex justify-center py-12">
				<div class="w-6 h-6 border-2 border-blue-500 border-t-transparent rounded-full animate-spin"></div>
			</div>
		{:else if permissions.length === 0}
			<div class="text-center py-12 text-gray-500">
				<p class="text-lg">No permissions found</p>
			</div>
		{:else}
			<div class="bg-white rounded-lg shadow-sm border border-gray-200 overflow-x-auto">
				<table class="w-full text-sm">
					<thead class="bg-gray-50 border-b border-gray-200">
						<tr>
							<th class="text-left px-4 py-3 font-semibold text-gray-600">Permission Name</th>
							<th class="text-left px-4 py-3 font-semibold text-gray-600">Description</th>
							<th class="text-right px-4 py-3 font-semibold text-gray-600">Actions</th>
						</tr>
					</thead>
					<tbody class="divide-y divide-gray-100">
						{#each permissions as perm (perm.id)}
							<tr class="hover:bg-gray-50">
								<td class="px-4 py-3 font-medium text-gray-900">{perm.permission_name}</td>
								<td class="px-4 py-3 text-gray-600 max-w-md truncate">{perm.permission_desc || '—'}</td>
								<td class="px-4 py-3 text-right">
									<button
										onclick={() => (deletePermConfirm = perm)}
										class="px-3 py-1.5 text-xs font-medium text-red-600 hover:bg-red-50 rounded-md transition-colors"
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
	{/if}

	<!-- ==================== GROUP PERMISSIONS TAB ==================== -->
	{#if activeTab === 'group-permissions'}
		<h2 class="text-lg font-semibold text-gray-900 mb-4">Group Permissions</h2>

		<div class="mb-6">
			<label class="block text-sm font-medium text-gray-700 mb-1">Select Group</label>
			<select
				bind:value={gpSelectedGroupId}
				class="w-full max-w-sm px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
			>
				<option value={null}>— Select a group —</option>
				{#each gpGroups as g (g.id)}
					<option value={g.id}>{g.groupname}</option>
				{/each}
			</select>
		</div>

		{#if gpError}
			<div class="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg text-sm mb-6">
				{gpError}
			</div>
		{/if}

		{#if gpSelectedGroupId}
			{#if gpLoading}
				<div class="flex justify-center py-12">
					<div class="w-6 h-6 border-2 border-blue-500 border-t-transparent rounded-full animate-spin"></div>
				</div>
			{:else}
				<div class="flex items-center justify-between mb-4">
					<h3 class="text-sm font-semibold text-gray-700">
						Assigned Permissions ({gpPermissionIds.length})
					</h3>
					<button
						onclick={openGpAddModal}
						class="px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded-lg hover:bg-blue-700 transition-colors"
					>
						+ Add Permission
					</button>
				</div>

				{#if gpDeleteError}
					<div class="bg-red-50 border border-red-200 text-red-700 px-3 py-2 rounded-lg text-sm mb-4">
						{gpDeleteError}
					</div>
				{/if}

				{#if gpPermissionIds.length === 0}
					<div class="text-center py-8 text-gray-500 bg-white rounded-lg shadow-sm border border-gray-200">
						<p>No permissions assigned to this group</p>
					</div>
				{:else}
					<div class="bg-white rounded-lg shadow-sm border border-gray-200 overflow-x-auto">
						<table class="w-full text-sm">
							<thead class="bg-gray-50 border-b border-gray-200">
								<tr>
									<th class="text-left px-4 py-3 font-semibold text-gray-600">Permission</th>
									<th class="text-right px-4 py-3 font-semibold text-gray-600">Actions</th>
								</tr>
							</thead>
							<tbody class="divide-y divide-gray-100">
								{#each gpPermissionIds as permId (permId)}
									<tr class="hover:bg-gray-50">
										<td class="px-4 py-3 font-medium text-gray-900">{getPermName(permId)}</td>
										<td class="px-4 py-3 text-right">
											<button
												onclick={() => (deleteGpConfirm = permId)}
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
			{/if}
		{:else}
			<div class="text-center py-12 text-gray-500">
				<p class="text-lg">Select a group to manage its permissions</p>
			</div>
		{/if}
	{/if}

	<!-- ==================== USER GROUPS TAB ==================== -->
	{#if activeTab === 'user-groups'}
		<h2 class="text-lg font-semibold text-gray-900 mb-4">User Groups</h2>

		<div class="mb-6 flex items-end gap-3">
			<div class="flex-1 max-w-sm">
				<label class="block text-sm font-medium text-gray-700 mb-1">User ID</label>
				<input
					type="number"
					bind:value={ugUserIdInput}
					class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
					placeholder="Enter user ID"
				/>
			</div>
			<button
				onclick={searchUserGroups}
				class="px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded-lg hover:bg-blue-700 transition-colors"
			>
				Search
			</button>
		</div>

		{#if ugError}
			<div class="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg text-sm mb-6">
				{ugError}
			</div>
		{/if}

		{#if ugSearchedUserId}
			{#if ugLoading}
				<div class="flex justify-center py-12">
					<div class="w-6 h-6 border-2 border-blue-500 border-t-transparent rounded-full animate-spin"></div>
				</div>
			{:else}
				<div class="flex items-center justify-between mb-4">
					<h3 class="text-sm font-semibold text-gray-700">
						Assigned Groups for User #{ugSearchedUserId} ({ugGroupIds.length})
					</h3>
					<button
						onclick={openUgAddModal}
						class="px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded-lg hover:bg-blue-700 transition-colors"
					>
						+ Add Group
					</button>
				</div>

				{#if ugDeleteError}
					<div class="bg-red-50 border border-red-200 text-red-700 px-3 py-2 rounded-lg text-sm mb-4">
						{ugDeleteError}
					</div>
				{/if}

				{#if ugGroupIds.length === 0}
					<div class="text-center py-8 text-gray-500 bg-white rounded-lg shadow-sm border border-gray-200">
						<p>User is not in any groups</p>
					</div>
				{:else}
					<div class="bg-white rounded-lg shadow-sm border border-gray-200 overflow-x-auto">
						<table class="w-full text-sm">
							<thead class="bg-gray-50 border-b border-gray-200">
								<tr>
									<th class="text-left px-4 py-3 font-semibold text-gray-600">Group</th>
									<th class="text-right px-4 py-3 font-semibold text-gray-600">Actions</th>
								</tr>
							</thead>
							<tbody class="divide-y divide-gray-100">
								{#each ugGroupIds as groupId (groupId)}
									<tr class="hover:bg-gray-50">
										<td class="px-4 py-3 font-medium text-gray-900">{getGroupName(groupId)}</td>
										<td class="px-4 py-3 text-right">
											<button
												onclick={() => (deleteUgConfirm = groupId)}
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
			{/if}
		{:else}
			<div class="text-center py-12 text-gray-500">
				<p class="text-lg">Enter a user ID to view their group memberships</p>
			</div>
		{/if}
	{/if}
</div>

<!-- ==================== GROUPS: Add/Edit Modal ==================== -->
{#if showGroupModal}
	<!-- svelte-ignore a11y_interactive_supports_focus -->
	<div
		class="fixed inset-0 z-50 flex items-center justify-center bg-black/40"
		onclick={closeGroupModal}
		onkeydown={(e) => e.key === 'Escape' && closeGroupModal()}
		role="dialog"
		aria-modal="true"
	>
		<div
			class="bg-white rounded-xl shadow-xl w-full max-w-lg mx-4"
			onclick={(e) => e.stopPropagation()}
			onkeydown={(e) => e.key === 'Escape' && closeGroupModal()}
			role="presentation"
		>
			<div class="px-6 py-4 border-b border-gray-200 flex items-center justify-between">
				<h2 class="text-lg font-semibold text-gray-900">
					{editingGroup ? 'Edit Group' : 'Add Group'}
				</h2>
				<button
					onclick={closeGroupModal}
					class="text-gray-400 hover:text-gray-600 transition-colors"
					aria-label="Close"
				>
					<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
					</svg>
				</button>
			</div>

			<div class="px-6 py-4 space-y-4">
				{#if groupFormError}
					<div class="bg-red-50 border border-red-200 text-red-700 px-3 py-2 rounded-lg text-sm">
						{groupFormError}
					</div>
				{/if}

				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">Group Name *</label>
					<input
						type="text"
						bind:value={groupForm.groupname}
						class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
						placeholder="Group name"
						required
					/>
				</div>

				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">Description</label>
					<textarea
						bind:value={groupForm.groupdescrip}
						class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
						placeholder="Description"
						rows="2"
					></textarea>
				</div>
			</div>

			<div class="px-6 py-4 border-t border-gray-200 flex justify-end gap-3">
				<button
					onclick={closeGroupModal}
					class="px-4 py-2 text-sm font-medium text-gray-700 bg-gray-100 rounded-lg hover:bg-gray-200 transition-colors"
				>
					Cancel
				</button>
				<button
					onclick={saveGroup}
					disabled={groupFormSaving || !groupForm.groupname}
					class="px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
				>
					{groupFormSaving ? 'Saving...' : editingGroup ? 'Update' : 'Create'}
				</button>
			</div>
		</div>
	</div>
{/if}

<!-- ==================== GROUPS: Delete Confirmation Modal ==================== -->
{#if deleteGroupConfirm}
	<!-- svelte-ignore a11y_interactive_supports_focus -->
	<div
		class="fixed inset-0 z-50 flex items-center justify-center bg-black/40"
		onclick={() => (deleteGroupConfirm = null)}
		onkeydown={(e) => e.key === 'Escape' && (deleteGroupConfirm = null)}
		role="dialog"
		aria-modal="true"
	>
		<div
			class="bg-white rounded-xl shadow-xl w-full max-w-sm mx-4 p-6"
			onclick={(e) => e.stopPropagation()}
			onkeydown={(e) => e.key === 'Escape' && (deleteGroupConfirm = null)}
			role="presentation"
		>
			<h3 class="text-lg font-semibold text-gray-900 mb-2">Confirm Delete</h3>
			<p class="text-sm text-gray-600 mb-6">
				Are you sure you want to delete group <strong>{deleteGroupConfirm.groupname}</strong>? This
				action cannot be undone.
			</p>
			<div class="flex justify-end gap-3">
				<button
					onclick={() => (deleteGroupConfirm = null)}
					class="px-4 py-2 text-sm font-medium text-gray-700 bg-gray-100 rounded-lg hover:bg-gray-200 transition-colors"
				>
					Cancel
				</button>
				<button
					onclick={confirmDeleteGroup}
					class="px-4 py-2 text-sm font-medium text-white bg-red-600 rounded-lg hover:bg-red-700 transition-colors"
				>
					Delete
				</button>
			</div>
		</div>
	</div>
{/if}

<!-- ==================== PERMISSIONS: Add Modal ==================== -->
{#if showPermModal}
	<!-- svelte-ignore a11y_interactive_supports_focus -->
	<div
		class="fixed inset-0 z-50 flex items-center justify-center bg-black/40"
		onclick={closePermModal}
		onkeydown={(e) => e.key === 'Escape' && closePermModal()}
		role="dialog"
		aria-modal="true"
	>
		<div
			class="bg-white rounded-xl shadow-xl w-full max-w-lg mx-4"
			onclick={(e) => e.stopPropagation()}
			onkeydown={(e) => e.key === 'Escape' && closePermModal()}
			role="presentation"
		>
			<div class="px-6 py-4 border-b border-gray-200 flex items-center justify-between">
				<h2 class="text-lg font-semibold text-gray-900">Add Permission</h2>
				<button
					onclick={closePermModal}
					class="text-gray-400 hover:text-gray-600 transition-colors"
					aria-label="Close"
				>
					<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
					</svg>
				</button>
			</div>

			<div class="px-6 py-4 space-y-4">
				{#if permFormError}
					<div class="bg-red-50 border border-red-200 text-red-700 px-3 py-2 rounded-lg text-sm">
						{permFormError}
					</div>
				{/if}

				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">Permission Name *</label>
					<input
						type="text"
						bind:value={permForm.permission_name}
						class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
						placeholder="Permission name"
						required
					/>
				</div>

				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">Description</label>
					<textarea
						bind:value={permForm.permission_desc}
						class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
						placeholder="Description"
						rows="2"
					></textarea>
				</div>
			</div>

			<div class="px-6 py-4 border-t border-gray-200 flex justify-end gap-3">
				<button
					onclick={closePermModal}
					class="px-4 py-2 text-sm font-medium text-gray-700 bg-gray-100 rounded-lg hover:bg-gray-200 transition-colors"
				>
					Cancel
				</button>
				<button
					onclick={savePermission}
					disabled={permFormSaving || !permForm.permission_name}
					class="px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
				>
					{permFormSaving ? 'Saving...' : 'Create'}
				</button>
			</div>
		</div>
	</div>
{/if}

<!-- ==================== PERMISSIONS: Delete Confirmation Modal ==================== -->
{#if deletePermConfirm}
	<!-- svelte-ignore a11y_interactive_supports_focus -->
	<div
		class="fixed inset-0 z-50 flex items-center justify-center bg-black/40"
		onclick={() => (deletePermConfirm = null)}
		onkeydown={(e) => e.key === 'Escape' && (deletePermConfirm = null)}
		role="dialog"
		aria-modal="true"
	>
		<div
			class="bg-white rounded-xl shadow-xl w-full max-w-sm mx-4 p-6"
			onclick={(e) => e.stopPropagation()}
			onkeydown={(e) => e.key === 'Escape' && (deletePermConfirm = null)}
			role="presentation"
		>
			<h3 class="text-lg font-semibold text-gray-900 mb-2">Confirm Delete</h3>
			<p class="text-sm text-gray-600 mb-6">
				Are you sure you want to delete permission
				<strong>{deletePermConfirm.permission_name}</strong>? This action cannot be undone.
			</p>
			<div class="flex justify-end gap-3">
				<button
					onclick={() => (deletePermConfirm = null)}
					class="px-4 py-2 text-sm font-medium text-gray-700 bg-gray-100 rounded-lg hover:bg-gray-200 transition-colors"
				>
					Cancel
				</button>
				<button
					onclick={confirmDeletePermission}
					class="px-4 py-2 text-sm font-medium text-white bg-red-600 rounded-lg hover:bg-red-700 transition-colors"
				>
					Delete
				</button>
			</div>
		</div>
	</div>
{/if}

<!-- ==================== GROUP PERMISSIONS: Add Modal ==================== -->
{#if showGpAddModal}
	<!-- svelte-ignore a11y_interactive_supports_focus -->
	<div
		class="fixed inset-0 z-50 flex items-center justify-center bg-black/40"
		onclick={closeGpAddModal}
		onkeydown={(e) => e.key === 'Escape' && closeGpAddModal()}
		role="dialog"
		aria-modal="true"
	>
		<div
			class="bg-white rounded-xl shadow-xl w-full max-w-md mx-4"
			onclick={(e) => e.stopPropagation()}
			onkeydown={(e) => e.key === 'Escape' && closeGpAddModal()}
			role="presentation"
		>
			<div class="px-6 py-4 border-b border-gray-200 flex items-center justify-between">
				<h2 class="text-lg font-semibold text-gray-900">Add Permission to Group</h2>
				<button
					onclick={closeGpAddModal}
					class="text-gray-400 hover:text-gray-600 transition-colors"
					aria-label="Close"
				>
					<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
					</svg>
				</button>
			</div>

			<div class="px-6 py-4 space-y-4">
				{#if gpAddError}
					<div class="bg-red-50 border border-red-200 text-red-700 px-3 py-2 rounded-lg text-sm">
						{gpAddError}
					</div>
				{/if}

				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">Permission</label>
					<select
						bind:value={gpAddPermId}
						class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
					>
						<option value={null}>— Select permission —</option>
						{#each availableGpPerms() as p (p.id)}
							<option value={p.id}>{p.permission_name}</option>
						{/each}
					</select>
					{#if availableGpPerms().length === 0}
						<p class="text-xs text-gray-500 mt-1">All permissions are already assigned.</p>
					{/if}
				</div>
			</div>

			<div class="px-6 py-4 border-t border-gray-200 flex justify-end gap-3">
				<button
					onclick={closeGpAddModal}
					class="px-4 py-2 text-sm font-medium text-gray-700 bg-gray-100 rounded-lg hover:bg-gray-200 transition-colors"
				>
					Cancel
				</button>
				<button
					onclick={addGroupPermission}
					disabled={gpAddSaving || !gpAddPermId}
					class="px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
				>
					{gpAddSaving ? 'Adding...' : 'Add'}
				</button>
			</div>
		</div>
	</div>
{/if}

<!-- ==================== GROUP PERMISSIONS: Remove Confirmation Modal ==================== -->
{#if deleteGpConfirm}
	<!-- svelte-ignore a11y_interactive_supports_focus -->
	<div
		class="fixed inset-0 z-50 flex items-center justify-center bg-black/40"
		onclick={() => (deleteGpConfirm = null)}
		onkeydown={(e) => e.key === 'Escape' && (deleteGpConfirm = null)}
		role="dialog"
		aria-modal="true"
	>
		<div
			class="bg-white rounded-xl shadow-xl w-full max-w-sm mx-4 p-6"
			onclick={(e) => e.stopPropagation()}
			onkeydown={(e) => e.key === 'Escape' && (deleteGpConfirm = null)}
			role="presentation"
		>
			<h3 class="text-lg font-semibold text-gray-900 mb-2">Confirm Remove</h3>
			<p class="text-sm text-gray-600 mb-6">
				Are you sure you want to remove permission
				<strong>{getPermName(deleteGpConfirm)}</strong> from this group?
			</p>
			<div class="flex justify-end gap-3">
				<button
					onclick={() => (deleteGpConfirm = null)}
					class="px-4 py-2 text-sm font-medium text-gray-700 bg-gray-100 rounded-lg hover:bg-gray-200 transition-colors"
				>
					Cancel
				</button>
				<button
					onclick={confirmDeleteGp}
					class="px-4 py-2 text-sm font-medium text-white bg-red-600 rounded-lg hover:bg-red-700 transition-colors"
				>
					Remove
				</button>
			</div>
		</div>
	</div>
{/if}

<!-- ==================== USER GROUPS: Add Modal ==================== -->
{#if showUgAddModal}
	<!-- svelte-ignore a11y_interactive_supports_focus -->
	<div
		class="fixed inset-0 z-50 flex items-center justify-center bg-black/40"
		onclick={closeUgAddModal}
		onkeydown={(e) => e.key === 'Escape' && closeUgAddModal()}
		role="dialog"
		aria-modal="true"
	>
		<div
			class="bg-white rounded-xl shadow-xl w-full max-w-md mx-4"
			onclick={(e) => e.stopPropagation()}
			onkeydown={(e) => e.key === 'Escape' && closeUgAddModal()}
			role="presentation"
		>
			<div class="px-6 py-4 border-b border-gray-200 flex items-center justify-between">
				<h2 class="text-lg font-semibold text-gray-900">Add User to Group</h2>
				<button
					onclick={closeUgAddModal}
					class="text-gray-400 hover:text-gray-600 transition-colors"
					aria-label="Close"
				>
					<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
					</svg>
				</button>
			</div>

			<div class="px-6 py-4 space-y-4">
				{#if ugAddError}
					<div class="bg-red-50 border border-red-200 text-red-700 px-3 py-2 rounded-lg text-sm">
						{ugAddError}
					</div>
				{/if}

				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">Group</label>
					<select
						bind:value={ugAddGroupId}
						class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
					>
						<option value={null}>— Select group —</option>
						{#each availableUgGroups() as g (g.id)}
							<option value={g.id}>{g.groupname}</option>
						{/each}
					</select>
					{#if availableUgGroups().length === 0}
						<p class="text-xs text-gray-500 mt-1">User is already in all groups.</p>
					{/if}
				</div>
			</div>

			<div class="px-6 py-4 border-t border-gray-200 flex justify-end gap-3">
				<button
					onclick={closeUgAddModal}
					class="px-4 py-2 text-sm font-medium text-gray-700 bg-gray-100 rounded-lg hover:bg-gray-200 transition-colors"
				>
					Cancel
				</button>
				<button
					onclick={addUserGroup}
					disabled={ugAddSaving || !ugAddGroupId}
					class="px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
				>
					{ugAddSaving ? 'Adding...' : 'Add'}
				</button>
			</div>
		</div>
	</div>
{/if}

<!-- ==================== USER GROUPS: Remove Confirmation Modal ==================== -->
{#if deleteUgConfirm}
	<!-- svelte-ignore a11y_interactive_supports_focus -->
	<div
		class="fixed inset-0 z-50 flex items-center justify-center bg-black/40"
		onclick={() => (deleteUgConfirm = null)}
		onkeydown={(e) => e.key === 'Escape' && (deleteUgConfirm = null)}
		role="dialog"
		aria-modal="true"
	>
		<div
			class="bg-white rounded-xl shadow-xl w-full max-w-sm mx-4 p-6"
			onclick={(e) => e.stopPropagation()}
			onkeydown={(e) => e.key === 'Escape' && (deleteUgConfirm = null)}
			role="presentation"
		>
			<h3 class="text-lg font-semibold text-gray-900 mb-2">Confirm Remove</h3>
			<p class="text-sm text-gray-600 mb-6">
				Are you sure you want to remove this user from group
				<strong>{getGroupName(deleteUgConfirm)}</strong>?
			</p>
			<div class="flex justify-end gap-3">
				<button
					onclick={() => (deleteUgConfirm = null)}
					class="px-4 py-2 text-sm font-medium text-gray-700 bg-gray-100 rounded-lg hover:bg-gray-200 transition-colors"
				>
					Cancel
				</button>
				<button
					onclick={confirmDeleteUg}
					class="px-4 py-2 text-sm font-medium text-white bg-red-600 rounded-lg hover:bg-red-700 transition-colors"
				>
					Remove
				</button>
			</div>
		</div>
	</div>
{/if}
