<script lang="ts">
	import { onMount } from 'svelte';
	import type { Restriction, Schedule } from '$lib/types';
	import { formatDateTime } from '$lib/utils/format';
	import { countries, getCountryName } from '$lib/data/countries';
	import DateTimePicker from '../DateTimePicker.svelte';

	export let restriction: Restriction;
	export let index: number;
	export let readonly: boolean = false;
	export let onRemove: (() => void) | undefined = undefined;
	export let availableScopes: string[] = [];

	// Local state for datetime inputs using flatpickr format "YYYY-MM-DD HH:MM"
	function timestampToDateTimeString(timestamp: number | undefined): string {
		if (!timestamp) return '';
		const d = new Date(timestamp * 1000);
		const year = d.getFullYear();
		const month = String(d.getMonth() + 1).padStart(2, '0');
		const day = String(d.getDate()).padStart(2, '0');
		const hours = String(d.getHours()).padStart(2, '0');
		const minutes = String(d.getMinutes()).padStart(2, '0');
		return `${year}-${month}-${day} ${hours}:${minutes}`;
	}

	function dateTimeStringToTimestamp(dateTimeStr: string): number | undefined {
		if (!dateTimeStr) return undefined;
		// Parse "YYYY-MM-DD HH:MM" format
		const d = new Date(dateTimeStr.replace(' ', 'T'));
		if (isNaN(d.getTime())) return undefined;
		return Math.floor(d.getTime() / 1000);
	}

	// Use reactive statements to update datetime strings when restriction changes (e.g., from template)
	$: nbfDateTimeStr = timestampToDateTimeString(restriction.nbf);
	$: expDateTimeStr = timestampToDateTimeString(restriction.exp);

	// Location restriction type: 'hosts', 'geoip_allow', 'geoip_disallow'
	let locationRestrictionType: 'hosts' | 'geoip_allow' | 'geoip_disallow' = 'geoip_allow';
	let locationTypeInitialized = false;
	
	// Initialize location restriction type based on existing data (only once)
	onMount(() => {
		if (restriction.hosts && restriction.hosts.length > 0) {
			locationRestrictionType = 'hosts';
		} else if (restriction.geoip_disallow && restriction.geoip_disallow.length > 0) {
			locationRestrictionType = 'geoip_disallow';
		} else {
			locationRestrictionType = 'geoip_allow';
		}
		locationTypeInitialized = true;
	});

	// Parse current scope string into array for checkbox state
	$: selectedScopes = restriction.scope ? restriction.scope.split(' ').filter(s => s.length > 0) : [];
	
	// If no scopes selected, all scopes are implicitly allowed
	$: allScopesAllowed = selectedScopes.length === 0;

	// Sync datetime changes back to restriction object
	function updateNbf(dateTimeStr: string) {
		nbfDateTimeStr = dateTimeStr;
		const timestamp = dateTimeStringToTimestamp(dateTimeStr);
		if (timestamp !== undefined) {
			restriction.nbf = timestamp;
		} else {
			delete restriction.nbf;
		}
		restriction = restriction; // Trigger reactivity
	}

	function updateExp(dateTimeStr: string) {
		expDateTimeStr = dateTimeStr;
		const timestamp = dateTimeStringToTimestamp(dateTimeStr);
		if (timestamp !== undefined) {
			restriction.exp = timestamp;
		} else {
			delete restriction.exp;
		}
		restriction = restriction; // Trigger reactivity
	}

	// Scope management
	function toggleScope(scope: string) {
		if (readonly) return;
		
		const currentScopes = restriction.scope ? restriction.scope.split(' ').filter(s => s.length > 0) : [];
		const idx = currentScopes.indexOf(scope);
		
		if (idx >= 0) {
			currentScopes.splice(idx, 1);
		} else {
			currentScopes.push(scope);
		}
		
		if (currentScopes.length > 0) {
			restriction.scope = currentScopes.join(' ');
		} else {
			delete restriction.scope;
		}
		// Trigger reactivity
		restriction = restriction;
	}

	// Hosts management
	let newHost = '';
	function addHost() {
		if (newHost.trim()) {
			if (!restriction.hosts) restriction.hosts = [];
			restriction.hosts = [...restriction.hosts, newHost.trim()];
			newHost = '';
		}
	}

	function removeHost(idx: number) {
		if (restriction.hosts) {
			restriction.hosts = restriction.hosts.filter((_, i) => i !== idx);
			if (restriction.hosts.length === 0) {
				delete restriction.hosts;
				restriction = restriction;
			}
		}
	}

	// Audience management
	let newAudience = '';
	function addAudience() {
		if (newAudience.trim()) {
			if (!restriction.audience) restriction.audience = [];
			restriction.audience = [...restriction.audience, newAudience.trim()];
			newAudience = '';
		}
	}

	function removeAudience(idx: number) {
		if (restriction.audience) {
			restriction.audience = restriction.audience.filter((_, i) => i !== idx);
			if (restriction.audience.length === 0) {
				delete restriction.audience;
				restriction = restriction;
			}
		}
	}

	// GeoIP allow management
	let selectedGeoAllow = '';
	function addGeoAllow() {
		if (selectedGeoAllow) {
			if (!restriction.geoip_allow) restriction.geoip_allow = [];
			if (!restriction.geoip_allow.includes(selectedGeoAllow)) {
				restriction.geoip_allow = [...restriction.geoip_allow, selectedGeoAllow];
			}
			selectedGeoAllow = '';
		}
	}
	
	function handleGeoAllowChange() {
		if (selectedGeoAllow) {
			addGeoAllow();
		}
	}

	function removeGeoAllow(idx: number) {
		if (restriction.geoip_allow) {
			restriction.geoip_allow = restriction.geoip_allow.filter((_, i) => i !== idx);
			if (restriction.geoip_allow.length === 0) {
				delete restriction.geoip_allow;
				restriction = restriction;
			}
		}
	}

	// GeoIP disallow management
	let selectedGeoDisallow = '';
	function addGeoDisallow() {
		if (selectedGeoDisallow) {
			if (!restriction.geoip_disallow) restriction.geoip_disallow = [];
			if (!restriction.geoip_disallow.includes(selectedGeoDisallow)) {
				restriction.geoip_disallow = [...restriction.geoip_disallow, selectedGeoDisallow];
			}
			selectedGeoDisallow = '';
		}
	}
	
	function handleGeoDisallowChange() {
		if (selectedGeoDisallow) {
			addGeoDisallow();
		}
	}

	function removeGeoDisallow(idx: number) {
		if (restriction.geoip_disallow) {
			restriction.geoip_disallow = restriction.geoip_disallow.filter((_, i) => i !== idx);
			if (restriction.geoip_disallow.length === 0) {
				delete restriction.geoip_disallow;
				restriction = restriction;
			}
		}
	}

	// Handle location restriction type change - don't clear other types to allow combining
	function handleLocationTypeChange() {
		// Just switch the view, don't clear data
		// Users can combine hosts with geoip restrictions
	}

	// --- Schedule management ---
	const FALLBACK_TIMEZONES = [
		'UTC', 'Europe/Berlin', 'Europe/London', 'Europe/Paris', 'Europe/Vienna', 'Europe/Zurich',
		'Europe/Amsterdam', 'Europe/Stockholm', 'Europe/Vilnius', 'America/New_York', 'America/Chicago',
		'America/Los_Angeles', 'Asia/Tokyo', 'Asia/Shanghai', 'Asia/Kolkata', 'Australia/Sydney',
		'Pacific/Auckland'
	];
	// The user's local timezone; used as the default for new schedules.
	const USER_TIMEZONE = (() => {
		try {
			return Intl.DateTimeFormat().resolvedOptions().timeZone || 'UTC';
		} catch {
			return 'UTC';
		}
	})();

	// Full list of IANA timezones from the browser, with a curated fallback.
	const ALL_TIMEZONES: string[] = (() => {
		try {
			const zones = Intl.supportedValuesOf?.('timeZone') ?? [];
			if (zones.length > 0) return zones;
		} catch {
			// ignore, fall through to fallback
		}
		return FALLBACK_TIMEZONES;
	})();

	$: timezoneGroups = (() => {
		const groups: Record<string, string[]> = {};
		for (const tz of ALL_TIMEZONES) {
			const idx = tz.indexOf('/');
			const label = idx > 0 ? tz.slice(0, idx) : 'Other';
			(groups[label] ??= []).push(tz);
		}
		return Object.entries(groups)
			.sort(([a], [b]) => a.localeCompare(b))
			.map(([label, zones]) => ({ label, zones }));
	})();

	const WEEKDAYS = [
		{ value: 1, label: 'Mon' },
		{ value: 2, label: 'Tue' },
		{ value: 3, label: 'Wed' },
		{ value: 4, label: 'Thu' },
		{ value: 5, label: 'Fri' },
		{ value: 6, label: 'Sat' },
		{ value: 7, label: 'Sun' }
	];
	const WEEKDAY_LABELS: Record<number, string> = { 1: 'Mon', 2: 'Tue', 3: 'Wed', 4: 'Thu', 5: 'Fri', 6: 'Sat', 7: 'Sun' };

	// Ensure schedule exists so bindings work
	$: if (restriction.schedule && !restriction.schedule.timezone) restriction.schedule.timezone = USER_TIMEZONE;

	$: scheduleActive = !!restriction.schedule;

	function toggleSchedule(active: boolean) {
		if (active) {
			restriction.schedule = { timezone: USER_TIMEZONE, days: [], days_of_month: [] };
		} else {
			delete restriction.schedule;
		}
		restriction = restriction;
	}

	function toggleDay(day: number) {
		if (readonly || !restriction.schedule) return;
		const days = restriction.schedule.days ?? [];
		const idx = days.indexOf(day);
		if (idx >= 0) {
			days.splice(idx, 1);
		} else {
			days.push(day);
		}
		if (days.length === 0) {
			delete restriction.schedule.days;
		} else {
			restriction.schedule.days = [...days].sort((a, b) => a - b);
		}
		restriction = restriction;
	}

	$: domInput = (restriction.schedule?.days_of_month ?? []).join(', ');

	function updateDom() {
		if (readonly || !restriction.schedule) return;
		const vals = domInput
			.split(',')
			.map((s) => parseInt(s.trim(), 10))
			.filter((n) => !isNaN(n) && n !== 0 && n >= -31 && n <= 31);
		const unique = [...new Set(vals)].sort((a, b) => a - b);
		if (unique.length > 0) {
			restriction.schedule.days_of_month = unique;
		} else {
			delete restriction.schedule.days_of_month;
		}
		restriction = restriction;
	}

	function updateWindowField(field: 'from' | 'to', val: string) {
		if (readonly || !restriction.schedule) return;
		if (val) {
			restriction.schedule[field] = val;
		} else {
			delete restriction.schedule[field];
		}
		restriction = restriction;
	}

	function clearWindow() {
		if (readonly || !restriction.schedule) return;
		delete restriction.schedule.from;
		delete restriction.schedule.to;
		restriction = restriction;
	}

	function updateEvery() {
		if (!restriction.schedule) return;
		if (restriction.schedule.every !== undefined && restriction.schedule.every > 1) {
			restriction.schedule.every = restriction.schedule.every;
		} else {
			delete restriction.schedule.every;
			delete restriction.schedule.anchor;
		}
		restriction = restriction;
	}

	function scheduleSummary(s: Schedule): string {
		const parts: string[] = [];
		if (s.days && s.days.length > 0) parts.push(s.days.map((d) => WEEKDAY_LABELS[d] ?? d).join(', '));
		if (s.days_of_month && s.days_of_month.length > 0) parts.push(`day-of-month: ${s.days_of_month.join(', ')}`);
		if (s.from && s.to) parts.push(`${s.from} - ${s.to}`);
		if (s.every) {
			parts.push(`every ${s.every} days${s.anchor ? ` (anchor ${s.anchor})` : ''}`);
		}
		parts.push(s.timezone || 'UTC');
		return parts.join(' · ') || 'any time';
	}
</script>

<div class="restriction-clause card mb-3">
	<div class="card-header d-flex justify-content-between align-items-center py-2">
		<span class="fw-bold">Restriction #{index + 1}</span>
		{#if !readonly && onRemove}
			<button type="button" class="btn btn-sm btn-outline-danger" on:click={onRemove} title="Remove restriction">
				<i class="fas fa-trash"></i>
			</button>
		{/if}
	</div>
	<div class="card-body">
		<div class="row g-3">
			<!-- Time restrictions -->
			<div class="col-md-6">
				<label class="form-label" for="restriction-{index}-nbf">
					<i class="fas fa-clock me-1"></i>
					Not Before
				</label>
				{#if readonly}
					<p class="form-control-plaintext">
						{restriction.nbf ? formatDateTime(restriction.nbf) : 'Not set'}
					</p>
				{:else}
					<DateTimePicker
						id="restriction-{index}-nbf"
						value={nbfDateTimeStr}
						placeholder="Select start date/time"
						on:change={(e) => updateNbf(e.detail)}
					/>
				{/if}
			</div>

			<div class="col-md-6">
				<label class="form-label" for="restriction-{index}-exp">
					<i class="fas fa-hourglass-end me-1"></i>
					Expires
				</label>
				{#if readonly}
					<p class="form-control-plaintext">
						{restriction.exp ? formatDateTime(restriction.exp) : 'Not set'}
					</p>
				{:else}
					<DateTimePicker
						id="restriction-{index}-exp"
						value={expDateTimeStr}
						placeholder="Select expiry date/time"
						on:change={(e) => updateExp(e.detail)}
					/>
				{/if}
			</div>

			<!-- Schedule -->
			<div class="col-12">
				<span class="form-label d-block">
					<i class="fas fa-calendar-alt me-1"></i>
					Schedule
				</span>
				{#if readonly}
					<p class="form-control-plaintext">
						{restriction.schedule ? scheduleSummary(restriction.schedule) : 'Not set'}
					</p>
				{:else}
					<div class="form-check mb-2">
						<input
							class="form-check-input"
							type="checkbox"
							id="restriction-{index}-schedule-enabled"
							checked={scheduleActive}
							on:change={(e) => toggleSchedule(e.currentTarget.checked)}
						/>
						<label class="form-check-label" for="restriction-{index}-schedule-enabled">
							Allow usage only on a recurring schedule
						</label>
					</div>
					{#if restriction.schedule}
						<div class="border rounded p-3 bg-body-tertiary">
							<div class="row g-3">
								<div class="col-md-6">
									<label class="form-label small mb-1" for="restriction-{index}-schedule-tz">
										Timezone
									</label>
									<select
										class="form-select form-select-sm"
										id="restriction-{index}-schedule-tz"
										bind:value={restriction.schedule.timezone}
									>
										{#each timezoneGroups as group}
											<optgroup label={group.label}>
												{#each group.zones as tz}
													<option value={tz}>{tz}</option>
												{/each}
											</optgroup>
										{/each}
									</select>
								</div>

								<div class="col-md-6">
									<span class="form-label small mb-1 d-block">Weekdays</span>
									<div class="d-flex flex-wrap gap-2">
										{#each WEEKDAYS as day}
											<label class="btn btn-sm border {restriction.schedule.days?.includes(day.value) ? 'btn-primary' : 'btn-outline-secondary'}"
												class:pe-none={readonly}>
												<input
													class="btn-check"
													type="checkbox"
													value={day.value}
													checked={restriction.schedule.days?.includes(day.value)}
													on:change={() => toggleDay(day.value)}
												/>
												{day.label}
											</label>
										{/each}
									</div>
									<small class="text-muted">Leave empty for every day of the week.</small>
								</div>

								<div class="col-md-6">
									<label class="form-label small mb-1" for="restriction-{index}-schedule-dom">
										Days of month
									</label>
									<input
										class="form-control form-control-sm"
										id="restriction-{index}-schedule-dom"
										placeholder="e.g. 1, 15, -1"
										bind:value={domInput}
										on:change={updateDom}
									/>
									<small class="text-muted">Comma-separated. Negative values count from the end of the month (-1 = last day).</small>
								</div>

								<div class="col-md-6">
									<span class="form-label small mb-1 d-block">Time window</span>
									<div class="d-flex align-items-center gap-2">
										<DateTimePicker
											id="restriction-{index}-schedule-from"
											value={restriction.schedule.from ?? ''}
											placeholder="Start"
											noCalendar
											dateFormat="H:i"
											altFormat="H:i"
											on:change={(e) => updateWindowField('from', e.detail)}
										/>
										<span>-</span>
										<DateTimePicker
											id="restriction-{index}-schedule-to"
											value={restriction.schedule.to ?? ''}
											placeholder="End"
											noCalendar
											dateFormat="H:i"
											altFormat="H:i"
											on:change={(e) => updateWindowField('to', e.detail)}
										/>
										{#if restriction.schedule.from || restriction.schedule.to}
											<button
												type="button"
												class="btn btn-sm btn-outline-secondary"
												on:click={clearWindow}
												title="Clear window (allow the whole day)"
											>
												<i class="fas fa-times"></i>
											</button>
										{/if}
									</div>
									<small class="text-muted">Leave empty for the whole day. An end before the start describes a window wrapping around midnight.</small>
								</div>

								<div class="col-md-6">
									<label class="form-label small mb-1" for="restriction-{index}-schedule-every">
										Every N days
									</label>
									<input
										class="form-control form-control-sm"
										type="number"
										id="restriction-{index}-schedule-every"
										min="2"
										placeholder="e.g. 2 for every other day"
										bind:value={restriction.schedule.every}
										on:change={updateEvery}
									/>
									<small class="text-muted">If set, usage is only allowed every N days.</small>
								</div>

								<div class="col-md-6">
									<label class="form-label small mb-1" for="restriction-{index}-schedule-anchor">
										Anchor date
									</label>
									<input
										class="form-control form-control-sm"
										type="date"
										id="restriction-{index}-schedule-anchor"
										bind:value={restriction.schedule.anchor}
									/>
									<small class="text-muted">Reference date for "every N days"; defaults to the creation date of the mytoken.</small>
								</div>
							</div>
						</div>
					{/if}
				{/if}
			</div>

			<!-- Scopes as checkable list -->
			<div class="col-12">
				<span class="form-label d-block">
					<i class="fas fa-shield-alt me-1"></i>
					Scopes
				</span>
				{#if availableScopes.length > 0}
					<div class="scope-list border rounded p-2">
						<table class="table table-sm table-striped mb-0">
							<tbody>
								{#each availableScopes as scope}
									{@const isSelected = selectedScopes.includes(scope)}
									<tr 
										class="scope-row"
										class:readonly
										on:click={() => !readonly && toggleScope(scope)}
										role="button"
										tabindex={readonly ? -1 : 0}
										on:keydown={(e) => (e.key === 'Enter' || e.key === ' ') && !readonly && toggleScope(scope)}
									>
										<td class="align-middle">{scope}</td>
										<td class="align-middle text-center" style="width: 40px;">
											{#if allScopesAllowed || isSelected}
												<i class="fas fa-check-circle text-success"></i>
											{:else}
												<i class="fas fa-times-circle text-danger"></i>
											{/if}
										</td>
										<td class="align-middle text-center" style="width: 40px;">
											<input 
												type="checkbox" 
												class="form-check-input"
												checked={isSelected}
												disabled={readonly}
												on:click|stopPropagation
												on:change={() => toggleScope(scope)}
											/>
										</td>
									</tr>
								{/each}
							</tbody>
						</table>
					</div>
					<small class="text-muted">
						{#if allScopesAllowed}
							No scopes selected - all scopes are allowed.
						{:else}
							Access Tokens obtained with this mytoken can only have the selected scope values.
						{/if}
					</small>
				{:else if readonly}
					<p class="form-control-plaintext">
						{restriction.scope || 'Not set'}
					</p>
				{:else}
					<input
						type="text"
						class="form-control"
						id="restriction-{index}-scope"
						placeholder="openid profile email"
						bind:value={restriction.scope}
					/>
					<small class="text-muted">Space-separated list of scopes (no scopes available from provider)</small>
				{/if}
			</div>

			<!-- Audience -->
			<div class="col-12">
				<label class="form-label" for="restriction-{index}-audience">
					<i class="fas fa-server me-1"></i>
					Audience
				</label>
				<div class="tags-container mb-2">
					{#each restriction.audience ?? [] as aud, idx}
						<span class="badge bg-info me-1 mb-1">
							{aud}
							{#if !readonly}
								<button
									type="button"
									class="btn-close btn-close-white ms-1"
									style="font-size: 0.6em;"
									on:click={() => removeAudience(idx)}
									title="Remove {aud}"
								></button>
							{/if}
						</span>
					{/each}
				</div>
				{#if !readonly}
					<div class="input-group input-group-sm">
						<input
							type="text"
							class="form-control"
							id="restriction-{index}-audience"
							placeholder="audience URI"
							bind:value={newAudience}
							on:keydown={(e) => e.key === 'Enter' && (e.preventDefault(), addAudience())}
						/>
						<button class="btn btn-outline-primary" type="button" on:click={addAudience} title="Add audience">
							<i class="fas fa-plus"></i>
						</button>
					</div>
					<small class="text-muted">If set, Access Tokens obtained with this mytoken can only be used at these audiences.</small>
				{/if}
			</div>

			<!-- Location-based restrictions -->
			<div class="col-12">
				<label class="form-label" for="restriction-{index}-location-type">
					<i class="fas fa-map-marker-alt me-1"></i>
					Restrict request location based on
				</label>
				{#if !readonly}
					<select 
						class="form-select mb-2" 
						id="restriction-{index}-location-type"
						bind:value={locationRestrictionType}
						on:change={handleLocationTypeChange}
					>
						<option value="hosts">Hostname, IP addresses, or subnets</option>
						<option value="geoip_allow">Allowed countries</option>
						<option value="geoip_disallow">Forbidden countries</option>
					</select>
				{/if}

				<div class="ms-3">
					{#if locationRestrictionType === 'hosts'}
						<!-- Hosts -->
						<span class="form-label small d-block">Allowed requests from these hosts</span>
						<div class="tags-container mb-2">
							{#each restriction.hosts ?? [] as host, idx}
								<span class="badge bg-secondary me-1 mb-1">
									{host}
									{#if !readonly}
										<button
											type="button"
											class="btn-close btn-close-white ms-1"
											style="font-size: 0.6em;"
											on:click={() => removeHost(idx)}
											title="Remove {host}"
										></button>
									{/if}
								</span>
							{/each}
						</div>
						{#if !readonly}
							<div class="input-group input-group-sm">
								<input
									type="text"
									class="form-control"
									placeholder="Hostname, IP, Subnet"
									bind:value={newHost}
									on:keydown={(e) => e.key === 'Enter' && (e.preventDefault(), addHost())}
								/>
								<button class="btn btn-outline-primary" type="button" on:click={addHost} title="Add host">
									<i class="fas fa-plus"></i>
								</button>
							</div>
							<small class="text-muted">If set, the mytoken can only be used from these hosts given by hostname, IP address or subnets.</small>
						{/if}
					{:else if locationRestrictionType === 'geoip_allow'}
						<!-- GeoIP Allow -->
						<span class="form-label small d-block">Allowed request countries</span>
						<div class="tags-container mb-2">
							{#each restriction.geoip_allow ?? [] as code, idx}
								<span class="badge bg-success me-1 mb-1">
									{getCountryName(code)} ({code})
									{#if !readonly}
										<button
											type="button"
											class="btn-close btn-close-white ms-1"
											style="font-size: 0.6em;"
											on:click={() => removeGeoAllow(idx)}
											title="Remove {getCountryName(code)}"
										></button>
									{/if}
								</span>
							{/each}
						</div>
						{#if !readonly}
							<select 
								class="form-select form-select-sm"
								bind:value={selectedGeoAllow}
								on:change={handleGeoAllowChange}
							>
								<option value="">-- Select a country --</option>
								{#each countries as country}
									<option value={country.code}>{country.name}</option>
								{/each}
							</select>
							<small class="text-muted">If set, the mytoken can only be used from these countries.</small>
						{/if}
					{:else if locationRestrictionType === 'geoip_disallow'}
						<!-- GeoIP Disallow -->
						<span class="form-label small d-block">Forbidden request countries</span>
						<div class="tags-container mb-2">
							{#each restriction.geoip_disallow ?? [] as code, idx}
								<span class="badge bg-danger me-1 mb-1">
									{getCountryName(code)} ({code})
									{#if !readonly}
										<button
											type="button"
											class="btn-close btn-close-white ms-1"
											style="font-size: 0.6em;"
											on:click={() => removeGeoDisallow(idx)}
											title="Remove {getCountryName(code)}"
										></button>
									{/if}
								</span>
							{/each}
						</div>
						{#if !readonly}
							<select 
								class="form-select form-select-sm"
								bind:value={selectedGeoDisallow}
								on:change={handleGeoDisallowChange}
							>
								<option value="">-- Select a country --</option>
								{#each countries as country}
									<option value={country.code}>{country.name}</option>
								{/each}
							</select>
							<small class="text-muted">If set, the mytoken cannot be used from these countries.</small>
						{/if}
					{/if}
				</div>
			</div>

			<!-- Usage limits -->
			<div class="col-md-6">
				<label class="form-label" for="restriction-{index}-usages-at">
					<i class="fas fa-key me-1"></i>
					Usages AT
				</label>
				{#if readonly}
					<p class="form-control-plaintext">
						{restriction.usages_AT ?? 'Unlimited'}
					</p>
				{:else}
					<input
						type="number"
						class="form-control"
						id="restriction-{index}-usages-at"
						min="0"
						placeholder="Unlimited"
						bind:value={restriction.usages_AT}
					/>
					<small class="text-muted">If set, the mytoken can only be used this often to request access tokens.</small>
				{/if}
			</div>

			<div class="col-md-6">
				<label class="form-label" for="restriction-{index}-usages-other">
					<i class="fas fa-sync me-1"></i>
					Usages Other
				</label>
				{#if readonly}
					<p class="form-control-plaintext">
						{restriction.usages_other ?? 'Unlimited'}
					</p>
				{:else}
					<input
						type="number"
						class="form-control"
						id="restriction-{index}-usages-other"
						min="0"
						placeholder="Unlimited"
						bind:value={restriction.usages_other}
					/>
					<small class="text-muted">If set, the mytoken can only be used this often for requests other than requesting access tokens.</small>
				{/if}
			</div>
		</div>
	</div>
</div>

<style>
	.restriction-clause {
		border: 1px solid var(--bs-border-color);
	}

	.card-header {
		background-color: var(--bs-tertiary-bg);
	}

	.tags-container {
		min-height: 24px;
	}

	.badge {
		font-weight: normal;
	}

	.form-control-plaintext {
		padding: 0.375rem 0;
		margin-bottom: 0;
	}

	.scope-list {
		background-color: var(--bs-tertiary-bg);
	}

	.scope-list table {
		margin-bottom: 0;
	}

	.scope-row {
		cursor: pointer;
		user-select: none;
	}

	.scope-row:hover:not(.readonly) {
		background-color: var(--bs-secondary-bg) !important;
	}

	.scope-row.readonly {
		cursor: default;
	}

	.scope-list tr:last-child td {
		border-bottom: none;
	}
</style>
