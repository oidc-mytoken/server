<script lang="ts">
	import { onMount } from 'svelte';
	import type { Grant } from '$lib/types';
	import { api, ApiClientError } from '$lib/api/client';
	import { ui } from '$lib/stores/ui';
	import LoadingSpinner from '../LoadingSpinner.svelte';

	let grants: Grant[] = [];
	let loading = true;
	let updating: string | null = null;

	// Grant type descriptions
	const grantDescriptions: Record<string, { label: string; description: string; icon: string }> = {
		'oidc_flow': {
			label: 'OIDC Flow',
			description: 'Create mytokens via browser-based OIDC login',
			icon: 'fa-openid'
		},
		'polling_code': {
			label: 'Polling Code',
			description: 'Create mytokens using device authorization flow',
			icon: 'fa-mobile-alt'
		},
		'transfer_code': {
			label: 'Transfer Code',
			description: 'Transfer mytokens between devices using short codes',
			icon: 'fa-exchange-alt'
		},
		'mytoken': {
			label: 'Mytoken',
			description: 'Create new mytokens using an existing mytoken',
			icon: 'fa-key'
		},
		'ssh': {
			label: 'SSH',
			description: 'Obtain mytokens using SSH key authentication',
			icon: 'fa-terminal'
		}
	};

	onMount(async () => {
		await loadGrants();
	});

	async function loadGrants() {
		loading = true;
		try {
			grants = await api.getGrants();
		} catch (error) {
			if (error instanceof ApiClientError) {
				ui.showError('Failed to load grants', error.description ?? error.code);
			}
		} finally {
			loading = false;
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
		} catch (error) {
			if (error instanceof ApiClientError) {
				ui.showError('Failed to update grant', error.description ?? error.code);
			}
		} finally {
			updating = null;
		}
	}

	function getGrantLabel(type: string): string {
		return grantDescriptions[type]?.label ?? type;
	}

	function getGrantDescription(type: string): string {
		return grantDescriptions[type]?.description ?? '';
	}

	function getGrantIcon(type: string): string {
		return grantDescriptions[type]?.icon ?? 'fa-check-circle';
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

		<div class="list-group">
			{#each grants as grant}
				<div class="list-group-item d-flex justify-content-between align-items-center">
					<div class="d-flex align-items-center">
						<div class="grant-icon me-3">
							<i class="fas {getGrantIcon(grant.type)} fa-lg" class:text-success={grant.enabled} class:text-muted={!grant.enabled}></i>
						</div>
						<div>
							<h6 class="mb-0">{getGrantLabel(grant.type)}</h6>
							<small class="text-muted">{getGrantDescription(grant.type)}</small>
						</div>
					</div>
					<div class="form-check form-switch">
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

	.list-group-item {
		border-left: none;
		border-right: none;
	}

	.list-group-item:first-child {
		border-top: none;
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
</style>
