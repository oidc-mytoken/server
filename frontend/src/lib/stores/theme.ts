import {derived, get, writable} from 'svelte/store';
import {browser} from '$app/environment';

export type ThemePreference = 'light' | 'dark' | 'auto';
export type EffectiveTheme = 'light' | 'dark';

const STORAGE_KEY = 'mytoken-theme';

/**
 * Get the system's preferred color scheme
 */
function getSystemTheme(): EffectiveTheme {
    if (!browser) return 'light';
    return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
}

/**
 * Load theme preference from localStorage
 */
function loadThemePreference(): ThemePreference {
    if (!browser) return 'auto';
    const stored = localStorage.getItem(STORAGE_KEY);
    if (stored === 'light' || stored === 'dark' || stored === 'auto') {
        return stored;
    }
    return 'auto';
}

/**
 * Save theme preference to localStorage
 */
function saveThemePreference(theme: ThemePreference): void {
    if (!browser) return;
    localStorage.setItem(STORAGE_KEY, theme);
}

/**
 * Apply theme to document
 */
function applyTheme(theme: EffectiveTheme): void {
    if (!browser) return;
    document.documentElement.setAttribute('data-bs-theme', theme);
}

// Store for the system's current theme preference (reactive to OS changes)
const systemTheme = writable<EffectiveTheme>(getSystemTheme());

// Initialize system theme listener
if (browser) {
    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
    const handleChange = (e: MediaQueryListEvent) => {
        systemTheme.set(e.matches ? 'dark' : 'light');
    };
    mediaQuery.addEventListener('change', handleChange);
}

// Store for user's theme preference
const themePreference = writable<ThemePreference>(loadThemePreference());

// Derived store for the effective theme (what's actually applied)
export const effectiveTheme = derived(
    [themePreference, systemTheme],
    ([$preference, $system]) => {
        if ($preference === 'auto') {
            return $system;
        }
        return $preference;
    }
);

// Subscribe to effective theme changes and apply to document
if (browser) {
    effectiveTheme.subscribe((theme) => {
        applyTheme(theme);
    });
}

/**
 * Theme store with methods to get/set preference
 */
function createThemeStore() {
    const {subscribe} = themePreference;

    return {
        subscribe,

        /**
         * Set the theme preference
         */
        set(preference: ThemePreference) {
            themePreference.set(preference);
            saveThemePreference(preference);
        },

        /**
         * Get the current preference
         */
        get(): ThemePreference {
            return get(themePreference);
        },

        /**
         * Cycle through themes: light -> dark -> auto -> light
         */
        cycle() {
            const current = get(themePreference);
            const next: ThemePreference =
                current === 'light' ? 'dark' : current === 'dark' ? 'auto' : 'light';
            this.set(next);
        },

        /**
         * Initialize theme on mount (call this in +layout.svelte onMount)
         * This ensures the theme is applied even if the store subscription
         * hasn't fired yet
         */
        initialize() {
            if (!browser) return;
            const preference = loadThemePreference();
            const effective = preference === 'auto' ? getSystemTheme() : preference;
            applyTheme(effective);
        }
    };
}

export const theme = createThemeStore();
