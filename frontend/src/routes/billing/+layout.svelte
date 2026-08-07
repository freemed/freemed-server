<script lang="ts">
	let { children } = $props();
</script>

<div class="max-w-7xl mx-auto">
	<div class="flex items-center gap-6 mb-6 border-b border-gray-200 pb-3">
		<h1 class="text-2xl font-bold text-gray-900">Billing</h1>
		<nav class="flex gap-1">
			<a href="/billing/claims" class="px-3 py-1.5 text-sm font-medium text-gray-600 hover:text-gray-900 rounded-md hover:bg-gray-100 transition-colors">Claims</a>
			<a href="/billing/ar" class="px-3 py-1.5 text-sm font-medium text-gray-600 hover:text-gray-900 rounded-md hover:bg-gray-100 transition-colors">AR Aging</a>
			<a href="/billing/remitt" class="px-3 py-1.5 text-sm font-medium text-gray-600 hover:text-gray-900 rounded-md hover:bg-gray-100 transition-colors">Remitt</a>
			<a href="/billing/superbills" class="px-3 py-1.5 text-sm font-medium text-gray-600 hover:text-gray-900 rounded-md hover:bg-gray-100 transition-colors">Superbills</a>
		</nav>
	</div>
	{@render children()}
</div>
