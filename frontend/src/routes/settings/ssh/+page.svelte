<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';

	// Redirect to settings page - SSH keys are now under the Grants tab
	onMount(() => {
		goto('/settings#ssh');
	});
</script>

<svelte:head>
	<title>mytoken - Redirecting...</title>
</svelte:head>

<div class="d-flex justify-content-center align-items-center" style="min-height: 50vh;">
	<div class="text-center">
		<div class="spinner-border text-primary mb-3" role="status">
			<span class="visually-hidden">Loading...</span>
		</div>
		<p class="text-muted">Redirecting to Settings...</p>
	</div>
</div>
