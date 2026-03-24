<script lang="ts">
	import { onMount } from 'svelte';
	import type { Tag } from '$lib/types';
	import { tags as tagsStore } from '$lib/stores/tags';
	import { isLoggedIn } from '$lib/stores/auth';
	import { discovery } from '$lib/stores/discovery';
	import { ui } from '$lib/stores/ui';
	import TagPill from '../TagPill.svelte';
	import LoadingSpinner from '../LoadingSpinner.svelte';

	let loading = true;
	let creating = false;
	let deleting: string | null = null;

	// New tag form
	let newTagName = '';
	let newTagColor = '#6c757d';

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
				newTagName = '';
				newTagColor = '#6c757d';
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
			message: `Are you sure you want to delete the tag "${tagName}"? This will remove it from all tokens.`,
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

	function getContrastColor(hexColor: string): string {
		// Convert hex to RGB
		const r = parseInt(hexColor.slice(1, 3), 16);
		const g = parseInt(hexColor.slice(3, 5), 16);
		const b = parseInt(hexColor.slice(5, 7), 16);
		// Calculate luminance
		const luminance = (0.299 * r + 0.587 * g + 0.114 * b) / 255;
		return luminance > 0.5 ? '#000000' : '#ffffff';
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
									bind:value={newTagColor}
									title="Choose tag color"
								/>
								<div class="color-presets">
									{#each colors as color}
										<button
											type="button"
											class="color-preset"
											style="background-color: {color.value}"
											title={color.label}
											on:click={() => newTagColor = color.value}
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
							<span 
								class="badge" 
								style="background-color: {newTagColor}; color: {getContrastColor(newTagColor)}"
							>
								{newTagName.trim()}
							</span>
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
					<div class="table-responsive">
						<table class="table table-hover align-middle mb-0">
							<thead>
								<tr>
									<th>Tag</th>
									<th>Color</th>
									<th class="text-end">Actions</th>
								</tr>
							</thead>
							<tbody>
								{#each tagList as tag}
									<tr>
										<td>
											<span 
												class="badge" 
												style="background-color: {tag.color ?? '#6c757d'}; color: {getContrastColor(tag.color ?? '#6c757d')}"
											>
												{tag.tag}
											</span>
										</td>
										<td>
											<div class="d-flex align-items-center gap-2">
												<span 
													class="color-swatch" 
													style="background-color: {tag.color ?? '#6c757d'}"
												></span>
												<code class="small">{tag.color ?? '#6c757d'}</code>
											</div>
										</td>
										<td class="text-end">
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
										</td>
									</tr>
								{/each}
							</tbody>
						</table>
					</div>
				{/if}
			</div>
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

	.color-swatch {
		width: 16px;
		height: 16px;
		border-radius: 4px;
		display: inline-block;
	}

	.table th {
		border-top: none;
		font-weight: 500;
	}
</style>
