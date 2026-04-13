<script lang="ts">
	import type { Calendar } from '$lib/types';
	import { api, ApiClientError } from '$lib/api/client';
	import { ui } from '$lib/stores/ui';
	import { tags as tagsStore } from '$lib/stores/tags';
	import { discovery } from '$lib/stores/discovery';
	import CopyButton from '../CopyButton.svelte';
	import TagPill from '../TagPill.svelte';

	export let calendars: Calendar[] = [];
	export let loading = false;
	export let onDelete: ((calendarId: string) => void) | undefined = undefined;
	export let onUpdate: ((calendar: Calendar) => void) | undefined = undefined;
	export let onCreate: (() => void) | undefined = undefined;

	// Edit modal state
	let editingCalendar: Calendar | null = null;
	let editDescription = '';
	let editTags: string[] = [];
	let saving = false;
	let showEditModal = false;

	// Create modal state
	let showCreateModal = false;
	let createDescription = '';
	let createTags: string[] = [];
	let creating = false;

	// Tag dropdown state (shared between edit and create)
	let newTagInput = '';

	// Get available tags from the store
	$: availableTags = $tagsStore.tags;

	// Tags not yet selected (for dropdown) - for edit modal
	$: availableTagsForEdit = availableTags.filter(t => !editTags.includes(t.tag));

	// Tags not yet selected (for dropdown) - for create modal
	$: availableTagsForCreate = availableTags.filter(t => !createTags.includes(t.tag));

	function getTagColor(tagName: string): string {
		const tag = availableTags.find(t => t.tag === tagName);
		return tag?.color ?? '';
	}

	// ========== Edit Modal Functions ==========
	function openEditModal(calendar: Calendar) {
		editingCalendar = calendar;
		editDescription = calendar.description ?? '';
		editTags = calendar.tags?.map(t => t.tag) ?? [];
		newTagInput = '';
		showEditModal = true;
	}

	function closeEditModal() {
		showEditModal = false;
		editingCalendar = null;
		editDescription = '';
		editTags = [];
		newTagInput = '';
		saving = false;
	}

	async function saveCalendar() {
		if (!editingCalendar) return;

		saving = true;
		try {
			const updated = await api.updateCalendar(editingCalendar.id, {
				description: editDescription,
				tags: editTags
			});
			ui.success('Calendar updated');
			onUpdate?.(updated);
			closeEditModal();
		} catch (error) {
			if (error instanceof ApiClientError) {
				ui.showError('Failed to update calendar', error.description ?? error.code);
			}
		} finally {
			saving = false;
		}
	}

	function removeEditTag(tagName: string) {
		editTags = editTags.filter(t => t !== tagName);
	}

	function addEditTag(tagName: string) {
		if (tagName && !editTags.includes(tagName)) {
			editTags = [...editTags, tagName];
		}
		newTagInput = '';
	}

	async function addNewTagToEdit() {
		const tagName = newTagInput.trim();
		if (!tagName) return;

		// Check if tag already exists
		const existingTag = availableTags.find(t => t.tag.toLowerCase() === tagName.toLowerCase());
		if (existingTag) {
			addEditTag(existingTag.tag);
			return;
		}

		// Create new tag
		if ($discovery.data?.usersettings_endpoint) {
			const success = await tagsStore.create({ 
				tag: tagName, 
				color: '#6c757d' 
			});
			if (success) {
				addEditTag(tagName);
			} else {
				ui.error('Failed to create tag');
			}
		}
	}

	// ========== Create Modal Functions ==========
	export function openCreateModal() {
		createDescription = '';
		createTags = [];
		newTagInput = '';
		showCreateModal = true;
	}

	function closeCreateModal() {
		showCreateModal = false;
		createDescription = '';
		createTags = [];
		newTagInput = '';
		creating = false;
	}

	async function createCalendar() {
		creating = true;
		try {
			await api.createCalendar({
				description: createDescription || undefined,
				tags: createTags.length > 0 ? createTags : undefined
			});
			ui.success('Calendar created');
			onCreate?.();
			closeCreateModal();
		} catch (error) {
			if (error instanceof ApiClientError) {
				ui.showError('Failed to create calendar', error.description ?? error.code);
			}
		} finally {
			creating = false;
		}
	}

	function removeCreateTag(tagName: string) {
		createTags = createTags.filter(t => t !== tagName);
	}

	function addCreateTag(tagName: string) {
		if (tagName && !createTags.includes(tagName)) {
			createTags = [...createTags, tagName];
		}
		newTagInput = '';
	}

	async function addNewTagToCreate() {
		const tagName = newTagInput.trim();
		if (!tagName) return;

		// Check if tag already exists
		const existingTag = availableTags.find(t => t.tag.toLowerCase() === tagName.toLowerCase());
		if (existingTag) {
			addCreateTag(existingTag.tag);
			return;
		}

		// Create new tag
		if ($discovery.data?.usersettings_endpoint) {
			const success = await tagsStore.create({ 
				tag: tagName, 
				color: '#6c757d' 
			});
			if (success) {
				addCreateTag(tagName);
			} else {
				ui.error('Failed to create tag');
			}
		}
	}

	// ========== Delete Function ==========
	async function handleDelete(calendar: Calendar) {
		const confirmed = await ui.confirm({
			title: 'Delete Calendar',
			message: `Are you sure you want to delete this calendar?`,
			confirmText: 'Delete',
			confirmVariant: 'danger'
		});
		
		if (!confirmed) return;

		try {
			await api.deleteCalendar(calendar.id);
			ui.success('Calendar deleted');
			onDelete?.(calendar.id);
		} catch (error) {
			if (error instanceof ApiClientError) {
				ui.showError('Failed to delete calendar', error.description ?? error.code);
			}
		}
	}

	// ========== URL Helpers ==========
	function getIcsUrl(calendar: Calendar): string {
		const path = calendar.ics_path ?? calendar.ics_url ?? '';
		if (path.startsWith('http')) {
			return path;
		}
		return `${window.location.origin}${path}`;
	}

	function getCalendarViewUrl(calendar: Calendar): string {
		const path = calendar.ics_path ?? calendar.ics_url ?? '';
		return `${path}/view`;
	}
</script>

<div class="calendar-list">
	{#if loading}
		<div class="text-center py-4">
			<div class="spinner-border text-primary" role="status">
				<span class="visually-hidden">Loading...</span>
			</div>
		</div>
	{:else if calendars.length === 0}
		<div class="text-center py-4 text-muted">
			<i class="fas fa-calendar-times fa-3x mb-3"></i>
			<p>No calendars found</p>
			<p class="small">
				Calendars show token expiration events in your calendar app.
			</p>
		</div>
	{:else}
		<div class="table-responsive">
			<table class="table table-hover align-middle">
				<thead>
					<tr>
						<th>Description</th>
						<th>Tags</th>
						<th style="width: 200px;">ICS URL</th>
						<th style="width: 150px;" class="text-end">Actions</th>
					</tr>
				</thead>
				<tbody>
					{#each calendars as calendar}
						<tr>
							<td>
								{calendar.description ?? '-'}
							</td>
							<td>
								{#if calendar.tags && calendar.tags.length > 0}
									{#each calendar.tags as tagInfo}
										<TagPill tag={tagInfo.tag} color={tagInfo.color} small />
									{/each}
								{:else}
									<span class="text-muted">-</span>
								{/if}
							</td>
							<td>
								<div class="d-flex align-items-center gap-2">
									<code class="text-truncate" style="max-width: 120px;" title={getIcsUrl(calendar)}>
										{getIcsUrl(calendar)}
									</code>
									<CopyButton value={getIcsUrl(calendar)} variant="icon" />
								</div>
							</td>
							<td class="text-end">
								<div class="btn-group btn-group-sm">
									<a 
										href={getCalendarViewUrl(calendar)}
										class="btn btn-outline-secondary"
										title="View Calendar"
									>
										<i class="fas fa-calendar-alt"></i>
									</a>
									<button
										type="button"
										class="btn btn-outline-secondary"
										title="Edit Calendar"
										onclick={() => openEditModal(calendar)}
									>
										<i class="fas fa-pencil-alt"></i>
									</button>
									<a 
										href={getIcsUrl(calendar)}
										class="btn btn-outline-primary"
										title="Download ICS"
										target="_blank"
									>
										<i class="fas fa-download"></i>
									</a>
									<button 
										type="button" 
										class="btn btn-outline-danger" 
										title="Delete"
										onclick={() => handleDelete(calendar)}
									>
										<i class="fas fa-trash"></i>
									</button>
								</div>
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>

		<div class="alert alert-info mt-3">
			<i class="fas fa-info-circle me-2"></i>
			<strong>How to use:</strong> Copy the ICS URL and add it as a subscription in your calendar application 
			(Google Calendar, Outlook, Apple Calendar, etc.).
		</div>
	{/if}
</div>

<!-- Edit Calendar Modal -->
{#if showEditModal && editingCalendar}
	<div class="modal show d-block" tabindex="-1" role="dialog">
		<div class="modal-dialog">
			<div class="modal-content">
				<div class="modal-header">
					<h5 class="modal-title">
						<i class="fas fa-edit me-2"></i>
						Edit Calendar
					</h5>
					<button 
						type="button" 
						class="btn-close" 
						aria-label="Close"
						onclick={closeEditModal}
					></button>
				</div>
				<div class="modal-body">
					<div class="mb-3">
						<label for="edit-description" class="form-label">Description</label>
						<textarea
							id="edit-description"
							class="form-control"
							rows="3"
							placeholder="Enter a description for this calendar..."
							bind:value={editDescription}
						></textarea>
					</div>

					<div class="mb-3">
						<label for="edit-tag-select" class="form-label">Tags</label>
						<p class="text-muted small mb-2">
							Tags filter which tokens appear in this calendar.
						</p>
						
						<!-- Selected tags with remove buttons -->
						{#if editTags.length > 0}
							<div class="selected-tags mb-2">
								{#each editTags as tagName}
									<TagPill 
										tag={tagName} 
										color={getTagColor(tagName)} 
										removable
										onRemove={() => removeEditTag(tagName)}
									/>
								{/each}
							</div>
						{/if}

						<!-- Tag dropdown + new tag input -->
						<div class="input-group">
							<select 
								id="edit-tag-select"
								class="form-select"
								bind:value={newTagInput}
								onchange={() => { if (newTagInput) addEditTag(newTagInput); }}
							>
								<option value="">Add existing tag...</option>
								{#each availableTagsForEdit as tag}
									<option value={tag.tag}>{tag.tag}</option>
								{/each}
							</select>
							<span class="input-group-text">or</span>
							<input 
								type="text" 
								class="form-control" 
								placeholder="New tag name"
								bind:value={newTagInput}
								onkeydown={(e) => { if (e.key === 'Enter') { e.preventDefault(); addNewTagToEdit(); } }}
							/>
							<button 
								type="button" 
								class="btn btn-outline-secondary"
								onclick={addNewTagToEdit}
								disabled={!newTagInput.trim()}
								title="Add tag"
							>
								<i class="fas fa-plus"></i>
							</button>
						</div>
						<small class="text-muted">Select an existing tag or type a new tag name and press Enter.</small>
					</div>
				</div>
				<div class="modal-footer">
					<button 
						type="button" 
						class="btn btn-secondary" 
						onclick={closeEditModal}
						disabled={saving}
					>
						Cancel
					</button>
					<button 
						type="button" 
						class="btn btn-primary" 
						onclick={saveCalendar}
						disabled={saving}
					>
						{#if saving}
							<span class="spinner-border spinner-border-sm me-1"></span>
							Saving...
						{:else}
							<i class="fas fa-save me-1"></i>
							Save Changes
						{/if}
					</button>
				</div>
			</div>
		</div>
	</div>
	<div class="modal-backdrop show"></div>
{/if}

<!-- Create Calendar Modal -->
{#if showCreateModal}
	<div class="modal show d-block" tabindex="-1" role="dialog">
		<div class="modal-dialog">
			<div class="modal-content">
				<div class="modal-header">
					<h5 class="modal-title">
						<i class="fas fa-plus-circle me-2"></i>
						Create Calendar
					</h5>
					<button 
						type="button" 
						class="btn-close" 
						aria-label="Close"
						onclick={closeCreateModal}
					></button>
				</div>
				<div class="modal-body">
					<div class="mb-3">
						<label for="create-description" class="form-label">Description</label>
						<textarea
							id="create-description"
							class="form-control"
							rows="3"
							placeholder="Enter a description for this calendar (optional)..."
							bind:value={createDescription}
						></textarea>
					</div>

					<div class="mb-3">
						<label for="create-tag-select" class="form-label">Tags</label>
						<p class="text-muted small mb-2">
							Tags filter which tokens appear in this calendar. Leave empty to include all tokens.
						</p>
						
						<!-- Selected tags with remove buttons -->
						{#if createTags.length > 0}
							<div class="selected-tags mb-2">
								{#each createTags as tagName}
									<TagPill 
										tag={tagName} 
										color={getTagColor(tagName)} 
										removable
										onRemove={() => removeCreateTag(tagName)}
									/>
								{/each}
							</div>
						{/if}

						<!-- Tag dropdown + new tag input -->
						<div class="input-group">
							<select 
								id="create-tag-select"
								class="form-select"
								bind:value={newTagInput}
								onchange={() => { if (newTagInput) addCreateTag(newTagInput); }}
							>
								<option value="">Add existing tag...</option>
								{#each availableTagsForCreate as tag}
									<option value={tag.tag}>{tag.tag}</option>
								{/each}
							</select>
							<span class="input-group-text">or</span>
							<input 
								type="text" 
								class="form-control" 
								placeholder="New tag name"
								bind:value={newTagInput}
								onkeydown={(e) => { if (e.key === 'Enter') { e.preventDefault(); addNewTagToCreate(); } }}
							/>
							<button 
								type="button" 
								class="btn btn-outline-secondary"
								onclick={addNewTagToCreate}
								disabled={!newTagInput.trim()}
								title="Add tag"
							>
								<i class="fas fa-plus"></i>
							</button>
						</div>
						<small class="text-muted">Select an existing tag or type a new tag name and press Enter.</small>
					</div>
				</div>
				<div class="modal-footer">
					<button 
						type="button" 
						class="btn btn-secondary" 
						onclick={closeCreateModal}
						disabled={creating}
					>
						Cancel
					</button>
					<button 
						type="button" 
						class="btn btn-primary" 
						onclick={createCalendar}
						disabled={creating}
					>
						{#if creating}
							<span class="spinner-border spinner-border-sm me-1"></span>
							Creating...
						{:else}
							<i class="fas fa-plus me-1"></i>
							Create Calendar
						{/if}
					</button>
				</div>
			</div>
		</div>
	</div>
	<div class="modal-backdrop show"></div>
{/if}

<style>
	.table th {
		font-weight: 500;
		border-top: none;
	}

	code {
		font-size: 0.8em;
		background-color: var(--bs-tertiary-bg);
		padding: 0.2em 0.4em;
		border-radius: 0.25rem;
	}

	.modal.show {
		background-color: rgba(0, 0, 0, 0.5);
	}

	.modal-backdrop {
		z-index: 1040;
	}

	.modal {
		z-index: 1050;
	}

	.selected-tags {
		display: flex;
		flex-wrap: wrap;
		gap: 0.25rem;
	}

	.input-group .form-select {
		max-width: 200px;
	}
</style>
