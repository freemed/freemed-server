import { browser } from '$app/environment';
import { logout } from './stores/auth.svelte';

const API_BASE = '/api';

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
	if (!browser) throw new Error('API calls only available in browser');

	const headers: Record<string, string> = {
		'Content-Type': 'application/json',
		...((options.headers as Record<string, string>) || {}),
	};
	// JWT is in httpOnly cookie — browser sends it automatically, no Authorization header needed.
	const res = await fetch(`${API_BASE}${path}`, { ...options, headers });
	if (res.status === 401) {
		logout();
		throw new Error('Session expired');
	}
	if (!res.ok) {
		const body = await res.text();
		throw new Error(`API error ${res.status}: ${body}`);
	}
	return res.json();
}

export const api = {
	get: <T = unknown>(path: string) => request<T>(path),
	post: <T = unknown>(path: string, data: unknown) =>
		request<T>(path, { method: 'POST', body: JSON.stringify(data) }),
	put: <T = unknown>(path: string, data: unknown) =>
		request<T>(path, { method: 'PUT', body: JSON.stringify(data) }),
	del: <T = unknown>(path: string) => request<T>(path, { method: 'DELETE' }),
};
