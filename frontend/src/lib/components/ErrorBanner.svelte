<script lang="ts">
	let {
		message = 'An error occurred.',
		onRetry = undefined as (() => void) | undefined,
	} = $props();
</script>

<div class="bg-red-50 border border-red-200 rounded-lg p-6 text-center">
	<svg class="w-10 h-10 text-red-400 mx-auto mb-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
		<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4.5c-.77-.833-2.694-.833-3.464 0L3.34 16.5c-.77.833.192 2.5 1.732 2.5z" />
	</svg>
	<p class="text-red-700 text-sm mb-3">{message}</p>
	{#if onRetry}
		<button
			onclick={onRetry}
			class="px-3 py-1.5 text-xs font-medium text-red-700 bg-red-100 rounded-md hover:bg-red-200 transition-colors"
		>
			Retry
		</button>
	{/if}
</div>
