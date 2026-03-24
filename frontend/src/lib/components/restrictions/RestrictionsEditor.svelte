<script lang="ts">
	import type { Restriction, RestrictionTemplate } from '$lib/types';
	import RestrictionClause from './RestrictionClause.svelte';
	import CollapsibleSection from '../CollapsibleSection.svelte';

	export let restrictions: Restriction[] = [];
	export let templates: RestrictionTemplate[] = [];
	export let readonly: boolean = false;
	export let collapsed: boolean = false;
	export let showTemplates: boolean = true;
	export let showJsonEditor: boolean = false;
	export let availableScopes: string[] = [];

	let selectedTemplate = '';
	let jsonEditorMode = false;
	let jsonText = '';
	let jsonError = '';

	$: if (jsonEditorMode) {
		jsonText = JSON.stringify(restrictions, null, 2);
	}

	// Icon status calculation
	interface RestrictionIcons {
		time: { color: 'success' | 'warning' | 'danger'; title: string };
		host: { color: 'success' | 'warning'; title: string };
		scope: { color: 'success' | 'warning'; title: string };
		aud: { color: 'success' | 'warning'; title: string };
		usages: { color: 'success' | 'warning'; title: string };
	}

	$: icons = calculateIcons(restrictions);

	function calculateIcons(restr: Restriction[]): RestrictionIcons {
		let clausesWithHost = 0;
		let clausesWithScope = 0;
		let clausesWithAud = 0;
		let clausesWithUsages = 0;
		let maxExpires = 0;
		let doesNotExpire = false;

		for (const r of restr) {
			if (r.scope !== undefined) {
				clausesWithScope++;
			}
			if (r.audience !== undefined && r.audience.length > 0) {
				clausesWithAud++;
			}
			if ((r.hosts !== undefined && r.hosts.length > 0) ||
				(r.geoip_allow !== undefined && r.geoip_allow.length > 0) ||
				(r.geoip_disallow !== undefined && r.geoip_disallow.length > 0)) {
				clausesWithHost++;
			}
			if (r.usages_other !== undefined || r.usages_AT !== undefined) {
				clausesWithUsages++;
			}
			if (r.exp === undefined || r.exp === 0) {
				doesNotExpire = true;
			} else if (r.exp > maxExpires) {
				maxExpires = r.exp;
			}
		}

		if (doesNotExpire) {
			maxExpires = 0;
		}

		const total = restr.length;

		// Time icon
		let timeIcon: RestrictionIcons['time'];
		if (maxExpires === 0) {
			timeIcon = { color: 'danger', title: 'This token does not expire!' };
		} else if ((maxExpires - Date.now() / 1000) > 3 * 24 * 3600) {
			timeIcon = { color: 'warning', title: 'This token is long-lived.' };
		} else {
			timeIcon = { color: 'success', title: 'This token expires within 3 days.' };
		}

		// Host icon
		const hostIcon: RestrictionIcons['host'] = clausesWithHost === total && total > 0
			? { color: 'success', title: 'The hosts from which this token can be used are restricted.' }
			: { color: 'warning', title: 'This token can be used from any host.' };

		// Scope icon
		const scopeIcon: RestrictionIcons['scope'] = clausesWithScope === total && total > 0
			? { color: 'success', title: 'This token has restrictions for scopes.' }
			: { color: 'warning', title: 'This token can use all configured scopes.' };

		// Audience icon
		const audIcon: RestrictionIcons['aud'] = clausesWithAud === total && total > 0
			? { color: 'success', title: 'This token can only obtain access tokens with restricted audiences.' }
			: { color: 'warning', title: 'This token can obtain access tokens with any audiences.' };

		// Usages icon
		const usagesIcon: RestrictionIcons['usages'] = clausesWithUsages === total && total > 0
			? { color: 'success', title: 'This token can only be used a limited number of times.' }
			: { color: 'warning', title: 'This token can be used an infinite number of times.' };

		return {
			time: timeIcon,
			host: hostIcon,
			scope: scopeIcon,
			aud: audIcon,
			usages: usagesIcon
		};
	}

	function addRestriction() {
		restrictions = [...restrictions, {}];
	}

	function removeRestriction(index: number) {
		restrictions = restrictions.filter((_, i) => i !== index);
	}

	function applyTemplate(template: RestrictionTemplate) {
		// Deep clone the template restrictions
		restrictions = JSON.parse(JSON.stringify(template.restrictions));
	}

	function handleTemplateChange(event: Event) {
		const target = event.target as HTMLSelectElement;
		const template = templates.find((t: RestrictionTemplate) => t.name === target.value);
		if (template) {
			applyTemplate(template);
		}
	}

	function toggleJsonEditor() {
		if (jsonEditorMode) {
			// Switching from JSON to GUI - try to parse
			try {
				restrictions = JSON.parse(jsonText);
				jsonError = '';
				jsonEditorMode = false;
			} catch (e) {
				jsonError = 'Invalid JSON: ' + (e as Error).message;
			}
		} else {
			// Switching from GUI to JSON
			jsonText = JSON.stringify(restrictions, null, 2);
			jsonError = '';
			jsonEditorMode = true;
		}
	}

	function handleJsonChange() {
		try {
			JSON.parse(jsonText);
			jsonError = '';
		} catch (e) {
			jsonError = 'Invalid JSON: ' + (e as Error).message;
		}
	}

	export function getRestrictions(): Restriction[] {
		if (jsonEditorMode && !jsonError) {
			return JSON.parse(jsonText);
		}
		return restrictions;
	}
</script>

<CollapsibleSection
	title="Restrictions"
	icon="fa-lock"
	{collapsed}
>
	<span slot="header-right" class="restriction-icons ms-2">
		<i 
			class="fas fa-clock me-1" 
			class:text-success={icons.time.color === 'success'}
			class:text-warning={icons.time.color === 'warning'}
			class:text-danger={icons.time.color === 'danger'}
			title={icons.time.title}
		></i>
		<i 
			class="fas fa-network-wired me-1" 
			class:text-success={icons.host.color === 'success'}
			class:text-warning={icons.host.color === 'warning'}
			title={icons.host.title}
		></i>
		<i 
			class="fas fa-shield-alt me-1" 
			class:text-success={icons.scope.color === 'success'}
			class:text-warning={icons.scope.color === 'warning'}
			title={icons.scope.title}
		></i>
		<i 
			class="fas fa-server me-1" 
			class:text-success={icons.aud.color === 'success'}
			class:text-warning={icons.aud.color === 'warning'}
			title={icons.aud.title}
		></i>
		<i 
			class="fas fa-less-than-equal" 
			class:text-success={icons.usages.color === 'success'}
			class:text-warning={icons.usages.color === 'warning'}
			title={icons.usages.title}
		></i>
	</span>
	{#if showTemplates && templates.length > 0 && !readonly}
		<div class="mb-3">
			<label for="restr-template" class="form-label">
				<i class="fas fa-layer-group me-1"></i>
				Apply Template
			</label>
			<select
				id="restr-template"
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

	{#if !readonly && showJsonEditor}
		<div class="mb-3">
			<button
				type="button"
				class="btn btn-sm"
				class:btn-outline-secondary={!jsonEditorMode}
				class:btn-secondary={jsonEditorMode}
				on:click={toggleJsonEditor}
			>
				<i class="fas fa-code me-1"></i>
				{jsonEditorMode ? 'Switch to GUI' : 'Edit as JSON'}
			</button>
		</div>
	{/if}

	{#if jsonEditorMode}
		<div class="json-editor mb-3">
			<textarea
				class="form-control font-monospace"
				rows="10"
				bind:value={jsonText}
				on:input={handleJsonChange}
				class:is-invalid={jsonError}
			></textarea>
			{#if jsonError}
				<div class="invalid-feedback">{jsonError}</div>
			{/if}
		</div>
	{:else}
		<div class="restrictions-list">
			{#each restrictions as restriction, index}
				<RestrictionClause
					bind:restriction
					{index}
					{readonly}
					{availableScopes}
					onRemove={readonly ? undefined : () => removeRestriction(index)}
				/>
			{/each}

			{#if restrictions.length === 0}
				<div class="alert alert-info mb-3">
					<i class="fas fa-info-circle me-2"></i>
					No restrictions defined. The token will have no usage restrictions.
				</div>
			{/if}

			{#if !readonly}
				<button type="button" class="btn btn-outline-primary" on:click={addRestriction}>
					<i class="fas fa-plus me-1"></i>
					Add Restriction Clause
				</button>
			{/if}
		</div>
	{/if}

	{#if restrictions.length > 0 && !readonly}
		<div class="mt-3">
			<small class="text-muted">
				<i class="fas fa-info-circle me-1"></i>
				Multiple restriction clauses are combined with OR logic. Each clause's conditions are combined with AND logic.
			</small>
		</div>
	{/if}
</CollapsibleSection>

<style>
	.json-editor textarea {
		font-size: 0.875rem;
	}

	.form-select {
		max-width: 300px;
	}

	.restriction-icons i {
		font-size: 1rem;
	}

	.restriction-icons .text-warning {
		opacity: 0.5;
	}
</style>
