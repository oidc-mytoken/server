<script lang="ts">
	import { createEventDispatcher, onMount } from 'svelte';
	import type { TokenInfoResponse, EventHistoryEntry, MytokenEntryTree, UsedRestriction, Restriction, WebCapability, Notification, Calendar, InitialMytokenRequest } from '$lib/types';
	import { api, ApiClientError } from '$lib/api/client';
	import { ui } from '$lib/stores/ui';
	import { formatDateTime, formatRelativeTime, formatTokenPreview } from '$lib/utils/format';
	import { getNotificationTypeIcon, getNotificationTypeShortLabel, isClassEnabled, getClassColor, getRootNotificationClasses } from '$lib/utils/notifications';
	import CapabilityTree from '../capabilities/CapabilityTree.svelte';
	import RestrictionsEditor from '../restrictions/RestrictionsEditor.svelte';
	import RotationSettings from '../RotationSettings.svelte';
	import CopyButton from '../CopyButton.svelte';
	import LoadingSpinner from '../LoadingSpinner.svelte';
	import TagPill from '../TagPill.svelte';
	import CollapsibleSection from '../CollapsibleSection.svelte';
	import NotificationClassIcon from '../notifications/NotificationClassIcon.svelte';

	// Input token
	export let initialToken: string = '';

	// Event dispatcher for parent communication
	const dispatch = createEventDispatcher<{
		createTransferCode: { token: string };
	}>();

	// State
	let token = '';
	let tokenInfo: TokenInfoResponse | null = null;
	let eventHistory: EventHistoryEntry[] = [];
	let subtokensTree: MytokenEntryTree | null = null;
	let loading = false;
	let loadingHistory = false;
	let loadingSubtokens = false;
	let activeTab: 'info' | 'json' | 'history' | 'subtokens' = 'info';

	// Capabilities from API for the tree display
	let webCapabilities: WebCapability[] = [];
	let loadingCapabilities = true;

	// Notifications and calendars
	let notifications: Notification[] = [];
	let calendars: Calendar[] = [];
	let loadingNotifications = false;
	let loadingCalendars = false;
	let notificationsError: string | null = null;
	let calendarsError: string | null = null;

	// Root notification classes for display
	const rootClasses = getRootNotificationClasses();

	// Subtokens expanded state
	let expandedSubtokens: Set<string> = new Set();

	// Flattened subtoken structure for table display
	interface FlattenedSubtoken {
		token: MytokenEntryTree['token'];
		depth: number;
		hasChildren: boolean;
		id: string;
		parentId: string | null;
	}

	// Flatten subtokens tree for display
	function flattenSubtokens(tree: MytokenEntryTree | null): FlattenedSubtoken[] {
		if (!tree || !tree.children) return [];
		
		const result: FlattenedSubtoken[] = [];
		let idCounter = 0;
		
		function traverse(node: MytokenEntryTree, depth: number, parentId: string | null) {
			const id = `subtoken-${idCounter++}`;
			const hasChildren = (node.children?.length ?? 0) > 0;
			
			result.push({
				token: node.token,
				depth,
				hasChildren,
				id,
				parentId
			});
			
			if (node.children) {
				for (const child of node.children) {
					traverse(child, depth + 1, id);
				}
			}
		}
		
		// Start from children of root (the root is the current token being introspected)
		for (const child of tree.children) {
			traverse(child, 0, null);
		}
		
		return result;
	}

	function toggleSubtokenExpand(id: string) {
		if (expandedSubtokens.has(id)) {
			expandedSubtokens.delete(id);
		} else {
			expandedSubtokens.add(id);
		}
		expandedSubtokens = new Set(expandedSubtokens);
	}

	function isSubtokenExpanded(id: string): boolean {
		return expandedSubtokens.has(id);
	}

	function isExpired(expiresAt: number | undefined): boolean {
		if (!expiresAt || expiresAt === 0) return false;
		return new Date(expiresAt * 1000) < new Date();
	}

	function formatExpiry(expiresAt: number | undefined): string {
		if (!expiresAt || expiresAt === 0) {
			return 'Does not expire';
		}
		return formatDateTime(expiresAt);
	}

	// Reactive flattened subtokens
	$: flattenedSubtokens = flattenSubtokens(subtokensTree);
	// Note: we pass expandedSubtokens as parameter to make this reactive to expand/collapse changes
	$: visibleSubtokens = getVisibleSubtokens(flattenedSubtokens, expandedSubtokens);

	function getVisibleSubtokens(items: FlattenedSubtoken[], expanded: Set<string>): FlattenedSubtoken[] {
		return items.filter(item => {
			if (item.parentId === null) return true;
			
			let current = item;
			while (current.parentId !== null) {
				if (!expanded.has(current.parentId)) {
					return false;
				}
				const parent = items.find(t => t.id === current.parentId);
				if (!parent) break;
				current = parent;
			}
			return true;
		});
	}

	// Track the last token we introspected to avoid duplicate requests
	let lastIntrospectedToken = '';

	// Track the last initialToken we processed to detect actual changes from parent
	let lastProcessedInitialToken = '';

	// Load capabilities on mount
	onMount(async () => {
		try {
			webCapabilities = await api.getAllCapabilities();
		} catch (error) {
			console.error('Failed to load capabilities:', error);
			webCapabilities = [];
		} finally {
			loadingCapabilities = false;
		}
	});

	// React to initialToken changes from parent and auto-introspect
	// This only triggers when the parent provides a new token, not when user edits
	$: if (initialToken && initialToken !== lastProcessedInitialToken) {
		lastProcessedInitialToken = initialToken;
		token = initialToken;
		introspect();
	}

	// Handle blur event - auto-introspect when focus leaves the input
	function handleBlur() {
		const trimmedToken = token.trim();
		if (trimmedToken && trimmedToken !== lastIntrospectedToken) {
			introspect();
		}
	}

	async function introspect() {
		const trimmedToken = token.trim();
		if (!trimmedToken) {
			return;
		}

		// Skip if we already introspected this token
		if (trimmedToken === lastIntrospectedToken && tokenInfo) {
			return;
		}

		loading = true;
		tokenInfo = null;
		eventHistory = [];
		subtokensTree = null;
		notifications = [];
		calendars = [];

		try {
			tokenInfo = await api.introspect(trimmedToken);
			lastIntrospectedToken = trimmedToken;
			ui.success('Token introspected successfully');
			
			// Load notifications and calendars in background
			loadNotifications();
			loadCalendars();
		} catch (error) {
			if (error instanceof ApiClientError) {
				ui.showError('Introspection Failed', error.description ?? error.code);
			} else {
				ui.showError('Error', (error as Error).message);
			}
		} finally {
			loading = false;
		}
	}

	async function loadEventHistory(forceReload = false) {
		if (!token) return;
		if (!forceReload && eventHistory.length > 0) return;

		loadingHistory = true;
		try {
			eventHistory = await api.getEventHistory(token);
		} catch (error) {
			if (error instanceof ApiClientError) {
				ui.error('Failed to load event history');
			}
		} finally {
			loadingHistory = false;
		}
	}

	async function loadSubtokens(forceReload = false) {
		if (!token) return;
		if (!forceReload && subtokensTree) return;

		loadingSubtokens = true;
		try {
			subtokensTree = await api.getSubtokens(token);
		} catch (error) {
			if (error instanceof ApiClientError) {
				ui.error('Failed to load subtokens');
			}
		} finally {
			loadingSubtokens = false;
		}
	}

	function handleTabChange(tab: typeof activeTab) {
		activeTab = tab;
		if (tab === 'history' && eventHistory.length === 0) {
			loadEventHistory();
		} else if (tab === 'subtokens' && !subtokensTree) {
			loadSubtokens();
		}
	}

	async function revokeToken() {
		const confirmed = await ui.confirm({
			title: 'Revoke Token',
			message: 'Are you sure you want to revoke this token? This cannot be undone.',
			confirmText: 'Revoke',
			confirmVariant: 'danger'
		});
		
		if (!confirmed) return;

		try {
			await api.revokeToken(token, false);
			ui.success('Token revoked successfully');
			tokenInfo = null;
			token = '';
			lastIntrospectedToken = '';
		} catch (error) {
			if (error instanceof ApiClientError) {
				ui.showError('Revocation Failed', error.description ?? error.code);
			}
		}
	}

	function handleCreateTransferCode() {
		dispatch('createTransferCode', { token: token.trim() });
	}

	function handleRecreate() {
		if (!tokenInfo) return;
		
		const tokenData = tokenInfo.token;
		const now = Math.floor(Date.now() / 1000);
		const offset = now - tokenData.iat;
		
		// Adjust restriction timestamps relative to now and remove usage tracking fields
		const adjustedRestrictions: Restriction[] | undefined = tokenData.restrictions?.map(r => {
			const adjusted: Restriction = {};
			if (r.nbf) adjusted.nbf = r.nbf + offset;
			if (r.exp) adjusted.exp = r.exp + offset;
			if (r.scope) adjusted.scope = r.scope;
			if (r.audience) adjusted.audience = [...r.audience];
			if (r.hosts) adjusted.hosts = [...r.hosts];
			if (r.geoip_allow) adjusted.geoip_allow = [...r.geoip_allow];
			if (r.geoip_disallow) adjusted.geoip_disallow = [...r.geoip_disallow];
			if (r.usages_AT !== undefined) adjusted.usages_AT = r.usages_AT;
			if (r.usages_other !== undefined) adjusted.usages_other = r.usages_other;
			return adjusted;
		});
		
		const request: InitialMytokenRequest = {
			name: tokenData.name,
			oidc_issuer: tokenData.oidc_iss,
			capabilities: tokenData.capabilities,
			restrictions: adjustedRestrictions,
			rotation: tokenData.rotation,
			// Determine token type from current token
			response_type: tokenInfo.token_type === 'short_token' ? 'short_token' : 'token'
		};
		
		// Add tags if present
		if (tokenInfo.tags && tokenInfo.tags.length > 0) {
			request.tags = tokenInfo.tags.map(t => ({
				tag: t.tag,
				include_children: t.include_children
			}));
		}
		
		// Encode and redirect to Create Mytoken tab
		const encoded = btoa(JSON.stringify(request));
		window.location.href = `/?r=${encoded}#mt`;
	}

	function getEventIcon(event: string): string {
		const eventLower = event.toLowerCase();
		if (eventLower.includes('created')) return 'fa-plus-circle text-success';
		if (eventLower.includes('revoked')) return 'fa-ban text-danger';
		if (eventLower.includes('at_created') || eventLower.includes('access_token')) return 'fa-key text-primary';
		if (eventLower.includes('rotated')) return 'fa-sync text-info';
		if (eventLower.includes('transferred')) return 'fa-exchange-alt text-warning';
		if (eventLower.includes('used')) return 'fa-check text-success';
		if (eventLower.includes('blocked') || eventLower.includes('denied')) return 'fa-times-circle text-danger';
		return 'fa-circle text-secondary';
	}

	function formatEventName(event: string): string {
		// Convert snake_case to Title Case with spaces
		return event
			.split('_')
			.map(word => word.charAt(0).toUpperCase() + word.slice(1).toLowerCase())
			.join(' ');
	}

	function parseUserAgent(userAgent: string | undefined): { icon: string; title: string } {
		if (!userAgent) return { icon: 'fas fa-question', title: 'Unknown' };
		
		const ua = userAgent.toLowerCase();
		
		// Check for known clients
		if (ua.includes('oidc-agent')) {
			return { icon: 'fas fa-terminal', title: 'oidc-agent' };
		}
		if (ua.includes('mytoken')) {
			return { icon: 'fas fa-key', title: 'mytoken client' };
		}
		
		// Check for curl
		if (ua.includes('curl')) {
			return { icon: 'fas fa-terminal', title: 'curl' };
		}
		
		// Browser detection
		if (ua.includes('firefox')) {
			return { icon: 'fab fa-firefox', title: 'Firefox' };
		}
		if (ua.includes('chrome') && !ua.includes('edg')) {
			return { icon: 'fab fa-chrome', title: 'Chrome' };
		}
		if (ua.includes('safari') && !ua.includes('chrome')) {
			return { icon: 'fab fa-safari', title: 'Safari' };
		}
		if (ua.includes('edg')) {
			return { icon: 'fab fa-edge', title: 'Edge' };
		}
		
		// OS detection fallback
		if (ua.includes('linux')) {
			return { icon: 'fab fa-linux', title: 'Linux Client' };
		}
		if (ua.includes('windows')) {
			return { icon: 'fab fa-windows', title: 'Windows Client' };
		}
		if (ua.includes('mac')) {
			return { icon: 'fab fa-apple', title: 'macOS Client' };
		}
		
		return { icon: 'fas fa-globe', title: userAgent };
	}

	// Convert UsedRestriction[] to Restriction[] for the RestrictionsEditor
	function toRestrictions(usedRestrictions: UsedRestriction[] | undefined): Restriction[] {
		if (!usedRestrictions) return [];
		return usedRestrictions.map(ur => ({
			nbf: ur.nbf,
			exp: ur.exp,
			scope: ur.scope,
			audience: ur.audience,
			hosts: ur.hosts,
			geoip_allow: ur.geoip_allow,
			geoip_disallow: ur.geoip_disallow,
			usages_AT: ur.usages_AT,
			usages_other: ur.usages_other
		}));
	}

	// Calculate total usages from restrictions
	function getTotalUsages(usedRestrictions: UsedRestriction[] | undefined): { at: string; other: string } {
		if (!usedRestrictions || usedRestrictions.length === 0) {
			return { at: 'Unlimited', other: 'Unlimited' };
		}

		let totalAT: number | null = null;
		let totalATDone = 0;
		let totalOther: number | null = null;
		let totalOtherDone = 0;

		for (const r of usedRestrictions) {
			if (r.usages_AT !== undefined) {
				totalAT = (totalAT ?? 0) + r.usages_AT;
			}
			if (r.usages_AT_done !== undefined) {
				totalATDone += r.usages_AT_done;
			}
			if (r.usages_other !== undefined) {
				totalOther = (totalOther ?? 0) + r.usages_other;
			}
			if (r.usages_other_done !== undefined) {
				totalOtherDone += r.usages_other_done;
			}
		}

		return {
			at: totalAT !== null ? `${totalATDone} / ${totalAT}` : 'Unlimited',
			other: totalOther !== null ? `${totalOtherDone} / ${totalOther}` : 'Unlimited'
		};
	}

	// Load notifications linked to this token
	async function loadNotifications() {
		if (!tokenInfo) return;
		
		loadingNotifications = true;
		notificationsError = null;
		try {
			notifications = await api.getNotifications();
		} catch (error) {
			console.error('Failed to load notifications:', error);
			notifications = [];
			if (error instanceof ApiClientError) {
				if (error.isAuthError()) {
					notificationsError = 'Session expired or invalid. Please log in again.';
				} else {
					notificationsError = error.description ?? error.code;
				}
			} else {
				notificationsError = (error as Error).message;
			}
		} finally {
			loadingNotifications = false;
		}
	}

	// Load calendars linked to this token
	async function loadCalendars() {
		if (!tokenInfo) return;
		
		loadingCalendars = true;
		calendarsError = null;
		try {
			calendars = await api.getCalendars();
		} catch (error) {
			console.error('Failed to load calendars:', error);
			calendars = [];
			if (error instanceof ApiClientError) {
				if (error.isAuthError()) {
					calendarsError = 'Session expired or invalid. Please log in again.';
				} else {
					calendarsError = error.description ?? error.code;
				}
			} else {
				calendarsError = (error as Error).message;
			}
		} finally {
			loadingCalendars = false;
		}
	}

	// Filter notifications that are linked to the current token
	function getLinkedNotifications(allNotifications: Notification[], info: TokenInfoResponse | null): {
		direct: Notification[];
		tagBased: Notification[];
		userWide: Notification[];
	} {
		if (!info) return { direct: [], tagBased: [], userWide: [] };
		
		const momId = info.mom_id ?? '';
		const tokenTags = info.tags ?? [];
		
		const direct: Notification[] = [];
		const tagBased: Notification[] = [];
		const userWide: Notification[] = [];
		
		for (const n of allNotifications) {
			if (n.user_wide) {
				userWide.push(n);
			} else if (n.subscribed_tokens?.includes(momId)) {
				direct.push(n);
			} else if (tokenTags.length > 0 && n.tags?.some(nt => tokenTags.some(tt => tt.tag === nt.tag))) {
				tagBased.push(n);
			}
		}
		
		return { direct, tagBased, userWide };
	}

	// Filter calendars that are linked to the current token
	function getLinkedCalendars(allCalendars: Calendar[], info: TokenInfoResponse | null): {
		direct: Calendar[];
		tagBased: Calendar[];
	} {
		if (!info) return { direct: [], tagBased: [] };
		
		const momId = info.mom_id ?? '';
		const tokenTags = info.tags ?? [];
		
		const direct: Calendar[] = [];
		const tagBased: Calendar[] = [];
		
		for (const c of allCalendars) {
			if (c.subscribed_tokens?.includes(momId)) {
				direct.push(c);
			} else if (tokenTags.length > 0 && c.tags?.some(ct => tokenTags.some(tt => tt.tag === ct.tag))) {
				tagBased.push(c);
			}
		}
		
		return { direct, tagBased };
	}

	// Reactive linked notifications and calendars
	$: linkedNotifications = getLinkedNotifications(notifications, tokenInfo);
	$: linkedCalendars = getLinkedCalendars(calendars, tokenInfo);
	$: hasLinkedNotifications = linkedNotifications.direct.length > 0 || linkedNotifications.tagBased.length > 0 || linkedNotifications.userWide.length > 0;
	$: hasLinkedCalendars = linkedCalendars.direct.length > 0 || linkedCalendars.tagBased.length > 0;
</script>

<div class="token-info">
	<!-- Token input -->
	<div class="mb-4">
		<label for="token-input" class="form-label">
			<i class="fas fa-search me-1"></i>
			Token to Introspect
		</label>
		<div class="input-group">
			<input
				type="text"
				id="token-input"
				class="form-control font-monospace"
				placeholder="Paste your mytoken here..."
				bind:value={token}
				on:keydown={(e) => e.key === 'Enter' && introspect()}
				on:blur={handleBlur}
			/>
			<CopyButton value={token} label="Copy" />
			<button class="btn btn-primary" type="button" on:click={introspect} disabled={loading}>
				{#if loading}
					<span class="spinner-border spinner-border-sm me-1"></span>
				{:else}
					<i class="fas fa-search me-1"></i>
				{/if}
				Introspect
			</button>
		</div>
	</div>

	{#if loading}
		<LoadingSpinner message="Loading token information..." />
	{:else if tokenInfo}
		<!-- Token info display -->
		<div class="token-info-display">
			<!-- Quick info header -->
			<div class="card mb-4">
				<div class="card-body">
					<div class="row align-items-center">
						<div class="col-md-8">
							<h5 class="mb-1">
								{tokenInfo.token.name ?? 'Unnamed Token'}
								{#if tokenInfo.valid}
									<span class="badge bg-success ms-2">Valid</span>
								{:else}
									<span class="badge bg-danger ms-2">Invalid</span>
								{/if}
							</h5>
							<p class="text-muted mb-0">
								<small>
									<i class="fas fa-fingerprint me-1"></i>
									{formatTokenPreview(tokenInfo.mom_id ?? '', 16)}
								</small>
							</p>
							{#if tokenInfo.tags && tokenInfo.tags.length > 0}
								<div class="mt-2">
									{#each tokenInfo.tags as tag}
										<TagPill 
											name={tag.tag} 
											color={tag.color} 
											small={true}
											includeChildren={tag.include_children}
										/>
									{/each}
								</div>
							{/if}
						</div>
						<div class="col-md-4 text-md-end mt-2 mt-md-0">
							<div class="btn-group">
								<button 
									class="btn btn-sm btn-outline-primary" 
									on:click={handleCreateTransferCode}
									title="Create transfer code for this token"
								>
									<i class="fas fa-exchange-alt me-1"></i>
									Transfer
								</button>
								<button 
									class="btn btn-sm btn-outline-secondary" 
									on:click={handleRecreate}
									title="Create a new mytoken with the same properties"
								>
									<i class="fas fa-copy me-1"></i>
									Re-create
								</button>
								<button class="btn btn-sm btn-outline-danger" on:click={revokeToken}>
									<i class="fas fa-ban me-1"></i>
									Revoke
								</button>
							</div>
						</div>
					</div>
				</div>
			</div>

			<!-- Tabs -->
			<ul class="nav nav-tabs mb-3">
				<li class="nav-item">
					<button
						class="nav-link"
						class:active={activeTab === 'info'}
						on:click={() => handleTabChange('info')}
					>
						<i class="fas fa-info-circle me-1"></i>
						Details
					</button>
				</li>
				<li class="nav-item">
					<button
						class="nav-link"
						class:active={activeTab === 'json'}
						on:click={() => handleTabChange('json')}
					>
						<i class="fas fa-code me-1"></i>
						JSON
					</button>
				</li>
				<li class="nav-item">
					<button
						class="nav-link"
						class:active={activeTab === 'history'}
						on:click={() => handleTabChange('history')}
					>
						<i class="fas fa-history me-1"></i>
						Event History
					</button>
				</li>
				<li class="nav-item">
					<button
						class="nav-link"
						class:active={activeTab === 'subtokens'}
						on:click={() => handleTabChange('subtokens')}
					>
						<i class="fas fa-sitemap me-1"></i>
						Subtokens
					</button>
				</li>
			</ul>

		<div class="tab-content">
			<div class="tab-pane" class:active={activeTab === 'info'} style:display={activeTab === 'info' ? undefined : 'none'}>
				<!-- Basic info -->
				<div class="card mb-3">
					<div class="card-header">
						<i class="fas fa-info-circle me-1"></i>
						Basic Information
					</div>
					<div class="card-body">
						<dl class="row mb-0">
							<dt class="col-sm-4">Token Version</dt>
							<dd class="col-sm-8">{tokenInfo.token.ver ?? 'N/A'}</dd>

							<dt class="col-sm-4">OIDC Issuer</dt>
							<dd class="col-sm-8">{tokenInfo.token.oidc_iss ?? 'N/A'}</dd>

							<dt class="col-sm-4">OIDC Subject</dt>
							<dd class="col-sm-8">{tokenInfo.token.oidc_sub ?? 'N/A'}</dd>

							<dt class="col-sm-4">Mytoken Issuer</dt>
							<dd class="col-sm-8">{tokenInfo.token.iss ?? 'N/A'}</dd>

							<dt class="col-sm-4">Audience</dt>
							<dd class="col-sm-8">{tokenInfo.token.aud ?? 'N/A'}</dd>

							<dt class="col-sm-4">Created (iat)</dt>
							<dd class="col-sm-8">
								{tokenInfo.token.iat ? formatDateTime(tokenInfo.token.iat) : 'N/A'}
							</dd>

							<dt class="col-sm-4">Not Before (nbf)</dt>
							<dd class="col-sm-8">
								{tokenInfo.token.nbf ? formatDateTime(tokenInfo.token.nbf) : 'N/A'}
							</dd>

							<dt class="col-sm-4">Expires (exp)</dt>
							<dd class="col-sm-8">
								{#if tokenInfo.token.exp}
									{formatDateTime(tokenInfo.token.exp)}
									<small class="text-muted">({formatRelativeTime(tokenInfo.token.exp)})</small>
								{:else}
									Never
								{/if}
							</dd>

							{#if tokenInfo.token.auth_time}
								<dt class="col-sm-4">Auth Time</dt>
								<dd class="col-sm-8">{formatDateTime(tokenInfo.token.auth_time)}</dd>
							{/if}

							<dt class="col-sm-4">Sequence Number</dt>
							<dd class="col-sm-8">{tokenInfo.token.seq_no ?? 0}</dd>

							<dt class="col-sm-4">AT Usages</dt>
							<dd class="col-sm-8">{getTotalUsages(tokenInfo.token.restrictions).at}</dd>

							<dt class="col-sm-4">Other Usages</dt>
							<dd class="col-sm-8">{getTotalUsages(tokenInfo.token.restrictions).other}</dd>
						</dl>
					</div>
				</div>

				<!-- Capabilities -->
				<div class="mb-3">
					{#if loadingCapabilities}
						<LoadingSpinner message="Loading capabilities..." />
					{:else}
						<CapabilityTree
							capabilities={webCapabilities}
							selectedCapabilities={tokenInfo.token.capabilities ?? []}
							readonly={true}
							collapsed={true}
							showTemplates={false}
						/>
					{/if}
				</div>

				<!-- Restrictions -->
				{#if tokenInfo.token.restrictions && tokenInfo.token.restrictions.length > 0}
					<div class="mb-3">
						<RestrictionsEditor
							restrictions={toRestrictions(tokenInfo.token.restrictions)}
							readonly={true}
							collapsed={true}
							showTemplates={false}
						/>
					</div>
				{/if}

				<!-- Rotation -->
				{#if tokenInfo.token.rotation}
					<div class="mb-3">
						<RotationSettings
							rotation={tokenInfo.token.rotation}
							readonly={true}
							collapsed={true}
							showTemplates={false}
						/>
					</div>
				{/if}

				<!-- Linked Notifications -->
				<CollapsibleSection
					title="Linked Notifications"
					icon="fa-bell"
					collapsed={true}
				>
					<span slot="header-right" class="ms-2">
						{#if loadingNotifications}
							<span class="spinner-border spinner-border-sm text-muted"></span>
						{:else if notificationsError}
							<span class="badge bg-warning text-dark" title={notificationsError}>
								<i class="fas fa-exclamation-triangle"></i>
							</span>
						{:else if hasLinkedNotifications}
							<span class="badge bg-info">
								{linkedNotifications.direct.length + linkedNotifications.tagBased.length + linkedNotifications.userWide.length}
							</span>
						{:else}
							<span class="badge bg-secondary">0</span>
						{/if}
					</span>

					{#if loadingNotifications}
						<div class="text-center py-3">
							<span class="spinner-border spinner-border-sm me-2"></span>
							Loading notifications...
						</div>
					{:else if notificationsError}
						<div class="alert alert-warning mb-0">
							<i class="fas fa-exclamation-triangle me-2"></i>
							{notificationsError}
						</div>
					{:else if hasLinkedNotifications}
						<!-- Directly subscribed notifications -->
						{#if linkedNotifications.direct.length > 0}
							<div class="mb-3">
								<h6 class="text-muted mb-2">
									<i class="fas fa-link me-1"></i>
									Directly Subscribed
								</h6>
								<div class="list-group list-group-flush">
									{#each linkedNotifications.direct as notif}
										<div class="list-group-item d-flex align-items-center gap-2 py-2">
											<span class="badge bg-secondary" title={notif.notification_type === 'mail' ? 'Email' : 'WebSocket'}>
												<i class="fas {getNotificationTypeIcon(notif.notification_type)}"></i>
											</span>
											<div class="notification-classes-icons">
												{#each rootClasses as cls}
													{@const status = isClassEnabled(notif, cls.id)}
													<span class="{getClassColor(status)}" title={cls.label}>
														<NotificationClassIcon icon={cls.icon} />
													</span>
												{/each}
											</div>
											{#if notif.tags && notif.tags.length > 0}
												<div class="tags ms-2">
													{#each notif.tags as tag}
														<TagPill name={tag.tag} color={tag.color} small />
													{/each}
												</div>
											{/if}
										</div>
									{/each}
								</div>
							</div>
						{/if}

						<!-- Tag-based notifications -->
						{#if linkedNotifications.tagBased.length > 0}
							<div class="mb-3">
								<h6 class="text-muted mb-2">
									<i class="fas fa-tags me-1"></i>
									Via Tags
								</h6>
								<div class="list-group list-group-flush">
									{#each linkedNotifications.tagBased as notif}
										<div class="list-group-item d-flex align-items-center gap-2 py-2">
											<span class="badge bg-secondary" title={notif.notification_type === 'mail' ? 'Email' : 'WebSocket'}>
												<i class="fas {getNotificationTypeIcon(notif.notification_type)}"></i>
											</span>
											<div class="notification-classes-icons">
												{#each rootClasses as cls}
													{@const status = isClassEnabled(notif, cls.id)}
													<span class="{getClassColor(status)}" title={cls.label}>
														<NotificationClassIcon icon={cls.icon} />
													</span>
												{/each}
											</div>
											{#if notif.tags && notif.tags.length > 0}
												<div class="tags ms-2">
													{#each notif.tags as tag}
														<TagPill name={tag.tag} color={tag.color} small />
													{/each}
												</div>
											{/if}
										</div>
									{/each}
								</div>
							</div>
						{/if}

						<!-- User-wide notifications -->
						{#if linkedNotifications.userWide.length > 0}
							<div class="mb-3">
								<h6 class="text-muted mb-2">
									<i class="fas fa-user me-1"></i>
									User-Wide
								</h6>
								<div class="list-group list-group-flush">
									{#each linkedNotifications.userWide as notif}
										<div class="list-group-item d-flex align-items-center gap-2 py-2">
											<span class="badge bg-secondary" title={notif.notification_type === 'mail' ? 'Email' : 'WebSocket'}>
												<i class="fas {getNotificationTypeIcon(notif.notification_type)}"></i>
											</span>
											<div class="notification-classes-icons">
												{#each rootClasses as cls}
													{@const status = isClassEnabled(notif, cls.id)}
													<span class="{getClassColor(status)}" title={cls.label}>
														<NotificationClassIcon icon={cls.icon} />
													</span>
												{/each}
											</div>
											<span class="badge bg-primary ms-auto">
												<i class="fas fa-user me-1"></i>
												All Tokens
											</span>
										</div>
									{/each}
								</div>
							</div>
						{/if}
					{:else}
						<div class="alert alert-secondary mb-0">
							<i class="fas fa-info-circle me-2"></i>
							No notifications linked to this token.
						</div>
					{/if}
				</CollapsibleSection>

				<!-- Linked Calendars -->
				<CollapsibleSection
					title="Linked Calendars"
					icon="fa-calendar"
					collapsed={true}
				>
					<span slot="header-right" class="ms-2">
						{#if loadingCalendars}
							<span class="spinner-border spinner-border-sm text-muted"></span>
						{:else if calendarsError}
							<span class="badge bg-warning text-dark" title={calendarsError}>
								<i class="fas fa-exclamation-triangle"></i>
							</span>
						{:else if hasLinkedCalendars}
							<span class="badge bg-info">
								{linkedCalendars.direct.length + linkedCalendars.tagBased.length}
							</span>
						{:else}
							<span class="badge bg-secondary">0</span>
						{/if}
					</span>

					{#if loadingCalendars}
						<div class="text-center py-3">
							<span class="spinner-border spinner-border-sm me-2"></span>
							Loading calendars...
						</div>
					{:else if calendarsError}
						<div class="alert alert-warning mb-0">
							<i class="fas fa-exclamation-triangle me-2"></i>
							{calendarsError}
						</div>
					{:else if hasLinkedCalendars}
						<!-- Directly subscribed calendars -->
						{#if linkedCalendars.direct.length > 0}
							<div class="mb-3">
								<h6 class="text-muted mb-2">
									<i class="fas fa-link me-1"></i>
									Directly Subscribed
								</h6>
								<div class="list-group list-group-flush">
									{#each linkedCalendars.direct as cal}
										<div class="list-group-item d-flex align-items-center gap-2 py-2">
											<i class="fas fa-calendar text-primary"></i>
											<span class="flex-grow-1">
												{cal.description || 'Unnamed Calendar'}
											</span>
											{#if cal.tags && cal.tags.length > 0}
												<div class="tags">
													{#each cal.tags as tag}
														<TagPill name={tag.tag} color={tag.color} small />
													{/each}
												</div>
											{/if}
											{#if cal.ics_url}
												<a 
													href={cal.ics_url} 
													class="btn btn-sm btn-outline-secondary"
													title="Download ICS"
													target="_blank"
												>
													<i class="fas fa-download"></i>
												</a>
											{/if}
										</div>
									{/each}
								</div>
							</div>
						{/if}

						<!-- Tag-based calendars -->
						{#if linkedCalendars.tagBased.length > 0}
							<div class="mb-3">
								<h6 class="text-muted mb-2">
									<i class="fas fa-tags me-1"></i>
									Via Tags
								</h6>
								<div class="list-group list-group-flush">
									{#each linkedCalendars.tagBased as cal}
										<div class="list-group-item d-flex align-items-center gap-2 py-2">
											<i class="fas fa-calendar text-primary"></i>
											<span class="flex-grow-1">
												{cal.description || 'Unnamed Calendar'}
											</span>
											{#if cal.tags && cal.tags.length > 0}
												<div class="tags">
													{#each cal.tags as tag}
														<TagPill name={tag.tag} color={tag.color} small />
													{/each}
												</div>
											{/if}
											{#if cal.ics_url}
												<a 
													href={cal.ics_url} 
													class="btn btn-sm btn-outline-secondary"
													title="Download ICS"
													target="_blank"
												>
													<i class="fas fa-download"></i>
												</a>
											{/if}
										</div>
									{/each}
								</div>
							</div>
						{/if}
					{:else}
						<div class="alert alert-secondary mb-0">
							<i class="fas fa-info-circle me-2"></i>
							No calendars linked to this token.
						</div>
					{/if}
				</CollapsibleSection>
			</div>

			<div class="tab-pane" class:active={activeTab === 'json'} style:display={activeTab === 'json' ? undefined : 'none'}>
				<!-- JSON output -->
				<div class="card">
					<div class="card-header d-flex justify-content-between align-items-center">
						<span>
							<i class="fas fa-code me-1"></i>
							Raw JSON Response
						</span>
						<CopyButton value={JSON.stringify(tokenInfo, null, 2)} label="Copy JSON" />
					</div>
					<div class="card-body p-0">
						<pre class="json-output mb-0"><code>{JSON.stringify(tokenInfo, null, 2)}</code></pre>
					</div>
				</div>
			</div>

			<div class="tab-pane" class:active={activeTab === 'history'} style:display={activeTab === 'history' ? undefined : 'none'}>
				<!-- Event history -->
				<div class="d-flex justify-content-between align-items-center mb-3">
					<h6 class="mb-0">
						<i class="fas fa-history me-1"></i>
						Event History
					</h6>
					<button 
						class="btn btn-sm btn-outline-secondary" 
						on:click={() => loadEventHistory(true)}
						disabled={loadingHistory}
					>
						{#if loadingHistory}
							<span class="spinner-border spinner-border-sm me-1"></span>
						{:else}
							<i class="fas fa-sync-alt me-1"></i>
						{/if}
						Reload
					</button>
				</div>
				{#if loadingHistory}
					<LoadingSpinner message="Loading event history..." />
				{:else if eventHistory.length > 0}
					<div class="table-responsive">
						<table class="table table-hover align-middle">
							<thead>
								<tr>
									<th>Event</th>
									<th>Comment</th>
									<th>Time</th>
									<th>IP</th>
									<th class="text-center">Client</th>
								</tr>
							</thead>
							<tbody>
								{#each eventHistory as event}
									{@const uaInfo = parseUserAgent(event.user_agent)}
									<tr>
										<td>
											<i class="fas {getEventIcon(event.event)} me-2"></i>
											{formatEventName(event.event)}
										</td>
										<td class="text-break">
											{event.comment || '-'}
										</td>
										<td class="text-nowrap">
											{formatDateTime(event.time)}
										</td>
										<td>
											<code>{event.ip || '-'}</code>
										</td>
										<td class="text-center">
											<i 
												class="{uaInfo.icon}" 
												title={event.user_agent || 'Unknown'}
											></i>
										</td>
									</tr>
								{/each}
							</tbody>
						</table>
					</div>
				{:else}
					<div class="alert alert-info">
						<i class="fas fa-info-circle me-2"></i>
						No event history available.
					</div>
				{/if}
			</div>

			<div class="tab-pane" class:active={activeTab === 'subtokens'} style:display={activeTab === 'subtokens' ? undefined : 'none'}>
				<!-- Subtokens -->
				{#if loadingSubtokens}
					<LoadingSpinner message="Loading subtokens..." />
				{:else if visibleSubtokens.length > 0}
					<div class="table-responsive">
						<table class="table table-hover align-middle subtokens-table">
							<thead>
								<tr>
									<th>Token Name</th>
									<th>Tags</th>
									<th>Created</th>
									<th>Created from IP</th>
									<th>Expires</th>
									<th>MOM ID</th>
									<th>
										<button
											class="btn btn-sm btn-outline-primary"
											on:click={() => loadSubtokens(true)}
											disabled={loadingSubtokens}
											title="Refresh"
										>
											<i class="fas fa-sync" class:fa-spin={loadingSubtokens}></i>
										</button>
									</th>
								</tr>
							</thead>
							<tbody>
								{#each visibleSubtokens as item (item.id)}
									{@const expired = isExpired(item.token.expires_at)}
									<tr class:token-expired={expired}>
										<td 
											class:token-fold={item.hasChildren}
											on:click={() => item.hasChildren && toggleSubtokenExpand(item.id)}
											on:keydown={(e) => e.key === 'Enter' && item.hasChildren && toggleSubtokenExpand(item.id)}
											role={item.hasChildren ? 'button' : undefined}
											tabindex={item.hasChildren ? 0 : undefined}
										>
											<span style="margin-left: {item.depth * 1.5}rem;"></span>
											{#if item.hasChildren}
												<i class="fas me-2" class:fa-caret-right={!isSubtokenExpanded(item.id)} class:fa-caret-down={isSubtokenExpanded(item.id)}></i>
											{:else}
												<span style="width: 1rem; display: inline-block;"></span>
											{/if}
											<span class:token-name-unnamed={!item.token.name}>
												{item.token.name || 'unnamed token'}
											</span>
										</td>
										<td>
											{#if item.token.tags && item.token.tags.length > 0}
												{#each item.token.tags as tag}
													<TagPill name={tag.tag} color={tag.color} small />
												{/each}
											{:else}
												<span class="text-muted">-</span>
											{/if}
										</td>
										<td>{formatDateTime(item.token.created)}</td>
										<td>{item.token.ip || '-'}</td>
										<td class:text-muted={!item.token.expires_at || item.token.expires_at === 0}>
											{formatExpiry(item.token.expires_at)}
											{#if expired}
												<span class="badge bg-secondary ms-1" title="This token has expired">Expired</span>
											{/if}
										</td>
										<td>
											<small class="text-muted font-monospace">
												{formatTokenPreview(item.token.mom_id, 8)}
											</small>
										</td>
										<td></td>
									</tr>
								{/each}
							</tbody>
						</table>
					</div>
					<div class="text-muted mt-2">
						<small>{flattenedSubtokens.length} subtoken{flattenedSubtokens.length !== 1 ? 's' : ''}</small>
					</div>
				{:else}
					<div class="d-flex justify-content-between align-items-center mb-3">
						<span></span>
						<button 
							class="btn btn-sm btn-outline-secondary" 
							on:click={() => loadSubtokens(true)}
							disabled={loadingSubtokens}
						>
							<i class="fas fa-sync-alt me-1"></i>
							Reload
						</button>
					</div>
					<div class="alert alert-info">
						<i class="fas fa-info-circle me-2"></i>
						No subtokens found.
					</div>
				{/if}
			</div>
		</div>
		</div>
	{:else}
		<!-- Empty state -->
		<div class="text-center py-5 text-muted">
			<i class="fas fa-search fa-3x mb-3"></i>
			<p>Enter a token above to view its details</p>
		</div>
	{/if}
</div>

<style>
	.nav-tabs .nav-link {
		color: var(--bs-secondary-color);
		border: none;
		border-bottom: 2px solid transparent;
	}

	.nav-tabs .nav-link:hover {
		color: var(--mytoken-primary);
		border-bottom-color: var(--bs-border-color);
	}

	.nav-tabs .nav-link.active {
		color: var(--mytoken-primary);
		background: transparent;
		border-bottom: 2px solid var(--mytoken-primary);
	}

	.card-header {
		background-color: var(--bs-tertiary-bg);
		font-weight: 500;
	}

	dt {
		font-weight: 500;
	}

	.json-output {
		background-color: var(--bs-dark);
		color: var(--bs-light);
		padding: 1rem;
		margin: 0;
		border-radius: 0 0 0.375rem 0.375rem;
		overflow-x: auto;
		max-height: 500px;
		font-size: 0.875rem;
	}

	.json-output code {
		color: inherit;
	}

	/* Subtokens table styles */
	.subtokens-table th {
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
		background-color: rgba(0, 0, 0, 0.02);
	}

	/* Notification and calendar styles */
	.notification-classes-icons {
		display: flex;
		gap: 0.5rem;
		font-size: 0.9em;
	}

	.notification-classes-icons span {
		opacity: 0.8;
	}

	.notification-classes-icons .text-muted {
		opacity: 0.4;
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
