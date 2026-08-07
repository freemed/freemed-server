import { browser } from '$app/environment';

// ---- Reactive auth state ----
let _authenticated = $state(false);
let _userType = $state<string | null>(null);
let _username = $state<string | null>(null);

// ---- Public API ----
export const auth = {
	get authenticated() { return _authenticated; },
	get userType() { return _userType; },
	get username() { return _username; },
};

export const isAuthenticated = () => _authenticated;
export const userType = () => _userType;
export const currentUsername = () => _username;

// ---- CSRF ----
function readCSRFCookie(): string | null {
	if (!browser) return null;
	const match = document.cookie.match(/(?:^|;\s*)csrf_token=([^;]*)/);
	return match ? match[1] : null;
}

export async function fetchCSRFToken(): Promise<string | null> {
	try {
		const res = await fetch('/auth/csrf');
		if (!res.ok) return null;
		const data = await res.json();
		return data.token || readCSRFCookie();
	} catch {
		return readCSRFCookie();
	}
}

// ---- Auth actions ----

export async function checkAuth(): Promise<boolean> {
	if (!browser) return false;
	try {
		const res = await fetch('/auth/me');
		if (!res.ok) {
			_authenticated = false;
			_userType = null;
			_username = null;
			return false;
		}
		const data = await res.json();
		_authenticated = true;
		_userType = data.user_type || null;
		_username = data.username || null;
		return true;
	} catch {
		_authenticated = false;
		_userType = null;
		_username = null;
		return false;
	}
}

export async function login(username: string, password: string): Promise<boolean> {
	const csrfToken = readCSRFCookie();
	const headers: Record<string, string> = { 'Content-Type': 'application/json' };
	if (csrfToken) {
		headers['X-CSRF-Token'] = csrfToken;
	}
	const res = await fetch('/auth/login', {
		method: 'POST',
		headers,
		body: JSON.stringify({ username, password }),
	});
	if (!res.ok) return false;

	// The JWT is now in an httpOnly cookie — verify by calling /auth/me
	const ok = await checkAuth();
	return ok;
}

export async function refreshToken(): Promise<boolean> {
	const res = await fetch('/auth/refresh_token');
	if (!res.ok) return false;
	return checkAuth();
}

export function logout() {
	if (browser) {
		fetch('/auth/logout', { method: 'DELETE' }).catch(() => {});
	}
	_authenticated = false;
	_userType = null;
	_username = null;
}
