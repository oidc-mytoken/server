<script lang="ts">
	import { generateTagColor } from '$lib/utils/color';

	export let name: string = '';  // Tag name (new prop)
	export let tag: string = '';   // Legacy prop (alias for name)
	export let color: string = '';  // Can be hex color (with or without #) or Bootstrap class
	export let removable: boolean = false;
	export let onRemove: (() => void) | undefined = undefined;
	export let small: boolean = false;
	export let includeChildren: boolean = false;  // Show the include_children icon
	export let onToggleChildren: (() => void) | undefined = undefined;  // Toggle include_children

	// Support both 'name' and 'tag' props
	$: displayName = name || tag;

	// Generate hash-based color as fallback when no color provided
	$: hashBasedColor = generateTagColor(displayName);

	// Normalize color - use hash-based color if empty/undefined
	$: trimmedColor = color && color.trim() ? color.trim() : '';

	// Check if color looks like a hex color (with or without #)
	// Hex colors: #abc, #aabbcc, abc, aabbcc (3 or 6 hex digits)
	$: isHexColor = trimmedColor ? /^#?[0-9A-Fa-f]{3}([0-9A-Fa-f]{3})?$/.test(trimmedColor) : false;

	// Normalize hex color to always have #
	$: normalizedHexColor = isHexColor ? (trimmedColor.startsWith('#') ? trimmedColor : `#${trimmedColor}`) : '';

	// Use hash-based color if no valid color provided
	$: effectiveColor = isHexColor ? normalizedHexColor : (trimmedColor || hashBasedColor);

	// Always use hex color styling (since we always have a hex color now)
	$: useHexStyle = isHexColor || !trimmedColor;
	$: textColor = useHexStyle ? getContrastColor(effectiveColor) : 'white';
	$: bgStyle = useHexStyle ? `background-color: ${effectiveColor}; color: ${textColor};` : '';
	$: bgClass = useHexStyle ? '' : `bg-${trimmedColor}`;

	function getContrastColor(hexColor: string): string {
		let hex = hexColor.replace('#', '');
		// Expand shorthand (e.g., "abc" -> "aabbcc")
		if (hex.length === 3) {
			hex = hex[0] + hex[0] + hex[1] + hex[1] + hex[2] + hex[2];
		}
		if (hex.length < 6) return '#ffffff';  // Invalid hex, default to white text
		const r = parseInt(hex.substr(0, 2), 16);
		const g = parseInt(hex.substr(2, 2), 16);
		const b = parseInt(hex.substr(4, 2), 16);
		const luminance = (0.299 * r + 0.587 * g + 0.114 * b) / 255;
		return luminance > 0.5 ? '#000000' : '#ffffff';
	}
</script>

{#if onToggleChildren}
	<button 
		type="button"
		class="badge badge-btn {bgClass}" 
		class:small-badge={small} 
		style={bgStyle}
		on:click={onToggleChildren}
		title={includeChildren ? "Click to exclude children" : "Click to include children"}
	>
		{displayName}
		<i 
			class="fas fa-sitemap ms-1" 
			class:text-muted={!includeChildren}
		></i>
	</button>
	{#if removable && onRemove}
		<button
			type="button"
			class="btn-close-inline"
			class:btn-close-white={textColor === '#ffffff'}
			class:btn-close-sm={small}
			aria-label="Remove tag"
			on:click|stopPropagation={onRemove}
		><i class="fas fa-times"></i></button>
	{/if}
{:else}
	<span 
		class="badge {bgClass}" 
		class:small-badge={small} 
		style={bgStyle}
	>
		{displayName}
		{#if includeChildren}
			<i class="fas fa-sitemap ms-1" title="Includes children"></i>
		{/if}
		{#if removable && onRemove}
			<button
				type="button"
				class="btn-close ms-1"
				class:btn-close-white={textColor === '#ffffff'}
				class:btn-close-sm={small}
				aria-label="Remove tag"
				on:click|stopPropagation={onRemove}
			></button>
		{/if}
	</span>
{/if}

<style>
	.badge {
		display: inline-flex;
		align-items: center;
		font-weight: 500;
		font-size: 1rem;
		padding: 0.5em 0.75em;
		margin-right: 0.35rem;
		margin-bottom: 0.35rem;
	}

	.badge-btn {
		cursor: pointer;
		border: none;
	}

	.badge-btn:hover {
		filter: brightness(0.9);
	}

	.small-badge {
		font-size: 0.75em;
		padding: 0.2em 0.5em;
	}

	.btn-close {
		font-size: 0.7em;
		padding: 0.25em;
		margin-left: 0.35em;
	}

	.btn-close-inline {
		background: none;
		border: none;
		color: inherit;
		font-size: 0.7em;
		padding: 0.25em;
		margin-left: -0.5em;
		margin-right: 0.35rem;
		cursor: pointer;
		opacity: 0.7;
	}

	.btn-close-inline:hover {
		opacity: 1;
	}

	.btn-close-sm {
		font-size: 0.5em;
	}

	.text-muted {
		opacity: 0.5;
	}
</style>
