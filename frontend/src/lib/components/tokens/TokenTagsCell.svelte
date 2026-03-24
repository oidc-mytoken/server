<script lang="ts">
	import type { MTTagInfo, Tag } from '$lib/types';
	import { api, ApiClientError } from '$lib/api/client';
	import { ui } from '$lib/stores/ui';
	import TagPill from '../TagPill.svelte';

	interface Props {
		momId: string;
		tokenTags: MTTagInfo[];
		availableTags: Tag[];
		onTagsChanged?: () => void;
	}

	let { momId, tokenTags, availableTags, onTagsChanged }: Props = $props();

	let showDropdown = $state(false);
	let loading = $state(false);
	let dropdownRef: HTMLDivElement | null = $state(null);

	// Tags that are not already on this token
	let addableTags = $derived(
		availableTags.filter(
			(t) => !tokenTags.some((tt) => tt.tag === t.tag)
		)
	);

	// Close dropdown when clicking outside
	function handleClickOutside(event: MouseEvent) {
		if (dropdownRef && !dropdownRef.contains(event.target as Node)) {
			showDropdown = false;
		}
	}

	$effect(() => {
		if (showDropdown) {
			document.addEventListener('click', handleClickOutside);
			return () => document.removeEventListener('click', handleClickOutside);
		}
	});

	async function removeTag(tagName: string) {
		loading = true;
		try {
			await api.removeTagFromMytoken(momId, tagName);
			ui.success(`Tag "${tagName}" removed`);
			onTagsChanged?.();
		} catch (err) {
			if (err instanceof ApiClientError) {
				ui.showError('Failed to remove tag', err.description ?? err.code);
			}
		} finally {
			loading = false;
		}
	}

	async function addTag(tagName: string) {
		loading = true;
		showDropdown = false;
		try {
			await api.addTagToMytoken(momId, tagName);
			ui.success(`Tag "${tagName}" added`);
			onTagsChanged?.();
		} catch (err) {
			if (err instanceof ApiClientError) {
				ui.showError('Failed to add tag', err.description ?? err.code);
			}
		} finally {
			loading = false;
		}
	}

	function toggleDropdown(event: MouseEvent) {
		event.stopPropagation();
		showDropdown = !showDropdown;
	}
</script>

<div class="token-tags-cell" bind:this={dropdownRef}>
	{#if tokenTags && tokenTags.length > 0}
		{#each tokenTags as tag}
			<TagPill 
				name={tag.tag} 
				color={tag.color} 
				small 
				removable 
				includeChildren={tag.include_children ?? false}
				onRemove={() => removeTag(tag.tag)}
			/>
		{/each}
	{/if}

	<!-- Add tag button (styled like a badge) -->
	{#if addableTags.length > 0}
		<div class="add-tag-wrapper">
			<button
				type="button"
				class="badge add-tag-badge"
				onclick={toggleDropdown}
				disabled={loading}
				title="Add tag"
			>
				{#if loading}
					<i class="fas fa-spinner fa-spin"></i>
				{:else}
					<i class="fas fa-plus"></i>
				{/if}
			</button>

			{#if showDropdown}
				<div class="tag-dropdown">
					{#each addableTags as tag}
						<button
							type="button"
							class="tag-dropdown-item"
							onclick={() => addTag(tag.tag)}
						>
							<TagPill name={tag.tag} color={tag.color} small />
						</button>
					{/each}
				</div>
			{/if}
		</div>
	{/if}

	{#if tokenTags.length === 0 && addableTags.length === 0}
		<span class="text-muted no-tags">-</span>
	{/if}
</div>

<style>
	.token-tags-cell {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: 0.25rem;
		position: relative;
	}

	.add-tag-wrapper {
		position: relative;
		display: inline-flex;
		align-items: center;
	}

	.add-tag-badge {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		font-weight: 500;
		font-size: 0.75em;
		padding: 0.2em 0.5em;
		margin-right: 0.35rem;
		margin-bottom: 0.35rem;
		background-color: transparent;
		border: 1px dashed var(--bs-secondary, #6c757d);
		color: var(--bs-secondary, #6c757d);
		cursor: pointer;
		transition: all 0.15s ease-in-out;
	}

	.add-tag-badge:hover:not(:disabled) {
		background-color: var(--bs-secondary, #6c757d);
		color: white;
		border-style: solid;
	}

	.add-tag-badge:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	.tag-dropdown {
		position: absolute;
		top: 100%;
		left: 0;
		z-index: 1000;
		min-width: 120px;
		max-height: 200px;
		overflow-y: auto;
		background: var(--bs-body-bg, #fff);
		border: 1px solid var(--bs-border-color, #dee2e6);
		border-radius: 0.375rem;
		box-shadow: 0 0.5rem 1rem rgba(0, 0, 0, 0.15);
		margin-top: 0.25rem;
	}

	.tag-dropdown-item {
		display: block;
		width: 100%;
		padding: 0.5rem 0.75rem;
		text-align: left;
		background: none;
		border: none;
		cursor: pointer;
	}

	.tag-dropdown-item:hover {
		background: var(--bs-tertiary-bg, #f8f9fa);
	}

	.no-tags {
		font-size: 0.875rem;
	}
</style>
