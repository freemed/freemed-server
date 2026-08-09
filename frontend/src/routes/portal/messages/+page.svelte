<script lang="ts">
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
	import ErrorBanner from '$lib/components/ErrorBanner.svelte';
	import EmptyState from '$lib/components/EmptyState.svelte';
	import { z } from 'zod';

	interface PortalMessage {
		id: number;
		message_time: string;
		subject: string;
		body: string;
		sender: string;
		is_read: number;
		urgency: number;
		tag: string | null;
		created_at: string;
	}

	let messages = $state<PortalMessage[]>([]);
	let loading = $state(true);
	let error = $state('');

	let showCompose = $state(false);
	let showMessage = $state<PortalMessage | null>(null);
	let sending = $state(false);
	let sendError = $state('');
	let successMsg = $state('');

	let composeSubject = $state('');
	let composeBody = $state('');
	let composeCategory = $state('');

	const composeSchema = z.object({
		subject: z.string().min(1, 'Subject is required').max(200, 'Subject is too long'),
		body: z.string().min(1, 'Message body is required').max(5000, 'Message is too long'),
	});

	let fieldErrors = $state<Record<string, string>>({});

	async function portalFetch<T>(path: string): Promise<T> {
		const res = await fetch(`/api/portal${path}`);
		if (!res.ok) {
			const body = await res.text();
			throw new Error(`API error ${res.status}: ${body}`);
		}
		return res.json();
	}

	async function loadMessages() {
		loading = true;
		error = '';
		try {
			messages = await portalFetch<PortalMessage[]>('/messages');
			messages = messages || [];
		} catch (e: any) {
			error = e.message || 'Failed to load messages';
		} finally {
			loading = false;
		}
	}

	$effect(() => {
		loadMessages();
	});

	function openCompose() {
		composeSubject = '';
		composeBody = '';
		composeCategory = '';
		fieldErrors = {};
		sendError = '';
		showCompose = true;
		showMessage = null;
	}

	function closeCompose() {
		showCompose = false;
	}

	function openMessage(msg: PortalMessage) {
		showMessage = msg;
		showCompose = false;
	}

	function closeMessage() {
		showMessage = null;
	}

	async function sendMessage(e: SubmitEvent) {
		e.preventDefault();
		sendError = '';
		fieldErrors = {};

		const result = composeSchema.safeParse({ subject: composeSubject, body: composeBody });
		if (!result.success) {
			for (const issue of result.error.issues) {
				fieldErrors = { ...fieldErrors, [issue.path[0] as string]: issue.message };
			}
			return;
		}

		sending = true;
		try {
			const res = await fetch('/api/portal/messages', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					subject: composeSubject,
					body: composeBody,
					category: composeCategory || undefined,
				}),
			});
			if (!res.ok) {
				const body = await res.json().catch(() => ({ message: 'Failed to send' }));
				throw new Error(body.message || 'Failed to send message');
			}
			closeCompose();
			successMsg = 'Message sent successfully!';
			await loadMessages();
			setTimeout(() => { successMsg = ''; }, 5000);
		} catch (e: any) {
			sendError = e.message || 'Failed to send message';
		} finally {
			sending = false;
		}
	}

	function formatDate(dateStr: string): string {
		if (!dateStr) return '—';
		const d = new Date(dateStr);
		if (isNaN(d.getTime())) return dateStr;
		return d.toLocaleDateString('en-US', {
			month: 'short',
			day: 'numeric',
			hour: 'numeric',
			minute: '2-digit',
		});
	}

	function urgencyLabel(urgency: number): string {
		switch (urgency) {
			case 1: return 'Normal';
			case 2: return 'Urgent';
			case 3: return 'Emergency';
			default: return 'Normal';
		}
	}

	function urgencyColor(urgency: number): string {
		switch (urgency) {
			case 2: return 'bg-yellow-100 text-yellow-800';
			case 3: return 'bg-red-100 text-red-800';
			default: return 'bg-gray-100 text-gray-600';
		}
	}
</script>

<div class="max-w-7xl mx-auto">
	<div class="mb-6 flex justify-between items-center">
		<div>
			<h1 class="text-2xl font-bold text-gray-900">Secure Messages</h1>
			<p class="text-sm text-gray-500 mt-1">Communicate with your healthcare team</p>
		</div>
		<div class="flex gap-2">
			<button onclick={loadMessages} class="px-3 py-1.5 text-sm text-gray-600 hover:bg-gray-100 rounded-md transition-colors">Refresh</button>
			<button
				onclick={openCompose}
				class="px-4 py-1.5 text-sm font-medium text-white bg-blue-600 rounded-md hover:bg-blue-700 transition-colors"
			>
				+ New Message
			</button>
		</div>
	</div>

	{#if successMsg}
		<div class="mb-4 bg-green-50 border border-green-200 rounded-lg p-3 flex items-center gap-2">
			<svg class="w-5 h-5 text-green-500 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
			</svg>
			<span class="text-sm text-green-700">{successMsg}</span>
		</div>
	{/if}

	{#if loading}
		<LoadingSpinner message="Loading messages…" />
	{:else if error}
		<ErrorBanner message={error} onRetry={loadMessages} />
	{:else if showCompose || showMessage}
		<!-- Message detail or compose view -->
		<div class="bg-white rounded-lg shadow-sm border border-gray-200">
			{#if showCompose}
				<div class="px-6 py-4 border-b border-gray-100 flex justify-between items-center">
					<h3 class="font-semibold text-gray-900">New Message</h3>
					<button onclick={closeCompose} class="text-sm text-gray-500 hover:text-gray-700">&larr; Back to messages</button>
				</div>
				<form onsubmit={sendMessage} class="p-6 space-y-4">
					{#if sendError}
						<div class="bg-red-50 border border-red-200 rounded-md p-3 text-sm text-red-700">{sendError}</div>
					{/if}
					<div>
						<label for="composeSubject" class="block text-sm font-medium text-gray-700 mb-1">Subject *</label>
						<input
							id="composeSubject"
							type="text"
							bind:value={composeSubject}
							class="w-full px-3 py-2 text-sm border rounded-md focus:outline-none focus:ring-1 focus:ring-blue-500 {fieldErrors.subject ? 'border-red-300' : 'border-gray-300'}"
							placeholder="Message subject"
						/>
						{#if fieldErrors.subject}
							<p class="text-red-600 text-xs mt-1">{fieldErrors.subject}</p>
						{/if}
					</div>
					<div>
						<label for="composeCategory" class="block text-sm font-medium text-gray-700 mb-1">Category</label>
						<select
							id="composeCategory"
							bind:value={composeCategory}
							class="w-full px-3 py-2 text-sm border border-gray-300 rounded-md focus:outline-none focus:ring-1 focus:ring-blue-500"
						>
							<option value="">General</option>
							<option value="appointment">Appointment</option>
							<option value="billing">Billing</option>
							<option value="prescription">Prescription</option>
							<option value="referral">Referral</option>
							<option value="medical_records">Medical Records</option>
						</select>
					</div>
					<div>
						<label for="composeBody" class="block text-sm font-medium text-gray-700 mb-1">Message *</label>
						<textarea
							id="composeBody"
							bind:value={composeBody}
							rows={6}
							class="w-full px-3 py-2 text-sm border rounded-md focus:outline-none focus:ring-1 focus:ring-blue-500 {fieldErrors.body ? 'border-red-300' : 'border-gray-300'}"
							placeholder="Type your message here…"
						></textarea>
						{#if fieldErrors.body}
							<p class="text-red-600 text-xs mt-1">{fieldErrors.body}</p>
						{/if}
					</div>
					<div class="flex justify-end gap-3">
						<button
							type="button"
							onclick={closeCompose}
							class="px-4 py-2 text-sm text-gray-600 hover:bg-gray-100 rounded-md transition-colors"
						>
							Cancel
						</button>
						<button
							type="submit"
							disabled={sending}
							class="px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-md hover:bg-blue-700 transition-colors disabled:opacity-50"
						>
							{sending ? 'Sending…' : 'Send Message'}
						</button>
					</div>
				</form>
			{:else if showMessage}
				<div class="px-6 py-4 border-b border-gray-100 flex justify-between items-center">
					<h3 class="font-semibold text-gray-900">{showMessage.subject}</h3>
					<button onclick={closeMessage} class="text-sm text-gray-500 hover:text-gray-700">&larr; Back to messages</button>
				</div>
				<div class="p-6">
					<div class="flex justify-between items-center mb-4">
						<div>
							<p class="text-sm font-medium text-gray-900">{showMessage.sender}</p>
							<p class="text-xs text-gray-500">{formatDate(showMessage.message_time)}</p>
						</div>
						<span class="inline-flex px-2 py-0.5 text-xs font-medium rounded-full {urgencyColor(showMessage.urgency)}">
							{urgencyLabel(showMessage.urgency)}
						</span>
					</div>
					<div class="bg-gray-50 rounded-lg p-4 text-sm text-gray-700 whitespace-pre-wrap">
						{showMessage.body}
					</div>
				</div>
			{/if}
		</div>
	{:else if messages.length === 0}
		<EmptyState
			title="No messages"
			message="Send a secure message to your healthcare team."
			actionLabel="New Message"
			onAction={openCompose}
		/>
	{:else}
		<div class="bg-white rounded-lg shadow-sm border border-gray-200">
			<div class="divide-y divide-gray-100">
				{#each messages as msg (msg.id)}
					<button
						onclick={() => openMessage(msg)}
						class="w-full text-left px-6 py-4 hover:bg-gray-50 transition-colors flex items-start gap-4"
					>
						<div class="shrink-0 mt-0.5">
							{#if !msg.is_read}
								<span class="block w-2.5 h-2.5 bg-blue-500 rounded-full"></span>
							{:else}
								<span class="block w-2.5 h-2.5 bg-gray-300 rounded-full"></span>
							{/if}
						</div>
						<div class="min-w-0 flex-1">
							<div class="flex justify-between items-baseline mb-1">
								<span class="text-sm font-medium text-gray-900 truncate">{msg.subject}</span>
								<span class="text-xs text-gray-400 shrink-0 ml-2">{formatDate(msg.message_time)}</span>
							</div>
							<div class="flex items-center gap-2">
								<span class="text-xs text-gray-500">{msg.sender}</span>
								<span class="inline-flex px-1.5 py-0.5 text-xs font-medium rounded-full {urgencyColor(msg.urgency)}">
									{urgencyLabel(msg.urgency)}
								</span>
							</div>
						</div>
					</button>
				{/each}
			</div>
		</div>
	{/if}
</div>
