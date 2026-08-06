<script lang="ts">
	import { goto } from '$app/navigation';
	import { api } from '$lib/api';

	interface User {
		id: number;
		username: string;
	}

	let users = $state<User[]>([]);
	let recipientId = $state<number | null>(null);
	let subject = $state('');
	let body = $state('');
	let loading = $state(true);
	let submitting = $state(false);
	let error = $state('');
	let fieldErrors = $state<Record<string, string>>({});
	let sent = $state(false);

	async function loadUsers() {
		loading = true;
		try {
			const data = await api.get<User[]>('/messages/list_users');
			users = data || [];
		} catch (e: any) {
			error = e.message || 'Failed to load users';
		} finally {
			loading = false;
		}
	}

	function validate(): boolean {
		const errs: Record<string, string> = {};
		if (!recipientId) errs.recipient = 'Please select a recipient';
		if (!subject.trim()) errs.subject = 'Subject is required';
		if (!body.trim()) errs.body = 'Message body is required';
		fieldErrors = errs;
		return Object.keys(errs).length === 0;
	}

	async function handleSubmit(e: Event) {
		e.preventDefault();
		if (!validate()) return;

		submitting = true;
		error = '';
		try {
			await api.post('/messages/send', {
				recipient: recipientId,
				subject: subject.trim(),
				body: body.trim(),
			});
			sent = true;
		} catch (e: any) {
			error = e.message || 'Failed to send message';
		} finally {
			submitting = false;
		}
	}

	loadUsers();
</script>

<div class="max-w-2xl mx-auto">
	<div class="flex items-center gap-3 mb-6">
		<button
			onclick={() => goto('/messages')}
			class="p-1.5 text-gray-500 hover:text-gray-700 hover:bg-gray-100 rounded-lg transition-colors"
			aria-label="Back to messages"
		>
			<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
			</svg>
		</button>
		<h1 class="text-2xl font-bold text-gray-900">Compose Message</h1>
	</div>

	{#if sent}
		<div class="bg-green-50 border border-green-200 rounded-lg p-6 text-center">
			<svg
				class="w-12 h-12 text-green-500 mx-auto mb-3"
				fill="none"
				stroke="currentColor"
				viewBox="0 0 24 24"
			>
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
			</svg>
			<h2 class="text-lg font-semibold text-green-800 mb-1">Message Sent</h2>
			<p class="text-sm text-green-600 mb-4">Your message has been sent successfully.</p>
			<div class="flex justify-center gap-3">
				<button
					onclick={() => goto('/messages')}
					class="px-4 py-2 bg-green-600 text-white text-sm font-medium rounded-lg hover:bg-green-700 transition-colors"
				>
					Back to Messages
				</button>
				<button
					onclick={() => {
						sent = false;
						recipientId = null;
						subject = '';
						body = '';
						fieldErrors = {};
					}}
					class="px-4 py-2 bg-white border border-gray-300 text-gray-700 text-sm font-medium rounded-lg hover:bg-gray-50 transition-colors"
				>
					Send Another
				</button>
			</div>
		</div>
	{:else}
		{#if error}
			<div class="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg text-sm mb-6">
				{error}
			</div>
		{/if}

		<form onsubmit={handleSubmit} class="bg-white rounded-lg border border-gray-200 p-6 space-y-5">
			<!-- Recipient -->
			<div>
				<label for="recipient" class="block text-sm font-medium text-gray-700 mb-1">
					Recipient <span class="text-red-500">*</span>
				</label>
				{#if loading}
					<div class="w-4 h-4 border-2 border-blue-500 border-t-transparent rounded-full animate-spin ml-1"></div>
				{:else}
					<select
						id="recipient"
						bind:value={recipientId}
						class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none transition-shadow"
						class:border-red-300={!!fieldErrors.recipient}
						class:border-gray-300={!fieldErrors.recipient}
					>
						<option value={null}>-- Select a recipient --</option>
						{#each users as user (user.id)}
							<option value={user.id}>{user.username}</option>
						{/each}
					</select>
					{#if fieldErrors.recipient}
						<p class="text-red-500 text-xs mt-1">{fieldErrors.recipient}</p>
					{/if}
				{/if}
			</div>

			<!-- Subject -->
			<div>
				<label for="subject" class="block text-sm font-medium text-gray-700 mb-1">
					Subject <span class="text-red-500">*</span>
				</label>
				<input
					id="subject"
					type="text"
					bind:value={subject}
					placeholder="Enter subject..."
					class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none transition-shadow"
					class:border-red-300={!!fieldErrors.subject}
					class:border-gray-300={!fieldErrors.subject}
				/>
				{#if fieldErrors.subject}
					<p class="text-red-500 text-xs mt-1">{fieldErrors.subject}</p>
				{/if}
			</div>

			<!-- Body -->
			<div>
				<label for="body" class="block text-sm font-medium text-gray-700 mb-1">
					Message <span class="text-red-500">*</span>
				</label>
				<textarea
					id="body"
					bind:value={body}
					rows="8"
					placeholder="Type your message..."
					class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none transition-shadow resize-y"
					class:border-red-300={!!fieldErrors.body}
					class:border-gray-300={!fieldErrors.body}
				></textarea>
				{#if fieldErrors.body}
					<p class="text-red-500 text-xs mt-1">{fieldErrors.body}</p>
				{/if}
			</div>

			<!-- Actions -->
			<div class="flex items-center justify-end gap-3 pt-2">
				<button
					type="button"
					onclick={() => goto('/messages')}
					class="px-4 py-2 border border-gray-300 text-gray-700 text-sm font-medium rounded-lg hover:bg-gray-50 transition-colors"
				>
					Cancel
				</button>
				<button
					type="submit"
					disabled={submitting}
					class="px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
				>
					{#if submitting}
						Sending...
					{:else}
						Send
					{/if}
				</button>
			</div>
		</form>
	{/if}
</div>
