<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import type { Capability, CapabilityTemplate, WebCapability } from '$lib/types';
	import CapabilityNode from './CapabilityNode.svelte';
	import CollapsibleSection from '../CollapsibleSection.svelte';

	export let capabilities: Capability[] | WebCapability[] = [];
	export let templates: CapabilityTemplate[] = [];
	export let readonly: boolean = false;
	export let collapsed: boolean = false;
	export let showTemplates: boolean = true;
	export let prefix: string = '';
	
	// For consent page: pass in selected capabilities as strings
	export let selectedCapabilities: string[] = [];

	const dispatch = createEventDispatcher<{ change: string[] }>();

	let selectedTemplate: string = '';

	// Convert WebCapability to Capability format if needed
	$: normalizedCapabilities = normalizeCapabilities(capabilities);
	
	// Track if we've done initial sync to avoid re-syncing on every change
	let initialSyncDone = false;
	
	// Store pending capabilities to apply when normalizedCapabilities becomes available
	let pendingCapabilities: string[] | null = null;
	
	// Initialize enabled state from selectedCapabilities (only once on mount/initial load)
	$: {
		if (!initialSyncDone && selectedCapabilities.length > 0 && normalizedCapabilities.length > 0) {
			syncEnabledState(normalizedCapabilities, selectedCapabilities);
			initialSyncDone = true;
		}
	}
	
	// Apply pending capabilities when normalizedCapabilities becomes available
	$: {
		if (pendingCapabilities !== null && normalizedCapabilities.length > 0) {
			const caps = pendingCapabilities;
			pendingCapabilities = null;
			applyPendingCapabilities(caps);
		}
	}
	
	function applyPendingCapabilities(capNames: string[]) {
		resetCapabilities(normalizedCapabilities);
		for (const capName of capNames) {
			const isReadOnly = capName.startsWith(READ_PREFIX);
			const actualName = isReadOnly ? capName.substring(READ_PREFIX.length) : capName;
			enableCapability(normalizedCapabilities, actualName, isReadOnly);
		}
		normalizedCapabilities = [...normalizedCapabilities];
		emitChange();
	}

	$: enabledCount = countEnabled(normalizedCapabilities);
	$: mainCaps = getMainCapabilities(normalizedCapabilities);
	$: colorCounts = countByColor(normalizedCapabilities);

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

	function isWebCapability(cap: Capability | WebCapability): cap is WebCapability {
		return 'read_write_capability' in cap;
	}

	function normalizeCapabilities(caps: Capability[] | WebCapability[]): Capability[] {
		if (!caps || caps.length === 0) return [];
		
		const first = caps[0];
		if (!isWebCapability(first)) {
			return caps as Capability[];
		}
		
		// Convert WebCapability[] to Capability[]
		return (caps as WebCapability[]).map(webCap => convertWebCapability(webCap));
	}

	function convertWebCapability(webCap: WebCapability | any): Capability {
		// read_write_capability can be either a string (capability name) or an object with name/description
		const rwCap = webCap.read_write_capability;
		if (!rwCap) {
			console.error('No read_write_capability found in:', webCap);
			return { name: 'unknown', enabled: false };
		}
		
		// Handle both formats:
		// 1. String format: "AT", "tokeninfo", etc.
		// 2. Object format: { name: "AT", description: "...", color_class: "...", capability_level: "..." }
		let name: string;
		let description: string = '';
		let colorClass: string = '';
		let capabilityLevel: string = '';
		
		if (typeof rwCap === 'string') {
			name = rwCap;
		} else {
			name = rwCap.name || 'unknown';
			description = rwCap.description || '';
			colorClass = rwCap.color_class || '';
			capabilityLevel = rwCap.capability_level || '';
		}
		
		// hasReadOnlyOption is true when there IS a read_only_capability (meaning we can toggle between modes)
		const hasReadOnlyOption = !!webCap.read_only_capability;
		
		const cap: Capability = {
			name,
			description,
			enabled: selectedCapabilities.includes(name),
			hasReadOnlyOption,
			colorClass,
			capabilityLevel
		};
		
		const children = webCap.children;
		if (children && children.length > 0) {
			cap.children = children.map((child: any) => convertWebCapability(child));
		}
		
		return cap;
	}

	const READ_PREFIX = 'read@';

	function syncEnabledState(caps: Capability[], selected: string[], parentEnabled: boolean = false, parentReadOnly: boolean = false) {
		for (const cap of caps) {
			// Check if this capability is in the selected list (with or without read@ prefix)
			const directlySelected = selected.includes(cap.name);
			const directlySelectedReadOnly = selected.includes(READ_PREFIX + cap.name);
			const isSelected = directlySelected || directlySelectedReadOnly;
			
			// A capability is enabled if:
			// 1. It's directly in the selected list (with or without read@ prefix), OR
			// 2. Its parent is enabled (inherits from parent)
			cap.enabled = isSelected || parentEnabled;
			
			// Set read-only mode if:
			// 1. It was selected with read@ prefix, OR
			// 2. Parent is in read-only mode (inherits from parent)
			if (cap.hasReadOnlyOption) {
				cap.readOnlyMode = directlySelectedReadOnly || (parentEnabled && parentReadOnly);
			}
			
			if (cap.children) {
				const childReadOnly = directlySelectedReadOnly || (parentEnabled && parentReadOnly);
				syncEnabledState(cap.children, selected, cap.enabled, childReadOnly);
			}
		}
	}

	function countEnabled(caps: Capability[]): number {
		let count = 0;
		for (const cap of caps) {
			if (cap.enabled) count++;
			if (cap.children) {
				count += countEnabled(cap.children);
			}
		}
		return count;
	}

	function applyTemplate(template: CapabilityTemplate) {
		// Reset all capabilities
		resetCapabilities(normalizedCapabilities);

		// Enable capabilities from template
		for (const capName of template.capabilities) {
			enableCapability(normalizedCapabilities, capName);
		}

		// Force reactivity and emit change
		normalizedCapabilities = [...normalizedCapabilities];
		emitChange();
	}

	function resetCapabilities(caps: Capability[]) {
		for (const cap of caps) {
			cap.enabled = false;
			if (cap.children) {
				resetCapabilities(cap.children);
			}
		}
	}

	function enableCapability(caps: Capability[], name: string, readOnlyMode: boolean = false): boolean {
		for (const cap of caps) {
			if (cap.name === name) {
				cap.enabled = true;
				if (cap.hasReadOnlyOption) {
					cap.readOnlyMode = readOnlyMode;
				}
				// Also enable all children (matching old interface behavior)
				enableAllChildren(cap, readOnlyMode);
				return true;
			}
			if (cap.children && enableCapability(cap.children, name, readOnlyMode)) {
				return true;
			}
		}
		return false;
	}
	
	function enableAllChildren(cap: Capability, readOnlyMode: boolean = false) {
		if (cap.children) {
			for (const child of cap.children) {
				child.enabled = true;
				if (child.hasReadOnlyOption) {
					child.readOnlyMode = readOnlyMode;
				}
				enableAllChildren(child, readOnlyMode);
			}
		}
	}

	function selectAll() {
		setAllCapabilities(normalizedCapabilities, true);
		normalizedCapabilities = [...normalizedCapabilities];
		emitChange();
	}

	function selectNone() {
		setAllCapabilities(normalizedCapabilities, false);
		normalizedCapabilities = [...normalizedCapabilities];
		emitChange();
	}

	function setAllCapabilities(caps: Capability[], enabled: boolean) {
		for (const cap of caps) {
			cap.enabled = enabled;
			if (cap.children) {
				setAllCapabilities(cap.children, enabled);
			}
		}
	}

	function handleTemplateChange(event: Event) {
		const target = event.target as HTMLSelectElement;
		const template = templates.find((t: CapabilityTemplate) => t.name === target.value);
		if (template) {
			applyTemplate(template);
		}
	}

	function handleCapabilityChange() {
		normalizedCapabilities = [...normalizedCapabilities];
		emitChange();
	}

	function emitChange() {
		const enabled = getEnabledCapabilities();
		dispatch('change', enabled);
	}

	export function getEnabledCapabilities(): string[] {
		const enabled: string[] = [];
		collectEnabled(normalizedCapabilities, enabled);
		// Filter out redundant capabilities (children covered by parents)
		return filterRedundantCapabilities(enabled);
	}

	/**
	 * Set enabled capabilities by name.
	 * Returns array of capability names that were not found.
	 */
	export function setEnabledCapabilities(capNames: string[]): string[] {
		// If normalizedCapabilities is not yet available, store for later
		if (normalizedCapabilities.length === 0) {
			pendingCapabilities = capNames;
			return [];
		}
		
		// Reset all capabilities first
		resetCapabilities(normalizedCapabilities);
		
		const unknownCapabilities: string[] = [];
		
		// Enable the specified capabilities
		for (const capName of capNames) {
			const isReadOnly = capName.startsWith(READ_PREFIX);
			const actualName = isReadOnly ? capName.substring(READ_PREFIX.length) : capName;
			const found = enableCapability(normalizedCapabilities, actualName, isReadOnly);
			if (!found) {
				unknownCapabilities.push(capName);
			}
		}
		
		// Force reactivity and emit change
		normalizedCapabilities = [...normalizedCapabilities];
		emitChange();
		
		return unknownCapabilities;
	}

	/**
	 * Alias for setEnabledCapabilities for convenience.
	 */
	export function setCapabilities(capNames: string[]): string[] {
		return setEnabledCapabilities(capNames);
	}

	function collectEnabled(caps: Capability[], enabled: string[]) {
		for (const cap of caps) {
			if (cap.enabled) {
				// If capability supports read/write and is in read-only mode, add prefix
				if (cap.hasReadOnlyOption && cap.readOnlyMode) {
					enabled.push(READ_PREFIX + cap.name);
				} else {
					enabled.push(cap.name);
				}
			}
			if (cap.children) {
				collectEnabled(cap.children, enabled);
			}
		}
	}
	
	// Filter out capabilities that are already covered by a parent capability
	function filterRedundantCapabilities(caps: string[]): string[] {
		return caps.filter((cap, index) => {
			// Check if this capability is a child of any other capability in the list
			for (let j = 0; j < caps.length; j++) {
				if (index === j) continue;
				if (isChildCapability(cap, caps[j])) {
					return false; // This cap is covered by another, filter it out
				}
			}
			return true;
		});
	}
	
	// Check if capability 'a' is a child of capability 'b'
	function isChildCapability(a: string, b: string): boolean {
		const aReadOnly = a.startsWith(READ_PREFIX);
		const bReadOnly = b.startsWith(READ_PREFIX);
		
		const aName = aReadOnly ? a.substring(READ_PREFIX.length) : a;
		const bName = bReadOnly ? b.substring(READ_PREFIX.length) : b;
		
		// A read-only parent cannot cover a read-write child
		// (but a read-write parent CAN cover a read-only child)
		if (bReadOnly && !aReadOnly) {
			return false;
		}
		
		const aParts = aName.split(':');
		const bParts = bName.split(':');
		
		// b must be shorter or equal to be a parent of a
		if (bParts.length > aParts.length) {
			return false;
		}
		
		// Check if b is a prefix of a
		for (let i = 0; i < bParts.length; i++) {
			if (aParts[i] !== bParts[i]) {
				return false;
			}
		}
		
		// If they're exactly the same, a is not a child of b (they're the same)
		if (aParts.length === bParts.length) {
			return false;
		}
		
		return true;
	}
</script>

<CollapsibleSection
	title="Capabilities"
	icon="fa-check-circle"
	{collapsed}
>
	<span slot="header-right" class="capability-header-info ms-2 d-flex align-items-center">
		<span class="capability-counts me-3">
			{#if colorCounts.green > 0}
				<span 
					class="badge bg-success rounded-pill me-1" 
					title="This mytoken has {colorCounts.green} normal capability{colorCounts.green !== 1 ? 'ies' : ''}."
				>
					{colorCounts.green}
				</span>
			{/if}
			{#if colorCounts.yellow > 0}
				<span 
					class="badge bg-warning text-dark rounded-pill me-1" 
					title="This mytoken has {colorCounts.yellow} powerful capability{colorCounts.yellow !== 1 ? 'ies' : ''}."
				>
					{colorCounts.yellow}
				</span>
			{/if}
			{#if colorCounts.red > 0}
				<span 
					class="badge bg-danger rounded-pill me-1" 
					title="This mytoken has {colorCounts.red} very powerful capability{colorCounts.red !== 1 ? 'ies' : ''}."
				>
					{colorCounts.red}
				</span>
			{/if}
		</span>
		<span class="capability-icons">
			<i 
				class="fab fa-openid me-1" 
				class:text-success={mainCaps.AT}
				class:text-muted={!mainCaps.AT}
				title={mainCaps.AT 
					? "This mytoken can be used to obtain OIDC Access Tokens." 
					: "This mytoken cannot be used to obtain OIDC Access Tokens."}
			></i>
			<i 
				class="fas fa-key me-1" 
				class:text-success={mainCaps.createMT}
				class:text-muted={!mainCaps.createMT}
				title={mainCaps.createMT 
					? "This mytoken can be used to create sub-mytokens." 
					: "This mytoken cannot be used to create sub-mytokens."}
			></i>
			<i 
				class="fas fa-info-circle me-1" 
				class:text-success={mainCaps.tokeninfo}
				class:text-muted={!mainCaps.tokeninfo}
				title={mainCaps.tokeninfo 
					? "This mytoken can be used to obtain tokeninfo about itself." 
					: "This mytoken cannot be used to obtain tokeninfo about itself."}
			></i>
			<i 
				class="fas fa-wrench me-1" 
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
		</span>
	</span>

	{#if showTemplates && templates.length > 0 && !readonly}
		<div class="mb-3">
			<label for="{prefix}cap-template" class="form-label">
				<i class="fas fa-layer-group me-1"></i>
				Apply Template
			</label>
			<select
				id="{prefix}cap-template"
				class="form-select"
				bind:value={selectedTemplate}
				on:change={handleTemplateChange}
			>
				<option value="">-- Select a template --</option>
				{#each templates as template}
					<option value={template.name}>{template.name}</option>
				{/each}
			</select>
		</div>
	{/if}

	{#if !readonly}
		<div class="mb-3">
			<div class="btn-group btn-group-sm" role="group">
				<button type="button" class="btn btn-outline-primary" on:click={selectAll}>
					<i class="fas fa-check-double me-1"></i>
					Select All
				</button>
				<button type="button" class="btn btn-outline-secondary" on:click={selectNone}>
					<i class="fas fa-times me-1"></i>
					Select None
				</button>
			</div>
		</div>
	{/if}

	<div class="capabilities-container">
		<ul class="list-group list-group-flush">
		{#each normalizedCapabilities as capability}
			<CapabilityNode {capability} {readonly} {prefix} depth={0} on:change={handleCapabilityChange} />
		{/each}
		</ul>
	</div>

	{#if enabledCount === 0 && !readonly}
		<div class="alert alert-warning mt-3 mb-0">
			<i class="fas fa-exclamation-triangle me-2"></i>
			No capabilities selected. The token will have minimal permissions.
		</div>
	{/if}
</CollapsibleSection>

<style>
	.capabilities-container {
		max-height: 400px;
		overflow-y: auto;
		border: 1px solid var(--bs-border-color);
		border-radius: 0.375rem;
	}

	.form-select {
		max-width: 300px;
	}

	.capability-icons .text-muted {
		opacity: 0.3;
	}

	.capability-icons i {
		font-size: 1rem;
	}

	.capability-counts .badge {
		font-size: 0.75em;
	}
</style>
