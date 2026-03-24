<script lang="ts">
	import { isLoggedIn, authInitialized } from '$lib/stores/auth';
	import { goto } from '$app/navigation';
	import GrantsSettings from '$lib/components/settings/GrantsSettings.svelte';
	import TagsSettings from '$lib/components/settings/TagsSettings.svelte';
	import NotificationSettings from '$lib/components/settings/NotificationSettings.svelte';

	let activeTab = 'grants';

	// Redirect to home if not logged in (only after auth is initialized)
	$: if ($authInitialized && !$isLoggedIn) {
		goto('/');
	}

	function setTab(tab: string) {
		activeTab = tab;
	}
</script>

<svelte:head>
	<title>mytoken - Settings</title>
</svelte:head>

<div class="settings-page">
	<h2 class="mb-4">
		<i class="fas fa-cog me-2"></i>
		Settings
	</h2>

	{#if $isLoggedIn}
		<ul class="nav nav-tabs mb-4" role="tablist">
			<li class="nav-item" role="presentation">
				<button
					class="nav-link"
					class:active={activeTab === 'grants'}
					type="button"
					role="tab"
					on:click={() => setTab('grants')}
				>
					<i class="fas fa-unlock-alt me-1"></i>
					Grants
				</button>
			</li>
			<li class="nav-item" role="presentation">
				<button
					class="nav-link"
					class:active={activeTab === 'tags'}
					type="button"
					role="tab"
					on:click={() => setTab('tags')}
				>
					<i class="fas fa-tags me-1"></i>
					Tags
				</button>
			</li>
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
			<li class="nav-item" role="presentation">
				<a class="nav-link" href="/settings/ssh">
					<i class="fas fa-key me-1"></i>
					SSH Keys
				</a>
			</li>
		</ul>

		<div class="tab-content">
			{#if activeTab === 'grants'}
				<GrantsSettings />
			{:else if activeTab === 'tags'}
				<TagsSettings />
			{:else if activeTab === 'notifications'}
				<NotificationSettings />
			{/if}
		</div>
	{:else}
		<div class="alert alert-warning">
			<i class="fas fa-exclamation-triangle me-2"></i>
			Please log in to access settings.
		</div>
	{/if}
</div>

<style>
	.nav-tabs .nav-link {
		color: #6c757d;
		border: none;
		border-bottom: 2px solid transparent;
	}

	.nav-tabs .nav-link:hover {
		color: #df691a;
		border-bottom-color: #dee2e6;
	}

	.nav-tabs .nav-link.active {
		color: #df691a;
		background: transparent;
		border-bottom: 2px solid #df691a;
	}
</style>
