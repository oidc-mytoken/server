<script lang="ts">
	import { page } from '$app/stores';

	const messages: Record<number, string> = {
		404: 'Page not found',
		405: 'Method not allowed',
		429: 'Too many requests - please slow down',
		500: 'Internal server error',
		501: 'Not implemented',
		505: 'HTTP version not supported'
	};

	$: status = $page.status;
	$: message = messages[status] || 'An error occurred';
</script>

<div class="error-page text-center py-5">
	<h1 class="display-1 text-muted">{status}</h1>
	<h2 class="mb-4">{message}</h2>
	{#if $page.error?.message}
		<p class="text-muted mb-4">{$page.error.message}</p>
	{/if}
	<a href="/" class="btn btn-primary">
		<i class="fas fa-home me-2"></i>
		Go Home
	</a>
</div>

<style>
	.error-page {
		min-height: 50vh;
		display: flex;
		flex-direction: column;
		justify-content: center;
		align-items: center;
	}

	.display-1 {
		font-size: 8rem;
		font-weight: 300;
	}
</style>
