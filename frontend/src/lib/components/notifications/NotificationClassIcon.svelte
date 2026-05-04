<script lang="ts">
	interface Props {
		icon: string;
		extraClass?: string;
	}

	let { icon, extraClass = '' }: Props = $props();

	const isComposite = $derived(icon.includes(' '));
	const iconParts = $derived(icon.split(' '));
</script>

{#if isComposite}
	<span class="nc-icon-stack {extraClass}">
		<i class="fas {iconParts[0]}"></i>
		<i class="fas {iconParts[1]} nc-icon-overlay"></i>
	</span>
{:else}
	<i class="fas {icon} {extraClass}"></i>
{/if}

<style>
	.nc-icon-stack {
		position: relative;
		display: inline-block;
		width: 1.25em;
		text-align: center;
	}
	.nc-icon-overlay {
		position: absolute;
		left: 0;
		right: 0;
		top: 0;
		opacity: 0.8;
	}
</style>
