<script lang="ts">
	import type { Capability } from '$lib/types';

	export let capabilities: Capability[] = [];

	// Count capabilities by color class
	$: counts = countByColor(capabilities);
	$: mainCaps = getMainCapabilities(capabilities);

	interface ColorCounts {
		green: number;
		yellow: number;
		red: number;
	}

	interface MainCaps {
		AT: boolean;
		createMT: boolean;
		tokeninfo: boolean;
		manageMT: boolean;
		settings: boolean;
	}

	function countByColor(caps: Capability[]): ColorCounts {
		const counts: ColorCounts = { green: 0, yellow: 0, red: 0 };
		
		function countRecursive(capList: Capability[]) {
			for (const cap of capList) {
				if (cap.enabled) {
					const colorClass = cap.colorClass ?? 'text-success';
					if (colorClass.includes('success')) counts.green++;
					else if (colorClass.includes('warning')) counts.yellow++;
					else if (colorClass.includes('danger')) counts.red++;
				}
				if (cap.children) {
					countRecursive(cap.children);
				}
			}
		}
		
		countRecursive(caps);
		return counts;
	}

	function getMainCapabilities(caps: Capability[]): MainCaps {
		const result: MainCaps = {
			AT: false,
			createMT: false,
			tokeninfo: false,
			manageMT: false,
			settings: false
		};

		function searchRecursive(capList: Capability[]) {
			for (const cap of capList) {
				if (cap.enabled) {
					if (cap.name === 'AT') result.AT = true;
					if (cap.name === 'create_mytoken') result.createMT = true;
					if (cap.name === 'tokeninfo' || cap.name.startsWith('tokeninfo:')) result.tokeninfo = true;
					if (cap.name === 'manage_mytokens' || cap.name.startsWith('manage_mytokens:')) result.manageMT = true;
					if (cap.name === 'settings' || cap.name.startsWith('settings:')) result.settings = true;
				}
				if (cap.children) {
					searchRecursive(cap.children);
				}
			}
		}

		searchRecursive(caps);
		return result;
	}
</script>

<div class="capability-summary d-flex justify-content-between align-items-center py-2">
	<!-- Counts by severity -->
	<div class="counts">
		{#if counts.green > 0}
			<span 
				class="badge bg-success rounded-pill me-1" 
				title="This mytoken has {counts.green} normal capability{counts.green !== 1 ? 'ies' : ''}."
			>
				{counts.green}
			</span>
		{/if}
		{#if counts.yellow > 0}
			<span 
				class="badge bg-warning text-dark rounded-pill me-1" 
				title="This mytoken has {counts.yellow} powerful capability{counts.yellow !== 1 ? 'ies' : ''}."
			>
				{counts.yellow}
			</span>
		{/if}
		{#if counts.red > 0}
			<span 
				class="badge bg-danger rounded-pill me-1" 
				title="This mytoken has {counts.red} very powerful capability{counts.red !== 1 ? 'ies' : ''}."
			>
				{counts.red}
			</span>
		{/if}
		{#if counts.green === 0 && counts.yellow === 0 && counts.red === 0}
			<span class="text-muted small">No capabilities selected</span>
		{/if}
	</div>

	<!-- Main capability icons -->
	<div class="icons fs-5">
		<i 
			class="fab fa-openid me-2" 
			class:text-success={mainCaps.AT}
			class:text-muted={!mainCaps.AT}
			title={mainCaps.AT 
				? "This mytoken can be used to obtain OIDC Access Tokens." 
				: "This mytoken cannot be used to obtain OIDC Access Tokens."}
		></i>
		<i 
			class="fas fa-key me-2" 
			class:text-success={mainCaps.createMT}
			class:text-muted={!mainCaps.createMT}
			title={mainCaps.createMT 
				? "This mytoken can be used to create sub-mytokens." 
				: "This mytoken cannot be used to create sub-mytokens."}
		></i>
		<i 
			class="fas fa-info-circle me-2" 
			class:text-success={mainCaps.tokeninfo}
			class:text-muted={!mainCaps.tokeninfo}
			title={mainCaps.tokeninfo 
				? "This mytoken can be used to obtain tokeninfo about itself." 
				: "This mytoken cannot be used to obtain tokeninfo about itself."}
		></i>
		<i 
			class="fas fa-wrench me-2" 
			class:text-success={mainCaps.manageMT}
			class:text-muted={!mainCaps.manageMT}
			title={mainCaps.manageMT 
				? "This mytoken can be used to manage other mytokens." 
				: "This mytoken cannot be used to manage other mytokens."}
		></i>
		<i 
			class="fas fa-cog" 
			class:text-success={mainCaps.settings}
			class:text-muted={!mainCaps.settings}
			title={mainCaps.settings 
				? "This mytoken can be used to change settings." 
				: "This mytoken cannot be used to change settings."}
		></i>
	</div>
</div>

<style>
	.capability-summary {
		border-bottom: 1px solid var(--bs-border-color);
		margin-bottom: 0.5rem;
	}

	.text-muted {
		opacity: 0.4;
	}

	.icons i {
		cursor: help;
	}

	.badge {
		font-size: 0.85em;
	}
</style>
