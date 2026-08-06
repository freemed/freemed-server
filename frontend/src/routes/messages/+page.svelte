<script lang="ts">
	import { goto } from '$app/navigation';
	import { api } from '$lib/api';

	interface Message {
		id: number;
		msgby: string;
		sender: string;
		msgtime: string;
		msgfor: string;
		msgpatient: number | null;
		msgperson: string;
		msgurgency: string;
		msgsubject: string;
		msgtext: string;
		msgread: number;
		msgunique: string;
		msgtag: string;
	}

	const Tab = {
		All: 'all',
		Unread: 'unread',
	} as const;

	type Tab = (typeof Tab)[keyof typeof Tab];

	let messages = $state<Message[]>([]);
	let loading = $state(true);
	let error = $state('');
	let activeTab = $state<Tab>(Tab.All);
	let expandedId = $state<number | null>(null);

	async function loadMessages() {
		loading = true;
		error = '';
		try {
			const data = await api.get<Message[]>('/messages/view');
			messages = data || [];
		} catch (e: any) {
			error = e.message || 'Failed to load messages';
			messages = [];
		} finally {
			loading = false;
		}
	}

	const filteredMessages = $derived(
		activeTab === Tab.Unread
			? messages.filter((m) => !m.msgread)
			: messages,
	);

	function toggleExpand(id: number) {
		expandedId = expandedId === id ? null : id;
	}

	function formatTime(ts: string): string {
		if (!ts) return '';
		try {
			const d = new Date(ts);
			if (isNaN(d.getTime())) return ts;
			return d.toLocaleString('en-US', {
				month: 'short',
				day: 'numeric',
				year: 'numeric',
				hour: 'numeric',
				minute: '2-digit',
			});
		} catch {
			return ts;
		}
	}

	function urgencyClass(urgency: string): string {
		switch (urgency?.toLowerCase()) {
			case 'high':
			case 'urgent':
				return 'text-red-600 font-semibold';
			case 'medium':
				return 'text-yellow-600';
			default:
				return 'text-gray-500';
		}
	}

	loadMessages();
</script>

<div class="max-w-5xl mx-auto">
	<div class="flex items-center justify-between mb-6">
		<h1 class="text-2xl font-bold text-gray-900">Messages</h1>
		<button
			onclick={() => goto('/messages/compose')}
			class="px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded-lg hover:bg-blue-700 transition-colors"
		>
			+ Compose
		</button>
	</div>

	<!-- Tabs -->
	<div class="flex gap-1 mb-4 border-b border-gray-200">
		<button
			onclick={() => (activeTab = Tab.All)}
			class="px-4 py-2 text-sm font-medium rounded-t-lg transition-colors {activeTab === Tab.All
				? 'border-b-2 border-blue-600 text-blue-600'
				: 'text-gray-500 hover:text-gray-700'}"
		>
			All
		</button>
		<button
			onclick={() => (activeTab = Tab.Unread)}
			class="px-4 py-2 text-sm font-medium rounded-t-lg transition-colors {activeTab === Tab.Unread
				? 'border-b-2 border-blue-600 text-blue-600'
				: 'text-gray-500 hover:text-gray-700'}"
		>
			Unread
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
	{:else if filteredMessages.length === 0}
		<div class="text-center py-12 text-gray-500">
			<p class="text-lg">No messages</p>
			<p class="text-sm mt-1">
				{activeTab === Tab.Unread ? 'No unread messages' : 'Your inbox is empty'}
			</p>
		</div>
	{:else}
		<div class="space-y-2">
			{#each filteredMessages as msg (msg.id)}
				<div class="bg-white rounded-lg border border-gray-200 overflow-hidden transition-shadow hover:shadow-sm">
					<!-- Summary row -->
					<button
						onclick={() => toggleExpand(msg.id)}
						class="w-full text-left px-4 py-3 flex items-center gap-4"
					>
						<!-- Read/unread dot -->
						<span
							class="flex-shrink-0 w-2.5 h-2.5 rounded-full"
							class:bg-blue-500={!msg.msgread}
							class:bg-gray-300={!!msg.msgread}
						></span>

						<div class="flex-1 min-w-0">
							<div class="flex items-center gap-2 mb-0.5">
								<span
									class="text-sm font-medium text-gray-900 truncate"
									class:font-bold={!msg.msgread}
								>
									{msg.msgsubject || '(no subject)'}
								</span>
								{#if msg.msgurgency}
									<span class="text-xs {urgencyClass(msg.msgurgency)}">
										{msg.msgurgency}
									</span>
								{/if}
							</div>
							<div class="flex items-center gap-3 text-xs text-gray-500">
								<span>{msg.sender || msg.msgby}</span>
								{#if msg.msgpatient}
									<span>Patient #{msg.msgpatient}</span>
								{/if}
								<span>{formatTime(msg.msgtime)}</span>
							</div>
						</div>

						<!-- Expand chevron -->
						<svg
							class="w-4 h-4 text-gray-400 flex-shrink-0 transition-transform {expandedId === msg.id ? 'rotate-180' : ''}"
							fill="none"
							stroke="currentColor"
							viewBox="0 0 24 24"
						>
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
						</svg>
					</button>

					<!-- Expanded detail -->
					{#if expandedId === msg.id}
						<div class="px-4 pb-4 pt-0 border-t border-gray-100">
							<div class="pt-3 space-y-2">
								<div class="flex flex-wrap gap-x-6 gap-y-1 text-xs text-gray-500">
									<span><strong>From:</strong> {msg.sender || msg.msgby}</span>
									<span><strong>To:</strong> {msg.msgfor}</span>
									{#if msg.msgperson}
										<span><strong>Regarding:</strong> {msg.msgperson}</span>
									{/if}
									{#if msg.msgtag}
										<span
											class="inline-block px-2 py-0.5 bg-gray-100 rounded-full text-xs text-gray-600"
										>
											{msg.msgtag}
										</span>
									{/if}
								</div>
								<div class="text-sm text-gray-700 whitespace-pre-wrap leading-relaxed">
									{msg.msgtext}
								</div>
							</div>
						</div>
					{/if}
				</div>
			{/each}
		</div>
	{/if}
</div>
