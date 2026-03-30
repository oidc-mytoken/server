import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [sveltekit()],
	server: {
		// Proxy API calls to Go backend during development
		proxy: {
			// API endpoints
			'/api': {
				target: 'http://localhost:8000',
				changeOrigin: true
			},
			// Discovery endpoints
			'/.well-known': {
				target: 'http://localhost:8000',
				changeOrigin: true
			},
			// Other backend endpoints
			'/redirect': {
				target: 'http://localhost:8000',
				changeOrigin: true
			},
			'/jwks': {
				target: 'http://localhost:8000',
				changeOrigin: true
			},
			'/c': {
				target: 'http://localhost:8000',
				changeOrigin: true
			},
			'/privacy': {
				target: 'http://localhost:8000',
				changeOrigin: true
			},
			'/calendars': {
				target: 'http://localhost:8000',
				changeOrigin: true
			},
			'/actions': {
				target: 'http://localhost:8000',
				changeOrigin: true
			},
			'/notifications': {
				target: 'http://localhost:8000',
				changeOrigin: true
			}
		}
	}
});
