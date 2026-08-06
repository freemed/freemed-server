<script lang="ts">
	import { page } from '$app/stores';
	import { api } from '$lib/api';

	interface ProgressNote {
		id: number;
		note_date: string;
		note_text: string;
		author?: string;
		note_type?: string;
	}

	let notes = $state<ProgressNote[]>([]);
	let loading = $state(true);
	let error = $state('');
	let patientId = $state('');

	$effect(() => {
		const id = $page.params.id;
		if (id) {
			patientId = id;
			loadNotes(id);
		}
	});

	async function loadNotes(id: string) {
		loading = true;
		error = '';
		try {
			const data = await api.get<ProgressNote[]>(`/patients/${id}/progress-notes`);
			notes = data || [];
		} catch (e: any) {
			error = e.message || 'Failed to load progress notes';
		} finally {
			loading = false;
		}
	}

	function formatDate(dateStr: string): string {
		if (!dateStr) return '';
		try {
			const d = new Date(dateStr);
			if (isNaN(d.getTime())) return dateStr;
			return d.toLocaleDateString('en-US', {
				year: 'numeric',
				month: 'long',
				day: 'numeric',
				hour: 'numeric',
				minute: '2-digit',
			});
		} catch {
			return dateStr;
		}
	}

	function truncateText(text: string, maxLen: number = 200): string {
		if (!text) return '';
		if (text.length <= maxLen) return text;
		return text.slice(0, maxLen).trimEnd() + '...';
	}
</script>

<div class="max-w-4xl mx-auto">
	<div class="flex items-center justify-between mb-6">
		<div>
			<h1 class="text-2xl font-bold text-gray-900">Progress Notes</h1>
			<p class="text-sm text-gray-500 mt-1">Patient #{patientId}</p>
		</div>
		<a
			href="/patients/{patientId}"
			class="text-sm text-blue-600 hover:text-blue-800 font-medium transition-colors"
		>
			&larr; Back to Patient
		</a>
	</div>

	{#if loading}
		<div class="flex justify-center py-12">
			<div class="w-6 h-6 border-2 border-blue-500 border-t-transparent rounded-full animate-spin"></div>
		</div>
	{:else if error}
		<div class="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg text-sm mb-6">
			{error}
		</div>
	{:else if notes.length === 0}
		<div class="text-center py-12 text-gray-500">
			<p class="text-lg">No progress notes found</p>
			<p class="text-sm mt-1">No notes have been recorded for this patient yet.</p>
		</div>
	{:else}
		<!-- Timeline -->
		<div class="relative">
			<!-- Vertical line -->
			<div class="absolute left-6 top-0 bottom-0 w-0.5 bg-gray-200"></div>

			<div class="space-y-6">
				{#each notes as note, i}
					<div class="relative pl-16">
						<!-- Timeline dot -->
						<div
							class="absolute left-4 top-5 w-4 h-4 rounded-full border-2 {i === 0 ? 'bg-blue-500 border-blue-500' : 'bg-white border-gray-300'} z-10"
						></div>

						<!-- Note card -->
						<div class="bg-white rounded-lg shadow-sm border border-gray-200 p-4">
							<div class="flex items-center justify-between mb-2">
								<div class="flex items-center gap-2">
									{#if note.note_type}
										<span class="inline-block px-2 py-0.5 text-xs font-medium bg-blue-50 text-blue-700 rounded">
											{note.note_type}
										</span>
									{/if}
									<span class="text-sm text-gray-500">
										{formatDate(note.note_date)}
									</span>
								</div>
								{#if note.author}
									<span class="text-xs text-gray-400">by {note.author}</span>
								{/if}
							</div>
							<p class="text-sm text-gray-700 whitespace-pre-wrap leading-relaxed">
								{truncateText(note.note_text)}
							</p>
							{#if note.note_text.length > 200}
								<button
									onclick={(e) => {
										const target = e.currentTarget;
										const paragraph = target.previousElementSibling as HTMLElement;
										if (paragraph) {
											const fullText = note.note_text;
											if (paragraph.textContent?.endsWith('...')) {
												paragraph.textContent = fullText;
												target.textContent = 'Show less';
											} else {
												paragraph.textContent = truncateText(fullText);
												target.textContent = 'Read more';
											}
										}
									}}
									class="mt-1 text-xs text-blue-600 hover:text-blue-800 font-medium"
								>
									Read more
								</button>
							{/if}
						</div>
					</div>
				{/each}
			</div>
		</div>
	{/if}
</div>
