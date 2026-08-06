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

// Module-level reactive state using Svelte 5 runes
let _token = $state<string | null>(loadToken());

export const authToken = {
	get current() { return _token; },
	set current(val: string | null) {
		_token = val;
		saveToken(val);
	}
};

export const isAuthenticated = () => !!_token;

export async function login(username: string, password: string): Promise<boolean> {
	const res = await fetch('/auth/login', {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ username, password }),
	});
	if (!res.ok) return false;
	const data = await res.json();
	authToken.current = data.token;
	return true;
}

export async function refreshToken(): Promise<boolean> {
	if (!_token) return false;
	const res = await fetch('/auth/refresh_token', {
		headers: { Authorization: `Bearer ${_token}` },
	});
	if (!res.ok) return false;
	const data = await res.json();
	authToken.current = data.token;
	return true;
}

export function logout() {
	const token = _token;
	authToken.current = null;
	if (token && browser) {
		fetch('/auth/logout', {
			method: 'DELETE',
			headers: { Authorization: `Bearer ${token}` },
		}).catch(() => {});
	}
}
