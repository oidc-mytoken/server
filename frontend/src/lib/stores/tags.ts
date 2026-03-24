import { writable } from 'svelte/store';
import type { Tag } from '$lib/types';

interface TagsState {
	tags: Tag[];
	loading: boolean;
	error: string | null;
	loaded: boolean;
}

function createTagsStore() {
	const { subscribe, set, update } = writable<TagsState>({
		tags: [],
		loading: false,
		error: null,
		loaded: false
	});

	return {
		subscribe,

		/**
		 * Fetch tags from the user settings endpoint (uses cookie auth)
		 */
		async fetch(usersettingsEndpoint: string): Promise<Tag[]> {
			update((state) => ({ ...state, loading: true, error: null }));

			try {
				const response = await fetch(`${usersettingsEndpoint}/tags`, {
					method: 'GET',
					headers: {
						'Content-Type': 'application/json'
					},
					credentials: 'include'
				});

				if (!response.ok) {
					// If unauthorized, just return empty tags - user may have logged out
					// or there was a token rotation issue
					if (response.status === 401) {
						console.warn('Tags fetch returned 401 - token may have been rotated');
						set({
							tags: [],
							loading: false,
							error: null,
							loaded: true
						});
						return [];
					}
					throw new Error(`Failed to fetch tags: ${response.status}`);
				}

				const data = await response.json();
				const tags: Tag[] = data.tags ?? [];

				set({
					tags,
					loading: false,
					error: null,
					loaded: true
				});

				return tags;
			} catch (e) {
				const errorMessage = e instanceof Error ? e.message : 'Unknown error';
				update((state) => ({
					...state,
					loading: false,
					error: errorMessage
				}));
				return [];
			}
		},

		/**
		 * Create a new tag (uses cookie auth)
		 */
		async create(usersettingsEndpoint: string, tag: Tag): Promise<boolean> {
			try {
				const response = await fetch(
					`${usersettingsEndpoint}/tags/${encodeURIComponent(tag.tag)}`,
					{
						method: 'POST',
						headers: {
							'Content-Type': 'application/json'
						},
						credentials: 'include',
						body: JSON.stringify(tag)
					}
				);

				if (!response.ok) {
					throw new Error(`Failed to create tag: ${response.status}`);
				}

				// Add to local state
				update((state) => ({
					...state,
					tags: [...state.tags, tag]
				}));

				return true;
			} catch (e) {
				console.error('Failed to create tag:', e);
				return false;
			}
		},

		/**
		 * Delete a tag (uses cookie auth)
		 */
		async delete(usersettingsEndpoint: string, tagName: string): Promise<boolean> {
			try {
				const response = await fetch(
					`${usersettingsEndpoint}/tags/${encodeURIComponent(tagName)}`,
					{
						method: 'DELETE',
						credentials: 'include'
					}
				);

				if (!response.ok) {
					throw new Error(`Failed to delete tag: ${response.status}`);
				}

				// Remove from local state
				update((state) => ({
					...state,
					tags: state.tags.filter((t) => t.tag !== tagName)
				}));

				return true;
			} catch (e) {
				console.error('Failed to delete tag:', e);
				return false;
			}
		},

		/**
		 * Clear the store
		 */
		clear() {
			set({
				tags: [],
				loading: false,
				error: null,
				loaded: false
			});
		}
	};
}

export const tags = createTagsStore();
