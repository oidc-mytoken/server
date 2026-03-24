<script lang="ts">
	import type { Rotation, RotationTemplate } from '$lib/types';
	import CollapsibleSection from './CollapsibleSection.svelte';
	import { formatDuration } from '$lib/utils/format';

	export let rotation: Rotation = {};
	export let templates: RotationTemplate[] = [];
	export let readonly: boolean = false;
	export let collapsed: boolean = true;
	export let showTemplates: boolean = true;

	let selectedTemplate = '';
	let lifetimeValue = rotation.lifetime ? Math.floor(rotation.lifetime / 3600) : 0;
	let lifetimeUnit: 'hours' | 'days' = 'hours';

	// Sync lifetime back to rotation object
	$: {
		if (lifetimeValue > 0) {
			rotation.lifetime = lifetimeValue * (lifetimeUnit === 'days' ? 86400 : 3600);
		} else {
			delete rotation.lifetime;
		}
	}

	function applyTemplate(template: RotationTemplate) {
		rotation = { ...template.rotation };
		if (rotation.lifetime) {
			// Convert to hours or days for display
			if (rotation.lifetime >= 86400 && rotation.lifetime % 86400 === 0) {
				lifetimeValue = rotation.lifetime / 86400;
				lifetimeUnit = 'days';
			} else {
				lifetimeValue = Math.floor(rotation.lifetime / 3600);
				lifetimeUnit = 'hours';
			}
		}
	}

	function handleTemplateChange(event: Event) {
		const target = event.target as HTMLSelectElement;
		const template = templates.find((t: RotationTemplate) => t.name === target.value);
		if (template) {
			applyTemplate(template);
		}
	}

	$: isEnabled = rotation.on_AT || rotation.on_other;

	// Track previous enabled state to auto-enable auto_revoke when rotation is first enabled
	let wasEnabled: boolean = false;
	$: {
		if (isEnabled && !wasEnabled) {
			// Rotation was just enabled, auto-enable auto_revoke
			rotation.auto_revoke = true;
		}
		wasEnabled = isEnabled ?? false;
	}

	// Icon status calculation
	interface RotationIcon {
		color: 'success' | 'primary' | 'info' | 'muted';
		title: string;
	}

	$: icon = calculateIcon(rotation);

	function calculateIcon(rot: Rotation): RotationIcon {
		const onAT = rot.on_AT ?? false;
		const onOther = rot.on_other ?? false;

		if (onAT && onOther) {
			return { color: 'success', title: 'This token is rotated whenever it is used.' };
		} else if (onAT) {
			return { color: 'primary', title: 'This token is rotated on access token requests.' };
		} else if (onOther) {
			return { color: 'info', title: 'This token is rotated on other requests than access token requests.' };
		} else {
			return { color: 'muted', title: 'This token is never rotated.' };
		}
	}
</script>

<CollapsibleSection
	title="Rotation"
	icon="fa-sync-alt"
	{collapsed}
>
	<span slot="header-right" class="rotation-icon ms-2">
		<i 
			class="fas fa-sync" 
			class:text-success={icon.color === 'success'}
			class:text-primary={icon.color === 'primary'}
			class:text-info={icon.color === 'info'}
			class:text-muted={icon.color === 'muted'}
			title={icon.title}
		></i>
	</span>
	{#if showTemplates && templates.length > 0 && !readonly}
		<div class="mb-3">
			<label for="rotation-template" class="form-label">
				<i class="fas fa-layer-group me-1"></i>
				Apply Template
			</label>
			<select
				id="rotation-template"
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

	<div class="rotation-settings">
		<div class="row g-3">
			<!-- Rotation triggers -->
			<div class="col-md-6">
				<div class="form-check">
					<input
						type="checkbox"
						class="form-check-input"
						id="rotation-on-at"
						bind:checked={rotation.on_AT}
						disabled={readonly}
					/>
					<label class="form-check-label" for="rotation-on-at">
						<i class="fas fa-key me-1" class:text-primary={rotation.on_AT} class:text-muted={!rotation.on_AT}></i>
						Rotate on Access Token creation
					</label>
				</div>
				<small class="ms-4" class:text-muted={!rotation.on_AT} class:text-success={rotation.on_AT}>
					{#if rotation.on_AT}
						Token rotates when used to obtain an access token
					{:else}
						Token does not rotate on access token creation
					{/if}
				</small>
			</div>

			<div class="col-md-6">
				<div class="form-check">
					<input
						type="checkbox"
						class="form-check-input"
						id="rotation-on-other"
						bind:checked={rotation.on_other}
						disabled={readonly}
					/>
					<label class="form-check-label" for="rotation-on-other">
						<i class="fas fa-sync me-1" class:text-info={rotation.on_other} class:text-muted={!rotation.on_other}></i>
						Rotate on other actions
					</label>
				</div>
				<small class="ms-4" class:text-muted={!rotation.on_other} class:text-success={rotation.on_other}>
					{#if rotation.on_other}
						Token rotates on introspection, subtoken creation, etc.
					{:else}
						Token does not rotate on other actions
					{/if}
				</small>
			</div>

			<!-- Lifetime -->
			{#if isEnabled}
				<div class="col-md-6">
					<label class="form-label" for="rotation-lifetime">
						<i class="fas fa-clock me-1"></i>
						New Token Lifetime
					</label>
					{#if readonly}
						<p class="form-control-plaintext">
							{rotation.lifetime ? formatDuration(rotation.lifetime) : 'Default'}
						</p>
					{:else}
						<div class="input-group">
							<input
								type="number"
								class="form-control"
								id="rotation-lifetime"
								min="0"
								placeholder="Default"
								bind:value={lifetimeValue}
							/>
							<select class="form-select" style="max-width: 100px;" bind:value={lifetimeUnit} aria-label="Lifetime unit">
								<option value="hours">Hours</option>
								<option value="days">Days</option>
							</select>
						</div>
						<small class="text-muted">
							How long the rotated token will be valid. Leave empty for default.
						</small>
					{/if}
				</div>

				<!-- Auto-revoke -->
				<div class="col-md-6">
					<div class="form-check mt-4">
						<input
							type="checkbox"
							class="form-check-input"
							id="rotation-auto-revoke"
							bind:checked={rotation.auto_revoke}
							disabled={readonly}
						/>
						<label class="form-check-label" for="rotation-auto-revoke">
							<i class="fas fa-ban me-1" class:text-danger={rotation.auto_revoke} class:text-muted={!rotation.auto_revoke}></i>
							Automatic revocation
						</label>
					</div>
					<small class="ms-4" class:text-muted={!rotation.auto_revoke} class:text-success={rotation.auto_revoke}>
						{#if rotation.auto_revoke}
							Automatic revocation on misuse detection is enabled
						{:else}
							Automatic revocation on misuse detection is disabled
						{/if}
					</small>
				</div>
			{/if}
		</div>

		{#if !isEnabled && !readonly}
			<div class="alert alert-info mt-3 mb-0">
				<i class="fas fa-info-circle me-2"></i>
				Rotation is disabled. Enable rotation to automatically refresh the token on use.
			</div>
		{/if}
		{#if !isEnabled && readonly}
			<div class="alert alert-secondary mt-3 mb-0">
				<i class="fas fa-times-circle me-2"></i>
				Rotation is not enabled for this token. The token will not be automatically refreshed.
			</div>
		{/if}
	</div>
</CollapsibleSection>

<style>
	.form-select {
		max-width: 300px;
	}

	.form-check-label {
		cursor: pointer;
	}

	.form-control-plaintext {
		padding: 0.375rem 0;
	}

	.rotation-icon i {
		font-size: 1rem;
	}

	.rotation-icon .text-muted {
		opacity: 0.5;
	}
</style>
