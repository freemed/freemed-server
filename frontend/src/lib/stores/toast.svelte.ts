export interface ToastMessage {
	id: number;
	message: string;
	type: 'info' | 'success' | 'error' | 'warning';
}

// ---- Reactive state ----
let _toasts = $state<ToastMessage[]>([]);
let _nextId = 0;

function addToast(message: string, type: ToastMessage['type'] = 'info', duration = 4000) {
	const id = _nextId++;
	_toasts = [..._toasts, { id, message, type }];
	setTimeout(() => {
		_toasts = _toasts.filter((t) => t.id !== id);
	}, duration);
}

// ---- Public API ----
export const toast = {
	get toasts() {
		return _toasts;
	},
	info: (msg: string) => addToast(msg, 'info'),
	success: (msg: string) => addToast(msg, 'success'),
	error: (msg: string) => addToast(msg, 'error'),
	warning: (msg: string) => addToast(msg, 'warning'),
};
