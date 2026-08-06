import { writable, get } from 'svelte/store';
import { browser } from '$app/environment';

const STORAGE_KEY = 'freemed_token';

function loadToken(): string | null {
	if (!browser) return null;
	return localStorage.getItem(STORAGE_KEY);
}

function saveToken(token: string | null) {
	if (!browser) return;
	if (token) {
		localStorage.setItem(STORAGE_KEY, token);
	} else {
		localStorage.removeItem(STORAGE_KEY);
	}
}

export const authToken = writable<string | null>(loadToken());
export const isAuthenticated = writable<boolean>(!!loadToken());

authToken.subscribe((token) => {
	saveToken(token);
	isAuthenticated.set(!!token);
});

export async function login(username: string, password: string): Promise<boolean> {
	const res = await fetch('/auth/login', {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ username, password }),
	});
	if (!res.ok) return false;
	const data = await res.json();
	authToken.set(data.token);
	return true;
}

export async function refreshToken(): Promise<boolean> {
	const token = get(authToken);
	if (!token) return false;
	const res = await fetch('/auth/refresh_token', {
		headers: { 'Authorization': `Bearer ${token}` },
	});
	if (!res.ok) return false;
	const data = await res.json();
	authToken.set(data.token);
	return true;
}

export function logout() {
	const token = get(authToken);
	authToken.set(null);
	if (token && browser) {
		fetch('/auth/logout', {
			method: 'DELETE',
			headers: { 'Authorization': `Bearer ${token}` },
		}).catch(() => {});
	}
}
