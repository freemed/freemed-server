import { redirect, type LoadEvent } from '@sveltejs/kit';
import { isAuthenticated } from '$lib/stores/auth.svelte';

export const load = ({ url }: LoadEvent) => {
	const publicPaths = ['/login'];
	if (!isAuthenticated && !publicPaths.includes(url.pathname)) {
		throw redirect(307, '/login');
	}
};
