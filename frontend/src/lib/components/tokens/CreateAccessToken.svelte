<script lang="ts">
	import { onMount } from 'svelte';
	import { browser } from '$app/environment';
	import { api, ApiClientError } from '$lib/api/client';
	import { isLoggedIn } from '$lib/stores/auth';
	import { providers } from '$lib/stores/discovery';
	import { ui } from '$lib/stores/ui';
	import { extractOidcIssuer } from '$lib/utils/jwt';
	import CopyButton from '../CopyButton.svelte';

	// OpenID Core specification default scopes (excluding offline_access which is not applicable for access tokens)
	const OPENID_CORE_SCOPES = ['openid', 'profile', 'email', 'address', 'phone'];

	// Form state
	let mytoken = ''; // Optional - if empty, cookie auth is used
	let scope = ''; // For additional custom scopes
	let audiences: string[] = [];
	let newAudience = '';
	let sessionIssuer: string | null = null;

	// Result state
	let accessToken = '';
	let tokenType = '';
	let expiresIn = 0;
	let loading = false;
	let showResult = false;

	onMount(() => {
		// Get session issuer from sessionStorage (set by auth store)
		if (browser) {
			sessionIssuer = sessionStorage.getItem('mytoken_oidc_issuer');
		}
	});

	// Extract issuer from pasted mytoken (if it's a JWT)
	$: mytokenIssuer = mytoken.trim() ? extractOidcIssuer(mytoken.trim()) : null;

	// Determine effective issuer: mytoken takes priority, then session, then null
	$: effectiveIssuer = mytokenIssuer ?? sessionIssuer ?? null;

	// Get scopes for the effective issuer
	$: availableScopes = getAvailableScopesForIssuer(effectiveIssuer);

	// Parse selected scopes from the scope string (scopes selected via checkboxes)
	$: selectedScopes = scope.split(' ').filter((x) => x);

	// Check if all scopes are allowed (none explicitly selected)
	$: allScopesAllowed = selectedScopes.length === 0;

	function getAvailableScopesForIssuer(issuer: string | null): string[] {
		if (!issuer) return OPENID_CORE_SCOPES;
		const provider = $providers.find((p) => p.issuer === issuer);
		const scopes = provider?.scopes_supported ?? OPENID_CORE_SCOPES;
		// Filter out offline_access as it's not applicable for access tokens
		return scopes.filter((s) => s !== 'offline_access');
	}

	async function handleSubmit() {
		// If not logged in and no token provided, error
		if (!$isLoggedIn && !mytoken.trim()) {
			ui.showError('Error', 'Please log in or enter a mytoken');
			return;
		}

		loading = true;
		showResult = false;
		accessToken = '';

		try {
			const response = await api.createAccessToken({
				grant_type: 'mytoken',
				mytoken: mytoken.trim() || undefined, // If empty, cookie auth is used
				scope: scope.trim() || undefined,
				audience: audiences.length > 0 ? audiences.join(' ') : undefined
			});

			accessToken = response.access_token;
			tokenType = response.token_type;
			expiresIn = response.expires_in ?? 0;
			showResult = true;
			ui.success('Access token created successfully!');
		} catch (error) {
			if (error instanceof ApiClientError) {
				ui.showError('Failed to create access token', error.description ?? error.code);
			} else {
				ui.showError('Error', (error as Error).message);
			}
		} finally {
			loading = false;
		}
	}

	function resetForm() {
		showResult = false;
		accessToken = '';
		scope = '';
		audiences = [];
		newAudience = '';
	}

	function toggleScope(s: string) {
		const scopes = scope.split(' ').filter((x) => x);
		if (scopes.includes(s)) {
			scope = scopes.filter((x) => x !== s).join(' ');
		} else {
			scope = [...scopes, s].join(' ');
		}
	}

	function addAudience() {
		if (newAudience.trim() && !audiences.includes(newAudience.trim())) {
			audiences = [...audiences, newAudience.trim()];
			newAudience = '';
		}
	}

	function removeAudience(idx: number) {
		audiences = audiences.filter((_, i) => i !== idx);
	}
</script>

<div class="create-access-token">
	{#if showResult}
		<!-- Result display -->
		<div class="result-section">
			<div class="alert alert-success">
				<h5 class="alert-heading">
					<i class="fas fa-check-circle me-2"></i>
					Access Token Created!
				</h5>
				<p class="mb-0">
					Your access token is ready. It will expire in {Math.floor(expiresIn / 60)} minutes.
				</p>
			</div>

			<div class="card mb-4">
				<div class="card-header d-flex justify-content-between align-items-center">
					<span>
						<i class="fas fa-key me-1"></i>
						Access Token
						<span class="badge bg-secondary ms-2">{tokenType}</span>
					</span>
					<CopyButton value={accessToken} label="Copy Token" />
				</div>
				<div class="card-body">
					<div class="token-display">
						<code class="text-break">{accessToken}</code>
					</div>
				</div>
			</div>

			<div class="d-flex gap-2">
				<button type="button" class="btn btn-primary" on:click={handleSubmit}>
					<i class="fas fa-sync me-1"></i>
					Get Another
				</button>
				<button type="button" class="btn btn-outline-secondary" on:click={resetForm}>
					<i class="fas fa-edit me-1"></i>
					Change Settings
				</button>
			</div>
		</div>
	{:else}
		<!-- Form -->
		<form on:submit|preventDefault={handleSubmit}>
			<!-- Mytoken input -->
			<div class="mb-4">
				<label for="mytoken-input" class="form-label">
					<i class="fas fa-ticket-alt me-1"></i>
					Mytoken
					{#if !$isLoggedIn}
						<span class="text-danger">*</span>
					{/if}
				</label>
				<textarea
					id="mytoken-input"
					class="form-control font-monospace"
					rows="3"
					placeholder={$isLoggedIn 
						? "Leave empty to use your session, or paste a specific mytoken..." 
						: "Paste your mytoken here..."}
					bind:value={mytoken}
					required={!$isLoggedIn}
				></textarea>
				{#if !$isLoggedIn && !mytoken}
					<small class="text-warning">
						<i class="fas fa-exclamation-triangle me-1"></i>
						You must provide a mytoken to obtain an access token when not logged in.
					</small>
				{:else if $isLoggedIn && !mytoken}
					<small class="text-muted">
						<i class="fas fa-info-circle me-1"></i>
						Using your current session (cookie auth).
					</small>
				{/if}
				{#if mytokenIssuer}
					<small class="text-success d-block">
						<i class="fas fa-check-circle me-1"></i>
						Detected provider: {mytokenIssuer}
					</small>
				{:else if mytoken.trim() && !mytokenIssuer}
					<small class="text-warning d-block">
						<i class="fas fa-exclamation-triangle me-1"></i>
						Could not detect provider from token (using {sessionIssuer ? 'session provider' : 'default scopes'}).
					</small>
				{/if}
			</div>

			<!-- Scope selection -->
			<div class="mb-4">
				<span class="form-label d-block">
					<i class="fas fa-list me-1"></i>
					Scope (optional)
				</span>

				{#if availableScopes.length > 0}
					<div class="scope-list border rounded p-2 mb-2">
						<table class="table table-sm table-striped mb-0">
							<tbody>
								{#each availableScopes as s}
									{@const isSelected = selectedScopes.includes(s)}
									<tr
										class="scope-row"
										on:click={() => toggleScope(s)}
										role="button"
										tabindex="0"
										on:keydown={(e) =>
											(e.key === 'Enter' || e.key === ' ') && toggleScope(s)}
									>
										<td class="align-middle">{s}</td>
										<td class="align-middle text-center" style="width: 40px;">
											{#if allScopesAllowed || isSelected}
												<i class="fas fa-check-circle text-success"></i>
											{:else}
												<i class="fas fa-times-circle text-danger"></i>
											{/if}
										</td>
										<td class="align-middle text-center" style="width: 40px;">
											<input
												type="checkbox"
												class="form-check-input"
												checked={isSelected}
												on:click|stopPropagation
												on:change={() => toggleScope(s)}
											/>
										</td>
									</tr>
								{/each}
							</tbody>
						</table>
					</div>
					<small class="text-muted d-block mb-2">
						{#if allScopesAllowed}
							No scopes selected - default scopes will be used.
						{:else}
							Only the selected scopes will be requested.
						{/if}
					</small>
				{/if}

				<input
					type="text"
					class="form-control"
					placeholder="Additional custom scopes (space-separated)"
					bind:value={scope}
				/>
				<small class="text-muted"> Enter additional scopes not listed above. </small>
			</div>

			<!-- Audience -->
			<div class="mb-4">
				<span class="form-label d-block">
					<i class="fas fa-server me-1"></i>
					Audience (optional)
				</span>
				<div class="tags-container mb-2">
					{#each audiences as aud, idx}
						<span class="badge bg-info me-1 mb-1">
							{aud}
							<button
								type="button"
								class="btn-close btn-close-white ms-1"
								style="font-size: 0.6em;"
								on:click={() => removeAudience(idx)}
								title="Remove {aud}"
							></button>
						</span>
					{/each}
				</div>
				<div class="input-group input-group-sm">
					<input
						type="text"
						class="form-control"
						placeholder="https://api.example.com"
						bind:value={newAudience}
						on:keydown={(e) => e.key === 'Enter' && (e.preventDefault(), addAudience())}
					/>
					<button
						class="btn btn-outline-primary"
						type="button"
						on:click={addAudience}
						title="Add audience"
					>
						<i class="fas fa-plus"></i>
					</button>
				</div>
				<small class="text-muted"> The intended audience(s) for the access token. </small>
			</div>

			<!-- Submit -->
			<div class="d-grid">
				<button type="submit" class="btn btn-primary btn-lg" disabled={loading}>
					{#if loading}
						<span class="spinner-border spinner-border-sm me-2" role="status"></span>
						Creating...
					{:else}
						<i class="fas fa-key me-2"></i>
						Get Access Token
					{/if}
				</button>
			</div>
		</form>
	{/if}
</div>

<style>
	.token-display {
		background-color: #f8f9fa;
		padding: 1rem;
		border-radius: 0.375rem;
		max-height: 200px;
		overflow-y: auto;
	}

	.token-display code {
		font-size: 0.875rem;
		word-break: break-all;
	}

	textarea.form-control {
		font-size: 0.875rem;
	}

	.scope-list {
		background-color: #f8f9fa;
	}

	.scope-list table {
		margin-bottom: 0;
	}

	.scope-row {
		cursor: pointer;
		user-select: none;
	}

	.scope-row:hover {
		background-color: #e9ecef !important;
	}

	.scope-list tr:last-child td {
		border-bottom: none;
	}

	.tags-container {
		min-height: 24px;
	}

	.badge {
		font-weight: normal;
	}
</style>
