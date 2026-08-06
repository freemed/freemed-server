<script lang="ts">
	import { page } from '$app/stores';
	let patientId = $state('');
	$effect(() => { patientId = $page.params.id; });
</script>

<div class="max-w-4xl mx-auto">
	<div class="flex items-center justify-between mb-6">
		<h1 class="text-2xl font-bold text-gray-900">Encounters</h1>
		<a href="/patients/{patientId}" class="text-sm text-blue-600 hover:text-blue-800">&larr; Back to Patient</a>
	</div>
	<div class="bg-white rounded-lg shadow-sm border border-gray-200 p-8 text-center text-gray-500">
		<p class="text-lg mb-2">Encounters module coming soon</p>
		<p class="text-sm">Patient encounter history and visit documentation will appear here.</p>
	</div>
</div>
