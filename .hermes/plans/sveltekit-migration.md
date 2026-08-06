# SvelteKit Frontend Migration Plan

> **For Hermes:** Use subagent-driven-development skill to implement this plan phase-by-phase.

**Goal:** Replace jQuery + Knockout + Bootstrap 4 frontend with SvelteKit + Tailwind CSS,
matching the stack used in A Penny Worth and Concordia projects.

**Architecture:** SvelteKit SPA with client-side routing, `$state`/`$derived` reactivity
replacing Knockout `observable`/`computed`, SvelteKit load functions replacing jQuery AJAX,
FullCalendar via `@fullcalendar/svelte`, Tailwind CSS replacing Bootstrap 4.

**Tech Stack:** Svelte 5 (runes), SvelteKit 2, Vite, Tailwind CSS 4, TypeScript, FullCalendar 6, `svelte-select`

---

## Auth Audit — Issues Found

### Critical

| # | Issue | Severity | Fix |
|---|-------|----------|-----|
| 1 | **Password hashing: MD5** | CRITICAL | Replace with bcrypt (cost 12). MD5 is trivially reversible for short passwords |
| 2 | **Token in URL parameter** (`query:token`) | CRITICAL | Remove from `TokenLookup`. Tokens in URLs leak into logs, history, referrer headers |
| 3 | **Logout via GET** (`auth.GET("/logout")`) | CRITICAL | Remove the GET handler. Logout MUST be POST/DELETE only. GET logout is CSRF-trivial |
| 4 | **No token invalidation on logout** | CRITICAL | Add JWT token blacklist in Redis (store jti until natural expiry). Logout must invalidate the token |

### High

| # | Issue | Severity | Fix |
|---|-------|----------|-----|
| 5 | **Authorization bypass** (`return true`) | HIGH | Implement RBAC. At minimum check `user_type` claim against required roles per endpoint |
| 6 | **Token storage: localStorage** | HIGH | Move to httpOnly cookie or use SvelteKit server-side session. localStorage is XSS-readable |
| 7 | **Token refresh: fixed 2min polling** | HIGH | Replace with on-demand refresh. Check token expiry before each request; refresh if < 2min remaining |

### Medium

| # | Issue | Severity | Fix |
|---|-------|----------|-----|
| 8 | **No token rotation tracking** | MEDIUM | Old tokens remain valid after refresh. Track `iat` in Redis; reject tokens older than the latest issued |
| 9 | **Session dual-storage confusion** | MEDIUM | Both JWT (localStorage) and Redis session. Simplify: either pure JWT with blacklist, or opaque session tokens |
| 10 | **No CSRF on login endpoint** | MEDIUM | Add CSRF token requirement for POST /auth/login. SvelteKit has built-in CSRF protection via `csrf` in form actions |

### Auth Fixes — Implementation Order

These fixes must be done BEFORE the SvelteKit migration begins, since the migration will build on the corrected auth flow:

1. **Replace MD5 with bcrypt** — in `model/user.go:CheckUserPassword` and any password-setting code
2. **Remove `query:token` from TokenLookup** — `"header:Authorization,cookie:jwt"`
3. **Add Redis token blacklist** — store `jti` on logout until token natural expiry
4. **Implement basic RBAC** — check `user_type` claim; map to endpoint access
5. **Remove GET /auth/logout** — DELETE only
6. **Add CSRF protection to login** — gin middleware or SvelteKit server-side proxy

---

## Phase 0: Auth Fixes (Backend — Prerequisite)

### Task 0.1: Replace MD5 with bcrypt

**Files:** `model/user.go`, `cmd/freemed-server/auth.go`

```go
import "golang.org/x/crypto/bcrypt"

func CheckUserPassword(username, password string) (int64, bool) {
    u, err := Queries.GetUserByUsername(context.Background(), username)
    if err != nil {
        return 0, false
    }
    if err := bcrypt.CompareHashAndPassword([]byte(u.Userpassword), []byte(password)); err != nil {
        return 0, false
    }
    return u.ID, true
}
```

**Migration:** Existing MD5 hashes must be upgraded. On login with MD5 match, re-hash with bcrypt and update the DB row.

### Task 0.2: Token blacklist on logout

**Files:** `common/session.go` (new method), `cmd/freemed-server/auth.go`

Add `BlacklistToken(jti string, ttl time.Duration)` to Redis connector. On logout:
1. Extract JWT `jti` claim
2. Store `jti` in Redis with TTL = token expiry
3. On every authenticated request, check if `jti` is blacklisted

### Task 0.3: Remove query:token + GET logout

**Files:** `cmd/freemed-server/auth.go`, `cmd/freemed-server/main.go`

### Task 0.4: Basic RBAC middleware

**Files:** New `cmd/freemed-server/rbac.go`

Check `user_type` claim. Map types to API path prefixes.

---

## Phase 1: SvelteKit Project Scaffold

### Task 1.1: Initialize SvelteKit project

```bash
npx sv create frontend --template minimal --types ts
cd frontend
npx svelte-add tailwindcss
npm install @fullcalendar/core @fullcalendar/daygrid @fullcalendar/timegrid \
  @fullcalendar/interaction @fullcalendar/svelte svelte-select
```

**Architecture:**
```
frontend/
├── src/
│   ├── app.html              # Shell HTML
│   ├── app.css               # Tailwind imports
│   ├── lib/
│   │   ├── api.ts            # API client (replaces jQuery AJAX wrappers)
│   │   ├── auth.ts           # Token management (replaces sessionAuth/sessionRenewal)
│   │   ├── stores/
│   │   │   └── auth.svelte.ts # Auth state store
│   │   └── components/
│   │       ├── Navbar.svelte
│   │       ├── LoginModal.svelte
│   │       └── Toast.svelte
│   └── routes/
│       ├── +layout.svelte     # Root layout (navbar, auth check)
│       ├── +layout.ts         # Auth guard load function
│       ├── +page.svelte       # Main/dashboard
│       ├── patients/
│       │   ├── +page.svelte   # Patient search
│       │   ├── new/+page.svelte
│       │   └── [id]/
│       │       ├── +page.svelte
│       │       └── progress-notes/+page.svelte
│       ├── scheduler/
│       │   └── +page.svelte
│       ├── messages/
│       │   ├── +page.svelte
│       │   └── compose/+page.svelte
│       ├── preferences/
│       │   └── +page.svelte
│       ├── settings/
│       │   └── +page.svelte
│       └── login/+page.svelte
├── static/
│   └── img/ (logo, favicon)
├── svelte.config.js
├── tailwind.config.ts
└── vite.config.ts
```

### Task 1.2: Configure API proxy

**File:** `vite.config.ts`

```typescript
export default defineConfig({
  plugins: [sveltekit()],
  server: {
    proxy: {
      '/api': 'http://localhost:3000',
      '/auth': 'http://localhost:3000'
    }
  }
});
```

### Task 1.3: Create API client

**File:** `src/lib/api.ts`

Replaces `$.ApiGET`, `$.ApiPOST`, `$.ApiDELETE`, `sessionAuth`, `displayError`:

```typescript
import { get } from 'svelte/store';
import { authToken, logout } from './stores/auth.svelte';

const API_BASE = '/api';

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
    const token = get(authToken);
    const headers: Record<string, string> = {
        'Content-Type': 'application/json',
        ...((options.headers as Record<string, string>) || {}),
    };
    if (token) {
        headers['Authorization'] = `Bearer ${token}`;
    }
    const res = await fetch(`${API_BASE}${path}`, { ...options, headers });
    if (res.status === 401) {
        logout();
        throw new Error('Session expired');
    }
    if (!res.ok) {
        throw new Error(`API error: ${res.status}`);
    }
    return res.json();
}

export const api = {
    get: <T>(path: string) => request<T>(path),
    post: <T>(path: string, data: unknown) =>
        request<T>(path, { method: 'POST', body: JSON.stringify(data) }),
    del: <T>(path: string) => request<T>(path, { method: 'DELETE' }),
};
```

### Task 1.4: Create auth store

**File:** `src/lib/stores/auth.svelte.ts`

Replaces `sessionId`, `storeSessionId`, `authenticated()`, `login()`, `logout()`, `sessionRenewalProcess`:

```typescript
import { writable } from 'svelte/store';
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
    if (token) {
        // Fire-and-forget: tell server to invalidate
        fetch('/auth/logout', {
            method: 'DELETE',
            headers: { 'Authorization': `Bearer ${token}` },
        });
    }
}
```

### Task 1.5: Root layout + auth guard

**File:** `src/routes/+layout.svelte`

Layout with navbar, login modal, toast container. Replaces `index.html` body.

**File:** `src/routes/+layout.ts`

```typescript
import { redirect } from '@sveltejs/kit';
import { get } from 'svelte/store';
import { isAuthenticated } from '$lib/stores/auth.svelte';

export const load = ({ url }) => {
    if (!get(isAuthenticated) && url.pathname !== '/login') {
        throw redirect(307, '/login');
    }
};
```

---

## Phase 2: Page Migration (one page per subagent)

### Task 2.1: Login page

**Source:** `ui/fragment/login-splash.html` + inline login form in `index.html`
**Target:** `src/routes/login/+page.svelte`

Tailwind-styled login form. Calls `login()` from auth store. On success, redirects to `/`.

**Components:** `LoginModal.svelte` — modal dialog with username/password, error display, loading state.

### Task 2.2: Main page (dashboard)

**Source:** `ui/fragment/main.html`
**Target:** `src/routes/+page.svelte`

Dashboard with stats, quick links. Loads data via `api.get()`.

### Task 2.3: Patient search

**Source:** `ui/fragment/patient.search.html` (79 lines)
**Target:** `src/routes/patients/+page.svelte`

- Replaces Select2 AJAX picklist with `svelte-select` async mode
- Knockout `data-bind="foreach: patients"` → Svelte `{#each patients as p}`
- Knockout `ko.observableArray` → Svelte `$state([])`

### Task 2.4: Patient view/edit

**Source:** `ui/fragment/patient.html` (146 lines) + `ui/fragment/patient.edit.html` (43 lines)
**Target:** `src/routes/patients/[id]/+page.svelte`

- Knockout `data-bind="value: fieldName"` → Svelte `bind:value={fieldName}`
- Knockout computed observables → Svelte `$derived()`
- `koOptionsProvider` → `api.get()` in `onMount`

### Task 2.5: Scheduler

**Source:** `ui/fragment/scheduler.html` (170 lines)
**Target:** `src/routes/scheduler/+page.svelte`

FullCalendar 6 via `@fullcalendar/svelte`. The Svelte wrapper provides:
```svelte
<FullCalendar options={calendarOptions} />
```

All the `$(document).ready()`, `eventClick`, `eventDrop`, `eventResize`, `events` callbacks become Svelte component props and event handlers.

### Task 2.6: Messages

**Source:** `ui/fragment/messages.html` (51 lines) + `ui/fragment/messages.compose.html` (34 lines)
**Target:** `src/routes/messages/+page.svelte` + `src/routes/messages/compose/+page.svelte`

Messages list with unread indicator. Compose form with recipient picklist.

### Task 2.7: Preferences + Settings

**Source:** `ui/fragment/preferences.html` (9 lines) + `ui/fragment/settings.html` (96 lines)
**Target:** `src/routes/preferences/+page.svelte` + `src/routes/settings/+page.svelte`

### Task 2.8: Progress notes

**Source:** `ui/fragment/patient.progressnotes.html` (57 lines)
**Target:** `src/routes/patients/[id]/progress-notes/+page.svelte`

---

## Phase 3: Shared Components

### Task 3.1: Navbar

**File:** `src/lib/components/Navbar.svelte`

Replaces the Bootstrap 4 navbar in `index.html`. Active state via `$page.url.pathname`.

### Task 3.2: Toast notifications

**File:** `src/lib/components/Toast.svelte`

Replaces Toastr. Simple store-based toast queue with Tailwind-styled toasts.

### Task 3.3: FullCalendar wrapper

**File:** `src/lib/components/Calendar.svelte`

Thin wrapper around `@fullcalendar/svelte` with FreeMED-specific defaults (bootstrap theme → Tailwind, business hours, event handlers).

---

## Phase 4: Polish & Deploy

### Task 4.1: Remove old frontend

Delete `ui/` directory. Update Go server to serve SvelteKit build output instead.

### Task 4.2: Dockerize frontend

Add nginx + SvelteKit static build to Docker Compose (matching A Penny Worth pattern: nginx serving SPA, proxying /api to Go backend).

### Task 4.3: Responsive audit

Verify mobile layout. Bootstrap 4 was responsive; Tailwind equivalent must match.

---

## Migration Order Summary

```
Phase 0: Auth fixes (backend)                  [1.0 day]
Phase 1: SvelteKit scaffold + API client       [1.0 day]
Phase 2: Page migration (8 pages, parallel)    [3.0 days]
Phase 3: Shared components                     [1.0 day]
Phase 4: Polish, deploy, remove old frontend   [1.0 day]
                                        Total: ~7.0 days
```

## Component Mapping

| Current | SvelteKit Equivalent |
|---------|---------------------|
| `$.ApiGET/POST/DELETE` | `api.get/post/del()` in `$lib/api.ts` |
| `sessionAuth` (beforeSend) | `Authorization` header in `api.ts` |
| `storeSessionId` + `$.sessionStorage` | `authToken` writable store + localStorage |
| `loginStateChange` | Reactive `$isAuthenticated` + Svelte `$effect` |
| `sessionRenewalProcess` (setInterval) | On-demand refresh in `api.ts` before requests |
| `loadPage(id)` | SvelteKit client-side routing (`goto()`) |
| `ko.cleanNode` + jQuery `.load()` | SvelteKit `+page.svelte` (auto-cleanup) |
| `ko.observable` / `ko.observableArray` | Svelte 5 `$state()` |
| `ko.computed` | Svelte 5 `$derived()` |
| `data-bind="text: x"` | `{x}` |
| `data-bind="value: x"` | `bind:value={x}` |
| `data-bind="foreach: items"` | `{#each items as item}` |
| `data-bind="visible: x"` | `{#if x}` |
| `ko.applyBindings(vm, node)` | Svelte auto-binds; no manual binding |
| Select2 AJAX picklist | `svelte-select` with `loadOptions` |
| Toastr | Custom `Toast.svelte` store |
| Tablesorter | Svelte sortable table or `@tanstack/svelte-table` |
| Bootstrap 4 classes | Tailwind CSS utility classes |
| `#mainFrame` div injection | SvelteKit `{@render children()}` in layout |
