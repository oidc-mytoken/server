<script lang="ts">
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { api, ApiClientError } from '$lib/api/client';
	import { isLoggedIn, auth, tokenInfo } from '$lib/stores/auth';
	import { tags as tagsStore } from '$lib/stores/tags';
	import { discovery } from '$lib/stores/discovery';
	import { ui } from '$lib/stores/ui';
	import { NOTIFICATION_CLASSES, type Notification, type MytokenEntryTree, type MytokenEntry, type TagInfo, type NotificationClass } from '$lib/types';
	import { getNotificationTypeIcon, getNotificationTypeLabel } from '$lib/utils/notifications';
	import { formatDateTime } from '$lib/utils/format';
	import { generateTagColor } from '$lib/utils/color';
	import TagPill from '$lib/components/TagPill.svelte';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
	import CollapsibleSection from '$lib/components/CollapsibleSection.svelte';
	import NotificationClassIcon from '$lib/components/notifications/NotificationClassIcon.svelte';

	// Get management code from URL
	$: managementCode = $page.params.mc;

	// Notification data
	let notification: Notification | null = null;
	let loading = true;
	let error: string | null = null;

	// Notification classes state
	let selectedClasses: string[] = [];
	let savingClasses = false;

	// Tags state
	let notificationTags: string[] = [];
	let savingTags = false;
	let newTagInput = '';

	// Tokens state (only for non-user-wide notifications)
	let subscribedTokens: MytokenEntry[] = [];
	let tagLinkedTokens: MytokenEntry[] = [];
	let allUserTokens: MytokenEntryTree[] = [];
	let loadingTokens = false;
	let showAddTokenModal = false;
	let selectedTokensToAdd: Set<string> = new Set();
	let includeChildrenMap: Map<string, boolean> = new Map();
	let expandedTokens: Set<string> = new Set();
	let addingTokens = false;
	let tokenSearchQuery = '';

	// Check if logged-in user is the owner of this notification
	// If not logged in, or notification doesn't have oidc info, treat as owner (no restrictions)
	$: isOwner = !$isLoggedIn || !notification?.oidc_iss || !notification?.oidc_sub || (
		$tokenInfo?.token?.oidc_iss === notification.oidc_iss &&
		$tokenInfo?.token?.oidc_sub === notification.oidc_sub
	);

	// Get root notification classes (those without a parent)
	const rootClasses = NOTIFICATION_CLASSES.filter(c => !c.parent);

	// Get child classes for a parent
	function getChildClasses(parentId: string): NotificationClass[] {
		return NOTIFICATION_CLASSES.filter((c: NotificationClass) => c.parent === parentId);
	}

	// Get all child class IDs recursively
	function getAllChildClassIds(parentId: string): string[] {
		const children = NOTIFICATION_CLASSES.filter(c => c.parent === parentId);
		let result: string[] = [];
		for (const child of children) {
			result.push(child.id);
			result = [...result, ...getAllChildClassIds(child.id)];
		}
		return result;
	}

	// Get all parent class IDs
	function getAllParentClassIds(classId: string): string[] {
		const cls = NOTIFICATION_CLASSES.find(c => c.id === classId);
		if (!cls?.parent) return [];
		return [cls.parent, ...getAllParentClassIds(cls.parent)];
	}

	// Toggle a notification class - matches main page behavior
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
		
		selectedClasses = newSelection;
	}

	// Available tags from user's tags
	$: availableTags = $tagsStore.tags || [];
	$: availableTagsForSelection = availableTags.filter(t => !notificationTags.includes(t.tag));

	// Flatten token tree for display
	interface FlattenedToken {
		token: MytokenEntry;
		depth: number;
		hasChildren: boolean;
		id: string;
	}

	// Sort trees by creation time (descending - newest first)
	function sortTreesByCreated(trees: MytokenEntryTree[]): MytokenEntryTree[] {
		const sorted = [...trees].sort((a, b) => (b.token.created || 0) - (a.token.created || 0));
		// Recursively sort children
		return sorted.map(tree => ({
			...tree,
			children: tree.children ? sortTreesByCreated(tree.children) : undefined
		}));
	}

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

	$: flattenedTokens = flattenTokenTrees(allUserTokens);

	// Check if a token is visible (considering tree expansion)
	function isTokenVisible(item: FlattenedToken, tokens: FlattenedToken[], expanded: Set<string>): boolean {
		// Root level tokens are always visible
		if (item.depth === 0) return true;
		
		// Check if all ancestors are expanded
		let currentDepth = item.depth - 1;
		let currentIndex = tokens.indexOf(item) - 1;
		
		while (currentDepth >= 0 && currentIndex >= 0) {
			const ancestor = tokens[currentIndex];
			if (ancestor.depth === currentDepth) {
				if (!expanded.has(ancestor.id)) {
					return false;
				}
				currentDepth--;
			}
			currentIndex--;
		}
		
		return true;
	}

	// Get visible tokens for the add modal (respecting expanded state and search)
	$: visibleTokens = flattenedTokens.filter(item => {
		// Exclude expired tokens (unless they are already directly subscribed)
		const directSubscribedIds = notification?.subscribed_tokens ?? [];
		if (isExpired(item.token) && !directSubscribedIds.includes(item.token.mom_id)) {
			return false;
		}
		
		// Filter by search query if present
		if (tokenSearchQuery) {
			const query = tokenSearchQuery.toLowerCase();
			const name = (item.token.name ?? '').toLowerCase();
			const momId = item.token.mom_id.toLowerCase();
			if (!name.includes(query) && !momId.includes(query)) {
				return false;
			}
		}
		
		// Check tree visibility
		return isTokenVisible(item, flattenedTokens, expandedTokens);
	});

	// Filter tokens not already directly subscribed for the add modal
	$: availableTokensToAdd = visibleTokens.filter(t => {
		const directSubscribedIds = notification?.subscribed_tokens ?? [];
		return !directSubscribedIds.includes(t.token.mom_id);
	});

	function isExpired(token: MytokenEntry): boolean {
		if (!token.expires_at || token.expires_at === 0) return false;
		return new Date(token.expires_at * 1000) < new Date();
	}

	// Load notification data
	async function loadNotification() {
		if (!managementCode) {
			error = 'No management code provided';
			loading = false;
			return;
		}
		loading = true;
		error = null;
		try {
			notification = await api.getNotificationByManagementCode(managementCode);
			selectedClasses = [...notification.notification_classes];
			notificationTags = notification.tags?.map(t => t.tag) || [];
			// Token loading is handled by the reactive statement that checks isOwner
		} catch (e) {
			if (e instanceof ApiClientError) {
				error = e.description ?? e.code;
			} else {
				error = 'Failed to load notification';
			}
		} finally {
			loading = false;
		}
	}

	// Load user's tokens for the subscribed tokens list
	async function loadTokens() {
		loadingTokens = true;
		try {
			allUserTokens = await api.listMytokensTree();
			updateTokenLists();
		} catch (e) {
			console.error('Failed to load tokens:', e);
		} finally {
			loadingTokens = false;
		}
	}

	// Update the subscribed and tag-linked token lists
	function updateTokenLists() {
		const allFlattened = flattenTokenTrees(allUserTokens);
		const directSubscribedIds = new Set(notification?.subscribed_tokens ?? []);
		
		// Build directly subscribed tokens list
		subscribedTokens = allFlattened
			.filter(t => directSubscribedIds.has(t.token.mom_id))
			.map(t => t.token);
		
		// Build tag-linked tokens list (tokens that have matching tags but aren't directly subscribed)
		if (notificationTags.length > 0) {
			tagLinkedTokens = allFlattened
				.filter(t => {
					// Exclude tokens that are directly subscribed
					if (directSubscribedIds.has(t.token.mom_id)) return false;
					// Check if token has any matching tags
					const tokenTags = t.token.tags?.map(tag => tag.tag) ?? [];
					return tokenTags.some(tt => notificationTags.includes(tt));
				})
				.map(t => t.token);
		} else {
			tagLinkedTokens = [];
		}
	}

	// React to notificationTags changes to update tag-linked tokens
	// Using JSON.stringify to ensure deep reactivity tracking
	$: notificationTagsKey = JSON.stringify(notificationTags);
	$: if (notification && allUserTokens.length > 0 && notificationTagsKey) {
		updateTokenLists();
	}

	// Save notification classes
	async function saveClasses() {
		if (!notification || !managementCode) return;
		savingClasses = true;
		try {
			await api.updateNotification(managementCode, {
				notification_classes: selectedClasses,
				tags: notificationTags
			});
			notification.notification_classes = [...selectedClasses];
			ui.success('Notification classes updated');
		} catch (e) {
			if (e instanceof ApiClientError) {
				ui.showError('Failed to update', e.description ?? e.code);
			}
		} finally {
			savingClasses = false;
		}
	}

	// Get tag color - check notification tags first, then available tags, fallback to hash-based
	function getTagColor(tagName: string): string {
		// First check the notification's own tag colors
		const notificationTag = notification?.tags?.find(t => t.tag === tagName);
		if (notificationTag?.color) {
			return notificationTag.color;
		}
		// Then check user's available tags
		const availableTag = availableTags.find(t => t.tag === tagName);
		if (availableTag?.color) {
			return availableTag.color;
		}
		// Fallback to hash-based color
		return generateTagColor(tagName);
	}

	// Add a tag to the notification
	function addTag(tag: string) {
		if (tag && !notificationTags.includes(tag)) {
			notificationTags = [...notificationTags, tag];
		}
		newTagInput = '';
	}

	// Add a new tag (create it if it doesn't exist)
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

	// Remove a tag from the notification
	function removeTag(tag: string) {
		notificationTags = notificationTags.filter(t => t !== tag);
	}

	// Save tags
	async function saveTags() {
		if (!notification || !managementCode) return;
		savingTags = true;
		try {
			await api.updateNotification(managementCode, {
				notification_classes: selectedClasses,
				tags: notificationTags
			});
			notification.tags = notificationTags.map(t => {
				const tagInfo = availableTags.find(at => at.tag === t);
				return { tag: t, color: tagInfo?.color || '#6c757d' };
			});
			ui.success('Tags updated');
		} catch (e) {
			if (e instanceof ApiClientError) {
				ui.showError('Failed to update tags', e.description ?? e.code);
			}
		} finally {
			savingTags = false;
		}
	}

	// Remove a token from the notification
	async function removeToken(momId: string) {
		if (!notification || !managementCode) return;
		
		// Check if this is the last token and there are no tags
		const isLastToken = subscribedTokens.length === 1;
		const hasTags = notificationTags.length > 0;
		
		if (isLastToken && !hasTags) {
			const confirmed = await ui.confirm({
				title: 'Delete Notification?',
				message: 'This is the last token subscribed to this notification and there are no tags. Removing it will delete the notification. Do you want to proceed?'
			});
			if (confirmed) {
				await deleteNotification();
			}
			return;
		}

		try {
			await api.removeTokenFromNotification(managementCode, momId);
			subscribedTokens = subscribedTokens.filter(t => t.mom_id !== momId);
			if (notification.subscribed_tokens) {
				notification.subscribed_tokens = notification.subscribed_tokens.filter(id => id !== momId);
			}
			ui.success('Token removed from notification');
		} catch (e) {
			if (e instanceof ApiClientError) {
				ui.showError('Failed to remove token', e.description ?? e.code);
			}
		}
	}

	// Open add token modal
	function openAddTokenModal() {
		selectedTokensToAdd = new Set();
		includeChildrenMap = new Map();
		tokenSearchQuery = '';
		expandedTokens = new Set();
		showAddTokenModal = true;
	}

	// Toggle token selection in add modal
	function toggleTokenSelection(momId: string) {
		if (selectedTokensToAdd.has(momId)) {
			selectedTokensToAdd.delete(momId);
			includeChildrenMap.delete(momId);
		} else {
			selectedTokensToAdd.add(momId);
			includeChildrenMap.set(momId, true); // Default to include children
		}
		selectedTokensToAdd = new Set(selectedTokensToAdd);
	}

	// Toggle include children for a token
	function toggleIncludeChildren(momId: string) {
		const current = includeChildrenMap.get(momId) ?? true;
		includeChildrenMap.set(momId, !current);
		includeChildrenMap = new Map(includeChildrenMap);
	}

	// Toggle token tree expansion
	function toggleExpand(id: string) {
		if (expandedTokens.has(id)) {
			expandedTokens.delete(id);
		} else {
			expandedTokens.add(id);
		}
		expandedTokens = new Set(expandedTokens);
	}

	// Add selected tokens to notification
	async function addTokensToNotification() {
		if (selectedTokensToAdd.size === 0 || !managementCode) return;
		
		addingTokens = true;
		try {
			for (const momId of selectedTokensToAdd) {
				const includeChildren = includeChildrenMap.get(momId) ?? true;
				await api.addTokenToNotification(managementCode, momId, includeChildren);
			}
			showAddTokenModal = false;
			await loadNotification(); // Reload to get updated subscribed tokens
			ui.success('Tokens added to notification');
		} catch (e) {
			if (e instanceof ApiClientError) {
				ui.showError('Failed to add tokens', e.description ?? e.code);
			}
		} finally {
			addingTokens = false;
		}
	}

	// Delete notification
	async function deleteNotification() {
		if (!managementCode) return;
		
		const confirmed = await ui.confirm({
			title: 'Delete Notification?',
			message: 'Are you sure you want to delete this notification? This action cannot be undone.'
		});
		if (!confirmed) return;

		try {
			await api.deleteNotification(managementCode);
			ui.success('Notification deleted');
			goto('/');
		} catch (e) {
			if (e instanceof ApiClientError) {
				ui.showError('Failed to delete notification', e.description ?? e.code);
			}
		}
	}

	// Handle login button click
	function handleLogin() {
		if (notification?.oidc_iss) {
			// Store current URL to return after login
			sessionStorage.setItem('returnUrl', window.location.href);
			api.login(notification.oidc_iss).then(res => {
				if (res.authorization_uri) {
					window.location.href = res.authorization_uri;
				} else if (res.consent_uri) {
					window.location.href = res.consent_uri;
				}
			}).catch(e => {
				ui.showError('Login failed', e.description ?? e.code);
			});
		}
	}

	// Initialize
	onMount(async () => {
		// Wait for discovery to load
		if (!$discovery.data) {
			await discovery.fetch();
		}
		// Load tags if logged in (force load even if already loaded to ensure fresh data)
		if ($isLoggedIn && $discovery.data?.usersettings_endpoint) {
			await tagsStore.fetch($discovery.data.usersettings_endpoint);
		}
		await loadNotification();
	});

	// Load tokens when user logs in as owner (if notification is already loaded and not user-wide)
	$: if ($isLoggedIn && isOwner && notification && !notification.user_wide && allUserTokens.length === 0 && !loadingTokens) {
		loadTokens();
	}
</script>

<svelte:head>
	<title>Notification Management - mytoken</title>
</svelte:head>

<div class="notification-management">
	<h2 class="mb-4">
		<i class="fas fa-bell me-2"></i>
		Notification Management
	</h2>

	{#if loading}
		<LoadingSpinner message="Loading notification..." />
	{:else if error}
		<div class="alert alert-danger">
			<i class="fas fa-exclamation-triangle me-2"></i>
			{error}
		</div>
		<a href="/" class="btn btn-primary">
			<i class="fas fa-home me-2"></i>
			Go Home
		</a>
	{:else if notification}
		<!-- Notification Summary -->
		<div class="card mb-4">
			<div class="card-body">
				<div class="d-flex align-items-center justify-content-between">
					<div>
						<span class="badge bg-secondary me-2">
							<i class="fas {getNotificationTypeIcon(notification.notification_type)} me-1"></i>
							{getNotificationTypeLabel(notification.notification_type)}
						</span>
						{#if notification.user_wide}
							<span class="badge bg-info">User-wide</span>
						{/if}
					</div>
					<button class="btn btn-outline-danger btn-sm" on:click={deleteNotification}>
						<i class="fas fa-trash me-1"></i>
						Delete Notification
					</button>
				</div>
			</div>
		</div>

		<!-- Warning for different user -->
		{#if $isLoggedIn && !isOwner}
			<div class="alert alert-warning mb-4">
				<i class="fas fa-exclamation-triangle me-2"></i>
				<strong>Different Account:</strong> You are logged in with a different account than the one that owns this notification. 
				Token and tag management is not available.
			</div>
		{/if}

		<!-- Notification Classes Section -->
		<CollapsibleSection title="Notification Classes" icon="fa-bell" collapsed={false}>
			<div class="notification-classes-list border rounded p-3" role="group">
				{#each rootClasses as cls}
					{@const childClasses = getChildClasses(cls.id)}
					{@const hasChildren = childClasses.length > 0}
					{@const isSelected = selectedClasses.includes(cls.id)}
					<div class="notification-class-item mb-2">
						<div class="form-check">
							<input
								type="checkbox"
								class="form-check-input"
								id="class-{cls.id}"
								checked={isSelected}
								on:change={() => toggleClass(cls.id)}
							/>
							<label class="form-check-label" for="class-{cls.id}">
								<NotificationClassIcon icon={cls.icon} extraClass="me-1{isSelected ? ' text-primary' : ''}" />
								<strong>{cls.label}</strong>
								<small class="text-muted ms-2">{cls.description}</small>
							</label>
						</div>

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
											id="class-{child.id}"
											checked={isChildSelected}
											on:change={() => toggleClass(child.id)}
										/>
									<label class="form-check-label" for="class-{child.id}">
										<NotificationClassIcon icon={child.icon} extraClass="me-1{isChildSelected ? ' text-primary' : ''}" />
											{child.label}
										</label>
									</div>

									{#if hasGrandchildren}
										<div class="ms-4">
											{#each grandchildClasses as grandchild}
												{@const isGrandchildSelected = selectedClasses.includes(grandchild.id)}
												<div class="form-check">
													<input
														type="checkbox"
														class="form-check-input"
														id="class-{grandchild.id}"
														checked={isGrandchildSelected}
														on:change={() => toggleClass(grandchild.id)}
													/>
												<label class="form-check-label" for="class-{grandchild.id}">
													<NotificationClassIcon icon={grandchild.icon} extraClass="me-1{isGrandchildSelected ? ' text-primary' : ''}" />
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
			<button 
				class="btn btn-primary mt-3" 
				on:click={saveClasses}
				disabled={savingClasses}
			>
				{#if savingClasses}
					<span class="spinner-border spinner-border-sm me-1"></span>
				{:else}
					<i class="fas fa-save me-1"></i>
				{/if}
				Save Classes
			</button>
		</CollapsibleSection>

		<!-- Tags Section -->
		<CollapsibleSection title="Tags" icon="fa-tags" collapsed={false}>
			<p class="text-muted mb-2">
				Tags automatically subscribe matching tokens to this notification.
			</p>
			
			<!-- Selected tags -->
			{#if notificationTags.length > 0}
				<div class="selected-tags mb-3">
					{#each notificationTags as tagName}
						<TagPill 
							tag={tagName} 
							color={getTagColor(tagName)}
							removable={$isLoggedIn && isOwner}
							onRemove={() => removeTag(tagName)}
						/>
					{/each}
				</div>
			{:else}
				<div class="text-muted mb-3">No tags selected</div>
			{/if}
			
			{#if $isLoggedIn && isOwner}
				<!-- Tag dropdown + new tag input -->
				<div class="input-group mb-3" style="max-width: 500px;">
					<select 
						class="form-select"
						bind:value={newTagInput}
						on:change={() => { if (newTagInput) addTag(newTagInput); }}
						style="max-width: 200px;"
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
						on:keydown={(e) => { if (e.key === 'Enter') { e.preventDefault(); addNewTag(); } }}
					/>
					<button 
						type="button" 
						class="btn btn-outline-secondary"
						on:click={addNewTag}
						disabled={!newTagInput.trim()}
						title="Add tag"
					>
						<i class="fas fa-plus"></i>
					</button>
				</div>
				<small class="text-muted d-block mb-3">Select an existing tag or type a new tag name and press Enter.</small>

				<button 
					class="btn btn-primary" 
					on:click={saveTags}
					disabled={savingTags}
				>
					{#if savingTags}
						<span class="spinner-border spinner-border-sm me-1"></span>
					{:else}
						<i class="fas fa-save me-1"></i>
					{/if}
					Save Tags
				</button>
			{:else if $isLoggedIn && !isOwner}
				<p class="text-muted mb-0">
					<i class="fas fa-info-circle me-1"></i>
					Tag management is only available for the notification owner.
				</p>
			{:else}
				<p class="text-muted mb-0">
					<i class="fas fa-info-circle me-1"></i>
					Log in to manage tags.
				</p>
			{/if}
		</CollapsibleSection>

		<!-- Subscribed Tokens Section -->
		{#if notification.user_wide}
			<CollapsibleSection title="Subscribed Tokens" icon="fa-key" collapsed={false}>
				<div class="alert alert-info mb-0">
					<i class="fas fa-info-circle me-2"></i>
					This is a <strong>user-wide notification</strong> and applies to all your mytokens.
				</div>
			</CollapsibleSection>
		{:else}
			<CollapsibleSection title="Subscribed Tokens" icon="fa-key" collapsed={false}>
				{#if !$isLoggedIn}
					<div class="alert alert-warning">
						<p class="mb-3">
							<i class="fas fa-info-circle me-2"></i>
							To view and manage subscribed tokens, please log in.
						</p>
						<button class="btn btn-primary" on:click={handleLogin}>
							<i class="fas fa-sign-in-alt me-2"></i>
							Login
						</button>
					</div>
				{:else if !isOwner}
					<div class="alert alert-info mb-0">
						<i class="fas fa-info-circle me-2"></i>
						Token management is only available for the notification owner.
					</div>
				{:else if loadingTokens}
					<LoadingSpinner message="Loading tokens..." />
				{:else}
					<!-- Directly Subscribed Tokens -->
					{#if subscribedTokens.length > 0}
						<h6 class="text-muted mb-2">
							<i class="fas fa-link me-1"></i>
							Directly Subscribed ({subscribedTokens.length})
						</h6>
						<div class="table-responsive mb-4">
							<table class="table table-hover align-middle">
								<thead>
									<tr>
										<th>Token Name</th>
										<th>Tags</th>
										<th>Created</th>
										<th>Expires</th>
										<th></th>
									</tr>
								</thead>
								<tbody>
									{#each subscribedTokens as token}
										{@const expired = isExpired(token)}
										<tr class:token-expired={expired}>
											<td>
												<span class:token-name-unnamed={!token.name}>
													{token.name || 'unnamed token'}
												</span>
											</td>
											<td>
												{#if token.tags && token.tags.length > 0}
													{#each token.tags.slice(0, 2) as tag}
														<TagPill tag={tag.tag} color={tag.color} small />
													{/each}
													{#if token.tags.length > 2}
														<span class="badge bg-secondary">+{token.tags.length - 2}</span>
													{/if}
												{:else}
													<span class="text-muted">-</span>
												{/if}
											</td>
											<td>{token.created ? formatDateTime(token.created) : '-'}</td>
											<td class:text-muted={!token.expires_at || token.expires_at === 0}>
												{token.expires_at ? formatDateTime(token.expires_at) : 'Does not expire'}
												{#if expired}
													<span class="badge bg-secondary ms-1" title="This token has expired">Expired</span>
												{/if}
											</td>
											<td>
												<button 
													class="btn btn-sm btn-outline-danger"
													on:click={() => removeToken(token.mom_id)}
													title="Remove from notification"
												>
													<i class="fas fa-trash"></i>
												</button>
											</td>
										</tr>
									{/each}
								</tbody>
							</table>
						</div>
					{/if}

					<!-- Tag-Linked Tokens -->
					{#if tagLinkedTokens.length > 0}
						<h6 class="text-muted mb-2">
							<i class="fas fa-tags me-1"></i>
							Linked by Tags ({tagLinkedTokens.length})
						</h6>
						<div class="table-responsive mb-4">
							<table class="table table-hover align-middle">
								<thead>
									<tr>
										<th>Token Name</th>
										<th>Matching Tags</th>
										<th>Created</th>
										<th>Expires</th>
									</tr>
								</thead>
								<tbody>
									{#each tagLinkedTokens as token}
										{@const matchingTags = token.tags?.filter(t => notificationTags.includes(t.tag)) ?? []}
										{@const expired = isExpired(token)}
										<tr class:token-expired={expired}>
											<td>
												<span class:token-name-unnamed={!token.name}>
													{token.name || 'unnamed token'}
												</span>
											</td>
											<td>
												{#each matchingTags as tag}
													<TagPill tag={tag.tag} color={tag.color} small />
												{/each}
											</td>
											<td>{token.created ? formatDateTime(token.created) : '-'}</td>
											<td class:text-muted={!token.expires_at || token.expires_at === 0}>
												{token.expires_at ? formatDateTime(token.expires_at) : 'Does not expire'}
												{#if expired}
													<span class="badge bg-secondary ms-1" title="This token has expired">Expired</span>
												{/if}
											</td>
										</tr>
									{/each}
								</tbody>
							</table>
						</div>
					{/if}

					<!-- No tokens message -->
					{#if subscribedTokens.length === 0 && tagLinkedTokens.length === 0}
						<div class="alert alert-light mb-4">
							<i class="fas fa-info-circle me-2"></i>
							No tokens subscribed to this notification. Add tokens directly or use tags to automatically include matching tokens.
						</div>
					{/if}

					<button class="btn btn-success" on:click={openAddTokenModal}>
						<i class="fas fa-plus-circle me-1"></i>
						Add Token
					</button>
				{/if}
			</CollapsibleSection>
		{/if}
	{/if}
</div>

<!-- Add Token Modal -->
{#if showAddTokenModal}
	<div class="modal show d-block" tabindex="-1" role="dialog">
		<div class="modal-dialog modal-lg modal-dialog-centered modal-dialog-scrollable" role="document">
			<div class="modal-content">
				<div class="modal-header">
					<h5 class="modal-title">
						<i class="fas fa-plus-circle me-2"></i>
						Add Tokens to Notification
					</h5>
					<button type="button" class="btn-close" on:click={() => showAddTokenModal = false} aria-label="Close"></button>
				</div>
				<div class="modal-body">
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
					{#if selectedTokensToAdd.size > 0}
						<div class="alert alert-info py-2 mb-2">
							<small>
								<i class="fas fa-check-circle me-1"></i>
								{selectedTokensToAdd.size} token{selectedTokensToAdd.size > 1 ? 's' : ''} selected
							</small>
						</div>
					{/if}

					{#if flattenedTokens.length === 0}
						<div class="alert alert-info mb-0">
							<i class="fas fa-info-circle me-2"></i>
							No tokens available.
						</div>
					{:else if availableTokensToAdd.length === 0 && !tokenSearchQuery}
						<div class="alert alert-info mb-0">
							<i class="fas fa-info-circle me-2"></i>
							All your tokens are already subscribed to this notification.
						</div>
					{:else if availableTokensToAdd.length === 0 && tokenSearchQuery}
						<div class="text-center text-muted py-3">
							<i class="fas fa-search me-1"></i>
							No tokens match your search
						</div>
					{:else}
						<div class="token-list-container">
							<table class="table table-hover align-middle mb-0">
								<thead class="sticky-top">
									<tr>
										<th style="width: 40px;"></th>
										<th>Name</th>
										<th style="width: 130px;">Created</th>
										<th style="width: 130px;">Expires</th>
										<th>Include Children</th>
									</tr>
								</thead>
								<tbody>
									{#each availableTokensToAdd as item}
										{@const token = item.token}
										{@const isSelected = selectedTokensToAdd.has(token.mom_id)}
										{@const expired = isExpired(token)}
										<tr class:table-active={isSelected} class:token-expired={expired}>
											<td>
												<input 
													type="checkbox" 
													class="form-check-input"
													checked={isSelected}
													on:change={() => toggleTokenSelection(token.mom_id)}
												/>
											</td>
											<td 
												class:token-fold={item.hasChildren}
												on:click={() => item.hasChildren && toggleExpand(item.id)}
												role={item.hasChildren ? 'button' : undefined}
											>
												<div class="d-flex align-items-center">
													<span style="margin-left: {item.depth * 1.2}rem;"></span>
													{#if item.hasChildren}
														<i class="fas me-1" class:fa-caret-right={!expandedTokens.has(item.id)} class:fa-caret-down={expandedTokens.has(item.id)}></i>
													{:else}
														<span style="width: 0.8rem; display: inline-block;"></span>
													{/if}
													<span class:token-name-unnamed={!token.name}>
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
											<td>{formatDateTime(token.created)}</td>
											<td class:text-muted={!token.expires_at || token.expires_at === 0}>
												{token.expires_at ? formatDateTime(token.expires_at) : 'Does not expire'}
												{#if expired}
													<span class="badge bg-secondary ms-1">Expired</span>
												{/if}
											</td>
											<td>
												{#if isSelected && item.hasChildren}
													<div class="form-check form-switch">
														<input
															type="checkbox"
															class="form-check-input"
															checked={includeChildrenMap.get(token.mom_id) ?? true}
															on:change={() => toggleIncludeChildren(token.mom_id)}
														/>
													</div>
												{/if}
											</td>
										</tr>
									{/each}
								</tbody>
							</table>
						</div>
					{/if}
				</div>
				<div class="modal-footer">
					<button type="button" class="btn btn-secondary" on:click={() => showAddTokenModal = false}>
						Cancel
					</button>
					<button 
						type="button" 
						class="btn btn-primary"
						on:click={addTokensToNotification}
						disabled={selectedTokensToAdd.size === 0 || addingTokens}
					>
						{#if addingTokens}
							<span class="spinner-border spinner-border-sm me-1"></span>
						{:else}
							<i class="fas fa-plus me-1"></i>
						{/if}
						Add Selected
					</button>
				</div>
			</div>
		</div>
	</div>
	<div class="modal-backdrop show"></div>
{/if}

<style>
	.notification-management {
		max-width: 900px;
		margin: 0 auto;
	}

	.notification-classes-list {
		max-height: 400px;
		overflow-y: auto;
		background-color: var(--bs-tertiary-bg);
	}

	.selected-tags {
		display: flex;
		flex-wrap: wrap;
		gap: 0.25rem;
		padding: 0.5rem;
		background-color: var(--bs-body-bg);
		border-radius: 0.375rem;
		border: 1px solid var(--bs-border-color);
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

	.token-list-container {
		max-height: 350px;
		overflow-y: auto;
		background-color: var(--bs-body-bg);
		border-radius: 0.375rem;
		border: 1px solid var(--bs-border-color);
	}

	.token-list-container .table {
		margin-bottom: 0;
	}

	.token-list-container th {
		font-weight: 500;
		font-size: 0.85rem;
		border-top: none;
	}

	.table th {
		border-top: none;
		font-weight: 500;
		white-space: nowrap;
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

	.table-active {
		background-color: rgba(var(--mytoken-primary-rgb), 0.1) !important;
	}

	.form-check-label {
		cursor: pointer;
	}

	/* Expired token styling - matches TokenList.svelte */
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

	/* Sticky header in modal needs proper background */
	.token-list-container thead.sticky-top {
		background-color: var(--bs-body-bg);
	}
</style>
