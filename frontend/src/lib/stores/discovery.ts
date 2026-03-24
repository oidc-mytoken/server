import { writable, derived, get } from 'svelte/store';
import { browser } from '$app/environment';
import type { DiscoveryDocument, Provider } from '$lib/types';

const STORAGE_KEY = 'mytoken_discovery';
const CACHE_DURATION = 5 * 60 * 1000; // 5 minutes

interface DiscoveryState {
	data: DiscoveryDocument | null;
	loading: boolean;
	error: string | null;
	lastFetched: number | null;
}

function createDiscoveryStore() {
	const { subscribe, set, update } = writable<DiscoveryState>({
		data: null,
		loading: false,
		error: null,
		lastFetched: null
	});

	return {
		subscribe,

		/**
		 * Load discovery document from cache if available and not stale
		 */
		loadFromCache(): boolean {
			if (!browser) return false;

			try {
				const cached = sessionStorage.getItem(STORAGE_KEY);
				if (cached) {
					const { data, timestamp } = JSON.parse(cached);
					const isStale = Date.now() - timestamp > CACHE_DURATION;

					if (!isStale && data) {
						set({
							data,
							loading: false,
							error: null,
							lastFetched: timestamp
						});
						return true;
					}
				}
			} catch (e) {
				console.warn('Failed to load discovery from cache:', e);
			}
			return false;
		},

		/**
		 * Fetch discovery document from server
		 */
		async fetch(): Promise<DiscoveryDocument | null> {
			update((state) => ({ ...state, loading: true, error: null }));

			try {
				const response = await fetch('/.well-known/mytoken-configuration');
				if (!response.ok) {
					throw new Error(`Failed to fetch discovery: ${response.status}`);
				}

				const data: DiscoveryDocument = await response.json();
				const timestamp = Date.now();

				// Save to cache
				if (browser) {
					try {
						sessionStorage.setItem(
							STORAGE_KEY,
							JSON.stringify({ data, timestamp })
						);
					} catch (e) {
						console.warn('Failed to cache discovery:', e);
					}
				}

				set({
					data,
					loading: false,
					error: null,
					lastFetched: timestamp
				});

				return data;
			} catch (e) {
				const errorMessage = e instanceof Error ? e.message : 'Unknown error';
				update((state) => ({
					...state,
					loading: false,
					error: errorMessage
				}));
				return null;
			}
		},

		/**
		 * Get a specific endpoint URL
		 */
		getEndpoint(name: keyof DiscoveryDocument): string | null {
			const state = get({ subscribe });
			if (!state.data) return null;
			const value = state.data[name];
			return typeof value === 'string' ? value : null;
		},

		/**
		 * Clear the store and cache
		 */
		clear() {
			if (browser) {
				sessionStorage.removeItem(STORAGE_KEY);
			}
			set({
				data: null,
				loading: false,
				error: null,
				lastFetched: null
			});
		}
	};
}

export const discovery = createDiscoveryStore();

// Derived stores for common values
export const providers = derived(
	discovery,
	($discovery) => $discovery.data?.providers_supported ?? []
);

export const mytokenEndpoint = derived(
	discovery,
	($discovery) => $discovery.data?.mytoken_endpoint ?? null
);

export const tokeninfoEndpoint = derived(
	discovery,
	($discovery) => $discovery.data?.tokeninfo_endpoint ?? null
);

export const notificationsEndpoint = derived(
	discovery,
	($discovery) => $discovery.data?.notifications_endpoint ?? null
);

export const usersettingsEndpoint = derived(
	discovery,
	($discovery) => $discovery.data?.usersettings_endpoint ?? null
);
