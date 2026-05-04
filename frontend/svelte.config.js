import adapter from '@sveltejs/adapter-static';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

/** @type {import('@sveltejs/kit').Config} */
const config = {
	preprocess: vitePreprocess(),

	kit: {
		adapter: adapter({
			pages: 'build',
			assets: 'build',
			fallback: 'index.html', // SPA fallback for client-side routing
			precompress: false,
			strict: true
		}),
		paths: {
			// Base path - can be configured if served from subdirectory
			base: ''
		},
		prerender: {
			handleHttpError: ({ path, message }) => {
				// Ignore missing favicon during prerender
				if (path === '/favicon.ico') {
					return;
				}
				throw new Error(message);
			}
		}
	}
};

export default config;
