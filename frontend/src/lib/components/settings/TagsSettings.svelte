<script lang="ts">
	import { onMount } from 'svelte';
	import type { Tag, TagInfo } from '$lib/types';
	import { tags as tagsStore } from '$lib/stores/tags';
	import { isLoggedIn } from '$lib/stores/auth';
	import { discovery } from '$lib/stores/discovery';
	import { ui } from '$lib/stores/ui';
	import LoadingSpinner from '../LoadingSpinner.svelte';
	import TagPill from '../TagPill.svelte';
	import { generateTagColor, normalizeColor } from '$lib/utils/color';

	let loading = true;
	let creating = false;
	let deleting: string | null = null;
	let updating: string | null = null;

	// New tag form
	let newTagName = '';
	let newTagColor = '#6c757d';
	let colorManuallySet = false;

	// Auto-generate color as user types tag name (unless manually overridden)
	$: if (!colorManuallySet && newTagName.trim()) {
		newTagColor = generateTagColor(newTagName.trim());
	}

	function onColorPickerChange(event: Event) {
		const input = event.target as HTMLInputElement;
		newTagColor = input.value;
		colorManuallySet = true;
	}

	function onPresetColorClick(color: string) {
		newTagColor = color;
		colorManuallySet = true;
	}

	// Editing state
	let editingTag: string | null = null;
	let editTagName = '';
	let editTagColor = '';

	// Predefined colors
	const colors = [
		{ value: '#6c757d', label: 'Gray' },
		{ value: '#007bff', label: 'Blue' },
		{ value: '#28a745', label: 'Green' },
		{ value: '#dc3545', label: 'Red' },
		{ value: '#ffc107', label: 'Yellow' },
		{ value: '#17a2b8', label: 'Cyan' },
		{ value: '#6f42c1', label: 'Purple' },
		{ value: '#fd7e14', label: 'Orange' },
		{ value: '#20c997', label: 'Teal' },
		{ value: '#e83e8c', label: 'Pink' }
	];

	$: tagList = $tagsStore.tags;

	onMount(async () => {
		await loadTags();
	});

	async function loadTags() {
		loading = true;
		try {
			if ($isLoggedIn && $discovery.data?.usersettings_endpoint) {
				await tagsStore.fetch($discovery.data.usersettings_endpoint);
			}
		} catch (error) {
			ui.error('Failed to load tags');
		} finally {
			loading = false;
		}
	}

	async function createTag() {
		if (!newTagName.trim()) {
			ui.showError('Error', 'Please enter a tag name');
			return;
		}

		if (tagList.some(t => t.tag.toLowerCase() === newTagName.trim().toLowerCase())) {
			ui.showError('Error', 'A tag with this name already exists');
			return;
		}

		creating = true;
		try {
			if ($isLoggedIn && $discovery.data?.usersettings_endpoint) {
				const tag: Tag = {
					tag: newTagName.trim(),
					color: newTagColor
				};
				await tagsStore.create($discovery.data.usersettings_endpoint, tag);
				ui.success(`Tag "${tag.tag}" created`);
				// Reset form
				newTagName = '';
				newTagColor = '#6c757d';
				colorManuallySet = false;
			}
		} catch (error) {
			ui.error('Failed to create tag');
		} finally {
			creating = false;
		}
	}

	async function deleteTag(tagName: string) {
		const confirmed = await ui.confirm({
			title: 'Delete Tag',
			message: `Are you sure you want to delete the tag "${tagName}"? This will remove it from all mytokens, calendars, and notifications.`,
			confirmText: 'Delete',
			confirmVariant: 'danger'
		});
		
		if (!confirmed) return;

		deleting = tagName;
		try {
			if ($isLoggedIn && $discovery.data?.usersettings_endpoint) {
				await tagsStore.delete($discovery.data.usersettings_endpoint, tagName);
				ui.success(`Tag "${tagName}" deleted`);
			}
		} catch (error) {
			ui.error('Failed to delete tag');
		} finally {
			deleting = null;
		}
	}

	function startEditing(tag: TagInfo) {
		editingTag = tag.tag;
		editTagName = tag.tag;
		// Normalize color for the color picker (requires # prefix)
		editTagColor = normalizeColor(tag.color);
	}

	function cancelEditing() {
		editingTag = null;
		editTagName = '';
		editTagColor = '';
	}

	async function saveTag() {
		if (!editingTag) return;
		
		if (!editTagName.trim()) {
			ui.showError('Error', 'Tag name cannot be empty');
			return;
		}

		// Check if new name conflicts with existing tag (except current one)
		if (editTagName.trim().toLowerCase() !== editingTag.toLowerCase() &&
		    tagList.some(t => t.tag.toLowerCase() === editTagName.trim().toLowerCase())) {
			ui.showError('Error', 'A tag with this name already exists');
			return;
		}

		const originalTag = tagList.find(t => t.tag === editingTag);
		if (!originalTag) return;

		// Check if anything changed
		const nameChanged = editTagName.trim() !== editingTag;
		const colorChanged = editTagColor !== (originalTag.color ?? '#6c757d');

		if (!nameChanged && !colorChanged) {
			cancelEditing();
			return;
		}

		updating = editingTag;
		try {
			if ($isLoggedIn && $discovery.data?.usersettings_endpoint) {
				const updates: { tag?: string; color?: string } = {};
				if (nameChanged) updates.tag = editTagName.trim();
				if (colorChanged) updates.color = editTagColor;

				const success = await tagsStore.update(
					$discovery.data.usersettings_endpoint,
					editingTag,
					updates
				);

				if (success) {
					ui.success('Tag updated');
					cancelEditing();
				} else {
					ui.error('Failed to update tag');
				}
			}
		} catch (error) {
			ui.error('Failed to update tag');
		} finally {
			updating = null;
		}
	}

	function handleKeydown(event: KeyboardEvent) {
		if (event.key === 'Enter') {
			saveTag();
		} else if (event.key === 'Escape') {
			cancelEditing();
		}
	}
</script>

<div class="tags-settings">
	<p class="text-muted mb-4">
		Tags help you organize your tokens. You can assign tags when creating tokens and use them to filter or manage tokens in bulk.
	</p>

	{#if loading}
		<LoadingSpinner message="Loading tags..." />
	{:else}
		<!-- Create new tag -->
		<div class="card mb-4">
			<div class="card-header">
				<h6 class="mb-0">
					<i class="fas fa-plus-circle me-2"></i>
					Create New Tag
				</h6>
			</div>
			<div class="card-body">
				<form on:submit|preventDefault={createTag}>
					<div class="row g-3 align-items-end">
						<div class="col-md-5">
							<label for="tag-name" class="form-label">Tag Name</label>
							<input
								type="text"
								id="tag-name"
								class="form-control"
								placeholder="e.g., production, testing"
								bind:value={newTagName}
								maxlength="50"
							/>
						</div>
						<div class="col-md-4">
							<label for="tag-color" class="form-label">Color</label>
							<div class="d-flex gap-2 align-items-center">
								<input
									type="color"
									id="tag-color"
									class="form-control form-control-color"
									value={newTagColor}
									on:input={onColorPickerChange}
									title="Choose tag color"
								/>
								<div class="color-presets">
									{#each colors as color}
										<button
											type="button"
											class="color-preset"
											style="background-color: {color.value}"
											title={color.label}
											on:click={() => onPresetColorClick(color.value)}
											class:active={newTagColor === color.value}
										></button>
									{/each}
								</div>
							</div>
						</div>
						<div class="col-md-3">
							<button type="submit" class="btn btn-primary w-100" disabled={creating || !newTagName.trim()}>
								{#if creating}
									<span class="spinner-border spinner-border-sm me-1"></span>
								{:else}
									<i class="fas fa-plus me-1"></i>
								{/if}
								Create Tag
							</button>
						</div>
					</div>
					{#if newTagName.trim()}
						<div class="mt-3">
							<span class="text-muted me-2">Preview:</span>
							<TagPill name={newTagName.trim()} color={newTagColor} />
						</div>
					{/if}
				</form>
			</div>
		</div>

		<!-- Existing tags -->
		<div class="card">
			<div class="card-header d-flex justify-content-between align-items-center">
				<h6 class="mb-0">
					<i class="fas fa-tags me-2"></i>
					Your Tags
				</h6>
				<span class="badge bg-secondary">{tagList.length} tags</span>
			</div>
			<div class="card-body">
				{#if tagList.length === 0}
					<div class="text-center py-4 text-muted">
						<i class="fas fa-tags fa-3x mb-3"></i>
						<p>No tags created yet</p>
						<p class="small">Create your first tag above to get started.</p>
					</div>
				{:else}
					<div class="tags-list">
						{#each tagList as tag}
							<div class="tag-item d-flex align-items-center justify-content-between p-3 border-bottom">
								{#if editingTag === tag.tag}
									<!-- Editing mode -->
									<div class="d-flex align-items-center gap-3 flex-grow-1">
										<input
											type="color"
											class="form-control form-control-color"
											bind:value={editTagColor}
											title="Choose tag color"
										/>
										<input
											type="text"
											class="form-control"
											bind:value={editTagName}
											on:keydown={handleKeydown}
											maxlength="50"
											style="max-width: 200px;"
										/>
										<span class="text-muted">→</span>
										<TagPill name={editTagName.trim() || '(empty)'} color={editTagColor} />
									</div>
									<div class="d-flex gap-2">
										<button
											type="button"
											class="btn btn-sm btn-success"
											title="Save changes"
											disabled={updating === tag.tag}
											on:click={saveTag}
										>
											{#if updating === tag.tag}
												<span class="spinner-border spinner-border-sm"></span>
											{:else}
												<i class="fas fa-check"></i>
											{/if}
										</button>
										<button
											type="button"
											class="btn btn-sm btn-outline-secondary"
											title="Cancel"
											on:click={cancelEditing}
										>
											<i class="fas fa-times"></i>
										</button>
									</div>
								{:else}
									<!-- Display mode -->
									<div class="d-flex align-items-center gap-3">
										<TagPill name={tag.tag} color={tag.color ?? ''} />
									</div>
									<div class="d-flex gap-2">
										<button
											type="button"
											class="btn btn-sm btn-outline-secondary"
											title="Edit tag"
											on:click={() => startEditing(tag)}
										>
											<i class="fas fa-edit"></i>
										</button>
										<button
											type="button"
											class="btn btn-sm btn-outline-danger"
											title="Delete tag"
											disabled={deleting === tag.tag}
											on:click={() => deleteTag(tag.tag)}
										>
											{#if deleting === tag.tag}
												<span class="spinner-border spinner-border-sm"></span>
											{:else}
												<i class="fas fa-trash"></i>
											{/if}
										</button>
									</div>
								{/if}
							</div>
						{/each}
					</div>
				{/if}
			</div>
		</div>

		<!-- Info box -->
		<div class="alert alert-info mt-4">
			<h6 class="alert-heading">
				<i class="fas fa-info-circle me-1"></i>
				About Tags
			</h6>
			<p class="mb-0">
				When you rename a tag, it will be automatically updated on all mytokens, calendars, and notifications that use it.
				Deleting a tag will remove it from all associated items.
			</p>
		</div>
	{/if}
</div>

<style>
	.card-header {
		background-color: #f8f9fa;
	}

	.form-control-color {
		width: 40px;
		height: 38px;
		padding: 0.25rem;
	}

	.color-presets {
		display: flex;
		gap: 4px;
		flex-wrap: wrap;
	}

	.color-preset {
		width: 20px;
		height: 20px;
		border: 2px solid transparent;
		border-radius: 4px;
		cursor: pointer;
		padding: 0;
	}

	.color-preset:hover {
		transform: scale(1.1);
	}

	.color-preset.active {
		border-color: #000;
		box-shadow: 0 0 0 2px rgba(0,0,0,0.2);
	}

	.tags-list {
		margin: -1rem;
	}

	.tag-item {
		transition: background-color 0.15s;
	}

	.tag-item:hover {
		background-color: #f8f9fa;
	}

	.tag-item:last-child {
		border-bottom: none !important;
	}
</style>
