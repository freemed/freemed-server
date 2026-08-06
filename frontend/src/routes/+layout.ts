import { redirect, type LoadEvent } from '@sveltejs/kit';
import { get } from 'svelte/store';
import { isAuthenticated } from '$lib/stores/auth';

export const load = ({ url }: LoadEvent) => {
	const publicPaths = ['/login'];
	if (!get(isAuthenticated) && !publicPaths.includes(url.pathname)) {
		throw redirect(307, '/login');
	}
};
