<script lang="ts">
	interface ToastMessage {
		id: number;
		message: string;
		type: 'info' | 'success' | 'error' | 'warning';
	}

	let toasts = $state<ToastMessage[]>([]);
	let nextId = 0;

	function addToast(message: string, type: ToastMessage['type'] = 'info', duration = 4000) {
		const id = nextId++;
		toasts = [...toasts, { id, message, type }];
		setTimeout(() => {
			toasts = toasts.filter(t => t.id !== id);
		}, duration);
	}

	// Expose to window for global access (backward compat with toastr pattern)
	if (typeof window !== 'undefined') {
		(window as any).toast = {
			info: (msg: string) => addToast(msg, 'info'),
			success: (msg: string) => addToast(msg, 'success'),
			error: (msg: string) => addToast(msg, 'error'),
			warning: (msg: string) => addToast(msg, 'warning'),
		};
	}

	const typeStyles: Record<string, string> = {
		info: 'bg-blue-50 border-blue-200 text-blue-800',
		success: 'bg-green-50 border-green-200 text-green-800',
		error: 'bg-red-50 border-red-200 text-red-800',
		warning: 'bg-yellow-50 border-yellow-200 text-yellow-800',
	};
</script>

{#if toasts.length > 0}
	<div class="fixed bottom-4 right-4 z-50 flex flex-col gap-2">
		{#each toasts as toast (toast.id)}
			<div class="px-4 py-2 rounded-lg border shadow-lg text-sm {typeStyles[toast.type]} transition-all">
				{toast.message}
			</div>
		{/each}
	</div>
{/if}
