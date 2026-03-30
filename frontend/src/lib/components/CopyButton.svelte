<script lang="ts">
	import { ui } from '$lib/stores/ui';

	export let value: string;
	export let label: string = 'Copy';
	export let successMessage: string = 'Copied to clipboard!';
	export let variant: 'button' | 'icon' = 'button';
	export let size: 'sm' | 'md' | 'lg' = 'md';

	let copied = false;
	let timeout: ReturnType<typeof setTimeout>;

	async function handleCopy() {
		try {
			await navigator.clipboard.writeText(value);
			copied = true;
			ui.success(successMessage);

			// Reset after 2 seconds
			clearTimeout(timeout);
			timeout = setTimeout(() => {
				copied = false;
			}, 2000);
		} catch (err) {
			ui.error('Failed to copy to clipboard');
			console.error('Copy failed:', err);
		}
	}
</script>

{#if variant === 'icon'}
	<button
		type="button"
		class="btn btn-link p-0 copy-icon"
		class:text-success={copied}
		title={copied ? 'Copied!' : label}
		on:click={handleCopy}
	>
		<i class="fas" class:fa-check={copied} class:fa-copy={!copied}></i>
	</button>
{:else}
	<button
		type="button"
		class="btn"
		class:btn-sm={size === 'sm'}
		class:btn-lg={size === 'lg'}
		class:btn-success={copied}
		class:btn-outline-secondary={!copied}
		on:click={handleCopy}
	>
		<i class="fas me-1" class:fa-check={copied} class:fa-copy={!copied}></i>
		{copied ? 'Copied!' : label}
	</button>
{/if}

<style>
	.copy-icon {
		font-size: 0.9em;
		text-decoration: none;
	}

	.copy-icon:hover {
		opacity: 0.8;
	}
</style>
