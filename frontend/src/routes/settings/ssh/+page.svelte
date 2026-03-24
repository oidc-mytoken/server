<script lang="ts">
	import { isLoggedIn, authInitialized } from '$lib/stores/auth';
	import { goto } from '$app/navigation';
	import SSHKeysSettings from '$lib/components/settings/SSHKeysSettings.svelte';

	// Redirect to home if not logged in (only after auth is initialized)
	$: if ($authInitialized && !$isLoggedIn) {
		goto('/');
	}
</script>

<svelte:head>
	<title>mytoken - SSH Keys</title>
</svelte:head>

<div class="ssh-page">
	<nav aria-label="breadcrumb" class="mb-3">
		<ol class="breadcrumb">
			<li class="breadcrumb-item"><a href="/settings">Settings</a></li>
			<li class="breadcrumb-item active" aria-current="page">SSH Keys</li>
		</ol>
	</nav>

	<h2 class="mb-4">
		<i class="fas fa-key me-2"></i>
		SSH Key Management
	</h2>

	{#if $isLoggedIn}
		<SSHKeysSettings />
	{:else}
		<div class="alert alert-warning">
			<i class="fas fa-exclamation-triangle me-2"></i>
			Please log in to manage SSH keys.
		</div>
	{/if}
</div>


