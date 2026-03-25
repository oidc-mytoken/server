import {writable} from 'svelte/store';
import type {Tag, TagInfo} from '$lib/types';
import {stripHashFromColor} from '$lib/utils/color';

interface TagsState {
	tags: TagInfo[];
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
		async fetch(usersettingsEndpoint: string): Promise<TagInfo[]> {
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
				const tags: TagInfo[] = data.tags ?? [];

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
		 * Note: The backend POST endpoint only creates the tag with default color.
		 * We need to use PUT to set the color afterwards.
		 */
		async create(usersettingsEndpoint: string, tag: Tag): Promise<boolean> {
			try {
				// First create the tag (color is not supported in POST)
				const createResponse = await fetch(
					`${usersettingsEndpoint}/tags/${encodeURIComponent(tag.tag)}`,
					{
						method: 'POST',
						headers: {
							'Content-Type': 'application/json'
						},
						credentials: 'include'
					}
				);

				if (!createResponse.ok) {
					throw new Error(`Failed to create tag: ${createResponse.status}`);
				}

				// If a color was specified, set it using PUT
				if (tag.color) {
					// Strip # from color for API (API expects color without #)
					const updateResponse = await fetch(
						`${usersettingsEndpoint}/tags/${encodeURIComponent(tag.tag)}`,
						{
							method: 'PUT',
							headers: {
								'Content-Type': 'application/json'
							},
							credentials: 'include',
							body: JSON.stringify({color: stripHashFromColor(tag.color)})
						}
					);

					if (!updateResponse.ok) {
						console.warn(`Tag created but failed to set color: ${updateResponse.status}`);
					}
				}

				// Add to local state as TagInfo (with color defaulting to gray if not specified)
				const tagInfo: TagInfo = {
					tag: tag.tag,
					color: tag.color ?? '#6c757d'
				};
				update((state) => ({
					...state,
					tags: [...state.tags, tagInfo]
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
		 * Update a tag (uses cookie auth)
		 * Can update name and/or color
		 */
		async update(usersettingsEndpoint: string, oldTagName: string, updates: {
			tag?: string;
			color?: string
		}): Promise<boolean> {
			try {
				// Strip # from color for API (API expects color without #)
				const apiUpdates = {
					...updates,
					color: updates.color ? stripHashFromColor(updates.color) : updates.color
				};

				const response = await fetch(
					`${usersettingsEndpoint}/tags/${encodeURIComponent(oldTagName)}`,
					{
						method: 'PUT',
						headers: {
							'Content-Type': 'application/json'
						},
						credentials: 'include',
						body: JSON.stringify(apiUpdates)
					}
				);

				if (!response.ok) {
					throw new Error(`Failed to update tag: ${response.status}`);
				}

				// Update local state (keep color with # for frontend)
				update((state) => ({
					...state,
					tags: state.tags.map((t) => {
						if (t.tag === oldTagName) {
							return {
								...t,
								tag: updates.tag ?? t.tag,
								color: updates.color ?? t.color
							};
						}
						return t;
					})
				}));

				return true;
			} catch (e) {
				console.error('Failed to update tag:', e);
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
