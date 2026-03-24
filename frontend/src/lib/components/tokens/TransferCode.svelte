<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import { api, ApiClientError } from '$lib/api/client';
	import { ui } from '$lib/stores/ui';
	import { formatDuration } from '$lib/utils/format';
	import CopyButton from '../CopyButton.svelte';

	// Event dispatcher for parent communication
	const dispatch = createEventDispatcher<{
		tokenReceived: { token: string };
	}>();

	// Props
	export let initialToken: string = '';

	// Mode: create or exchange - default to exchange (TC -> MT)
	let mode: 'create' | 'exchange' = 'exchange';

	// Create mode state
	let mytoken = ''; // Optional - if empty, cookie auth is used
	let transferCode = '';
	let expiresIn = 0;
	let createLoading = false;
	let createSuccess = false;

	// Exchange mode state
	let inputTransferCode = '';
	let receivedToken = '';
	let exchangeLoading = false;
	let exchangeSuccess = false;

	// Track the last initialToken we processed to detect actual changes from parent
	let lastProcessedInitialToken = '';

	// React to initialToken changes from parent
	$: if (initialToken && initialToken !== lastProcessedInitialToken) {
		lastProcessedInitialToken = initialToken;
		mytoken = initialToken;
		mode = 'create';
		// Reset state when new token is provided
		createSuccess = false;
		transferCode = '';
	}

	async function createTransferCode() {
		// Token is required
		if (!mytoken.trim()) {
			ui.showError('Error', 'Please enter a mytoken to transfer');
			return;
		}

		createLoading = true;
		createSuccess = false;

		try {
			const response = await api.createTransferCode({
				grant_type: 'mytoken',
				mytoken: mytoken.trim()
			});

			transferCode = response.transfer_code;
			expiresIn = response.expires_in;
			createSuccess = true;
			ui.success('Transfer code created!');
		} catch (error) {
			if (error instanceof ApiClientError) {
				ui.showError('Failed to create transfer code', error.description ?? error.code);
			} else {
				ui.showError('Error', (error as Error).message);
			}
		} finally {
			createLoading = false;
		}
	}

	async function exchangeTransferCode() {
		if (!inputTransferCode.trim()) {
			ui.showError('Error', 'Please enter a transfer code');
			return;
		}

		exchangeLoading = true;
		exchangeSuccess = false;

		try {
			const response = await api.exchangeTransferCode(inputTransferCode.trim());

			receivedToken = response.mytoken;
			exchangeSuccess = true;
			ui.success('Token received successfully!');

			// Dispatch event so parent can navigate to Token Info if desired
			dispatch('tokenReceived', { token: receivedToken });
		} catch (error) {
			if (error instanceof ApiClientError) {
				ui.showError('Failed to exchange transfer code', error.description ?? error.code);
			} else {
				ui.showError('Error', (error as Error).message);
			}
		} finally {
			exchangeLoading = false;
		}
	}

	function resetCreate() {
		createSuccess = false;
		transferCode = '';
	}

	function resetExchange() {
		exchangeSuccess = false;
		receivedToken = '';
		inputTransferCode = '';
	}
</script>

<div class="transfer-code">
	<!-- Mode tabs -->
	<ul class="nav nav-pills mb-4">
		<li class="nav-item">
			<button
				class="nav-link"
				class:active={mode === 'exchange'}
				on:click={() => (mode = 'exchange')}
			>
				<i class="fas fa-download me-1"></i>
				Redeem Transfer Code
			</button>
		</li>
		<li class="nav-item">
			<button
				class="nav-link"
				class:active={mode === 'create'}
				on:click={() => (mode = 'create')}
			>
				<i class="fas fa-upload me-1"></i>
				Create Transfer Code
			</button>
		</li>
	</ul>

	{#if mode === 'create'}
		<!-- Create transfer code -->
		{#if createSuccess}
			<div class="result-section">
				<div class="alert alert-success">
					<h5 class="alert-heading">
						<i class="fas fa-check-circle me-2"></i>
						Transfer Code Created!
					</h5>
					<p class="mb-0">
						Share this code with the recipient. It expires in {formatDuration(expiresIn)}.
					</p>
				</div>

				<div class="card mb-4">
					<div class="card-header d-flex justify-content-between align-items-center">
						<span>
							<i class="fas fa-exchange-alt me-1"></i>
							Transfer Code
						</span>
						<CopyButton value={transferCode} label="Copy Code" />
					</div>
					<div class="card-body text-center">
						<div class="transfer-code-display">
							<code class="display-6">{transferCode}</code>
						</div>
						<p class="text-muted mt-2 mb-0">
							<i class="fas fa-clock me-1"></i>
							Valid for {formatDuration(expiresIn)}
						</p>
					</div>
				</div>

				<button type="button" class="btn btn-primary" on:click={resetCreate}>
					<i class="fas fa-plus me-1"></i>
					Create Another
				</button>
			</div>
		{:else}
			<form on:submit|preventDefault={createTransferCode}>
				<div class="mb-4">
					<label for="mytoken-create" class="form-label">
						<i class="fas fa-ticket-alt me-1"></i>
						Mytoken to Transfer
					</label>
					<textarea
						id="mytoken-create"
						class="form-control font-monospace"
						rows="3"
						placeholder="Paste the mytoken you want to transfer..."
						bind:value={mytoken}
						required
					></textarea>
					<small class="text-muted">
						Enter the mytoken you want to transfer to another device.
					</small>
				</div>

				<div class="alert alert-info">
					<i class="fas fa-info-circle me-2"></i>
					<strong>About Transfer Codes:</strong>
					Transfer codes allow you to securely move a mytoken to another device.
					The code is short-lived and can only be used once.
				</div>

				<div class="d-grid">
					<button type="submit" class="btn btn-primary btn-lg" disabled={createLoading || !mytoken.trim()}>
						{#if createLoading}
							<span class="spinner-border spinner-border-sm me-2"></span>
							Creating...
						{:else}
							<i class="fas fa-upload me-2"></i>
							Create Transfer Code
						{/if}
					</button>
				</div>
			</form>
		{/if}
	{:else}
		<!-- Exchange transfer code -->
		{#if exchangeSuccess}
			<div class="result-section">
				<div class="alert alert-success">
					<h5 class="alert-heading">
						<i class="fas fa-check-circle me-2"></i>
						Token Received!
					</h5>
					<p class="mb-0">
						The mytoken has been transferred to this device and set as your current token.
					</p>
				</div>

				<div class="card mb-4">
					<div class="card-header d-flex justify-content-between align-items-center">
						<span>
							<i class="fas fa-key me-1"></i>
							Received Mytoken
						</span>
						<CopyButton value={receivedToken} label="Copy Token" />
					</div>
					<div class="card-body">
						<div class="token-display">
							<code class="text-break">{receivedToken}</code>
						</div>
					</div>
				</div>

				<div class="d-flex gap-2">
					<a href="/#info" class="btn btn-primary">
						<i class="fas fa-info-circle me-1"></i>
						View Token Info
					</a>
					<button type="button" class="btn btn-outline-secondary" on:click={resetExchange}>
						<i class="fas fa-exchange-alt me-1"></i>
						Redeem Another
					</button>
				</div>
			</div>
		{:else}
			<form on:submit|preventDefault={exchangeTransferCode}>
				<div class="mb-4">
					<label for="transfer-code-input" class="form-label">
						<i class="fas fa-exchange-alt me-1"></i>
						Transfer Code
					</label>
					<input
						type="text"
						id="transfer-code-input"
						class="form-control form-control-lg text-center font-monospace"
						placeholder="Enter transfer code"
						bind:value={inputTransferCode}
						required
						autocomplete="off"
					/>
					<small class="text-muted">
						Enter the transfer code you received from another device.
					</small>
				</div>

				<div class="d-grid">
					<button type="submit" class="btn btn-primary btn-lg" disabled={exchangeLoading}>
						{#if exchangeLoading}
							<span class="spinner-border spinner-border-sm me-2"></span>
							Redeeming...
						{:else}
							<i class="fas fa-download me-2"></i>
							Redeem Transfer Code
						{/if}
					</button>
				</div>
			</form>
		{/if}
	{/if}
</div>

<style>
	.nav-pills .nav-link {
		color: #6c757d;
	}

	.nav-pills .nav-link.active {
		background-color: #df691a;
	}

	.transfer-code-display {
		padding: 1.5rem;
		background-color: #f8f9fa;
		border-radius: 0.5rem;
	}

	.transfer-code-display code {
		letter-spacing: 0.25em;
		font-weight: 600;
	}

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
</style>
