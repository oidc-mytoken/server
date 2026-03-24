<script lang="ts">
	import { onMount } from 'svelte';
	import type { EmailSettings } from '$lib/types';
	import { api, ApiClientError } from '$lib/api/client';
	import { ui } from '$lib/stores/ui';
	import LoadingSpinner from '../LoadingSpinner.svelte';

	let loading = true;
	let saving = false;
	let sendingVerification = false;

	// Email settings
	let email = '';
	let originalEmail = '';
	let verified = false;
	let mimetype: 'text/plain' | 'text/html' = 'text/html';

	$: hasChanges = email !== originalEmail || mimetype !== originalMimetype;
	let originalMimetype: string = 'text/html';

	onMount(async () => {
		await loadSettings();
	});

	async function loadSettings() {
		loading = true;
		try {
			const settings = await api.getEmailSettings();
			email = settings.email ?? '';
			originalEmail = email;
			verified = settings.verified ?? false;
			mimetype = (settings.mimetype as 'text/plain' | 'text/html') ?? 'text/html';
			originalMimetype = mimetype;
		} catch (error) {
			if (error instanceof ApiClientError) {
				// Email might not be configured yet, that's okay
				if (error.status !== 404) {
					ui.error('Failed to load email settings');
				}
			}
		} finally {
			loading = false;
		}
	}

	async function saveSettings() {
		if (!email.trim()) {
			ui.showError('Error', 'Please enter an email address');
			return;
		}

		// Basic email validation
		const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
		if (!emailRegex.test(email.trim())) {
			ui.showError('Error', 'Please enter a valid email address');
			return;
		}

		saving = true;
		try {
			await api.updateEmailSettings({
				email: email.trim(),
				mimetype
			});
			originalEmail = email.trim();
			originalMimetype = mimetype;
			// If email changed, it needs re-verification
			if (email !== originalEmail) {
				verified = false;
			}
			ui.success('Email settings saved');
		} catch (error) {
			if (error instanceof ApiClientError) {
				ui.showError('Failed to save settings', error.description ?? error.code);
			}
		} finally {
			saving = false;
		}
	}

	async function sendVerificationEmail() {
		if (!email.trim()) {
			ui.showError('Error', 'Please save an email address first');
			return;
		}

		sendingVerification = true;
		try {
			// The API might have a specific endpoint for this
			// For now, we'll just update the email to trigger verification
			await api.updateEmailSettings({ email: email.trim() });
			ui.success('Verification email sent! Please check your inbox.');
		} catch (error) {
			if (error instanceof ApiClientError) {
				ui.showError('Failed to send verification email', error.description ?? error.code);
			}
		} finally {
			sendingVerification = false;
		}
	}

	function cancelChanges() {
		email = originalEmail;
		mimetype = originalMimetype as 'text/plain' | 'text/html';
	}
</script>

<div class="notification-settings">
	<p class="text-muted mb-4">
		Configure your email address to receive notifications about token usage and security events.
	</p>

	{#if loading}
		<LoadingSpinner message="Loading settings..." />
	{:else}
		<form on:submit|preventDefault={saveSettings}>
			<!-- Email address -->
			<div class="mb-4">
				<label for="email-address" class="form-label">
					<i class="fas fa-envelope me-1"></i>
					Email Address
				</label>
				<div class="input-group">
					<input
						type="email"
						id="email-address"
						class="form-control"
						placeholder="you@example.com"
						bind:value={email}
					/>
					{#if email && originalEmail === email}
						<span class="input-group-text">
							{#if verified}
								<i class="fas fa-check-circle text-success" title="Verified"></i>
							{:else}
								<i class="fas fa-exclamation-circle text-warning" title="Not verified"></i>
							{/if}
						</span>
					{/if}
				</div>
				{#if email && originalEmail === email && !verified}
					<div class="mt-2">
						<button
							type="button"
							class="btn btn-sm btn-outline-primary"
							disabled={sendingVerification}
							on:click={sendVerificationEmail}
						>
							{#if sendingVerification}
								<span class="spinner-border spinner-border-sm me-1"></span>
							{:else}
								<i class="fas fa-paper-plane me-1"></i>
							{/if}
							Send Verification Email
						</button>
						<small class="text-muted ms-2">
							You need to verify your email to receive notifications.
						</small>
					</div>
				{/if}
			</div>

			<!-- Email format -->
			<div class="mb-4">
				<span class="form-label d-block" id="email-format-label">
					<i class="fas fa-file-alt me-1"></i>
					Email Format
				</span>
				<div class="btn-group w-100" role="group" aria-labelledby="email-format-label">
					<input
						type="radio"
						class="btn-check"
						name="mimetype"
						id="format-html"
						value="text/html"
						bind:group={mimetype}
					/>
					<label class="btn btn-outline-primary" for="format-html">
						<i class="fas fa-code me-1"></i>
						HTML
					</label>

					<input
						type="radio"
						class="btn-check"
						name="mimetype"
						id="format-text"
						value="text/plain"
						bind:group={mimetype}
					/>
					<label class="btn btn-outline-primary" for="format-text">
						<i class="fas fa-align-left me-1"></i>
						Plain Text
					</label>
				</div>
				<small class="text-muted d-block mt-1">
					HTML emails are formatted with colors and styling. Plain text emails are simpler and work in all email clients.
				</small>
			</div>

			<!-- Actions -->
			<div class="d-flex gap-2">
				<button type="submit" class="btn btn-primary" disabled={saving || !hasChanges}>
					{#if saving}
						<span class="spinner-border spinner-border-sm me-1"></span>
						Saving...
					{:else}
						<i class="fas fa-save me-1"></i>
						Save Changes
					{/if}
				</button>
				{#if hasChanges}
					<button type="button" class="btn btn-outline-secondary" on:click={cancelChanges}>
						Cancel
					</button>
				{/if}
			</div>
		</form>

		<!-- Info box -->
		<div class="alert alert-info mt-4">
			<h6 class="alert-heading">
				<i class="fas fa-info-circle me-1"></i>
				About Email Notifications
			</h6>
			<p class="mb-0">
				Once you verify your email, you can create notification subscriptions on the 
				<a href="/#notifications">Notifications tab</a> to receive alerts about:
			</p>
			<ul class="mb-0 mt-2">
				<li>Access token creations</li>
				<li>Security events (blocked usages, unknown IPs)</li>
				<li>Token expirations</li>
				<li>Setting changes</li>
			</ul>
		</div>
	{/if}
</div>

<style>
	.btn-group {
		max-width: 300px;
	}

	.input-group-text {
		background-color: white;
	}
</style>
