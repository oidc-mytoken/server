<script lang="ts">
	import { onMount } from 'svelte';
	import { isLoggedIn } from '$lib/stores/auth';
	import { discovery } from '$lib/stores/discovery';
	import type { InitialMytokenRequest } from '$lib/types';
	import AboutTab from '$lib/components/AboutTab.svelte';
	import CreateMytoken from '$lib/components/tokens/CreateMytoken.svelte';
	import CreateAccessToken from '$lib/components/tokens/CreateAccessToken.svelte';
	import TokenInfo from '$lib/components/tokens/TokenInfo.svelte';
	import TokenList from '$lib/components/tokens/TokenList.svelte';
	import TransferCode from '$lib/components/tokens/TransferCode.svelte';
	import Notifications from '$lib/components/notifications/Notifications.svelte';

	// Tab state - default to 'about' for new visitors
	let activeTab = 'about';

	// Token to pass to TokenInfo when created
	let tokenForInfo = '';

	// Token to pass to TransferCode when requested from TokenInfo
	let tokenForTransfer = '';

	// Notifications subtab (notifications or calendars)
	let notificationsSubtab: 'notifications' | 'calendars' = 'notifications';

	// Initial mytoken request from URL parameter (?r=<base64>)
	let initialMytokenRequest: InitialMytokenRequest | null = null;
	let requestParamError: string | null = null;

	// Check if notifications are enabled
	$: notificationsEnabled = !!$discovery.data?.notifications_endpoint;

	// Tabs available to all users
	const publicTabs = ['about', 'mt', 'at', 'info', 'transfer'];
	
	// Tabs only available when logged in
	const authTabs = ['list', 'notifications'];

	// Handle URL hash for tab navigation and ?r= parameter for Create Mytoken
	onMount(() => {
		// Parse ?r= parameter for Create Mytoken pre-population
		const urlParams = new URLSearchParams(window.location.search);
		const requestParam = urlParams.get('r');
		
		if (requestParam) {
			try {
				const decoded = atob(requestParam);
				const parsed = JSON.parse(decoded) as InitialMytokenRequest;
				
				// Pass to CreateMytoken - validation happens there after providers are loaded
				initialMytokenRequest = parsed;
				activeTab = 'mt';
			} catch (e) {
				console.error('Failed to parse request parameter:', e);
				requestParamError = 'Invalid request parameter format.';
			}
			
			// Clean URL (remove ?r= parameter, keep hash)
			const hash = window.location.hash;
			window.history.replaceState(null, '', window.location.pathname + (hash || '#mt'));
		}
		
		// Handle URL hash for tab navigation
		const hash = window.location.hash.slice(1);
		if (hash) {
			// Handle #calendars as a shortcut to notifications tab with calendars subtab
			if (hash === 'calendars') {
				if ($isLoggedIn) {
					activeTab = 'notifications';
					notificationsSubtab = 'calendars';
				} else {
					activeTab = 'about';
					window.history.replaceState(null, '', '#about');
				}
			// Check if it's a valid tab
			} else if (publicTabs.includes(hash)) {
				activeTab = hash;
			} else if (authTabs.includes(hash)) {
				// Auth-required tabs - redirect to about if not logged in
				if ($isLoggedIn) {
					activeTab = hash;
				} else {
					activeTab = 'about';
					window.history.replaceState(null, '', '#about');
				}
			}
		}
	});

	function setTab(tab: string) {
		// Prevent navigation to auth-required tabs if not logged in
		if (authTabs.includes(tab) && !$isLoggedIn) {
			return;
		}
		activeTab = tab;
		// Reset notifications subtab when navigating to notifications tab normally
		if (tab === 'notifications') {
			notificationsSubtab = 'notifications';
		}
		window.history.replaceState(null, '', `#${tab}`);
	}

	function handleNavigate(event: CustomEvent<{ tab: string }>) {
		setTab(event.detail.tab);
	}

	function handleTokenCreated(event: CustomEvent<{ token: string; tokenType: string }>) {
		tokenForInfo = event.detail.token;
		setTab('info');
		// Clear tokenForInfo after a brief delay to allow TokenInfo to process it
		setTimeout(() => {
			tokenForInfo = '';
		}, 100);
	}

	function handleCreateTransferCode(event: CustomEvent<{ token: string }>) {
		tokenForTransfer = event.detail.token;
		setTab('transfer');
		// Clear tokenForTransfer after a brief delay to allow TransferCode to process it
		setTimeout(() => {
			tokenForTransfer = '';
		}, 100);
	}

	function handleTokenReceived(event: CustomEvent<{ token: string }>) {
		tokenForInfo = event.detail.token;
		setTab('info');
		setTimeout(() => {
			tokenForInfo = '';
		}, 100);
	}
</script>

<svelte:head>
	<title>mytoken - Home</title>
</svelte:head>

<div class="home-page">
	<!-- Unified tabbed interface for all users -->
	<ul class="nav nav-tabs mb-4" role="tablist">
		<!-- About tab - always first -->
		<li class="nav-item" role="presentation">
			<button
				class="nav-link"
				class:active={activeTab === 'about'}
				type="button"
				role="tab"
				on:click={() => setTab('about')}
			>
				<i class="fas fa-home me-1"></i>
				About
			</button>
		</li>

		<!-- Create Mytoken - available to all -->
		<li class="nav-item" role="presentation">
			<button
				class="nav-link"
				class:active={activeTab === 'mt'}
				type="button"
				role="tab"
				on:click={() => setTab('mt')}
			>
				<i class="fas fa-plus-circle me-1"></i>
				Create Mytoken
			</button>
		</li>

		<!-- Get Access Token - available to all (requires mytoken when logged out) -->
		<li class="nav-item" role="presentation">
			<button
				class="nav-link"
				class:active={activeTab === 'at'}
				type="button"
				role="tab"
				on:click={() => setTab('at')}
			>
				<i class="fas fa-key me-1"></i>
				Get Access Token
			</button>
		</li>

		<!-- Token Info - available to all -->
		<li class="nav-item" role="presentation">
			<button
				class="nav-link"
				class:active={activeTab === 'info'}
				type="button"
				role="tab"
				on:click={() => setTab('info')}
			>
				<i class="fas fa-info-circle me-1"></i>
				Token Info
			</button>
		</li>

		<!-- Transfer Code - available to all -->
		<li class="nav-item" role="presentation">
			<button
				class="nav-link"
				class:active={activeTab === 'transfer'}
				type="button"
				role="tab"
				on:click={() => setTab('transfer')}
			>
				<i class="fas fa-exchange-alt me-1"></i>
				Transfer Code
			</button>
		</li>

		<!-- My Tokens - logged in only -->
		{#if $isLoggedIn}
			<li class="nav-item" role="presentation">
				<button
					class="nav-link"
					class:active={activeTab === 'list'}
					type="button"
					role="tab"
					on:click={() => setTab('list')}
				>
					<i class="fas fa-list me-1"></i>
					My Tokens
				</button>
			</li>
		{/if}

		<!-- Notifications - logged in only, if enabled -->
		{#if $isLoggedIn && notificationsEnabled}
			<li class="nav-item" role="presentation">
				<button
					class="nav-link"
					class:active={activeTab === 'notifications'}
					type="button"
					role="tab"
					on:click={() => setTab('notifications')}
				>
					<i class="fas fa-bell me-1"></i>
					Notifications
				</button>
			</li>
		{/if}
	</ul>

	<!-- Error alert for invalid request parameter -->
	{#if requestParamError}
		<div class="alert alert-danger alert-dismissible fade show" role="alert">
			<i class="fas fa-exclamation-circle me-2"></i>
			{requestParamError}
			<button type="button" class="btn-close" aria-label="Close" on:click={() => requestParamError = null}></button>
		</div>
	{/if}

	<!-- Tab content -->
	<div class="tab-content">
		{#if activeTab === 'about'}
			<div class="tab-pane active" role="tabpanel">
				<AboutTab on:navigate={handleNavigate} />
			</div>
		{:else if activeTab === 'mt'}
			<div class="tab-pane active" role="tabpanel">
				<CreateMytoken 
					initialRequest={initialMytokenRequest}
					on:created={handleTokenCreated} 
				/>
			</div>
		{:else if activeTab === 'at'}
			<div class="tab-pane active" role="tabpanel">
				<CreateAccessToken />
			</div>
		{:else if activeTab === 'info'}
			<div class="tab-pane active" role="tabpanel">
				<TokenInfo initialToken={tokenForInfo} on:createTransferCode={handleCreateTransferCode} />
			</div>
		{:else if activeTab === 'transfer'}
			<div class="tab-pane active" role="tabpanel">
				<TransferCode initialToken={tokenForTransfer} on:tokenReceived={handleTokenReceived} />
			</div>
		{:else if activeTab === 'list' && $isLoggedIn}
			<div class="tab-pane active" role="tabpanel">
				<TokenList />
			</div>
		{:else if activeTab === 'notifications' && $isLoggedIn && notificationsEnabled}
			<div class="tab-pane active" role="tabpanel">
				<Notifications initialSubtab={notificationsSubtab} />
			</div>
		{/if}
	</div>
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
</style>
