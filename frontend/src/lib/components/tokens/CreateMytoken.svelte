<script lang="ts">
	import { onMount, tick, createEventDispatcher } from 'svelte';
	import type { Restriction, Rotation, WebCapability, CapabilityTemplate, RestrictionTemplate, RotationTemplate, MytokenProfile, CreateMytokenTag, CreateMytokenRequest, InitialMytokenRequest } from '$lib/types';
	import { api, ApiClientError } from '$lib/api/client';
	import { discovery, providers } from '$lib/stores/discovery';
	import { isLoggedIn, auth } from '$lib/stores/auth';
	import { ui } from '$lib/stores/ui';
	import { tags } from '$lib/stores/tags';
	import CapabilityTree from '../capabilities/CapabilityTree.svelte';
	import RestrictionsEditor from '../restrictions/RestrictionsEditor.svelte';
	import RotationSettings from '../RotationSettings.svelte';
	import CollapsibleSection from '../CollapsibleSection.svelte';
	import TagPill from '../TagPill.svelte';
	import CopyButton from '../CopyButton.svelte';
	import LoadingSpinner from '../LoadingSpinner.svelte';

	const dispatch = createEventDispatcher<{ created: { token: string; tokenType: string } }>();

	// Props
	export let initialRequest: InitialMytokenRequest | null = null;

	// Form state
	let selectedProvider = '';
	let tokenName = '';
	let tokenType: 'token' | 'short_token' | 'auto' = 'token';
	let maxTokenLength: number | undefined = undefined;
	let webCapabilities: WebCapability[] = [];
	let restrictions: Restriction[] = [];
	let rotation: Rotation = {};
	let selectedTags: CreateMytokenTag[] = [];

	// Profile/template state
	let profiles: MytokenProfile[] = [];
	let capabilityTemplates: CapabilityTemplate[] = [];
	let restrictionTemplates: RestrictionTemplate[] = [];
	let rotationTemplates: RotationTemplate[] = [];
	let selectedProfile = '';

	// UI state
	let loading = false;
	let loadingData = true;
	let polling = false;
	let pollingCode = '';
	let pollingInterval = 5;        // Server-provided polling interval (seconds)
	let pollingExpiresAt = 0;       // Timestamp when polling expires
	let consentUri = '';            // Consent URI to display during polling
	let createdToken = '';
	let createdTokenType = 'token'; // 'token' | 'short_token' | 'transfer_code'
	let showResult = false;

	// Component refs
	let capabilityTreeRef: CapabilityTree;
	let restrictionsEditorRef: RestrictionsEditor;

	// Derived: available scopes based on selected provider
	$: availableScopes = getAvailableScopes(selectedProvider);
	
	function getAvailableScopes(providerIssuer: string): string[] {
		if (!providerIssuer) return [];
		const provider = $providers.find(p => p.issuer === providerIssuer);
		return provider?.scopes_supported ?? [];
	}
	const DEFAULT_PROFILE_NAME = 'web-default';
	let defaultProfileApplied = false;

	/**
	 * Wait for providers to be loaded from discovery
	 */
	async function waitForProviders(maxWaitMs: number = 5000): Promise<boolean> {
		const startTime = Date.now();
		while ($providers.length === 0 && Date.now() - startTime < maxWaitMs) {
			await new Promise(resolve => setTimeout(resolve, 50));
		}
		return $providers.length > 0;
	}

	// Initialize capabilities and templates
	onMount(async () => {
		loadingData = true;
		try {
			// Fetch capabilities with full structure
			webCapabilities = await api.getAllCapabilities();
			
			// Try to fetch profiles (may fail if not configured)
			try {
				const serverProfiles = await api.getProfiles();
				profiles = serverProfiles.map(p => {
					const payload = p.payload as any;
					return {
						name: p.name,
						capabilities: payload?.capabilities,
						restrictions: payload?.restrictions,
						rotation: payload?.rotation,
						tags: payload?.tags
					};
				});
			} catch {
				profiles = [];
			}
			
			// Try to fetch templates (may fail if not configured)
			try {
				const capTemplates = await api.getCapabilityTemplates();
				capabilityTemplates = capTemplates.map(p => ({
					name: p.name,
					capabilities: Array.isArray(p.payload) ? p.payload as string[] : []
				}));
			} catch {
				capabilityTemplates = [];
			}
			
			try {
				const restrTemplates = await api.getRestrictionTemplates();
				restrictionTemplates = restrTemplates.map(p => ({
					name: p.name,
					restrictions: Array.isArray(p.payload) ? p.payload as Restriction[] : []
				}));
			} catch {
				restrictionTemplates = [];
			}
			
			try {
				const rotTemplates = await api.getRotationTemplates();
				rotationTemplates = rotTemplates.map(p => ({
					name: p.name,
					rotation: (p.payload as Rotation) ?? {}
				}));
			} catch {
				rotationTemplates = [];
			}
		} catch (err) {
			console.error('Error loading capabilities:', err);
			// Fall back to empty capabilities
			webCapabilities = [];
		} finally {
			loadingData = false;
			
			// Wait for DOM to update so capabilityTreeRef is available
			await tick();
			// Wait again to ensure child component's reactive statements have processed
			await tick();
			
			// Load tags before applying initial request so tag colors are available
			// Wait for discovery if not yet loaded (might still be loading from layout)
			if ($isLoggedIn && !$tags.loaded) {
				// Wait for discovery endpoint to be available
				let waitAttempts = 0;
				while (!$discovery.data?.usersettings_endpoint && waitAttempts < 50) {
					await new Promise(resolve => setTimeout(resolve, 50));
					waitAttempts++;
				}
				
				if ($discovery.data?.usersettings_endpoint) {
					await tags.fetch($discovery.data.usersettings_endpoint);
					// Wait for store update to propagate
					await tick();
				}
			}
			
			// Apply initial request if provided (from URL parameter), otherwise apply default profile
			if (initialRequest) {
				// Wait for providers to be loaded before validating issuer
				await waitForProviders();
				const success = applyInitialRequest(initialRequest);
				if (!success) {
					// Validation failed (e.g., invalid issuer), apply default profile instead
					applyDefaultProfile();
				}
			} else {
				applyDefaultProfile();
			}
		}
		
		// Auto-select current session provider if logged in
		if ($isLoggedIn && $auth.oidcIssuer) {
			selectedProvider = $auth.oidcIssuer;
		}
	});

	function applyDefaultProfile() {
		if (defaultProfileApplied) return;
		
		const defaultProfile = profiles.find(p => p.name === DEFAULT_PROFILE_NAME);
		if (defaultProfile) {
			selectedProfile = DEFAULT_PROFILE_NAME;
			applyProfile(defaultProfile);
			defaultProfileApplied = true;
		}
	}

	/**
	 * Apply initial request data from URL parameter to populate form
	 * Returns true if successful, false if there was an error (e.g., invalid issuer)
	 */
	function applyInitialRequest(request: InitialMytokenRequest): boolean {
		// Validate oidc_issuer if provided
		if (request.oidc_issuer) {
			// Normalize URLs for comparison (remove trailing slash)
			const normalizeUrl = (url: string) => url.replace(/\/+$/, '');
			const requestIssuer = normalizeUrl(request.oidc_issuer);
			const providerExists = $providers.some(p => normalizeUrl(p.issuer) === requestIssuer);
			if (!providerExists) {
				ui.showError(
					'Unsupported OpenID Provider',
					`The OpenID Provider "${request.oidc_issuer}" is not supported by this mytoken instance.`
				);
				return false;
			}
		}

		// Clear selected profile since we're using custom values
		selectedProfile = '';
		defaultProfileApplied = true; // Prevent default profile from being applied
		
		if (request.name) {
			tokenName = request.name;
		}
		
		if (request.oidc_issuer) {
			selectedProvider = request.oidc_issuer;
		}
		
		if (request.response_type) {
			tokenType = request.response_type;
		}
		
		if (request.capabilities && request.capabilities.length > 0 && capabilityTreeRef) {
			const unknownCaps = capabilityTreeRef.setCapabilities(request.capabilities);
			if (unknownCaps.length > 0) {
				ui.warning(`Unknown capabilities ignored: ${unknownCaps.join(', ')}`);
			}
		}
		
		if (request.restrictions) {
			restrictions = JSON.parse(JSON.stringify(request.restrictions));
		}
		
		if (request.rotation) {
			rotation = { ...request.rotation };
		}
		
		if (request.tags) {
			selectedTags = request.tags.map(t => 
				typeof t === 'string' 
					? { tag: t, include_children: false }
					: { tag: t.tag, include_children: t.include_children ?? false }
			);
		}

		return true;
	}

	function applyProfile(profile: MytokenProfile) {
		// Start from a clean state
		if (capabilityTreeRef) {
			capabilityTreeRef.setEnabledCapabilities([]);
		}
		restrictions = [];
		rotation = {};
		selectedTags = [];
		
		// Apply capabilities if defined
		if (profile.capabilities && capabilityTreeRef) {
			capabilityTreeRef.setEnabledCapabilities(profile.capabilities);
		}
		
		// Apply restrictions if defined
		if (profile.restrictions) {
			restrictions = JSON.parse(JSON.stringify(profile.restrictions));
		}
		
		// Apply rotation if defined
		if (profile.rotation) {
			rotation = { ...profile.rotation };
		}
		
		// Apply tags if defined
		if (profile.tags) {
			selectedTags = profile.tags.map(t => ({ tag: t }));
		}
	}

	function handleProfileChange(event: Event) {
		const target = event.target as HTMLSelectElement;
		const profile = profiles.find(p => p.name === target.value);
		if (profile) {
			applyProfile(profile);
		}
	}

	async function handleSubmit() {
		loading = true;
		showResult = false;
		createdToken = '';
		createdTokenType = 'token';

		try {
			const enabledCapabilities = capabilityTreeRef?.getEnabledCapabilities() ?? [];
			const finalRestrictions = restrictionsEditorRef?.getRestrictions() ?? restrictions;

		const request: CreateMytokenRequest = {
			grant_type: 'oidc_flow',
			oidc_flow: 'authorization_code',
			redirect_type: 'native',
			application_name: 'mytoken webinterface',
			oidc_issuer: selectedProvider || undefined,
			name: tokenName || undefined,
			capabilities: enabledCapabilities.length > 0 ? enabledCapabilities : undefined,
			restrictions: finalRestrictions.length > 0 ? finalRestrictions : undefined,
			rotation: Object.keys(rotation).length > 0 ? rotation : undefined,
			tags: selectedTags.length > 0 ? selectedTags : undefined,
			response_type: tokenType !== 'token' ? tokenType : undefined,
			max_token_len: tokenType === 'auto' && maxTokenLength !== undefined && maxTokenLength > 0 ? maxTokenLength : undefined
		};

			const response = await api.createMytoken(request);

			if ('polling_code' in response) {
				// OIDC flow - need to redirect and poll
				pollingCode = (response as any).polling_code;
				pollingInterval = (response as any).interval ?? 5;
				const expiresIn = (response as any).expires_in ?? 300;
				pollingExpiresAt = Date.now() + (expiresIn * 1000);
				consentUri = (response as any).consent_uri;

				if (consentUri) {
					// Open consent URI in new window
					window.open(consentUri, '_blank');
					startPolling();
				}
			} else if (response.mytoken) {
				// Direct response with token
				handleTokenCreated(response.mytoken, (response as any).mytoken_type ?? 'token');
			}
		} catch (error) {
			if (error instanceof ApiClientError) {
				ui.showError('Failed to create mytoken', error.description ?? error.code);
			} else {
				ui.showError('Error', (error as Error).message);
			}
		} finally {
			loading = false;
		}
	}

	async function startPolling() {
		polling = true;

		const poll = async () => {
			// Check if expired
			if (Date.now() >= pollingExpiresAt) {
				polling = false;
				consentUri = '';
				ui.showError('Timeout', 'The authorization request has expired. Please try again.');
				return;
			}

			if (!polling) return;

			try {
				const response = await api.pollMytoken(pollingCode);
				// Success (200 OK) - we got the token
				polling = false;
				consentUri = '';

				const tokenType = (response as any).mytoken_type ?? 'token';
				const token = tokenType === 'transfer_code'
					? (response as any).transfer_code
					: (response as any).mytoken;

				handleTokenCreated(token, tokenType);
			} catch (error) {
				if (error instanceof ApiClientError) {
					if (error.code === 'authorization_pending') {
						// Still waiting - continue polling with server-provided interval
						setTimeout(poll, pollingInterval * 1000);
					} else {
						// Real error - map to user-friendly message
						polling = false;
						consentUri = '';
						const message = getErrorMessage(error.code, error.description);
						ui.showError('Authorization Failed', message);
					}
				} else {
					polling = false;
					consentUri = '';
					ui.showError('Polling Error', (error as Error).message);
				}
			}
		};

		poll();
	}

	function getErrorMessage(code: string, description?: string): string {
		switch (code) {
			case 'access_denied':
				return 'Authorization was denied.';
			case 'expired_token':
				return 'The polling code has expired. Please try again.';
			case 'invalid_grant':
				return 'Invalid authorization.';
			case 'invalid_token':
				return 'Invalid polling code.';
			default:
				return description ?? code;
		}
	}

	function handleTokenCreated(token: string, tokenType: string = 'token') {
		createdToken = token;
		createdTokenType = tokenType;
		showResult = true;
		ui.success('Mytoken created successfully!');

		// Dispatch event to parent so it can switch to info tab
		dispatch('created', { token, tokenType });
	}

	function cancelPolling() {
		polling = false;
		pollingCode = '';
		consentUri = '';
	}

	function resetForm() {
		showResult = false;
		createdToken = '';
		tokenName = '';
		tokenType = 'token';
		maxTokenLength = undefined;
		selectedTags = [];
		restrictions = [];
		rotation = {};
	}

	// Reactive lookup of tag colors - returns color for a given tag name
	// Using $tags.tags directly ensures reactivity when tags are loaded
	function getTagColor(tagName: string, tagsArray: typeof $tags.tags): string {
		const tagData = tagsArray.find(t => t.tag === tagName);
		return tagData?.color ?? '';
	}

	function toggleIncludeChildren(tagName: string) {
		const index = selectedTags.findIndex(t => t.tag === tagName);
		if (index >= 0) {
			selectedTags[index].include_children = !selectedTags[index].include_children;
			selectedTags = [...selectedTags];
		}
	}

	function toggleTag(tagName: string, includeChildren: boolean = false) {
		const existingIndex = selectedTags.findIndex(t => t.tag === tagName);
		if (existingIndex >= 0) {
			selectedTags = selectedTags.filter((_, i) => i !== existingIndex);
		} else {
			selectedTags = [...selectedTags, { tag: tagName, include_children: includeChildren }];
		}
	}

	function isTagSelected(tagName: string): boolean {
		return selectedTags.some(t => t.tag === tagName);
	}

	// Tag modal state
	let showTagModal = false;
	let newTagName = '';
	let tagIncludeChildren = false;
	let selectedExistingTag = '';
	let creatingNewTag = false;

	function openTagModal() {
		showTagModal = true;
		newTagName = '';
		tagIncludeChildren = false;
		selectedExistingTag = '';
		creatingNewTag = false;
	}

	function closeTagModal() {
		showTagModal = false;
	}

	async function addTagFromModal() {
		let tagName = '';
		
		if (creatingNewTag) {
			tagName = newTagName.trim();
			if (!tagName) return;
			
			// Create the new tag on the server
			if ($discovery.data?.usersettings_endpoint) {
				const success = await tags.create($discovery.data.usersettings_endpoint, { tag: tagName });
				if (!success) {
					ui.showError('Error', 'Failed to create tag');
					return;
				}
			}
		} else {
			tagName = selectedExistingTag;
			if (!tagName) return;
		}
		
		// Add to selected tags if not already selected
		if (!isTagSelected(tagName)) {
			selectedTags = [...selectedTags, { tag: tagName, include_children: tagIncludeChildren }];
		}
		
		closeTagModal();
	}

	// Get available tags (not already selected)
	$: availableTags = $tags.tags.filter(t => !isTagSelected(t.tag));
</script>

<div class="create-mytoken">
	{#if showResult}
		<!-- Result display -->
		<div class="result-section">
			<div class="alert alert-success">
				<h5 class="alert-heading">
					<i class="fas fa-check-circle me-2"></i>
					Mytoken Created Successfully!
					<span class="badge ms-2" class:bg-primary={createdTokenType === 'token'}
						  class:bg-info={createdTokenType === 'short_token'}
						  class:bg-warning={createdTokenType === 'transfer_code'}>
						{#if createdTokenType === 'short_token'}
							Short Token
						{:else if createdTokenType === 'transfer_code'}
							Transfer Code
						{:else}
							JWT
						{/if}
					</span>
				</h5>
				<p>Your new mytoken has been created. Please copy it now - it won't be shown again.</p>
			</div>

			<div class="card mb-4">
				<div class="card-header d-flex justify-content-between align-items-center">
					<span>Your Mytoken</span>
					<CopyButton value={createdToken} label="Copy Token" />
				</div>
				<div class="card-body">
					<div class="token-display">
						<code class="text-break">{createdToken}</code>
					</div>
				</div>
			</div>

			<div class="d-flex gap-2">
				<button type="button" class="btn btn-primary" on:click={resetForm}>
					<i class="fas fa-plus me-1"></i>
					Create Another
				</button>
				<button type="button" class="btn btn-outline-secondary" on:click={() => (showResult = false)}>
					<i class="fas fa-edit me-1"></i>
					Edit Settings
				</button>
			</div>
		</div>
	{:else if polling}
		<!-- Polling state -->
		<div class="polling-section text-center py-5">
			<LoadingSpinner size="lg" message="Waiting for authorization..." />
			<p class="mt-3 text-muted">
				Please complete the login in the popup window.
			</p>

			{#if consentUri}
				<div class="alert alert-info text-start mt-4">
					<p class="mb-2">
						<i class="fas fa-external-link-alt me-2"></i>
						If the authorization window didn't open, click the link below:
					</p>
					<a href={consentUri} target="_blank" rel="noopener noreferrer" class="consent-uri-link">
						{consentUri}
					</a>
				</div>
			{/if}

			<button type="button" class="btn btn-outline-secondary mt-3" on:click={cancelPolling}>
				<i class="fas fa-times me-1"></i>
				Cancel
			</button>
		</div>
	{:else}
		<!-- Form -->
		<form on:submit|preventDefault={handleSubmit}>
			<!-- Profile selection -->
			{#if profiles.length > 0}
				<div class="mb-4">
					<label for="profile" class="form-label">
						<i class="fas fa-bookmark me-1"></i>
						Profile
					</label>
					<select 
						id="profile" 
						class="form-select" 
						bind:value={selectedProfile}
						on:change={handleProfileChange}
					>
						<option value="">-- Select a profile to prefill --</option>
						{#each profiles as profile}
							<option value={profile.name}>{profile.name}</option>
						{/each}
					</select>
					<small class="text-muted">
						Select a profile to prefill the form with predefined settings.
					</small>
				</div>
			{/if}

			<!-- Provider selection -->
			{#if $providers.length > 0}
				<div class="mb-4">
					<label for="provider" class="form-label">
						<i class="fab fa-openid me-1"></i>
						OpenID Provider
					</label>
					<select id="provider" class="form-select" bind:value={selectedProvider} required={!$isLoggedIn}>
						{#if !$isLoggedIn}
							<option value="">-- Select a provider --</option>
						{/if}
						{#each $providers as provider}
							<option value={provider.issuer}>
								{provider.name ?? provider.issuer}
							</option>
						{/each}
					</select>
					<small class="text-muted">
						OpenID Provider for which this mytoken can obtain access tokens.
					</small>
				</div>
			{/if}

			<!-- Token name -->
			<div class="mb-4">
				<label for="token-name" class="form-label">
					<i class="fas fa-tag me-1"></i>
					Token Name (optional)
				</label>
				<input
					type="text"
					id="token-name"
					class="form-control"
					placeholder="My Token"
					bind:value={tokenName}
				/>
				<small class="text-muted">
					A name to help you identify this token later.
				</small>
			</div>

			<!-- Token Type -->
			<div class="mb-4">
				<label for="token-type" class="form-label">
					<i class="fas fa-file-code me-1"></i>
					Mytoken Type
				</label>
				<select id="token-type" class="form-select" bind:value={tokenType}>
					<option value="token">JWT</option>
					<option value="short_token">Short Token</option>
					<option value="auto">Depending on size</option>
				</select>
				{#if tokenType === 'auto'}
					<div class="mt-2">
						<label for="max-token-length" class="form-label">
							Maximum Token Length
						</label>
						<input
							type="number"
							id="max-token-length"
							class="form-control"
							min="0"
							placeholder="Leave empty for default"
							bind:value={maxTokenLength}
						/>
						<small class="text-muted">
							Adapt the mytoken type so that the returned mytoken is not longer than this value.
						</small>
					</div>
				{/if}
			</div>

			<!-- Tags -->
			{#if $isLoggedIn}
				<div class="mb-4">
					<span class="form-label d-block" id="tags-selection-label">
						<i class="fas fa-tags me-1"></i>
						Tags (optional)
					</span>
					
					{#if selectedTags.length > 0}
						<div class="selected-tags mb-2">
							{#each selectedTags as tagInfo (tagInfo.tag)}
								<TagPill 
									tag={tagInfo.tag} 
									color={getTagColor(tagInfo.tag, $tags.tags)}
									includeChildren={tagInfo.include_children}
									removable
									onRemove={() => toggleTag(tagInfo.tag)}
									onToggleChildren={() => toggleIncludeChildren(tagInfo.tag)}
								/>
							{/each}
						</div>
						<small class="text-muted d-block mb-2">
							<i class="fas fa-sitemap me-1"></i> Click a tag to toggle "include children".
						</small>
					{:else}
						<p class="text-muted mb-2">No tags added.</p>
					{/if}
					
					<button type="button" class="btn btn-sm btn-outline-success" on:click={openTagModal}>
						<i class="fas fa-plus me-1"></i>
						Add Tag
					</button>
					<small class="form-text text-muted d-block mt-1">
						Tags help you organize and filter your mytokens.
					</small>
				</div>
			{/if}

			<!-- Capabilities -->
			<div class="mb-4">
				{#if loadingData}
					<LoadingSpinner message="Loading capabilities..." />
				{:else}
					<CapabilityTree
						bind:this={capabilityTreeRef}
						capabilities={webCapabilities}
						templates={capabilityTemplates}
						collapsed={false}
						showTemplates={capabilityTemplates.length > 0}
					/>
				{/if}
			</div>

			<!-- Restrictions -->
			<div class="mb-4">
				<RestrictionsEditor
					bind:this={restrictionsEditorRef}
					bind:restrictions
					templates={restrictionTemplates}
					collapsed={true}
					showTemplates={restrictionTemplates.length > 0}
					showJsonEditor={true}
					{availableScopes}
				/>
			</div>

			<!-- Rotation -->
			<div class="mb-4">
				<RotationSettings
					bind:rotation
					templates={rotationTemplates}
					collapsed={true}
					showTemplates={rotationTemplates.length > 0}
				/>
			</div>

			<!-- Submit -->
			<div class="d-grid">
				<button type="submit" class="btn btn-primary btn-lg" disabled={loading}>
					{#if loading}
						<span class="spinner-border spinner-border-sm me-2" role="status"></span>
						Creating...
					{:else}
						<i class="fas fa-plus-circle me-2"></i>
						Create Mytoken
					{/if}
				</button>
			</div>
		</form>
	{/if}
</div>

<!-- Add Tag Modal -->
{#if showTagModal}
	<div class="modal-backdrop fade show"></div>
	<div class="modal fade show d-block" tabindex="-1" role="dialog">
		<div class="modal-dialog modal-dialog-centered" role="document">
			<div class="modal-content">
				<div class="modal-header">
					<h5 class="modal-title">
						<i class="fas fa-tag me-2"></i>
						Add Tag to Mytoken
					</h5>
					<button type="button" class="btn-close" on:click={closeTagModal} aria-label="Close"></button>
				</div>
				<div class="modal-body">
					{#if availableTags.length > 0 || creatingNewTag}
						{#if !creatingNewTag}
							<div class="mb-3">
								<label for="tag-selector" class="form-label">Select Tag</label>
								<select 
									id="tag-selector" 
									class="form-select"
									bind:value={selectedExistingTag}
								>
									<option value="">-- Select a tag --</option>
									{#each availableTags as tag}
										<option value={tag.tag}>{tag.tag}</option>
									{/each}
								</select>
							</div>
							<div class="text-center mb-3">
								<button 
									type="button" 
									class="btn btn-sm btn-link"
									on:click={() => creatingNewTag = true}
								>
									<i class="fas fa-plus me-1"></i>
									Create New Tag
								</button>
							</div>
						{:else}
							<div class="mb-3">
								<label for="new-tag-name" class="form-label">New Tag Name</label>
								<input 
									type="text" 
									id="new-tag-name" 
									class="form-control"
									placeholder="Enter tag name"
									bind:value={newTagName}
								/>
							</div>
							<div class="text-center mb-3">
								<button 
									type="button" 
									class="btn btn-sm btn-link"
									on:click={() => creatingNewTag = false}
								>
									<i class="fas fa-arrow-left me-1"></i>
									Select Existing Tag
								</button>
							</div>
						{/if}
					{:else}
						<div class="mb-3">
							<label for="new-tag-name" class="form-label">New Tag Name</label>
							<input 
								type="text" 
								id="new-tag-name" 
								class="form-control"
								placeholder="Enter tag name"
								bind:value={newTagName}
							/>
							<small class="text-muted">No existing tags available. Create a new one.</small>
						</div>
					{/if}
					
					<div class="form-check">
						<input 
							type="checkbox" 
							class="form-check-input" 
							id="tag-include-children"
							bind:checked={tagIncludeChildren}
						/>
						<label class="form-check-label" for="tag-include-children">
							<i class="fas fa-sitemap me-1"></i>
							Include children
						</label>
						<small class="form-text text-muted d-block">
							If checked, subtokens created from this mytoken will inherit this tag.
						</small>
					</div>
				</div>
				<div class="modal-footer">
					<button type="button" class="btn btn-secondary" on:click={closeTagModal}>
						Cancel
					</button>
					<button 
						type="button" 
						class="btn btn-success"
						on:click={addTagFromModal}
						disabled={(creatingNewTag || availableTags.length === 0) ? !newTagName.trim() : !selectedExistingTag}
					>
						<i class="fas fa-tag me-1"></i>
						Add Tag
					</button>
				</div>
			</div>
		</div>
	</div>
{/if}

<style>
	.token-display {
		background-color: var(--bs-tertiary-bg);
		padding: 1rem;
		border-radius: 0.375rem;
		max-height: 200px;
		overflow-y: auto;
	}

	.token-display code {
		font-size: 0.875rem;
		word-break: break-all;
	}

	.consent-uri-link {
		word-break: break-all;
		display: block;
	}
</style>
