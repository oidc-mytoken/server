import {writable} from 'svelte/store';
import type {Tag, TagInfo} from '$lib/types';
import {api} from '$lib/api/client';
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
		async fetch(): Promise<TagInfo[]> {
			update((state) => ({ ...state, loading: true, error: null }));

			try {
				let tags: TagInfo[];
				try {
					tags = await api.getTags();
				} catch (err: any) {
					if (err.status === 401) {
						console.warn('Tags fetch returned 401 - token may have been rotated');
						set({
							tags: [],
							loading: false,
							error: null,
							loaded: true
						});
						return [];
					}
					throw err;
				}

				set({
					tags,
					loading: false,
					error: null,
					loaded: true
				});

				return tags;
			} catch (err: any) {
				const errorMessage = err instanceof Error ? err.message : 'An unknown error occurred';
				update((state) => ({ ...state, loading: false, error: errorMessage }));
				console.error('Failed to fetch tags:', err);
				return [];
			}
		},

		/**
		 * Create a new tag (uses cookie auth)
		 */
		async create(tag: Tag): Promise<boolean> {
			try {
				await api.createTag(tag.tag, stripHashFromColor(tag.color || ''));
				
				let newTag: TagInfo = {
					tag: tag.tag,
					color: tag.color ?? ''
				};

				// Add to local state
				update((state) => ({
					...state,
					tags: [...state.tags, newTag]
				}));

				return true;
			} catch (err) {
				console.error('Error creating tag:', err);
				return false;
			}
		},

		/**
		 * Delete a tag
		 */
		async delete(tagName: string): Promise<boolean> {
			try {
				await api.deleteTag(tagName);

				// Remove from local state
				update((state) => ({
					...state,
					tags: state.tags.filter((t) => t.tag !== tagName)
				}));

				return true;
			} catch (err) {
				console.error('Error deleting tag:', err);
				return false;
			}
		},

		/**
		 * Update tag color or name
		 */
		async update(oldTagName: string, updates: {
			color?: string;
			tag?: string;
		}): Promise<boolean> {
			try {
				// Strip # from color for API (API expects color without #)
				const apiUpdates = {
					...updates,
					color: updates.color ? stripHashFromColor(updates.color) : updates.color
				};

				await api.updateTag(oldTagName, apiUpdates.color, apiUpdates.tag);

				// Update local state (keep color with # for frontend)
				update((state) => ({
					...state,
					tags: state.tags.map((t) => {
						if (t.tag === oldTagName) {
							return {
								...t,
								...updates,
								color: updates.color ?? t.color
							};
						}
						return t;
					})
				}));

				return true;
			} catch (err) {
				console.error('Error renaming/updating tag:', err);
				return false;
			}
		},

		/**
		 * Rename a tag (convenience wrapper around update)
		 */
		async rename(oldTagName: string, updates: Partial<Tag>): Promise<boolean> {
			return this.update(oldTagName, updates);
		}
	};
}

export const tags = createTagsStore();
