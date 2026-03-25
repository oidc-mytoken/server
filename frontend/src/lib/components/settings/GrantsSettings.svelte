<script lang="ts">
	import { onMount } from 'svelte';
	import type { Grant, SSHKeyInfo, Capability, Restriction, WebCapability } from '$lib/types';
	import { api, ApiClientError } from '$lib/api/client';
	import { ui } from '$lib/stores/ui';
	import LoadingSpinner from '../LoadingSpinner.svelte';
	import CapabilityTree from '../capabilities/CapabilityTree.svelte';
	import RestrictionsEditor from '../restrictions/RestrictionsEditor.svelte';

	// Props
	export let initialExpandSSH = false;

	let grants: Grant[] = [];
	let loading = true;
	let updating: string | null = null;
	let expandedGrant: string | null = initialExpandSSH ? 'ssh' : null;

	// SSH-specific state
	let sshGrantEnabled = false;
	let sshKeys: SSHKeyInfo[] = [];
	let loadingSSH = false;
	let addingSSH = false;
	let deletingSSH: string | null = null;

	// Capabilities for SSH key form
	let webCapabilities: WebCapability[] = [];
	let loadingCapabilities = false;
	// Default capabilities: AT, tokeninfo, create_mytoken (create subtoken)
	let selectedCapabilities = ['AT', 'tokeninfo', 'create_mytoken'];

	// New SSH key form
	let newKeyName = '';
	let newKeyValue = '';
	let showAdvanced = false;
	let sshKeyFileInput: HTMLInputElement;

	// SSH key add workflow state
	let sshAuthorizationUri: string | null = null;
	let sshPollingCode: string | null = null;
	let sshPollingInterval: number | null = null;
	let sshResult: { ssh_user?: string; ssh_host_config?: string } | null = null;

	// Component refs
	let capabilityTreeRef: CapabilityTree;
	let restrictionsEditorRef: RestrictionsEditor;

	// Available grant types to display (matching the Mustache frontend which only shows SSH)
	// Other grant types (oidc_flow, polling_code, transfer_code, mytoken) are not user-configurable
	const availableGrants: Array<{ type: string; label: string; description: string; icon: string; expandable: boolean }> = [
		{
			type: 'ssh',
			label: 'SSH',
			description: 'The SSH grant type allows you to link an ssh key and use ssh authentication for various actions.',
			icon: 'fa-terminal',
			expandable: true
		}
	];

	onMount(async () => {
		await loadGrants();
	});

	async function loadGrants() {
		loading = true;
		try {
			// Fetch enabled grants from API
			const enabledGrants = await api.getGrants();
			
			// Build grants list from available grants, with enabled status from API
			grants = availableGrants.map(ag => ({
				type: ag.type,
				enabled: enabledGrants.some(eg => eg.type === ag.type && eg.enabled)
			}));
			
			// Load SSH info
			await loadSSHInfo();
		} catch (error) {
			if (error instanceof ApiClientError) {
				ui.showError('Failed to load grants', error.description ?? error.code);
			}
		} finally {
			loading = false;
		}
	}

	async function loadSSHInfo() {
		loadingSSH = true;
		try {
			const info = await api.getSSHInfo();
			sshGrantEnabled = info.grant_enabled;
			sshKeys = info.ssh_keys ?? [];
			// Update grants state to reflect SSH grant status
			grants = grants.map(g => 
				g.type === 'ssh' ? { ...g, enabled: sshGrantEnabled } : g
			);
		} catch (error) {
			if (error instanceof ApiClientError) {
				console.error('Failed to load SSH info:', error);
			}
		} finally {
			loadingSSH = false;
		}
	}

	async function loadCapabilities() {
		if (webCapabilities.length > 0) return; // Already loaded
		
		loadingCapabilities = true;
		try {
			webCapabilities = await api.getAllCapabilities();
		} catch (error) {
			if (error instanceof ApiClientError) {
				console.error('Failed to load capabilities:', error);
			}
		} finally {
			loadingCapabilities = false;
		}
	}

	async function toggleGrant(grant: Grant) {
		updating = grant.type;
		try {
			if (grant.enabled) {
				await api.disableGrant(grant.type);
				ui.success(`${getGrantLabel(grant.type)} disabled`);
			} else {
				await api.enableGrant(grant.type);
				ui.success(`${getGrantLabel(grant.type)} enabled`);
			}
			// Update local state
			grants = grants.map(g => 
				g.type === grant.type ? { ...g, enabled: !g.enabled } : g
			);
			// Update SSH-specific state
			if (grant.type === 'ssh') {
				sshGrantEnabled = !grant.enabled;
			}
		} catch (error) {
			if (error instanceof ApiClientError) {
				ui.showError('Failed to update grant', error.description ?? error.code);
			}
		} finally {
			updating = null;
		}
	}

	function toggleExpand(grantType: string) {
		if (expandedGrant === grantType) {
			expandedGrant = null;
		} else {
			expandedGrant = grantType;
		}
	}

	function getGrantInfo(type: string) {
		return availableGrants.find(g => g.type === type);
	}

	function getGrantLabel(type: string): string {
		return getGrantInfo(type)?.label ?? type;
	}

	function getGrantDescription(type: string): string {
		return getGrantInfo(type)?.description ?? '';
	}

	function getGrantIcon(type: string): string {
		return getGrantInfo(type)?.icon ?? 'fa-check-circle';
	}

	function isExpandable(type: string): boolean {
		return getGrantInfo(type)?.expandable ?? false;
	}

	// SSH Key Management Functions
	async function addSSHKey() {
		if (!newKeyName.trim()) {
			ui.showError('Error', 'Please enter a name for the SSH key');
			return;
		}

		if (!newKeyValue.trim()) {
			ui.showError('Error', 'Please paste your SSH public key');
			return;
		}

		// Basic SSH key validation
		if (!newKeyValue.trim().startsWith('ssh-') && 
		    !newKeyValue.trim().startsWith('ecdsa-') &&
		    !newKeyValue.trim().startsWith('sk-')) {
			ui.showError('Error', 'Invalid SSH public key format. Key should start with ssh-rsa, ssh-ed25519, etc.');
			return;
		}

		addingSSH = true;
		try {
			const request: { name?: string; ssh_key: string; restrictions?: unknown[]; capabilities?: string[] } = {
				name: newKeyName.trim(),
				ssh_key: newKeyValue.trim()
			};

			// Add capabilities if any selected
			if (capabilityTreeRef) {
				const caps = capabilityTreeRef.getEnabledCapabilities();
				if (caps.length > 0) {
					request.capabilities = caps;
				}
			}

			// Add restrictions if any defined
			if (restrictionsEditorRef) {
				const restr = restrictionsEditorRef.getRestrictions();
				if (restr.length > 0) {
					request.restrictions = restr;
				}
			}

			const response = await api.addSSHKey(request);
			
			if (response.consent_uri) {
				sshAuthorizationUri = response.consent_uri;
				sshPollingCode = response.polling_code ?? null;
				// Start polling for completion
				if (sshPollingCode) {
					startPolling(response.interval);
				}
			} else {
				ui.success(`SSH key "${newKeyName}" added`);
				resetSSHForm();
				await loadSSHInfo();
			}
		} catch (error) {
			if (error instanceof ApiClientError) {
				ui.showError('Failed to add SSH key', error.description ?? error.code);
			}
		} finally {
			addingSSH = false;
		}
	}

	function startPolling(interval?: number) {
		if (!sshPollingCode) return;
		
		// Default to 5 seconds if not specified, interval from server is in seconds
		const pollInterval = interval ? interval * 1000 : 5000;
		
		sshPollingInterval = window.setInterval(async () => {
			try {
				const result = await api.pollSSHKey(sshPollingCode!);
				// Success - ssh_user is required in success response
				if (result.ssh_user) {
					stopPolling();
					// Clear authorization URI first so the success result shows
					sshAuthorizationUri = null;
					sshPollingCode = null;
					sshResult = result;
					ui.success('SSH key added successfully!');
					await loadSSHInfo();
				}
			} catch (error) {
				if (error instanceof ApiClientError) {
					// Handle specific error codes as per RFC 8628
					switch (error.code) {
						case 'authorization_pending':
							// Keep polling - user hasn't completed authorization yet
							break;
						case 'access_denied':
							stopPolling();
							ui.showError('Authorization Denied', 'You denied the authorization request.');
							resetSSHForm();
							break;
						case 'expired_token':
							stopPolling();
							ui.showError('Code Expired', 'The code expired. Please try again.');
							resetSSHForm();
							break;
						case 'invalid_grant':
						case 'invalid_token':
							stopPolling();
							ui.showError('Invalid Code', 'The code has already been used.');
							resetSSHForm();
							break;
						default:
							stopPolling();
							ui.showError('Failed to complete SSH key setup', error.description ?? error.code);
							resetSSHForm();
							break;
					}
				} else {
					stopPolling();
					ui.showError('Error', 'An unexpected error occurred');
					resetSSHForm();
				}
			}
		}, pollInterval);
	}

	function stopPolling() {
		if (sshPollingInterval) {
			window.clearInterval(sshPollingInterval);
			sshPollingInterval = null;
		}
	}

	function resetSSHForm() {
		newKeyName = '';
		newKeyValue = '';
		showAdvanced = false;
		sshAuthorizationUri = null;
		sshPollingCode = null;
		sshResult = null;
		stopPolling();
	}

	async function deleteSSHKey(keyFP: string, keyName?: string) {
		const displayName = keyName || keyFP;
		const confirmed = await ui.confirm({
			title: 'Delete SSH Key',
			message: `Are you sure you want to delete the SSH key "${displayName}"?`,
			confirmText: 'Delete',
			confirmVariant: 'danger'
		});
		
		if (!confirmed) return;

		deletingSSH = keyFP;
		try {
			await api.deleteSSHKey(keyFP);
			ui.success(`SSH key "${displayName}" deleted`);
			sshKeys = sshKeys.filter(k => k.ssh_key_fp !== keyFP);
		} catch (error) {
			if (error instanceof ApiClientError) {
				ui.showError('Failed to delete SSH key', error.description ?? error.code);
			}
		} finally {
			deletingSSH = null;
		}
	}

	function formatTimestamp(timestamp: number): string {
		return new Date(timestamp * 1000).toLocaleString();
	}

	function copyToClipboard(text: string) {
		navigator.clipboard.writeText(text);
		ui.success('Copied to clipboard');
	}

	function extractHostName(hostConfig: string): string {
		// Extract the Host name from the SSH config
		// Format: "Host mytoken-example\n    HostName ..."
		const match = hostConfig.match(/^Host\s+(\S+)/m);
		return match ? match[1] : 'mytoken-host';
	}

	function handleSSHKeyFileUpload(event: Event) {
		const input = event.target as HTMLInputElement;
		const file = input.files?.[0];
		if (!file) return;

		const reader = new FileReader();
		reader.onload = (e) => {
			const content = e.target?.result as string;
			if (content) {
				newKeyValue = content.trim();
				// Try to extract key name from file name if not already set
				if (!newKeyName.trim() && file.name) {
					// Remove .pub extension if present
					const baseName = file.name.replace(/\.pub$/, '').replace(/\.txt$/, '');
					newKeyName = baseName;
				}
			}
		};
		reader.onerror = () => {
			ui.showError('Error', 'Failed to read file');
		};
		reader.readAsText(file);
		
		// Reset file input so the same file can be selected again
		input.value = '';
	}
</script>

<div class="grants-settings">
	{#if loading}
		<LoadingSpinner message="Loading grants..." />
	{:else if grants.length === 0}
		<div class="alert alert-info">
			<i class="fas fa-info-circle me-2"></i>
			No grant types available.
		</div>
	{:else}
		<p class="text-muted mb-4">
			Enable or disable different methods for creating and obtaining mytokens.
		</p>

		<div class="accordion" id="grantsAccordion">
			{#each grants as grant, index}
				<div class="accordion-item">
					<div class="accordion-header">
						<!-- svelte-ignore a11y_click_events_have_key_events -->
						<!-- svelte-ignore a11y_no_static_element_interactions -->
						<div 
							class="grant-header d-flex justify-content-between align-items-center p-3"
							class:expandable={isExpandable(grant.type)}
							on:click={(e) => {
								// Only toggle if expandable and not clicking on the switch
								const target = e.target as HTMLElement;
								if (isExpandable(grant.type) && !target.closest('.form-check')) {
									toggleExpand(grant.type);
								}
							}}
						>
							<div class="d-flex align-items-center flex-grow-1">
								<div class="grant-icon me-3">
									<i class="fas {getGrantIcon(grant.type)} fa-lg" 
									   class:text-success={grant.enabled} 
									   class:text-muted={!grant.enabled}></i>
								</div>
								<div class="flex-grow-1">
									<span class="grant-label fw-semibold">{getGrantLabel(grant.type)}</span>
									<br />
									<small class="text-muted">{getGrantDescription(grant.type)}</small>
								</div>
							</div>
							<div class="d-flex align-items-center gap-3">
								<div class="form-check form-switch mb-0" on:click|stopPropagation>
									<input
										type="checkbox"
										class="form-check-input"
										id="grant-{grant.type}"
										checked={grant.enabled}
										disabled={updating === grant.type}
										on:change={() => toggleGrant(grant)}
										role="switch"
									/>
									<label class="form-check-label visually-hidden" for="grant-{grant.type}">
										{grant.enabled ? 'Disable' : 'Enable'} {getGrantLabel(grant.type)}
									</label>
									{#if updating === grant.type}
										<span class="spinner-border spinner-border-sm ms-2" role="status"></span>
									{/if}
								</div>
								{#if isExpandable(grant.type)}
									<span class="expand-icon">
										{#if expandedGrant === grant.type}
											<i class="fas fa-chevron-up"></i>
										{:else}
											<i class="fas fa-chevron-down"></i>
										{/if}
									</span>
								{/if}
							</div>
						</div>
					</div>
					
					{#if isExpandable(grant.type) && expandedGrant === grant.type}
						<div class="accordion-collapse">
							<div class="accordion-body">
								{#if grant.type === 'ssh'}
									<!-- SSH Keys Management -->
									{#if loadingSSH}
										<LoadingSpinner message="Loading SSH keys..." />
									{:else}
										<p class="text-muted small mb-3">
											To add a new SSH key, configure what can be done with it (capabilities), upload the key, 
											then complete the OIDC authorization flow to activate it.
										</p>

										<!-- SSH Grant Status -->
										<div class="mb-4">
											<div class="d-flex align-items-center gap-2">
												{#if sshGrantEnabled}
													<span class="badge bg-success">
														<i class="fas fa-check-circle me-1"></i>
														SSH Grant Enabled
													</span>
												{:else}
													<span class="badge bg-secondary">
														<i class="fas fa-times-circle me-1"></i>
														SSH Grant Disabled
													</span>
													<small class="text-muted">Enable the SSH grant to add SSH keys</small>
												{/if}
											</div>
										</div>

										{#if sshGrantEnabled}
											<!-- SSH Key Workflow States -->
											{#if sshResult}
												<!-- Success Result - show this first! -->
												<div class="card mb-4 border-success">
													<div class="card-header bg-success text-white">
														<h6 class="mb-0">
															<i class="fas fa-check-circle me-2"></i>
															SSH Key Added Successfully
														</h6>
													</div>
													<div class="card-body">
														{#if sshResult.ssh_user}
															<div class="alert alert-warning">
																<strong>Important:</strong> Save your SSH username and the configuration below! We cannot display them again.
															</div>
															<p>
																Your SSH username: 
																<code class="fs-5">{sshResult.ssh_user}</code>
																<button 
																	class="btn btn-sm btn-outline-secondary ms-2"
																	title="Copy username"
																	on:click={() => copyToClipboard(sshResult?.ssh_user ?? '')}
																>
																	<i class="fas fa-copy"></i>
																</button>
															</p>
														{/if}
														{#if sshResult.ssh_host_config}
															<div class="mb-3">
																<span class="form-label d-block">SSH Host Configuration:</span>
																<p class="text-muted small">Add this to your <code>~/.ssh/config</code> file:</p>
																<div class="position-relative">
																	<pre class="bg-dark text-light p-3 rounded mb-0"><code>{sshResult.ssh_host_config}</code></pre>
																	<button 
																		class="btn btn-sm btn-outline-light position-absolute top-0 end-0 m-2"
																		title="Copy to clipboard"
																		aria-label="Copy SSH host configuration to clipboard"
																		on:click={() => copyToClipboard(sshResult?.ssh_host_config ?? '')}
																	>
																		<i class="fas fa-copy"></i>
																	</button>
																</div>
															</div>

															<!-- Usage examples -->
															<div class="alert alert-info mt-3">
																<h6 class="alert-heading">
																	<i class="fas fa-terminal me-1"></i>
																	Using Your SSH Key
																</h6>
																<p class="mb-2">After adding the host configuration to your SSH config, you can use the following commands:</p>
																<p class="mb-1"><strong>Get an Access Token:</strong></p>
																<pre class="bg-dark text-light p-2 rounded mb-2"><code>ssh {extractHostName(sshResult.ssh_host_config)} AT</code></pre>
																<p class="mb-1"><strong>Get a Mytoken:</strong></p>
																<pre class="bg-dark text-light p-2 rounded mb-0"><code>ssh {extractHostName(sshResult.ssh_host_config)} MT</code></pre>
															</div>
														{/if}
														<button class="btn btn-primary" on:click={resetSSHForm}>
															Done
														</button>
													</div>
												</div>
											{:else if sshAuthorizationUri}
												<!-- Authorization Flow -->
												<div class="card mb-4 border-primary">
													<div class="card-header bg-primary text-white">
														<h6 class="mb-0">
															<i class="fas fa-external-link-alt me-2"></i>
															Complete Authorization
														</h6>
													</div>
													<div class="card-body">
														<p>To add your SSH key, follow this link and authenticate:</p>
														<div class="mb-3">
															<a href={sshAuthorizationUri} target="_blank" rel="noopener noreferrer" class="btn btn-lg btn-primary">
																<i class="fas fa-external-link-alt me-2"></i>
																Authenticate with OpenID Provider
															</a>
														</div>
														<div class="d-flex align-items-center gap-2">
															<span class="spinner-border spinner-border-sm text-primary"></span>
															<span class="text-muted">Waiting for authorization...</span>
														</div>
														<hr />
														<button class="btn btn-outline-secondary" on:click={resetSSHForm}>
															Cancel
														</button>
													</div>
												</div>
											{:else}
												<!-- Add SSH Key Form -->
												<div class="card mb-4">
													<div class="card-header">
														<h6 class="mb-0">
															<i class="fas fa-plus-circle me-2"></i>
															Add SSH Key
														</h6>
													</div>
													<div class="card-body">
														<form on:submit|preventDefault={addSSHKey}>
															<div class="mb-3">
																<label for="ssh-key-name" class="form-label">Key Name</label>
																<input
																	type="text"
																	id="ssh-key-name"
																	class="form-control"
																	placeholder="e.g., My Laptop"
																	bind:value={newKeyName}
																	maxlength="100"
																/>
																<small class="text-muted">A friendly name to identify this key.</small>
															</div>

															<div class="mb-3">
																<label for="ssh-key-value" class="form-label">SSH Public Key</label>
																<div class="input-group">
																	<textarea
																		id="ssh-key-value"
																		class="form-control font-monospace"
																		rows="3"
																		placeholder="ssh-rsa AAAAB3... or ssh-ed25519 AAAAC3..."
																		bind:value={newKeyValue}
																	></textarea>
																	<button 
																		type="button" 
																		class="btn btn-outline-secondary"
																		on:click={() => sshKeyFileInput?.click()}
																		title="Upload key file"
																	>
																		<i class="fas fa-upload"></i>
																	</button>
																</div>
																<input 
																	type="file" 
																	class="d-none" 
																	accept=".pub,.txt"
																	bind:this={sshKeyFileInput}
																	on:change={handleSSHKeyFileUpload}
																/>
																<small class="text-muted">
																	Paste your public key or upload a file (usually <code>~/.ssh/id_rsa.pub</code> or <code>~/.ssh/id_ed25519.pub</code>).
																</small>
															</div>

															<!-- Advanced options toggle -->
															<div class="mb-3">
																<button
																	type="button"
																	class="btn btn-link p-0 text-decoration-none"
																	on:click={() => {
																		showAdvanced = !showAdvanced;
																		if (showAdvanced) {
																			loadCapabilities();
																		}
																	}}
																>
																	<i class="fas" class:fa-chevron-down={showAdvanced} class:fa-chevron-right={!showAdvanced}></i>
																	Advanced Options
																</button>
															</div>

															{#if showAdvanced}
																<div class="advanced-options border rounded p-3 mb-3">
																	<p class="text-muted small mb-3">
																		Configure the capabilities and restrictions for tokens created with this SSH key.
																		Default capabilities (AT, tokeninfo, create subtoken) are pre-selected.
																	</p>

																	<div class="mb-3">
																		{#if loadingCapabilities}
																			<LoadingSpinner message="Loading capabilities..." />
																		{:else}
																			<CapabilityTree
																				bind:this={capabilityTreeRef}
																				capabilities={webCapabilities}
																				{selectedCapabilities}
																				collapsed={false}
																				showTemplates={false}
																			/>
																		{/if}
																	</div>

																	<div class="mb-3">
																		<RestrictionsEditor
																			bind:this={restrictionsEditorRef}
																			collapsed={true}
																			showTemplates={false}
																		/>
																	</div>
																</div>
															{/if}

															<button type="submit" class="btn btn-primary" disabled={addingSSH || !newKeyName.trim() || !newKeyValue.trim()}>
																{#if addingSSH}
																	<span class="spinner-border spinner-border-sm me-1"></span>
																	Adding...
																{:else}
																	<i class="fas fa-plus me-1"></i>
																	Add SSH Key
																{/if}
															</button>
														</form>
													</div>
												</div>
											{/if}

											<!-- SSH Keys Table -->
											<div class="card">
												<div class="card-header d-flex justify-content-between align-items-center">
													<h6 class="mb-0">
														<i class="fas fa-key me-2"></i>
														Your SSH Keys
													</h6>
													<span class="badge bg-secondary">{sshKeys.length} keys</span>
												</div>
												<div class="card-body p-0">
													{#if sshKeys.length === 0}
														<div class="text-center py-4 text-muted">
															<i class="fas fa-key fa-2x mb-2 opacity-50"></i>
															<p class="mb-0">No SSH keys registered</p>
														</div>
													{:else}
														<div class="table-responsive">
															<table class="table table-hover mb-0">
																<thead>
																	<tr>
																		<th>Name</th>
																		<th>Fingerprint</th>
																		<th>Created</th>
																		<th>Last Used</th>
																		<th></th>
																	</tr>
																</thead>
																<tbody>
																	{#each sshKeys as key}
																		<tr>
																			<td>{key.name || '(unnamed)'}</td>
																			<td><code class="small">{key.ssh_key_fp || 'N/A'}</code></td>
																			<td class="small">{formatTimestamp(key.created)}</td>
																			<td class="small">{key.last_used ? formatTimestamp(key.last_used) : 'Never'}</td>
																			<td>
																				<button
																					type="button"
																					class="btn btn-sm btn-outline-danger"
																					title="Delete key"
																					disabled={deletingSSH === key.ssh_key_fp}
																					on:click={() => deleteSSHKey(key.ssh_key_fp ?? '', key.name)}
																				>
																					{#if deletingSSH === key.ssh_key_fp}
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
									{/if}
								{/if}
							</div>
						</div>
					{/if}
				</div>
			{/each}
		</div>

		<div class="alert alert-warning mt-4">
			<i class="fas fa-exclamation-triangle me-2"></i>
			<strong>Note:</strong> Disabling a grant type will prevent you from using that method to create new mytokens.
			Existing mytokens are not affected.
		</div>
	{/if}
</div>

<style>
	.grant-icon {
		width: 40px;
		text-align: center;
	}

	.grant-header {
		background-color: var(--bs-accordion-bg, #fff);
		border-bottom: 1px solid var(--bs-accordion-border-color, rgba(0,0,0,.125));
	}

	.grant-header.expandable {
		cursor: pointer;
	}

	.grant-header.expandable:hover {
		background-color: #f8f9fa;
	}

	.expand-icon {
		color: #6c757d;
		padding: 0.25rem 0.5rem;
	}

	.accordion-item:last-child .grant-header {
		border-bottom: none;
	}

	.accordion-item:last-child .accordion-collapse .accordion-body {
		border-top: 1px solid var(--bs-accordion-border-color, rgba(0,0,0,.125));
	}

	.form-switch .form-check-input {
		width: 3em;
		height: 1.5em;
		cursor: pointer;
	}

	.form-switch .form-check-input:checked {
		background-color: #28a745;
		border-color: #28a745;
	}

	.accordion-body {
		background-color: #f8f9fa;
	}

	.advanced-options {
		background-color: #fff;
	}

	textarea.form-control {
		font-size: 0.875rem;
	}

	pre {
		margin-bottom: 0;
	}

	code {
		font-size: 0.85em;
	}

	.table th {
		border-top: none;
		font-weight: 600;
		font-size: 0.875rem;
	}
</style>
