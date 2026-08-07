<script lang="ts">
	let {
		title = 'No data found',
		message = '',
		actionLabel = '',
		onAction = undefined as (() => void) | undefined,
	} = $props();
</script>

<div class="text-center py-5">
	<svg class="text-muted mb-3" width="48" height="48" fill="none" stroke="currentColor" viewBox="0 0 24 24">
		<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-2.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4" />
	</svg>
	<p class="text-muted">{title}</p>
	{#if message}
		<p class="text-muted small mt-1">{message}</p>
	{/if}
	{#if actionLabel && onAction}
		<button class="btn btn-outline-primary btn-sm mt-2" onclick={onAction}>
			{actionLabel}
		</button>
	{/if}
</div>
