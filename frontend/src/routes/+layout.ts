import { redirect, type LoadEvent } from '@sveltejs/kit';
import { authToken } from '$lib/stores/auth.svelte';

export const load = ({ url }: LoadEvent) => {
	const publicPaths = ['/login'];
	if (!authToken.current && !publicPaths.includes(url.pathname)) {
		throw redirect(307, '/login');
	}
};
