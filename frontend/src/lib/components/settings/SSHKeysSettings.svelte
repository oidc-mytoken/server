<script lang="ts">
	import { onMount } from 'svelte';
	import type { SSHKeyInfo, Capability, Restriction } from '$lib/types';
	import { api, ApiClientError } from '$lib/api/client';
	import { ui } from '$lib/stores/ui';
	import LoadingSpinner from '../LoadingSpinner.svelte';
	import CapabilityTree from '../capabilities/CapabilityTree.svelte';
	import RestrictionsEditor from '../restrictions/RestrictionsEditor.svelte';
	import CollapsibleSection from '../CollapsibleSection.svelte';

	// NOTE: This component is deprecated. SSH key management is now part of GrantsSettings.
	// Users accessing /settings/ssh will be redirected to /settings#ssh

	let sshKeys: SSHKeyInfo[] = [];
	let loading = true;
	let adding = false;
	let deleting: string | null = null;

	// New key form
	let newKeyName = '';
	let newKeyValue = '';
	let newKeyCapabilities: Capability[] = [];
	let newKeyRestrictions: Restriction[] = [];
	let showAdvanced = false;

	// Component refs
	let capabilityTreeRef: CapabilityTree;
	let restrictionsEditorRef: RestrictionsEditor;

	onMount(async () => {
		await loadSSHKeys();
	});

	async function loadSSHKeys() {
		loading = true;
		try {
			sshKeys = await api.getSSHKeys();
		} catch (error) {
			if (error instanceof ApiClientError) {
				ui.error('Failed to load SSH keys');
			}
		} finally {
			loading = false;
		}
	}

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
		if (!newKeyValue.trim().startsWith('ssh-') && !newKeyValue.trim().startsWith('ecdsa-')) {
			ui.showError('Error', 'Invalid SSH public key format. Key should start with ssh-rsa, ssh-ed25519, etc.');
			return;
		}

		adding = true;
		try {
			const key: { name?: string; ssh_key: string; restrictions?: Restriction[]; capabilities?: string[] } = {
				name: newKeyName.trim(),
				ssh_key: newKeyValue.trim()
			};

			// Add capabilities if any selected
			if (capabilityTreeRef) {
				const caps = capabilityTreeRef.getEnabledCapabilities();
				if (caps.length > 0) {
					key.capabilities = caps;
				}
			}

			// Add restrictions if any defined
			if (restrictionsEditorRef) {
				const restr = restrictionsEditorRef.getRestrictions();
				if (restr.length > 0) {
					key.restrictions = restr;
				}
			}

			await api.addSSHKey(key);
			ui.success(`SSH key "${key.name}" added`);
			
			// Reset form
			newKeyName = '';
			newKeyValue = '';
			newKeyCapabilities = [];
			newKeyRestrictions = [];
			showAdvanced = false;
			
			// Reload list
			await loadSSHKeys();
		} catch (error) {
			if (error instanceof ApiClientError) {
				ui.showError('Failed to add SSH key', error.description ?? error.code);
			}
		} finally {
			adding = false;
		}
	}

	async function deleteSSHKey(keyName: string) {
		const confirmed = await ui.confirm({
			title: 'Delete SSH Key',
			message: `Are you sure you want to delete the SSH key "${keyName}"?`,
			confirmText: 'Delete',
			confirmVariant: 'danger'
		});
		
		if (!confirmed) return;

		deleting = keyName;
		try {
			await api.deleteSSHKey(keyName);
			ui.success(`SSH key "${keyName}" deleted`);
			sshKeys = sshKeys.filter(k => k.name !== keyName);
		} catch (error) {
			if (error instanceof ApiClientError) {
				ui.showError('Failed to delete SSH key', error.description ?? error.code);
			}
		} finally {
			deleting = null;
		}
	}

	function formatKeyPreview(key: string): string {
		const parts = key.split(' ');
		if (parts.length >= 2) {
			const keyType = parts[0];
			const keyData = parts[1];
			const comment = parts.slice(2).join(' ');
			const preview = keyData.length > 20 
				? keyData.slice(0, 10) + '...' + keyData.slice(-10)
				: keyData;
			return `${keyType} ${preview}${comment ? ' ' + comment : ''}`;
		}
		return key.length > 50 ? key.slice(0, 50) + '...' : key;
	}
</script>

<div class="ssh-keys-settings">
	<p class="text-muted mb-4">
		SSH keys allow you to obtain mytokens using SSH authentication instead of browser-based OIDC login.
		This is useful for automated scripts and command-line tools.
	</p>

	{#if loading}
		<LoadingSpinner message="Loading SSH keys..." />
	{:else}
		<!-- Add new SSH key -->
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
						<label for="key-name" class="form-label">Key Name</label>
						<input
							type="text"
							id="key-name"
							class="form-control"
							placeholder="e.g., My Laptop"
							bind:value={newKeyName}
							maxlength="100"
						/>
						<small class="text-muted">A friendly name to identify this key.</small>
					</div>

					<div class="mb-3">
						<label for="key-value" class="form-label">Public Key</label>
						<textarea
							id="key-value"
							class="form-control font-monospace"
							rows="3"
							placeholder="ssh-rsa AAAAB3... or ssh-ed25519 AAAAC3..."
							bind:value={newKeyValue}
						></textarea>
						<small class="text-muted">
							Paste your public key (usually found in <code>~/.ssh/id_rsa.pub</code> or <code>~/.ssh/id_ed25519.pub</code>).
						</small>
					</div>

					<!-- Advanced options toggle -->
					<div class="mb-3">
						<button
							type="button"
							class="btn btn-link p-0 text-decoration-none"
							on:click={() => showAdvanced = !showAdvanced}
						>
							<i class="fas" class:fa-chevron-down={showAdvanced} class:fa-chevron-right={!showAdvanced}></i>
							Advanced Options
						</button>
					</div>

					{#if showAdvanced}
						<div class="advanced-options border rounded p-3 mb-3">
							<p class="text-muted small mb-3">
								Optionally restrict the capabilities and usage of tokens created with this SSH key.
							</p>

							<div class="mb-3">
								<CapabilityTree
									bind:this={capabilityTreeRef}
									bind:capabilities={newKeyCapabilities}
									collapsed={true}
									showTemplates={false}
								/>
							</div>

							<div class="mb-3">
								<RestrictionsEditor
									bind:this={restrictionsEditorRef}
									bind:restrictions={newKeyRestrictions}
									collapsed={true}
									showTemplates={false}
								/>
							</div>
						</div>
					{/if}

					<button type="submit" class="btn btn-primary" disabled={adding || !newKeyName.trim() || !newKeyValue.trim()}>
						{#if adding}
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

		<!-- Existing SSH keys -->
		<div class="card">
			<div class="card-header d-flex justify-content-between align-items-center">
				<h6 class="mb-0">
					<i class="fas fa-key me-2"></i>
					Your SSH Keys
				</h6>
				<span class="badge bg-secondary">{sshKeys.length} keys</span>
			</div>
			<div class="card-body">
				{#if sshKeys.length === 0}
					<div class="text-center py-4 text-muted">
						<i class="fas fa-key fa-3x mb-3"></i>
						<p>No SSH keys registered</p>
						<p class="small">Add your first SSH key above to enable SSH-based token creation.</p>
					</div>
				{:else}
					<div class="list-group list-group-flush">
						{#each sshKeys as key}
							<div class="list-group-item d-flex justify-content-between align-items-start">
								<div class="flex-grow-1">
									<h6 class="mb-1">{key.name || '(unnamed)'}</h6>
									<code class="small text-muted">{key.ssh_key_fp || key.ssh_key || 'N/A'}</code>
									<div class="mt-1">
										<small class="text-muted">
											Created: {new Date(key.created * 1000).toLocaleString()}
											{#if key.last_used}
												| Last used: {new Date(key.last_used * 1000).toLocaleString()}
											{/if}
										</small>
									</div>
								</div>
								<button
									type="button"
									class="btn btn-sm btn-outline-danger"
									title="Delete key"
									disabled={deleting === (key.ssh_key_fp ?? key.name ?? '')}
									on:click={() => deleteSSHKey(key.ssh_key_fp ?? '')}
								>
									{#if deleting === (key.ssh_key_fp ?? key.name ?? '')}
										<span class="spinner-border spinner-border-sm"></span>
									{:else}
										<i class="fas fa-trash"></i>
									{/if}
								</button>
							</div>
						{/each}
					</div>
				{/if}
			</div>
		</div>

		<!-- Help box -->
		<div class="alert alert-info mt-4">
			<h6 class="alert-heading">
				<i class="fas fa-terminal me-1"></i>
				Using SSH Keys
			</h6>
			<p>Once you've registered an SSH key, you can obtain mytokens using:</p>
			<pre class="mb-0 bg-dark text-light p-2 rounded"><code>ssh mytoken@{window.location.hostname}</code></pre>
		</div>
	{/if}
</div>

<style>
	.card-header {
		background-color: var(--bs-tertiary-bg);
	}

	textarea.form-control {
		font-size: 0.875rem;
	}

	.advanced-options {
		background-color: var(--bs-tertiary-bg);
	}

	.list-group-item {
		border-left: none;
		border-right: none;
	}

	.list-group-item:first-child {
		border-top: none;
	}

	.list-group-item:last-child {
		border-bottom: none;
	}

	pre {
		margin-bottom: 0;
	}

	code {
		font-size: 0.85em;
	}
</style>
