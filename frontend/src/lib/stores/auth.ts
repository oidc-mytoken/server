import {derived, writable} from 'svelte/store';
import {browser} from '$app/environment';
import type {TokenInfoResponse} from '$lib/types';

const ISSUER_STORAGE_KEY = 'mytoken_oidc_issuer';
const SCOPES_STORAGE_KEY = 'mytoken_scopes';

interface AuthState {
	isLoggedIn: boolean;
	tokenInfo: TokenInfoResponse | null;
	oidcIssuer: string | null;
	scopes: string[];
	loading: boolean;
	initialized: boolean;  // True after first checkLogin completes
	error: string | null;
}

function createAuthStore() {
	const { subscribe, set, update } = writable<AuthState>({
		isLoggedIn: false,
		tokenInfo: null,
		oidcIssuer: null,
		scopes: [],
		loading: false,
		initialized: false,
		error: null
	});

	return {
		subscribe,

		/**
		 * Check if logged in via cookie by calling introspect
		 * This mirrors the Mustache frontend's checkIfLoggedIn behavior
		 */
		async checkLogin(tokeninfoEndpoint: string): Promise<boolean> {
			update((s) => ({ ...s, loading: true }));

			try {
				const response = await fetch(tokeninfoEndpoint, {
					method: 'POST',
					headers: {
						'Content-Type': 'application/json'
					},
					credentials: 'include',
					body: JSON.stringify({
						action: 'introspect'
					})
				});

				if (!response.ok) {
					throw new Error('Not logged in');
				}

				const data = await response.json();
				const tokenInfo: TokenInfoResponse = data;
				const token = data.token;

				// Extract issuer and scopes from token info
				const issuer = token?.oidc_iss ?? null;
				const scopes = extractScopesFromToken(token);

				// Save to sessionStorage for quick access
				if (browser && issuer) {
					sessionStorage.setItem(ISSUER_STORAGE_KEY, issuer);
					sessionStorage.setItem(SCOPES_STORAGE_KEY, JSON.stringify(scopes));
				}

				update((s) => ({
					...s,
					isLoggedIn: true,
					tokenInfo,
					oidcIssuer: issuer,
					scopes,
					loading: false,
					initialized: true,
					error: null
				}));

				return true;
			} catch {
				// Not logged in
				update((s) => ({
					...s,
					isLoggedIn: false,
					tokenInfo: null,
					loading: false,
					initialized: true
				}));
				return false;
			}
		},

		/**
		 * Set logged in state with token info
		 */
		setLoggedIn(tokenInfo: TokenInfoResponse, issuer?: string, scopes?: string[]) {
			if (browser) {
				if (issuer) {
					sessionStorage.setItem(ISSUER_STORAGE_KEY, issuer);
				}
				if (scopes) {
					sessionStorage.setItem(SCOPES_STORAGE_KEY, JSON.stringify(scopes));
				}
			}

			update((state) => ({
				...state,
				isLoggedIn: true,
				tokenInfo,
				oidcIssuer: issuer ?? state.oidcIssuer,
				scopes: scopes ?? state.scopes,
				error: null
			}));
		},

		/**
		 * Load cached issuer/scopes from storage (for quick UI updates)
		 */
		loadFromStorage() {
			if (!browser) return;

			try {
				const issuer = sessionStorage.getItem(ISSUER_STORAGE_KEY);
				const scopes = sessionStorage.getItem(SCOPES_STORAGE_KEY);

				if (issuer) {
					update((state) => ({
						...state,
						oidcIssuer: issuer,
						scopes: scopes ? JSON.parse(scopes) : []
					}));
				}
			} catch (e) {
				console.warn('Failed to load auth from storage:', e);
			}
		},

		/**
		 * Clear auth state and storage
		 */
		logout() {
			if (browser) {
				sessionStorage.removeItem(ISSUER_STORAGE_KEY);
				sessionStorage.removeItem(SCOPES_STORAGE_KEY);
			}

			set({
				isLoggedIn: false,
				tokenInfo: null,
				oidcIssuer: null,
				scopes: [],
				loading: false,
				initialized: true,  // Keep initialized true - we know we're logged out
				error: null
			});
		},

		/**
		 * Set loading state
		 */
		setLoading(loading: boolean) {
			update((state) => ({ ...state, loading }));
		},

		/**
		 * Set error state
		 */
		setError(error: string | null) {
			update((state) => ({ ...state, error, loading: false }));
		}
	};
}

/**
 * Extract maximum scopes from token restrictions
 */
function extractScopesFromToken(token: any): string[] {
	if (!token?.restrictions) return [];

	const allScopes = new Set<string>();
	for (const restriction of token.restrictions) {
		if (restriction.scope) {
			const scopes = restriction.scope.split(' ');
			scopes.forEach((s: string) => allScopes.add(s));
		}
	}
	return Array.from(allScopes);
}

export const auth = createAuthStore();

// Derived stores
export const isLoggedIn = derived(auth, ($auth) => $auth.isLoggedIn);
export const authInitialized = derived(auth, ($auth) => $auth.initialized);
export const tokenInfo = derived(auth, ($auth) => $auth.tokenInfo);
export const oidcIssuer = derived(auth, ($auth) => $auth.oidcIssuer);
