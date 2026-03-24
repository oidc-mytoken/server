<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import type { Capability } from '$lib/types';

	export let capability: Capability;
	export let readonly: boolean = false;
	export let depth: number = 0;
	export let prefix: string = '';

	const dispatch = createEventDispatcher<{ 
		change: void; 
		uncheckParent: void;
		setParentReadOnly: void;
	}>();

	$: hasChildren = capability.children && capability.children.length > 0;
	
	// Check if any child (recursively) is enabled
	function hasEnabledChildren(cap: Capability): boolean {
		if (!cap.children) return false;
		for (const child of cap.children) {
			if (child.enabled) return true;
			if (hasEnabledChildren(child)) return true;
		}
		return false;
	}
	
	// Collapse capabilities with children by default in readonly mode when:
	// 1. The capability itself is enabled (all children are implicitly enabled), OR
	// 2. The capability and none of its children are enabled
	// This makes the tree more compact when viewing token info
	let expandedInitialized = false;
	let expanded = true;
	$: {
		if (!expandedInitialized && hasChildren && readonly) {
			// In readonly mode, collapse if:
			// - capability is enabled (children are implicitly enabled), OR
			// - capability and all children are disabled
			const anyChildEnabled = hasEnabledChildren(capability);
			expanded = !capability.enabled && anyChildEnabled;
			expandedInitialized = true;
		}
	}
	$: inputId = `${prefix}cap-${capability.name.replace(/[^a-zA-Z0-9]/g, '-')}`;
	
	// Get color class for the power level icon
	$: iconColorClass = capability.colorClass || 'text-success';
	
	// Read/write mode: true = read-write (full access), false = read-only
	// Default to read-write mode (not read-only)
	$: isReadWriteMode = !(capability.readOnlyMode ?? false);

	function handleChange(event: Event) {
		const target = event.target as HTMLInputElement;
		const activated = target.checked;
		capability.enabled = activated;

		if (activated) {
			// When checked: also check all children
			if (hasChildren) {
				enableChildren(capability);
			}
		} else {
			// When unchecked: 
			// 1. Uncheck all children
			if (hasChildren) {
				disableChildren(capability);
			}
			// 2. Notify parent to uncheck itself
			dispatch('uncheckParent');
		}
		
		dispatch('change');
	}

	function enableChildren(cap: Capability) {
		if (cap.children) {
			for (const child of cap.children) {
				child.enabled = true;
				enableChildren(child);
			}
		}
	}
	
	function disableChildren(cap: Capability) {
		if (cap.children) {
			for (const child of cap.children) {
				child.enabled = false;
				disableChildren(child);
			}
		}
	}

	function toggleExpand() {
		expanded = !expanded;
	}

	function handleChildChange() {
		dispatch('change');
	}
	
	function handleChildUncheckParent() {
		// A child was unchecked, so we need to uncheck this capability too
		capability.enabled = false;
		// Do NOT disable other children - only this capability gets unchecked
		// Propagate up to our parent
		dispatch('uncheckParent');
		dispatch('change');
	}
	
	function handleChildSetParentReadOnly() {
		// A child was set to read-only, so we need to set this capability to read-only too
		if (capability.hasReadOnlyOption && !capability.readOnlyMode) {
			capability.readOnlyMode = true;
			// Propagate up to our parent
			dispatch('setParentReadOnly');
		}
		dispatch('change');
	}
	
	function toggleReadWriteMode(event: Event) {
		event.preventDefault();
		event.stopPropagation();
		
		// Toggle the mode
		capability.readOnlyMode = !capability.readOnlyMode;
		
		if (capability.readOnlyMode) {
			// Switching to read-only: also switch all children to read-only
			if (hasChildren) {
				setChildrenReadOnly(capability, true);
			}
			// Notify parent to switch to read-only as well
			dispatch('setParentReadOnly');
		} else {
			// Switching to read-write: also switch all children to read-write
			if (hasChildren) {
				setChildrenReadOnly(capability, false);
			}
			// Note: Do NOT propagate read-write up to parent
		}
		
		dispatch('change');
	}
	
	function setChildrenReadOnly(cap: Capability, readOnly: boolean) {
		if (cap.children) {
			for (const child of cap.children) {
				if (child.hasReadOnlyOption) {
					child.readOnlyMode = readOnly;
				}
				setChildrenReadOnly(child, readOnly);
			}
		}
	}
</script>

<li class="list-group-item capability-item" style="padding-left: {depth * 1.5}rem;">
	<div class="d-flex align-items-center justify-content-between">
		<div class="d-flex align-items-center flex-grow-1">
			{#if hasChildren}
				<button
					type="button"
					class="btn btn-sm btn-link p-0 me-2 expand-btn"
					on:click={toggleExpand}
					aria-expanded={expanded}
					aria-label={expanded ? 'Collapse' : 'Expand'}
				>
					<i class="fas" class:fa-chevron-down={expanded} class:fa-chevron-right={!expanded}></i>
				</button>
			{:else}
				<span class="expand-placeholder me-2"></span>
			{/if}

			<div class="form-check mb-0 flex-grow-1">
				<input
					type="checkbox"
					class="form-check-input"
					id={inputId}
					checked={capability.enabled ?? false}
					disabled={readonly}
					on:change={handleChange}
				/>
				<label class="form-check-label" for={inputId}>
					<span class="capability-name">{capability.name}</span>
					{#if capability.hasReadOnlyOption && !readonly}
						<button 
							type="button"
							class="rw-toggle ms-2 btn btn-sm p-0" 
							on:click={toggleReadWriteMode}
							title={isReadWriteMode ? 'Full access. Click to switch to read-only.' : 'Read-only access. Click to switch to full access.'}
						>
							{#if isReadWriteMode}
								<i class="fas fa-pencil-alt"></i>
							{:else}
								<i class="fas fa-book-open"></i>
							{/if}
						</button>
					{:else if capability.hasReadOnlyOption && readonly && capability.enabled}
						<!-- Show read/write indicator in readonly mode -->
						<span 
							class="rw-indicator ms-2" 
							title={isReadWriteMode ? 'Full access (read & write)' : 'Read-only access'}
						>
							{#if isReadWriteMode}
								<i class="fas fa-pencil-alt text-warning"></i>
							{:else}
								<i class="fas fa-book-open text-info"></i>
							{/if}
						</span>
					{/if}
				</label>
			</div>
		</div>
		
		<div class="d-flex align-items-center">
			{#if capability.description}
				<small class="text-muted me-2 capability-desc d-none d-md-inline">{capability.description}</small>
			{/if}
			
			{#if capability.capabilityLevel}
				<i 
					class="fas fa-exclamation-circle {iconColorClass}" 
					title={capability.capabilityLevel}
				></i>
			{/if}
		</div>
	</div>

	{#if hasChildren && expanded}
		<ul class="list-group list-group-flush mt-2 children-list">
			{#each capability.children ?? [] as child}
				<svelte:self
					capability={child}
					{readonly}
					depth={depth + 1}
					{prefix}
					on:change={handleChildChange}
					on:uncheckParent={handleChildUncheckParent}
					on:setParentReadOnly={handleChildSetParentReadOnly}
				/>
			{/each}
		</ul>
	{/if}
</li>

<style>
	.capability-item {
		border: none;
		border-bottom: 1px solid #f0f0f0;
		padding: 0.5rem 1rem;
	}

	.capability-item:last-child {
		border-bottom: none;
	}

	.expand-btn {
		width: 1.5rem;
		text-decoration: none;
		color: #6c757d;
	}

	.expand-btn:hover {
		color: #df691a;
	}

	.expand-placeholder {
		width: 1.5rem;
		display: inline-block;
	}

	.children-list {
		margin-left: 0.5rem;
		border-left: 2px solid #e9ecef;
	}

	.form-check-label {
		cursor: pointer;
	}

	.form-check-input:disabled + .form-check-label {
		cursor: not-allowed;
		opacity: 0.7;
	}
	
	.capability-name {
		font-weight: 500;
	}
	
	.capability-desc {
		font-size: 0.8rem;
		max-width: 300px;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
	
	.rw-toggle {
		cursor: pointer;
		padding: 2px 6px !important;
		border-radius: 4px;
		background-color: #e9ecef;
		display: inline-flex;
		align-items: center;
		justify-content: center;
		border: none;
		line-height: 1;
	}
	
	.rw-toggle:hover {
		background-color: #dee2e6;
	}
	
	.rw-toggle i {
		font-size: 0.75rem;
		color: #333;
	}

	.rw-indicator {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		padding: 2px 6px;
		border-radius: 4px;
		background-color: #f8f9fa;
		line-height: 1;
	}

	.rw-indicator i {
		font-size: 0.75rem;
	}
</style>
