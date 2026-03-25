<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import type { Notification, MytokenEntry, MytokenEntryTree } from '$lib/types';
	import { NOTIFICATION_CLASSES } from '$lib/types';
	import { api, ApiClientError } from '$lib/api/client';
	import { ui } from '$lib/stores/ui';
	import { tags as tagsStore } from '$lib/stores/tags';
	import { discovery } from '$lib/stores/discovery';
	import { isLoggedIn } from '$lib/stores/auth';
	import { formatDateTime } from '$lib/utils/format';
	import TagPill from '../TagPill.svelte';
	import LoadingSpinner from '../LoadingSpinner.svelte';

	const dispatch = createEventDispatcher<{ 
		saved: Notification | void;
		cancel: void;
	}>();

	// Props
	export let show = false;
	export let notification: Notification | null = null; // null = create mode, otherwise edit mode

	$: isEditMode = notification !== null;
	$: modalTitle = isEditMode ? 'Edit Notification' : 'New Notification';

	// Form state
	let selectedClasses: string[] = [];
	let subscriptionMode: 'tags' | 'tokens' | 'user_wide' = 'tags';
	let selectedTags: string[] = [];
	let selectedTokenIds: string[] = [];

	// Flattened token entry with depth info for tree display
	interface FlattenedToken {
		token: MytokenEntry;
		depth: number;
		hasChildren: boolean;
		id: string;
	}

	// UI state
	let saving = false;
	let tokenTrees: MytokenEntryTree[] = [];
	let flattenedTokens: FlattenedToken[] = [];
	let loadingTokens = false;
	let newTagInput = '';
	let tokenSearchQuery = '';
	let expandedTokens: Set<string> = new Set();

	// Get available tags from the store
	$: availableTags = $tagsStore.tags;
	$: availableTagsForSelection = availableTags.filter(t => !selectedTags.includes(t.tag));

	// Sort trees by creation time (descending - newest first)
	function sortTreesByCreated(trees: MytokenEntryTree[]): MytokenEntryTree[] {
		const sorted = [...trees].sort((a, b) => (b.token.created || 0) - (a.token.created || 0));
		// Recursively sort children
		return sorted.map(tree => ({
			...tree,
			children: tree.children ? sortTreesByCreated(tree.children) : undefined
		}));
	}

	// Flatten token trees into a displayable list with depth
	function flattenTokenTrees(trees: MytokenEntryTree[]): FlattenedToken[] {
		const result: FlattenedToken[] = [];
		const sortedTrees = sortTreesByCreated(trees);
		
		const traverse = (tree: MytokenEntryTree, depth: number) => {
			const hasChildren = (tree.children?.length ?? 0) > 0;
			result.push({
				token: tree.token,
				depth,
				hasChildren,
				id: tree.token.mom_id
			});
			
			if (tree.children) {
				for (const child of tree.children) {
					traverse(child, depth + 1);
				}
			}
		};
		
		sortedTrees.forEach(t => traverse(t, 0));
		return result;
	}

	// Filter and get visible tokens based on search query and expanded state
	$: visibleTokens = flattenedTokens.filter(item => {
		// First check if token matches search query
		if (tokenSearchQuery) {
			const query = tokenSearchQuery.toLowerCase();
			const name = (item.token.name ?? '').toLowerCase();
			const momId = item.token.mom_id.toLowerCase();
			if (!name.includes(query) && !momId.includes(query)) {
				return false;
			}
		}
		
		// Check if parent is expanded (for tree visibility)
		if (item.depth === 0) return true;
		
		// Find all ancestors and check if they're expanded
		const ancestors = flattenedTokens.filter(t => t.depth < item.depth);
		let currentDepth = item.depth - 1;
		let currentIndex = flattenedTokens.indexOf(item) - 1;
		
		while (currentDepth >= 0 && currentIndex >= 0) {
			const ancestor = flattenedTokens[currentIndex];
			if (ancestor.depth === currentDepth) {
				if (!expandedTokens.has(ancestor.id)) {
					return false;
				}
				currentDepth--;
			}
			currentIndex--;
		}
		
		return true;
	});

	// Initialize form when modal opens or notification changes
	$: if (show) {
		initializeForm();
	}

	function initializeForm() {
		if (notification) {
			// Edit mode - populate from notification
			selectedClasses = [...notification.notification_classes];
			selectedTags = notification.tags?.map(t => t.tag) ?? [];
			
			// Determine subscription mode from notification
			if (notification.user_wide) {
				subscriptionMode = 'user_wide';
			} else if (notification.tags && notification.tags.length > 0) {
				subscriptionMode = 'tags';
			} else if (notification.subscribed_tokens && notification.subscribed_tokens.length > 0) {
				subscriptionMode = 'tokens';
				selectedTokenIds = [...notification.subscribed_tokens];
			} else {
				subscriptionMode = 'tags';
			}
		} else {
			// Create mode - reset form
			selectedClasses = [];
			subscriptionMode = 'tags';
			selectedTags = [];
			selectedTokenIds = [];
			expandedTokens = new Set();
		}
		newTagInput = '';
		tokenSearchQuery = '';
	}

	// Toggle token tree expansion
	function toggleTokenExpand(tokenId: string) {
		if (expandedTokens.has(tokenId)) {
			expandedTokens.delete(tokenId);
		} else {
			expandedTokens.add(tokenId);
		}
		expandedTokens = new Set(expandedTokens); // Trigger reactivity
	}

	function isTokenExpanded(tokenId: string): boolean {
		return expandedTokens.has(tokenId);
	}

	// Format expiry date
	function formatExpiry(token: MytokenEntry): string {
		if (!token.expires_at || token.expires_at === 0) {
			return 'Does not expire';
		}
		return formatDateTime(token.expires_at);
	}

	function isExpired(token: MytokenEntry): boolean {
		if (!token.expires_at || token.expires_at === 0) return false;
		return new Date(token.expires_at * 1000) < new Date();
	}

	// Load tokens when switching to tokens mode
	async function loadTokens() {
		if (tokenTrees.length > 0 || !$isLoggedIn) return;

		loadingTokens = true;
		try {
			tokenTrees = await api.listMytokensTree();
			flattenedTokens = flattenTokenTrees(tokenTrees);
		} catch (error) {
			ui.error('Failed to load tokens');
		} finally {
			loadingTokens = false;
		}
	}

	$: if (subscriptionMode === 'tokens') {
		loadTokens();
	}

	// ========== Notification Classes ==========
	// Group classes by parent
	$: topLevelClasses = NOTIFICATION_CLASSES.filter(c => !c.parent);
	$: getChildClasses = (parentId: string) => NOTIFICATION_CLASSES.filter(c => c.parent === parentId);

	function getAllChildClassIds(parentId: string): string[] {
		const children = NOTIFICATION_CLASSES.filter(c => c.parent === parentId);
		let result: string[] = [];
		for (const child of children) {
			result.push(child.id);
			result = [...result, ...getAllChildClassIds(child.id)];
		}
		return result;
	}

	function getAllParentClassIds(classId: string): string[] {
		const cls = NOTIFICATION_CLASSES.find(c => c.id === classId);
		if (!cls?.parent) return [];
		return [cls.parent, ...getAllParentClassIds(cls.parent)];
	}

	function toggleClass(classId: string) {
		const isCurrentlySelected = selectedClasses.includes(classId);
		let newSelection = [...selectedClasses];
		
		if (isCurrentlySelected) {
			// Unchecking: remove this class and all its children
			const childIds = getAllChildClassIds(classId);
			const toRemove = new Set([classId, ...childIds]);
			newSelection = newSelection.filter(c => !toRemove.has(c));
			
			// Also uncheck all parent classes (since they're no longer fully enabled)
			const parentIds = getAllParentClassIds(classId);
			newSelection = newSelection.filter(c => !parentIds.includes(c));
		} else {
			// Checking: add this class and all its children
			const childIds = getAllChildClassIds(classId);
			const toAdd = [classId, ...childIds];
			newSelection = [...new Set([...newSelection, ...toAdd])];
		}
		
		// Trigger reactivity by assigning new array
		selectedClasses = newSelection;
	}

	function isClassSelected(classId: string): boolean {
		return selectedClasses.includes(classId);
	}

	// ========== Tags ==========
	function getTagColor(tagName: string): string {
		const tag = availableTags.find(t => t.tag === tagName);
		return tag?.color ?? '';
	}

	function removeTag(tagName: string) {
		selectedTags = selectedTags.filter(t => t !== tagName);
	}

	function addTag(tagName: string) {
		if (tagName && !selectedTags.includes(tagName)) {
			selectedTags = [...selectedTags, tagName];
		}
		newTagInput = '';
	}

	async function addNewTag() {
		const tagName = newTagInput.trim();
		if (!tagName) return;

		// Check if tag already exists
		const existingTag = availableTags.find(t => t.tag.toLowerCase() === tagName.toLowerCase());
		if (existingTag) {
			addTag(existingTag.tag);
			return;
		}

		// Create new tag
		if ($discovery.data?.usersettings_endpoint) {
			const success = await tagsStore.create($discovery.data.usersettings_endpoint, { 
				tag: tagName, 
				color: '#6c757d' 
			});
			if (success) {
				addTag(tagName);
			} else {
				ui.error('Failed to create tag');
			}
		}
	}

	// ========== Token Selection ==========
	function isTokenSelected(momId: string): boolean {
		return selectedTokenIds.includes(momId);
	}

	// Get all child token IDs for a given token
	function getAllChildTokenIds(momId: string): string[] {
		const result: string[] = [];
		const tokenIndex = flattenedTokens.findIndex(t => t.id === momId);
		if (tokenIndex === -1) return result;
		
		const parentDepth = flattenedTokens[tokenIndex].depth;
		
		// Collect all tokens that come after this one with greater depth
		for (let i = tokenIndex + 1; i < flattenedTokens.length; i++) {
			const token = flattenedTokens[i];
			if (token.depth <= parentDepth) {
				// We've moved past the children
				break;
			}
			result.push(token.id);
		}
		
		return result;
	}

	function toggleTokenSelection(momId: string) {
		const childIds = getAllChildTokenIds(momId);
		const allIds = [momId, ...childIds];
		
		if (selectedTokenIds.includes(momId)) {
			// Uncheck: remove this token and all children
			selectedTokenIds = selectedTokenIds.filter(id => !allIds.includes(id));
		} else {
			// Check: add this token and all children
			selectedTokenIds = [...new Set([...selectedTokenIds, ...allIds])];
		}
	}

	// ========== Form Submission ==========
	async function handleSubmit() {
		if (selectedClasses.length === 0) {
			ui.showError('Error', 'Please select at least one notification class');
			return;
		}

		if (subscriptionMode === 'tags' && selectedTags.length === 0) {
			ui.showError('Error', 'Please select at least one tag');
			return;
		}

		if (subscriptionMode === 'tokens' && selectedTokenIds.length === 0) {
			ui.showError('Error', 'Please select at least one token');
			return;
		}

		saving = true;

		try {
			if (isEditMode && notification) {
				// Update existing notification
				await api.updateNotification(notification.management_code, {
					notification_classes: selectedClasses,
					tags: subscriptionMode === 'tags' ? selectedTags : undefined
				});
				ui.success('Notification updated');
				dispatch('saved');
			} else {
				// Create new notification
				// For tokens mode, we need to create one notification per token (API limitation)
				if (subscriptionMode === 'tokens') {
					for (const tokenId of selectedTokenIds) {
						const request: any = {
							notification_type: 'mail',
							notification_classes: selectedClasses,
							mom_id: tokenId
						};
						await api.createNotification(request);
					}
					ui.success(`Created ${selectedTokenIds.length} notification subscription${selectedTokenIds.length > 1 ? 's' : ''}`);
				} else {
					const request: any = {
						notification_type: 'mail',
						notification_classes: selectedClasses
					};

					if (subscriptionMode === 'user_wide') {
						request.user_wide = true;
					} else if (subscriptionMode === 'tags') {
						request.tags = selectedTags;
					}

					await api.createNotification(request);
					ui.success('Notification created');
				}
				dispatch('saved');
			}
			closeModal();
		} catch (error) {
			if (error instanceof ApiClientError) {
				ui.showError(isEditMode ? 'Failed to update notification' : 'Failed to create notification', error.description ?? error.code);
			}
		} finally {
			saving = false;
		}
	}

	function closeModal() {
		show = false;
		dispatch('cancel');
	}
</script>

{#if show}
	<div class="modal show d-block" tabindex="-1" role="dialog">
		<div class="modal-dialog modal-xl modal-dialog-scrollable">
			<div class="modal-content">
				<div class="modal-header">
					<h5 class="modal-title">
						<i class="fas {isEditMode ? 'fa-edit' : 'fa-plus-circle'} me-2"></i>
						{modalTitle}
					</h5>
					<button 
						type="button" 
						class="btn-close" 
						aria-label="Close"
						onclick={closeModal}
					></button>
				</div>
				<div class="modal-body">
					<!-- Notification Classes -->
					<div class="mb-4">
						<span class="form-label fw-bold d-block" id="modal-classes-label">
							<i class="fas fa-list-check me-1"></i>
							Notification Classes
						</span>
						<div class="notification-classes-list border rounded p-3" role="group" aria-labelledby="modal-classes-label">
							{#each topLevelClasses as cls}
								{@const childClasses = getChildClasses(cls.id)}
								{@const hasChildren = childClasses.length > 0}
								{@const isSelected = selectedClasses.includes(cls.id)}
								<div class="notification-class-item mb-2">
									<div class="form-check">
										<input 
											type="checkbox" 
											class="form-check-input" 
											id="modal-class-{cls.id}"
											checked={isSelected}
											onchange={() => toggleClass(cls.id)}
										/>
										<label class="form-check-label" for="modal-class-{cls.id}">
											<i class="fas {cls.icon} me-1" class:text-primary={isSelected}></i>
											<strong>{cls.label}</strong>
											<small class="text-muted ms-2">{cls.description}</small>
										</label>
									</div>

									<!-- Child classes -->
									{#if hasChildren}
										<div class="ms-4 mt-1">
											{#each childClasses as child}
												{@const grandchildClasses = getChildClasses(child.id)}
												{@const hasGrandchildren = grandchildClasses.length > 0}
												{@const isChildSelected = selectedClasses.includes(child.id)}
												<div class="form-check">
													<input 
														type="checkbox" 
														class="form-check-input" 
														id="modal-class-{child.id}"
														checked={isChildSelected}
														onchange={() => toggleClass(child.id)}
													/>
													<label class="form-check-label" for="modal-class-{child.id}">
														<i class="fas {child.icon} me-1" class:text-primary={isChildSelected}></i>
														{child.label}
													</label>
												</div>

												<!-- Grandchildren -->
												{#if hasGrandchildren}
													<div class="ms-4">
														{#each grandchildClasses as grandchild}
															{@const isGrandchildSelected = selectedClasses.includes(grandchild.id)}
															<div class="form-check">
																<input 
																	type="checkbox" 
																	class="form-check-input" 
																	id="modal-class-{grandchild.id}"
																	checked={isGrandchildSelected}
																	onchange={() => toggleClass(grandchild.id)}
																/>
																<label class="form-check-label" for="modal-class-{grandchild.id}">
																	<i class="fas {grandchild.icon} me-1" class:text-primary={isGrandchildSelected}></i>
																	{grandchild.label}
																</label>
															</div>
														{/each}
													</div>
												{/if}
											{/each}
										</div>
									{/if}
								</div>
							{/each}
						</div>
						{#if selectedClasses.length > 0}
							<small class="text-muted mt-1 d-block">
								{selectedClasses.length} class{selectedClasses.length > 1 ? 'es' : ''} selected
							</small>
						{/if}
					</div>

					<!-- Token Subscription -->
					<div class="mb-4">
						<span class="form-label fw-bold d-block" id="modal-apply-to-label">
							<i class="fas fa-link me-1"></i>
							Apply To
						</span>
						<ul class="nav nav-pills mb-3" aria-labelledby="modal-apply-to-label">
							<li class="nav-item">
								<button 
									type="button" 
									class="nav-link" 
									class:active={subscriptionMode === 'tags'}
									onclick={() => subscriptionMode = 'tags'}
								>
									<i class="fas fa-tags me-1"></i>
									By Tags
								</button>
							</li>
							<li class="nav-item">
								<button 
									type="button" 
									class="nav-link" 
									class:active={subscriptionMode === 'tokens'}
									onclick={() => subscriptionMode = 'tokens'}
								>
									<i class="fas fa-key me-1"></i>
									Select Tokens
								</button>
							</li>
							<li class="nav-item">
								<button 
									type="button" 
									class="nav-link" 
									class:active={subscriptionMode === 'user_wide'}
									onclick={() => subscriptionMode = 'user_wide'}
								>
									<i class="fas fa-user me-1"></i>
									All Tokens
								</button>
							</li>
						</ul>

						<div class="subscription-content border rounded p-3">
							{#if subscriptionMode === 'tags'}
								<p class="text-muted mb-2">
									Select tags to automatically subscribe matching tokens:
								</p>
								
								<!-- Selected tags with remove buttons -->
								{#if selectedTags.length > 0}
									<div class="selected-tags mb-2">
										{#each selectedTags as tagName}
											<TagPill 
												tag={tagName} 
												color={getTagColor(tagName)} 
												removable
												onRemove={() => removeTag(tagName)}
											/>
										{/each}
									</div>
								{/if}

								<!-- Tag dropdown + new tag input -->
								<div class="input-group">
									<select 
										class="form-select"
										bind:value={newTagInput}
										onchange={() => { if (newTagInput) addTag(newTagInput); }}
									>
										<option value="">Add existing tag...</option>
										{#each availableTagsForSelection as tag}
											<option value={tag.tag}>{tag.tag}</option>
										{/each}
									</select>
									<span class="input-group-text">or</span>
									<input 
										type="text" 
										class="form-control" 
										placeholder="New tag name"
										bind:value={newTagInput}
										onkeydown={(e) => { if (e.key === 'Enter') { e.preventDefault(); addNewTag(); } }}
									/>
									<button 
										type="button" 
										class="btn btn-outline-secondary"
										onclick={addNewTag}
										disabled={!newTagInput.trim()}
										title="Add tag"
									>
										<i class="fas fa-plus"></i>
									</button>
								</div>
								<small class="text-muted">Select an existing tag or type a new tag name and press Enter.</small>

							{:else if subscriptionMode === 'tokens'}
								<p class="text-muted mb-2">
									Select tokens to subscribe for notifications:
								</p>
								
								{#if loadingTokens}
									<LoadingSpinner size="sm" message="Loading tokens..." />
								{:else if tokenTrees.length > 0}
									<!-- Search filter -->
									<div class="mb-3">
										<input 
											type="text" 
											class="form-control form-control-sm" 
											placeholder="Search tokens..."
											bind:value={tokenSearchQuery}
										/>
									</div>

									<!-- Selected count -->
									{#if selectedTokenIds.length > 0}
										<div class="alert alert-info py-2 mb-2">
											<small>
												<i class="fas fa-check-circle me-1"></i>
												{selectedTokenIds.length} token{selectedTokenIds.length > 1 ? 's' : ''} selected
											</small>
										</div>
									{/if}

									<!-- Token list with tree structure -->
									<div class="token-list-container">
										<table class="table table-sm table-hover mb-0">
											<thead class="sticky-top bg-white">
												<tr>
													<th style="width: 40px;"></th>
													<th>Name</th>
													<th style="width: 130px;">Created</th>
													<th style="width: 130px;">Expires</th>
												</tr>
											</thead>
											<tbody>
												{#each visibleTokens as item}
													{@const token = item.token}
													{@const expired = isExpired(token)}
													<tr class:table-active={isTokenSelected(token.mom_id)} class:text-muted={expired}>
														<td>
															<input 
																type="checkbox" 
																class="form-check-input"
																checked={isTokenSelected(token.mom_id)}
																onchange={() => toggleTokenSelection(token.mom_id)}
															/>
														</td>
														<td 
															class:token-fold={item.hasChildren}
															onclick={() => item.hasChildren && toggleTokenExpand(item.id)}
															role={item.hasChildren ? 'button' : undefined}
														>
															<div class="d-flex align-items-center">
																<span style="margin-left: {item.depth * 1.2}rem;"></span>
																{#if item.hasChildren}
																	<i class="fas me-1" class:fa-caret-right={!isTokenExpanded(item.id)} class:fa-caret-down={isTokenExpanded(item.id)}></i>
																{:else}
																	<span style="width: 0.8rem; display: inline-block;"></span>
																{/if}
																<span class="token-name" class:token-name-unnamed={!token.name}>
																	{token.name || 'unnamed token'}
																</span>
																{#if token.tags && token.tags.length > 0}
																	<span class="ms-2">
																		{#each token.tags.slice(0, 2) as tag}
																			<TagPill tag={tag.tag} color={tag.color} small />
																		{/each}
																		{#if token.tags.length > 2}
																			<span class="badge bg-secondary">+{token.tags.length - 2}</span>
																		{/if}
																	</span>
																{/if}
															</div>
														</td>
														<td>
															<small class="text-muted">
																{formatDateTime(token.created)}
															</small>
														</td>
														<td>
															<small class:text-danger={expired} class:text-muted={!expired}>
																{formatExpiry(token)}
															</small>
														</td>
													</tr>
												{/each}
											</tbody>
										</table>
									</div>

									{#if visibleTokens.length === 0}
										<div class="text-center text-muted py-3">
											<i class="fas fa-search me-1"></i>
											No tokens match your search
										</div>
									{/if}
								{:else}
									<div class="alert alert-info mb-0">
										<i class="fas fa-info-circle me-2"></i>
										No tokens available.
									</div>
								{/if}

							{:else if subscriptionMode === 'user_wide'}
								<div class="alert alert-info mb-0">
									<i class="fas fa-info-circle me-2"></i>
									This notification will apply to <strong>all your tokens</strong>, including any tokens created in the future.
								</div>
							{/if}
						</div>
					</div>
				</div>
				<div class="modal-footer">
					<button 
						type="button" 
						class="btn btn-secondary" 
						onclick={closeModal}
						disabled={saving}
					>
						Cancel
					</button>
					<button 
						type="button" 
						class="btn btn-primary" 
						onclick={handleSubmit}
						disabled={saving || selectedClasses.length === 0}
						title={selectedClasses.length === 0 ? 'Please select at least one notification class' : ''}
					>
						{#if saving}
							<span class="spinner-border spinner-border-sm me-1"></span>
							{isEditMode ? 'Saving...' : 'Creating...'}
						{:else}
							<i class="fas {isEditMode ? 'fa-save' : 'fa-plus'} me-1"></i>
							{isEditMode ? 'Save Changes' : 'Create Notification'}
						{/if}
					</button>
				</div>
			</div>
		</div>
	</div>
	<div class="modal-backdrop show"></div>
{/if}

<style>
	.modal.show {
		background-color: rgba(0, 0, 0, 0.5);
	}

	.modal-backdrop {
		z-index: 1040;
	}

	.modal {
		z-index: 1050;
	}

	.notification-classes-list {
		background-color: var(--bs-tertiary-bg);
	}

	.nav-pills .nav-link {
		color: var(--bs-secondary-color);
	}

	.nav-pills .nav-link.active {
		background-color: var(--mytoken-primary);
	}

	.subscription-content {
		background-color: var(--bs-tertiary-bg);
	}

	.selected-tags {
		display: flex;
		flex-wrap: wrap;
		gap: 0.25rem;
		padding: 0.5rem;
		background-color: var(--bs-body-bg);
		border-radius: 0.375rem;
		margin-bottom: 0.5rem;
	}

	.input-group .form-select {
		max-width: 200px;
	}

	.form-check-label {
		cursor: pointer;
	}

	.token-list-container {
		max-height: 300px;
		overflow-y: auto;
		background-color: var(--bs-body-bg);
		border-radius: 0.375rem;
	}

	.token-list-container .table {
		margin-bottom: 0;
	}

	.token-list-container th {
		font-weight: 500;
		font-size: 0.85rem;
		border-top: none;
	}

	.token-name {
		font-weight: 500;
	}

	.token-name-unnamed {
		font-style: italic;
		color: var(--bs-secondary-color);
	}

	.token-fold {
		cursor: pointer;
	}

	.token-fold:hover {
		background-color: rgba(var(--bs-body-color-rgb), 0.05);
	}

	.table-active {
		background-color: rgba(var(--mytoken-primary-rgb), 0.1) !important;
	}
</style>
