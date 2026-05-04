<script lang="ts">
	export let size: 'sm' | 'md' | 'lg' = 'md';
	export let color: string = 'primary';
	export let message: string = '';
	export let inline: boolean = false;
</script>

<div class="loading-container" class:inline class:text-center={!inline}>
	<div
		class="spinner-border text-{color}"
		class:spinner-border-sm={size === 'sm'}
		role="status"
	>
		<span class="visually-hidden">Loading...</span>
	</div>
	{#if message}
		<p class="loading-message text-muted" class:ms-2={inline} class:mt-2={!inline}>
			{message}
		</p>
	{/if}
</div>

<style>
	.loading-container {
		padding: 1rem;
	}

	.inline {
		display: inline-flex;
		align-items: center;
		padding: 0;
	}

	.loading-message {
		margin-bottom: 0;
	}

	.spinner-border-sm {
		width: 1rem;
		height: 1rem;
	}
</style>
