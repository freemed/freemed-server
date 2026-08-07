<script lang="ts">
	let { message = 'Loading…' } = $props();
</script>

<div class="d-flex flex-column align-items-center justify-content-center py-5">
	<div class="spinner-border text-primary mb-3" role="status">
		<span class="visually-hidden">Loading…</span>
	</div>
	<p class="text-muted small">{message}</p>
</div>
