<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import type { ConsentData, Restriction, Rotation, WebCapability, CreateMytokenTag } from '$lib/types';
	import { api, ApiClientError } from '$lib/api/client';
	import { ui } from '$lib/stores/ui';
	import { tags } from '$lib/stores/tags';
	import { generateTagColor } from '$lib/utils/color';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
	import CapabilityTree from '$lib/components/capabilities/CapabilityTree.svelte';
	import RestrictionsEditor from '$lib/components/restrictions/RestrictionsEditor.svelte';
	import RotationSettings from '$lib/components/RotationSettings.svelte';
	import TagPill from '$lib/components/TagPill.svelte';

	let loading = $state(true);
	let submitting = $state(false);
	let error = $state('');
	let consentData = $state<ConsentData | null>(null);

	// Editable fields
	let tokenName = $state('');
	let selectedCapabilities: string[] = $state([]);
	let restrictions: Restriction[] = $state([]);
	let rotation: Rotation | undefined = $state(undefined);
	let selectedTags: CreateMytokenTag[] = $state([]);

	let consentCode = $derived($page.params.code);

	onMount(() => {
		loadConsentData();
	});



	async function loadConsentData() {
		if (!consentCode) {
			error = 'No consent code provided';
			loading = false;
			return;
		}

		loading = true;
		error = '';

		try {
			const data = await api.getConsentData(consentCode);
			consentData = data;
			
			// Initialize editable fields from consent data
			tokenName = data.token_name || '';
			selectedCapabilities = [...(data.capabilities || [])];
			restrictions = data.restrictions ? [...data.restrictions] : [];
			rotation = data.rotation ? { ...data.rotation } : undefined;
			selectedTags = data.tags ? [...data.tags] : [];
		} catch (err) {
			console.error('Error loading consent data:', err);
			if (err instanceof ApiClientError) {
				if (err.status === 404) {
					error = 'This consent link has expired or is invalid.';
				} else {
					error = err.description ?? err.code;
				}
			} else {
				error = (err as Error).message;
			}
		} finally {
			loading = false;
			console.log('Finally block, loading set to false:', loading);
		}
	}

	// Create a reactive map of tag colors
	let tagColors = $derived.by(() => {
		const colorMap = new Map<string, string>();
		for (const tag of $tags.tags) {
			if (tag.color) {
				colorMap.set(tag.tag, tag.color);
			}
		}
		return colorMap;
	});

	function getTagColor(tagName: string): string {
		// Use color from user's tags store if available, otherwise generate from tag name hash
		return tagColors.get(tagName) ?? generateTagColor(tagName);
	}

	async function handleApprove() {
		if (!consentData || !consentCode) return;

		submitting = true;

		try {
			const response = await api.approveConsent(consentCode, {
				oidc_issuer: consentData.oidc_issuer,
				capabilities: selectedCapabilities,
				restrictions: restrictions.length > 0 ? restrictions : undefined,
				rotation: rotation,
				name: tokenName || undefined,
				tags: selectedTags.length > 0 ? selectedTags : undefined
			});

			// Redirect to authorization URL
			window.location.href = response.authorization_uri;
		} catch (err) {
			submitting = false;
			if (err instanceof ApiClientError) {
				ui.showError('Consent Failed', err.description ?? err.code);
			} else {
				ui.showError('Error', (err as Error).message);
			}
		}
	}

	async function handleDecline() {
		if (!consentCode) return;
		
		submitting = true;

		try {
			const response = await api.declineConsent(consentCode);
			// Redirect to the returned URL
			window.location.href = response.url;
		} catch (err) {
			submitting = false;
			if (err instanceof ApiClientError) {
				ui.showError('Error', err.description ?? err.code);
			} else {
				// If decline fails, just go home
				goto('/');
			}
		}
	}

	function handleCapabilitiesChange(event: CustomEvent<string[]>) {
		selectedCapabilities = event.detail;
	}

	// Get provider name from issuer URL
	function getProviderName(issuer: string): string {
		try {
			const url = new URL(issuer);
			return url.hostname;
		} catch {
			return issuer;
		}
	}

	// Tag helper functions
	function toggleIncludeChildren(tagName: string) {
		const index = selectedTags.findIndex(t => t.tag === tagName);
		if (index >= 0) {
			selectedTags[index].include_children = !selectedTags[index].include_children;
			selectedTags = [...selectedTags];
		}
	}

	function removeTag(tagName: string) {
		selectedTags = selectedTags.filter(t => t.tag !== tagName);
	}
</script>

<svelte:head>
	<title>mytoken - Authorization Request</title>
</svelte:head>

<div class="consent-page">
	<div class="row justify-content-center">
		<div class="col-lg-10 col-xl-8">
			{#if loading}
				<div class="text-center py-5">
					<LoadingSpinner message="Loading authorization request..." />
				</div>
			{:else if error}
				<div class="card shadow">
					<div class="card-body text-center py-5">
						<i class="fas fa-exclamation-triangle fa-4x text-danger mb-4"></i>
						<h4>Authorization Request Error</h4>
						<p class="text-muted mb-4">{error}</p>
						<a href="/" class="btn btn-primary">
							<i class="fas fa-home me-2"></i>
							Go to Home
						</a>
					</div>
				</div>
			{:else if consentData}
				<div class="card shadow">
					<div class="card-header">
						<h4 class="mb-0">
							<i class="fas fa-shield-alt me-2"></i>
							Authorization Request
							{#if consentData.application_name}
								<small class="text-white-50 ms-2">({consentData.application_name})</small>
							{/if}
						</h4>
					</div>
					<div class="card-body">
						<!-- Issuer Info -->
						<div class="alert alert-info mb-4">
							<i class="fas fa-info-circle me-2"></i>
							You are about to create a mytoken for 
							<strong>{getProviderName(consentData.oidc_issuer)}</strong>
						</div>

						<!-- Token Name -->
						<div class="mb-4">
							<label for="token-name" class="form-label">
								<i class="fas fa-tag me-1"></i>
								Token Name (optional)
							</label>
							<input
								type="text"
								id="token-name"
								class="form-control"
								placeholder="Give this token a memorable name..."
								bind:value={tokenName}
							/>
							<div class="form-text">
								A name helps you identify this token later.
							</div>
						</div>

						<!-- Tags -->
						{#if selectedTags.length > 0}
							<div class="mb-4">
								<span class="form-label d-block">
									<i class="fas fa-tags me-1"></i>
									Tags
								</span>
								<div class="selected-tags">
									{#each selectedTags as tagInfo}
										<TagPill 
											tag={tagInfo.tag}
											color={getTagColor(tagInfo.tag)}
											includeChildren={tagInfo.include_children}
											removable
											onRemove={() => removeTag(tagInfo.tag)}
											onToggleChildren={() => toggleIncludeChildren(tagInfo.tag)}
										/>
									{/each}
								</div>
								<small class="text-muted d-block mt-1">
									<i class="fas fa-sitemap me-1"></i> Click a tag to toggle "include children".
								</small>
							</div>
						{/if}

						<!-- Capabilities -->
						<div class="mb-4">
							<CapabilityTree
								capabilities={consentData.all_capabilities}
								selectedCapabilities={selectedCapabilities}
								on:change={handleCapabilitiesChange}
								collapsed={true}
							/>
						</div>

						<!-- Restrictions -->
						<div class="mb-4">
							<RestrictionsEditor
								bind:restrictions={restrictions}
								collapsed={true}
							/>
						</div>

						<!-- Rotation -->
						<div class="mb-4">
							<RotationSettings
								bind:rotation={rotation}
								collapsed={true}
							/>
						</div>

						<hr class="my-4" />

						<!-- Action Buttons -->
						<div class="d-grid gap-2 d-md-flex justify-content-md-end">
							<button
								class="btn btn-outline-secondary"
								onclick={handleDecline}
								disabled={submitting}
							>
								<i class="fas fa-times me-2"></i>
								Decline
							</button>
							<button
								class="btn btn-success"
								onclick={handleApprove}
								disabled={submitting}
							>
								{#if submitting}
									<span class="spinner-border spinner-border-sm me-2" role="status"></span>
									Processing...
								{:else}
									<i class="fas fa-check me-2"></i>
									Authorize
								{/if}
							</button>
						</div>
					</div>
				</div>

				<!-- Info Box -->
				<div class="card mt-4">
					<div class="card-body">
						<h6 class="card-title">
							<i class="fas fa-info-circle text-info me-2"></i>
							What happens next?
						</h6>
						<p class="card-text text-muted mb-0">
							After clicking "Authorize", you will be redirected to 
							<strong>{getProviderName(consentData.oidc_issuer)}</strong> to authenticate.
							Once authenticated, a mytoken will be created with the capabilities and 
							restrictions you've configured above.
						</p>
					</div>
				</div>
			{/if}
		</div>
	</div>
</div>

<style>
	.card {
		border: none;
	}

	.card-header {
		background-color: #df691a;
		color: white;
	}

	.consent-page {
		padding-bottom: 2rem;
	}
</style>
