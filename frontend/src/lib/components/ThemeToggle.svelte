<script lang="ts">
	import { theme, effectiveTheme, type ThemePreference } from '$lib/stores/theme';

	let dropdownOpen = false;

	const themes: { value: ThemePreference; label: string; icon: string }[] = [
		{ value: 'light', label: 'Light', icon: 'fa-sun' },
		{ value: 'dark', label: 'Dark', icon: 'fa-moon' },
		{ value: 'auto', label: 'Auto', icon: 'fa-desktop' }
	];

	function selectTheme(value: ThemePreference) {
		theme.set(value);
		dropdownOpen = false;
	}

	function getCurrentIcon(preference: ThemePreference): string {
		const t = themes.find((t) => t.value === preference);
		return t?.icon ?? 'fa-sun';
	}

	function handleKeydown(event: KeyboardEvent) {
		if (event.key === 'Escape') {
			dropdownOpen = false;
		}
	}

	function handleClickOutside(event: MouseEvent) {
		const target = event.target as HTMLElement;
		if (!target.closest('.theme-toggle')) {
			dropdownOpen = false;
		}
	}
</script>

<svelte:window on:click={handleClickOutside} on:keydown={handleKeydown} />

<div class="theme-toggle dropdown">
	<button
		class="btn btn-link nav-link px-2"
		type="button"
		aria-expanded={dropdownOpen}
		aria-label="Toggle theme"
		on:click={() => (dropdownOpen = !dropdownOpen)}
	>
		<i class="fas {getCurrentIcon($theme)} fa-fw"></i>
	</button>

	{#if dropdownOpen}
		<ul class="dropdown-menu dropdown-menu-end show">
			{#each themes as t}
				<li>
					<button
						class="dropdown-item d-flex align-items-center"
						class:active={$theme === t.value}
						on:click={() => selectTheme(t.value)}
					>
						<i class="fas {t.icon} fa-fw me-2"></i>
						{t.label}
						{#if $theme === t.value}
							<i class="fas fa-check ms-auto"></i>
						{/if}
					</button>
				</li>
			{/each}
		</ul>
	{/if}
</div>

<style>
	.theme-toggle {
		position: relative;
	}

	.theme-toggle .btn-link {
		color: rgba(255, 255, 255, 0.85);
		text-decoration: none;
	}

	.theme-toggle .btn-link:hover,
	.theme-toggle .btn-link:focus {
		color: #fff;
	}

	.dropdown-menu {
		position: absolute;
		top: 100%;
		right: 0;
		min-width: 8rem;
		padding: 0.25rem 0;
		margin-top: 0.25rem;
		z-index: 1000;
	}

	.dropdown-item {
		cursor: pointer;
	}

	.dropdown-item.active {
		background-color: var(--bs-primary);
		color: #fff;
	}

	.dropdown-item:not(.active):hover {
		background-color: var(--bs-tertiary-bg);
	}
</style>
