<script lang="ts">
	export let title: string;
	export let collapsed: boolean = false;
	export let icon: string = '';
	export let badgeText: string = '';
	export let badgeColor: string = 'secondary';

	function toggle() {
		collapsed = !collapsed;
	}
</script>

<div class="collapsible-section">
	<button
		type="button"
		class="collapsible-header btn btn-link text-decoration-none w-100 text-start p-0"
		on:click={toggle}
		aria-expanded={!collapsed}
	>
		<span class="d-flex align-items-center">
			<i class="fas fa-chevron-right me-2 collapse-icon" class:rotated={!collapsed}></i>
			{#if icon}
				<i class="fas {icon} me-2"></i>
			{/if}
			<span class="flex-grow-1">{title}</span>
			<slot name="header-right">
				{#if badgeText}
					<span class="badge bg-{badgeColor} ms-2">{badgeText}</span>
				{/if}
			</slot>
		</span>
	</button>

	<div class="collapsible-content" class:collapsed>
		<div class="content-inner pt-3">
			<slot />
		</div>
	</div>
</div>

<style>
	.collapsible-header {
		color: inherit;
		font-weight: 500;
	}

	.collapsible-header:hover {
		color: #df691a;
	}

	.collapse-icon {
		transition: transform 0.2s ease;
		font-size: 0.8em;
	}

	.collapse-icon.rotated {
		transform: rotate(90deg);
	}

	.collapsible-content {
		overflow: hidden;
		max-height: 2000px;
		transition: max-height 0.3s ease, opacity 0.3s ease;
		opacity: 1;
	}

	.collapsible-content.collapsed {
		max-height: 0;
		opacity: 0;
	}

	.content-inner {
		border-left: 2px solid var(--bs-border-color);
		padding-left: 1rem;
		margin-left: 0.5rem;
	}
</style>
