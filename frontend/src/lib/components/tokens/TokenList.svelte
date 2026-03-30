<script lang="ts">
	import { onMount } from 'svelte';
	import type { MytokenEntry, MytokenEntryTree, MTTagInfo, Tag } from '$lib/types';
	import { api, ApiClientError } from '$lib/api/client';
	import { ui } from '$lib/stores/ui';
	import { tags } from '$lib/stores/tags';
	import { usersettingsEndpoint, notificationsEndpoint } from '$lib/stores/discovery';
	import { formatDateTime } from '$lib/utils/format';
	import LoadingSpinner from '../LoadingSpinner.svelte';
	import TagPill from '../TagPill.svelte';
	import TokenTagsCell from './TokenTagsCell.svelte';
	import EventHistoryModal from '../EventHistoryModal.svelte';
	import TokenNotificationModal from './TokenNotificationModal.svelte';

	let eventHistoryModal: EventHistoryModal;
	
	// Available tags from user settings (for adding to tokens)
	let availableTags: Tag[] = $derived($tags.tags);
	
	// Notification modal state
	let showNotificationModal = $state(false);
	let notificationToken: MytokenEntry | null = $state(null);
	
	// Check if notifications are supported
	let notificationsSupported = $derived(!!$notificationsEndpoint);

	// Flattened entry with depth info for display
	interface FlattenedToken {
		token: MytokenEntry;
		depth: number;
		hasChildren: boolean;
		id: string;
		parentId: string | null;
	}

	// Sortable columns
	type SortColumn = 'name' | 'created' | 'ip' | 'expires';
	type SortDirection = 'asc' | 'desc';

	let tokenTrees: MytokenEntryTree[] = $state([]);
	let loading = $state(true);
	let error = $state('');

	// Track expanded state for each token
	let expandedTokens: Set<string> = $state(new Set());

	// Filtering
	let searchQuery = $state('');
	let selectedTag = $state('');

	// Sorting - default to created date descending (newest first)
	let sortColumn: SortColumn = $state('created');
	let sortDirection: SortDirection = $state('desc');

	// Collect all unique tags for the filter dropdown
	let allTags = $derived.by(() => {
		const tags = new Map<string, MTTagInfo>();
		const traverse = (tree: MytokenEntryTree) => {
			if (tree.token.tags) {
				for (const tag of tree.token.tags) {
					tags.set(tag.tag, tag);
				}
			}
			if (tree.children) {
				tree.children.forEach(traverse);
			}
		};
		tokenTrees.forEach(traverse);
		return tags;
	});

	// Flatten tree to array with depth info
	let flattenedTokens = $derived.by(() => {
		const result: FlattenedToken[] = [];
		let idCounter = 0;
		
		const traverse = (tree: MytokenEntryTree, depth: number, parentId: string | null) => {
			const id = `token-${idCounter++}`;
			const hasChildren = (tree.children?.length ?? 0) > 0;
			
			result.push({
				token: tree.token,
				depth,
				hasChildren,
				id,
				parentId
			});
			
			if (tree.children) {
				for (const child of tree.children) {
					traverse(child, depth + 1, id);
				}
			}
		};
		
		tokenTrees.forEach(t => traverse(t, 0, null));
		return result;
	});

	// Check if a token is visible (parent is expanded or it's a root)
	function isVisible(token: FlattenedToken, tokens: FlattenedToken[], expanded: Set<string>): boolean {
		if (token.parentId === null) return true;
		
		// Find all ancestors and check if they're expanded
		let current = token;
		while (current.parentId !== null) {
			if (!expanded.has(current.parentId)) {
				return false;
			}
			// Find parent
			const parent = tokens.find(t => t.id === current.parentId);
			if (!parent) break;
			current = parent;
		}
		return true;
	}

	// Sort comparator function
	function compareTokens(a: FlattenedToken, b: FlattenedToken, column: SortColumn, direction: SortDirection): number {
		let comparison = 0;
		
		switch (column) {
			case 'name': {
				const nameA = (a.token.name || 'unnamed token').toLowerCase();
				const nameB = (b.token.name || 'unnamed token').toLowerCase();
				comparison = nameA.localeCompare(nameB);
				break;
			}
			case 'created': {
				comparison = (a.token.created || 0) - (b.token.created || 0);
				break;
			}
			case 'ip': {
				const ipA = a.token.ip || '';
				const ipB = b.token.ip || '';
				comparison = ipA.localeCompare(ipB);
				break;
			}
			case 'expires': {
				// Tokens that don't expire go to the end
				const expA = a.token.expires_at || Number.MAX_SAFE_INTEGER;
				const expB = b.token.expires_at || Number.MAX_SAFE_INTEGER;
				comparison = expA - expB;
				break;
			}
		}
		
		return direction === 'asc' ? comparison : -comparison;
	}

	// Toggle sort column/direction
	function toggleSort(column: SortColumn) {
		if (sortColumn === column) {
			// Toggle direction if same column
			sortDirection = sortDirection === 'asc' ? 'desc' : 'asc';
		} else {
			// New column, default to ascending (except for dates which default to descending)
			sortColumn = column;
			sortDirection = (column === 'created' || column === 'expires') ? 'desc' : 'asc';
		}
	}

	// Get sort icon for a column
	function getSortIcon(column: SortColumn): string {
		if (sortColumn !== column) return 'fa-sort';
		return sortDirection === 'asc' ? 'fa-sort-up' : 'fa-sort-down';
	}

	// Filter and sort visible tokens
	let visibleTokens = $derived.by(() => {
		// First filter
		const filtered = flattenedTokens.filter(t => {
			// Name filter
			if (searchQuery) {
				const q = searchQuery.toLowerCase();
				const name = (t.token.name || 'unnamed token').toLowerCase();
				if (!name.includes(q)) {
					return false;
				}
			}
			// Tag filter
			if (selectedTag) {
				if (!t.token.tags?.some(tokenTag => tokenTag.tag === selectedTag)) {
					return false;
				}
			}
			// Visibility (parent expanded)
			return isVisible(t, flattenedTokens, expandedTokens);
		});
		
		// Then sort (only root-level tokens, preserving tree structure)
		// We need to sort root tokens and keep children with their parents
		const rootTokens = filtered.filter(t => t.depth === 0);
		const sortedRoots = [...rootTokens].sort((a, b) => compareTokens(a, b, sortColumn, sortDirection));
		
		// Rebuild the list maintaining parent-child relationships
		const result: FlattenedToken[] = [];
		for (const root of sortedRoots) {
			result.push(root);
			// Add all descendants of this root (they're already in order from flattening)
			const descendants = filtered.filter(t => {
				if (t.depth === 0) return false;
				// Check if this token is a descendant of the current root
				let current = t;
				while (current.parentId !== null) {
					const parent = flattenedTokens.find(p => p.id === current.parentId);
					if (!parent) break;
					if (parent.id === root.id) return true;
					current = parent;
				}
				return false;
			});
			result.push(...descendants);
		}
		
		return result;
	});

	onMount(async () => {
		await loadTokens();
		// Fetch available tags once for the tag selector
		const endpoint = $usersettingsEndpoint;
		if (endpoint && !$tags.loaded) {
			tags.fetch(endpoint);
		}
	});

	async function loadTokens() {
		loading = true;
		error = '';

		try {
			tokenTrees = await api.listMytokensTree();
			// Expand all root tokens by default
			expandedTokens = new Set();
		} catch (err) {
			if (err instanceof ApiClientError) {
				if (err.status === 401) {
					error = 'Please log in to view your tokens';
				} else {
					error = err.description ?? err.code;
				}
			} else {
				error = (err as Error).message;
			}
		} finally {
			loading = false;
		}
	}

	function toggleExpand(id: string) {
		if (expandedTokens.has(id)) {
			expandedTokens.delete(id);
		} else {
			expandedTokens.add(id);
		}
		// Create new Set to trigger reactivity with $state
		expandedTokens = new Set(expandedTokens);
	}

	function isExpanded(id: string): boolean {
		return expandedTokens.has(id);
	}

	function isExpired(token: MytokenEntry): boolean {
		if (!token.expires_at || token.expires_at === 0) return false;
		return new Date(token.expires_at * 1000) < new Date();
	}

	function formatExpiry(token: MytokenEntry): string {
		if (!token.expires_at || token.expires_at === 0) {
			return 'Does not expire';
		}
		return formatDateTime(token.expires_at);
	}

	async function revokeToken(momId: string, name?: string) {
		const displayName = name || 'unnamed token';
		const confirmed = await ui.confirm({
			title: 'Revoke Token',
			message: `Are you sure you want to revoke "${displayName}"? This cannot be undone.`,
			confirmText: 'Revoke',
			confirmVariant: 'danger'
		});
		
		if (!confirmed) return;

		try {
			await api.revokeToken(momId, false);
			ui.success(`Token "${displayName}" revoked`);
			await loadTokens(); // Reload the list
		} catch (err) {
			if (err instanceof ApiClientError) {
				ui.showError('Revocation Failed', err.description ?? err.code);
			}
		}
	}

	function showHistory(momId: string, name?: string) {
		eventHistoryModal.open(momId, name);
	}

	function showNotifications(token: MytokenEntry) {
		notificationToken = token;
		showNotificationModal = true;
	}
</script>

<div class="token-list">
	{#if loading}
		<LoadingSpinner message="Loading your tokens..." />
	{:else if error}
		<div class="alert alert-danger">
			<i class="fas fa-exclamation-triangle me-2"></i>
			{error}
		</div>
	{:else if tokenTrees.length === 0}
		<div class="text-center py-5 text-muted">
			<i class="fas fa-key fa-3x mb-3"></i>
			<p>You don't have any tokens yet</p>
			<a href="/#mt" class="btn btn-primary">
				<i class="fas fa-plus-circle me-1"></i>
				Create Your First Token
			</a>
		</div>
	{:else}
		<!-- Token table -->
		<div class="table-responsive">
			<table class="table table-hover align-middle">
				<thead>
					<tr>
						<th style="min-width: 30%;">
							<span 
								class="sortable-header" 
								role="button"
								tabindex="0"
								onclick={() => toggleSort('name')} 
								onkeydown={(e) => e.key === 'Enter' && toggleSort('name')}
								title="Sort by name"
							>
								Token Name
								<i class="fas {getSortIcon('name')} ms-1"></i>
							</span>
							<input
								type="text"
								class="form-control form-control-sm d-inline-block ms-2"
								style="width: 150px;"
								placeholder="Search by name"
								bind:value={searchQuery}
							/>
						</th>
						<th>
							<span>Tags</span>
							<select
								class="form-select form-select-sm d-inline-block ms-2"
								style="width: auto;"
								bind:value={selectedTag}
							>
								<option value="">All</option>
								{#each [...allTags.values()] as tag}
									<option value={tag.tag}>{tag.tag}</option>
								{/each}
							</select>
						</th>
						<th>
							<span 
								class="sortable-header" 
								role="button"
								tabindex="0"
								onclick={() => toggleSort('created')} 
								onkeydown={(e) => e.key === 'Enter' && toggleSort('created')}
								title="Sort by creation date"
							>
								Created
								<i class="fas {getSortIcon('created')} ms-1"></i>
							</span>
						</th>
						<th>
							<span 
								class="sortable-header" 
								role="button"
								tabindex="0"
								onclick={() => toggleSort('ip')} 
								onkeydown={(e) => e.key === 'Enter' && toggleSort('ip')}
								title="Sort by IP"
							>
								Created from IP
								<i class="fas {getSortIcon('ip')} ms-1"></i>
							</span>
						</th>
						<th>
							<span 
								class="sortable-header" 
								role="button"
								tabindex="0"
								onclick={() => toggleSort('expires')} 
								onkeydown={(e) => e.key === 'Enter' && toggleSort('expires')}
								title="Sort by expiration"
							>
								Expires
								<i class="fas {getSortIcon('expires')} ms-1"></i>
							</span>
						</th>
						<th>
							<button
								class="btn btn-sm btn-outline-primary"
								onclick={loadTokens}
								disabled={loading}
								title="Refresh"
							>
								<i class="fas fa-sync" class:fa-spin={loading}></i>
							</button>
						</th>
					</tr>
				</thead>
				<tbody>
					{#each visibleTokens as item (item.id)}
						{@const token = item.token}
						{@const expired = isExpired(token)}
						<tr class:token-expired={expired}>
							<td 
								class:token-fold={item.hasChildren}
								onclick={() => item.hasChildren && toggleExpand(item.id)}
								onkeydown={(e) => e.key === 'Enter' && item.hasChildren && toggleExpand(item.id)}
								role={item.hasChildren ? 'button' : undefined}
								tabindex={item.hasChildren ? 0 : undefined}
							>
								<span style="margin-left: {item.depth * 1.5}rem;"></span>
								{#if item.hasChildren}
									<i class="fas me-2" class:fa-caret-right={!isExpanded(item.id)} class:fa-caret-down={isExpanded(item.id)}></i>
								{:else}
									<span style="width: 1rem; display: inline-block;"></span>
								{/if}
								<span class:token-name-unnamed={!token.name}>
									{token.name || 'unnamed token'}
								</span>
							</td>
							<td>
								<TokenTagsCell 
									momId={token.mom_id}
									tokenTags={token.tags ?? []}
									availableTags={availableTags}
									onTagsChanged={loadTokens}
								/>
							</td>
							<td>{formatDateTime(token.created)}</td>
							<td>{token.ip || '-'}</td>
							<td class:text-muted={!token.expires_at || token.expires_at === 0}>
								{formatExpiry(token)}
								{#if expired}
									<span class="badge bg-secondary ms-1" title="This token has expired">Expired</span>
								{/if}
							</td>
							<td>
								<div class="btn-group btn-group-sm">
									{#if notificationsSupported}
										<button
											type="button"
											class="btn btn-outline-secondary"
											class:text-muted={expired}
											title={expired ? 'Token expired' : 'Notifications'}
											onclick={() => showNotifications(token)}
											disabled={expired}
										>
											<i class="fas fa-bell"></i>
										</button>
									{/if}
									<button
										type="button"
										class="btn btn-outline-secondary"
										title="Event History"
										onclick={() => showHistory(token.mom_id, token.name)}
									>
										<i class="fas fa-history"></i>
									</button>
									<button
										type="button"
										class="btn btn-outline-danger"
										title="Revoke token"
										onclick={() => revokeToken(token.mom_id, token.name)}
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

		<!-- Summary -->
		<div class="text-muted mt-3">
			<small>
				{#if searchQuery || selectedTag}
					Showing {visibleTokens.length} filtered tokens
				{:else}
					{tokenTrees.length} root token{tokenTrees.length !== 1 ? 's' : ''}, {flattenedTokens.length} total
				{/if}
			</small>
		</div>
	{/if}
</div>

<EventHistoryModal bind:this={eventHistoryModal} />

<TokenNotificationModal 
	bind:show={showNotificationModal}
	token={notificationToken}
	onChanged={loadTokens}
/>

<style>
	.table th {
		border-top: none;
		font-weight: 500;
		white-space: nowrap;
	}

	.table th input,
	.table th select {
		font-weight: normal;
	}

	.sortable-header {
		cursor: pointer;
		user-select: none;
	}

	.sortable-header:hover {
		color: var(--mytoken-primary);
	}

	.sortable-header i {
		font-size: 0.75em;
		opacity: 0.5;
	}

	.sortable-header i.fa-sort-up,
	.sortable-header i.fa-sort-down {
		opacity: 1;
		color: var(--mytoken-primary);
	}

	.token-name-unnamed {
		font-style: italic;
		color: var(--bs-secondary-color);
	}

	.token-fold {
		cursor: pointer;
	}

	.token-fold:hover {
		background-color: rgba(var(--bs-body-color-rgb), 0.02);
	}

	.btn-group-sm .btn {
		padding: 0.25rem 0.5rem;
	}

	/* Expired token styling */
	:global(tr.token-expired) {
		opacity: 0.6;
		background-color: rgba(var(--bs-secondary-rgb), 0.05);
	}

	:global(tr.token-expired td) {
		color: var(--bs-secondary-color);
	}

	:global([data-bs-theme='dark'] tr.token-expired) {
		background-color: rgba(0, 0, 0, 0.15);
	}
</style>
