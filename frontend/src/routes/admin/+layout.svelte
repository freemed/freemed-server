<script lang="ts">
	let { children } = $props();
</script>

<div class="max-w-7xl mx-auto">
	<div class="flex items-center gap-4 mb-6 border-b border-gray-200 pb-3">
		<h1 class="text-2xl font-bold text-gray-900">Administration</h1>
		<nav class="flex gap-1">
			<a href="/admin/users" class="px-3 py-1.5 text-sm font-medium text-gray-600 hover:text-gray-900 rounded-md hover:bg-gray-100">Users</a>
		</nav>
	</div>
	{@render children()}
</div>
