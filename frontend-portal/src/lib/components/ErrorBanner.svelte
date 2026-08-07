<script lang="ts">
	let {
		message = 'An error occurred.',
		onRetry = undefined as (() => void) | undefined,
	} = $props();
</script>

<div class="alert alert-danger d-flex align-items-center justify-content-between" role="alert">
	<div class="d-flex align-items-center">
		<svg class="me-2 flex-shrink-0" width="20" height="20" fill="none" stroke="currentColor" viewBox="0 0 24 24">
			<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4.5c-.77-.833-2.694-.833-3.464 0L3.34 16.5c-.77.833.192 2.5 1.732 2.5z" />
		</svg>
		<span>{message}</span>
	</div>
	{#if onRetry}
		<button class="btn btn-outline-danger btn-sm" onclick={onRetry}>
			Retry
		</button>
	{/if}
</div>
