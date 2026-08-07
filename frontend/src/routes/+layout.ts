import type { LoadEvent } from '@sveltejs/kit';

// SPA fallback mode: no server-side auth checks.
// Auth redirect is handled client-side in +layout.svelte.
export const load = ({ url }: LoadEvent) => {
	return { pathname: url.pathname };
};
