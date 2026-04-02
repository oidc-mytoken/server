<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import { providers } from '$lib/stores/discovery';
	import type { Provider } from '$lib/types';

	export let selectedProvider: string = '';
	export let required: boolean = false;
	export let disabled: boolean = false;
	export let placeholder: string = '-- Select a provider --';
	export let id: string = 'provider-selector';

	const dispatch = createEventDispatcher<{ change: string }>();

	let isOpen = false;
	let search = '';
	let searchInput: HTMLInputElement | null = null;
	let dropdownContainer: HTMLDivElement | null = null;
	let highlightedIndex = -1;

	$: filteredProviders = $providers.filter((provider) => {
		if (!search) return true;
		const searchLower = search.toLowerCase();
		const name = (provider.name ?? '').toLowerCase();
		const issuer = provider.issuer.toLowerCase();
		return name.includes(searchLower) || issuer.includes(searchLower);
	});

	$: selectedProviderName = getProviderDisplayName(selectedProvider);

	function getProviderDisplayName(issuer: string): string {
		if (!issuer) return '';
		const provider = $providers.find(p => p.issuer === issuer);
		return provider?.name ?? issuer;
	}

	function toggleDropdown() {
		if (disabled) return;
		isOpen = !isOpen;
		if (isOpen) {
			search = '';
			highlightedIndex = -1;
			setTimeout(() => searchInput?.focus(), 0);
		}
	}

	function closeDropdown() {
		isOpen = false;
		search = '';
		highlightedIndex = -1;
	}

	function selectProvider(provider: Provider) {
		selectedProvider = provider.issuer;
		dispatch('change', provider.issuer);
		closeDropdown();
	}

	function handleKeydown(event: KeyboardEvent) {
		if (!isOpen) {
			if (event.key === 'Enter' || event.key === ' ' || event.key === 'ArrowDown') {
				event.preventDefault();
				toggleDropdown();
			}
			return;
		}

		switch (event.key) {
			case 'Escape':
				event.preventDefault();
				closeDropdown();
				break;
			case 'ArrowDown':
				event.preventDefault();
				highlightedIndex = Math.min(highlightedIndex + 1, filteredProviders.length - 1);
				scrollToHighlighted();
				break;
			case 'ArrowUp':
				event.preventDefault();
				highlightedIndex = Math.max(highlightedIndex - 1, 0);
				scrollToHighlighted();
				break;
			case 'Enter':
				event.preventDefault();
				if (highlightedIndex >= 0 && highlightedIndex < filteredProviders.length) {
					selectProvider(filteredProviders[highlightedIndex]);
				} else if (filteredProviders.length === 1) {
					selectProvider(filteredProviders[0]);
				}
				break;
		}
	}

	function scrollToHighlighted() {
		if (highlightedIndex >= 0) {
			const items = dropdownContainer?.querySelectorAll('.provider-item');
			items?.[highlightedIndex]?.scrollIntoView({ block: 'nearest' });
		}
	}

	function handleClickOutside(event: MouseEvent) {
		const target = event.target as HTMLElement;
		if (dropdownContainer && !dropdownContainer.contains(target)) {
			closeDropdown();
		}
	}

	$: if (isOpen) {
		document.addEventListener('click', handleClickOutside);
	} else {
		document.removeEventListener('click', handleClickOutside);
	}
</script>

<div class="provider-selector" bind:this={dropdownContainer}>
	<button
		type="button"
		class="form-select text-start d-flex align-items-center"
		class:is-invalid={required && !selectedProvider}
		{id}
		{disabled}
		on:click={toggleDropdown}
		on:keydown={handleKeydown}
		aria-haspopup="listbox"
		aria-expanded={isOpen}
	>
		{#if selectedProvider}
			<span class="flex-grow-1 text-truncate">
				{selectedProviderName}
			</span>
			{#if $providers.find(p => p.issuer === selectedProvider)?.oidfed}
				<i class="fas fa-project-diagram ms-2 text-info" title="Discovered via OpenID Federation"></i>
			{/if}
		{:else}
			<span class="text-muted flex-grow-1">{placeholder}</span>
		{/if}
	</button>

	{#if isOpen}
		<div class="dropdown-menu show w-100" role="listbox">
			<div class="px-2 pb-2">
				<input
					type="text"
					class="form-control form-control-sm"
					placeholder="Search providers..."
					bind:value={search}
					bind:this={searchInput}
					on:keydown={handleKeydown}
					on:click|stopPropagation
				/>
			</div>
			<div class="dropdown-divider"></div>
			<div class="provider-list">
				{#each filteredProviders as provider, index}
					<button
						type="button"
						class="dropdown-item provider-item d-flex align-items-center"
						class:active={index === highlightedIndex}
						class:selected={provider.issuer === selectedProvider}
						on:click={() => selectProvider(provider)}
						on:mouseenter={() => highlightedIndex = index}
						role="option"
						aria-selected={provider.issuer === selectedProvider}
					>
						<span class="flex-grow-1 text-truncate">{provider.name ?? provider.issuer}</span>
						{#if provider.oidfed}
							<i class="fas fa-project-diagram ms-2 text-info" title="Discovered via OpenID Federation"></i>
						{/if}
						{#if provider.issuer === selectedProvider}
							<i class="fas fa-check ms-2 text-success"></i>
						{/if}
					</button>
				{:else}
					<div class="dropdown-item text-muted">
						{#if search}
							No providers match "{search}"
						{:else}
							No providers available
						{/if}
					</div>
				{/each}
			</div>
		</div>
	{/if}
</div>

{#if $providers.some(p => p.oidfed)}
	<small class="text-muted d-block mt-1">
		<i class="fas fa-project-diagram text-info"></i> = Discovered via OpenID Federation
	</small>
{/if}

<style>
	.provider-selector {
		position: relative;
	}

	.provider-selector .dropdown-menu {
		position: absolute;
		top: 100%;
		left: 0;
		z-index: 1000;
		max-height: 350px;
		overflow: hidden;
		display: flex;
		flex-direction: column;
	}

	.provider-list {
		max-height: 250px;
		overflow-y: auto;
	}

	.provider-item {
		cursor: pointer;
	}

	.provider-item.active {
		background-color: var(--bs-primary);
		color: white;
	}

	.provider-item.active .text-info {
		color: white !important;
	}

	.provider-item.selected:not(.active) {
		background-color: rgba(var(--bs-primary-rgb), 0.1);
	}

	.form-select {
		cursor: pointer;
	}

	.form-select:disabled {
		cursor: not-allowed;
	}
</style>
