import { browser } from '$app/environment';

// ---- Reactive portal auth state ----
let _authenticated = $state(false);
let _patientId = $state<string | null>(null);
let _firstName = $state<string | null>(null);
let _lastName = $state<string | null>(null);
let _dateOfBirth = $state<string | null>(null);

// ---- Public API ----
export const portalAuth = {
	get authenticated() { return _authenticated; },
	get patientId() { return _patientId; },
	get firstName() { return _firstName; },
	get lastName() { return _lastName; },
	get dateOfBirth() { return _dateOfBirth; },
	get displayName() {
		return [_firstName, _lastName].filter(Boolean).join(' ') || 'Patient';
	},
};

// ---- Auth actions ----

export async function checkAuth(): Promise<boolean> {
	if (!browser) return false;
	try {
		const res = await fetch('/portal/auth/me');
		if (!res.ok) {
			_authenticated = false;
			_patientId = null;
			_firstName = null;
			_lastName = null;
			_dateOfBirth = null;
			return false;
		}
		const data = await res.json();
		_authenticated = true;
		_patientId = data.patient_id || null;
		_firstName = data.first_name || null;
		_lastName = data.last_name || null;
		_dateOfBirth = data.date_of_birth || null;
		return true;
	} catch {
		_authenticated = false;
		_patientId = null;
		_firstName = null;
		_lastName = null;
		_dateOfBirth = null;
		return false;
	}
}

export async function login(patientId: string, dateOfBirth: string, pin: string): Promise<boolean> {
	const res = await fetch('/portal/auth/login', {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ patient_id: patientId, date_of_birth: dateOfBirth, pin }),
	});
	if (!res.ok) return false;

	// JWT now in httpOnly cookie — verify by calling /portal/auth/me
	const ok = await checkAuth();
	return ok;
}

export function logout() {
	if (browser) {
		fetch('/portal/auth/logout', { method: 'DELETE' }).catch(() => {});
	}
	_authenticated = false;
	_patientId = null;
	_firstName = null;
	_lastName = null;
	_dateOfBirth = null;
}
