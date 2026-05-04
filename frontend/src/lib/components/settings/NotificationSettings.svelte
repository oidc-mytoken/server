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
	let preferHtml = true;
	let originalPreferHtml = true;

	$: hasChanges = email !== originalEmail || preferHtml !== originalPreferHtml;

	onMount(async () => {
		await loadSettings();
	});

	async function loadSettings() {
		loading = true;
		try {
			const settings = await api.getEmailSettings();
			email = settings.email_address ?? '';
			originalEmail = email;
			verified = settings.email_verified ?? false;
			preferHtml = settings.prefer_html_mail ?? true;
			originalPreferHtml = preferHtml;
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
			const emailChanged = email.trim() !== originalEmail;
			const mimeChanged = preferHtml !== originalPreferHtml;
			
			// Build request with only changed fields
			const request: { email_address?: string; prefer_html_mail?: boolean } = {};
			if (emailChanged) {
				request.email_address = email.trim();
			}
			if (mimeChanged) {
				request.prefer_html_mail = preferHtml;
			}
			
			await api.updateEmailSettings(request);
			originalEmail = email.trim();
			originalPreferHtml = preferHtml;
			// If email changed, it needs re-verification
			if (emailChanged) {
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
			// Updating the email address triggers a verification email
			await api.updateEmailSettings({ email_address: email.trim() });
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
		preferHtml = originalPreferHtml;
	}
</script>

<div class="notification-settings">
	<p class="text-muted mb-4">
		Configure your email address to receive notifications about token usage and security events.
	</p>

	{#if loading}
		<LoadingSpinner message="Loading settings..." />
	{:else}
		<!-- Email Settings -->
		<div class="card mb-4">
			<div class="card-header">
				<span class="fw-semibold">
					<i class="fas fa-envelope me-2"></i>
					Email Settings
					{#if verified}
						<span class="badge bg-success ms-2">Verified</span>
					{:else if email}
						<span class="badge bg-warning text-dark ms-2">Unverified</span>
					{/if}
				</span>
			</div>
			<div class="card-body">
				<form on:submit|preventDefault={saveSettings}>
					<!-- Email address -->
					<div class="mb-4">
						<label for="email-address" class="form-label">
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
							Email Format
						</span>
						<div class="btn-group" role="group" aria-labelledby="email-format-label">
							<input
								type="radio"
								class="btn-check"
								name="preferHtml"
								id="format-html"
								checked={preferHtml}
								on:change={() => preferHtml = true}
							/>
							<label class="btn btn-outline-primary" for="format-html">
								<i class="fas fa-code me-1"></i>
								HTML
							</label>

							<input
								type="radio"
								class="btn-check"
								name="preferHtml"
								id="format-text"
								checked={!preferHtml}
								on:change={() => preferHtml = false}
							/>
							<label class="btn btn-outline-primary" for="format-text">
								<i class="fas fa-align-left me-1"></i>
								Plain Text
							</label>
						</div>
						<small class="text-muted d-block mt-1">
							HTML emails are formatted with colors and styling. Plain text emails are simpler.
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
			</div>
		</div>

		<!-- Calendars Link -->
		<div class="card mb-4">
			<div class="card-header">
				<span class="fw-semibold">
					<i class="fas fa-calendar-alt me-2"></i>
					Calendars
				</span>
			</div>
			<div class="card-body">
			<p class="text-muted mb-3">
				Create ICS calendars to track token expirations. Subscribe to these calendars in your calendar app.
			</p>
			<a href="/#calendars" class="btn btn-outline-primary">
				<i class="fas fa-external-link-alt me-1"></i>
				Manage Calendars
			</a>
			</div>
		</div>

		<!-- Info box -->
		<div class="alert alert-info">
			<h6 class="alert-heading">
				<i class="fas fa-info-circle me-1"></i>
				About Notifications
			</h6>
			<p class="mb-2">
				Once you verify your email, you can create notification subscriptions on the 
				<a href="/#notifications">Notifications tab</a> to receive alerts about:
			</p>
			<ul class="mb-0">
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
