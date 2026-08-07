<script lang="ts">
	import { toast } from '$lib/stores/toast.svelte';

	const typeStyles: Record<string, string> = {
		info: 'bg-blue-50 border-blue-200 text-blue-800',
		success: 'bg-green-50 border-green-200 text-green-800',
		error: 'bg-red-50 border-red-200 text-red-800',
		warning: 'bg-yellow-50 border-yellow-200 text-yellow-800',
	};
</script>

{#if toast.toasts.length > 0}
	<div class="fixed bottom-4 right-4 z-50 flex flex-col gap-2">
		{#each toast.toasts as t (t.id)}
			<div class="px-4 py-2 rounded-lg border shadow-lg text-sm {typeStyles[t.type]} transition-all">
				{t.message}
			</div>
		{/each}
	</div>
{/if}
