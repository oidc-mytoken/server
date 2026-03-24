<script lang="ts">
	import { onMount } from 'svelte';
	import type { Notification, Calendar } from '$lib/types';
	import { api, ApiClientError } from '$lib/api/client';
	import { isLoggedIn } from '$lib/stores/auth';
	import { tags } from '$lib/stores/tags';
	import { discovery } from '$lib/stores/discovery';
	import { ui } from '$lib/stores/ui';
	import NotificationList from './NotificationList.svelte';
	import NotificationModal from './NotificationModal.svelte';
	import CalendarList from './CalendarList.svelte';
	import LoadingSpinner from '../LoadingSpinner.svelte';

	// State
	let notifications: Notification[] = [];
	let calendars: Calendar[] = [];
	let loading = true;
	let activeTab: 'notifications' | 'calendars' = 'notifications';
	let loadError: string | null = null;
	
	// Modal states
	let showCreateNotificationModal = false;
	let calendarListRef: CalendarList;

	onMount(async () => {
		await loadData();
	});

	async function loadData() {
		loading = true;
		loadError = null;
		try {
			// Load tags if not already loaded
			if ($isLoggedIn && $discovery.data?.usersettings_endpoint) {
				if (!$tags.loaded) {
					await tags.fetch($discovery.data.usersettings_endpoint);
				}
			}

			// Load notifications and calendars
			const [notifResult, calResult] = await Promise.all([
				api.getNotifications(),
				api.getCalendars()
			]);

			notifications = notifResult;
			calendars = calResult;
		} catch (error) {
			console.error('Failed to load notifications/calendars:', error);
			if (error instanceof ApiClientError) {
				if (error.isAuthError()) {
					loadError = 'Session expired or invalid. Please log in again.';
				} else {
					loadError = error.description ?? error.code;
				}
			} else {
				loadError = (error as Error).message;
			}
			notifications = [];
			calendars = [];
		} finally {
			loading = false;
		}
	}

	function handleNotificationDeleted(managementCode: string) {
		notifications = notifications.filter(n => n.management_code !== managementCode);
	}

	function handleNotificationUpdated(updated: Notification | void) {
		// Reload to get the updated notification data
		loadData();
	}

	function handleNotificationCreated() {
		showCreateNotificationModal = false;
		loadData(); // Reload to get the new notification
	}

	function handleCalendarDeleted(calendarId: string) {
		calendars = calendars.filter(c => c.id !== calendarId);
	}

	function handleCalendarUpdated(updated: Calendar) {
		calendars = calendars.map(c => c.id === updated.id ? updated : c);
	}

	function handleCalendarCreated() {
		loadData();
	}
</script>

<div class="notifications-container">
	<!-- Header with tabs -->
	<div class="d-flex justify-content-between align-items-center mb-4">
		<ul class="nav nav-pills">
			<li class="nav-item">
				<button 
					class="nav-link" 
					class:active={activeTab === 'notifications'}
					onclick={() => activeTab = 'notifications'}
				>
					<i class="fas fa-bell me-1"></i>
					Notifications
					{#if notifications.length > 0}
						<span class="badge bg-secondary ms-1">{notifications.length}</span>
					{/if}
				</button>
			</li>
			<li class="nav-item">
				<button 
					class="nav-link" 
					class:active={activeTab === 'calendars'}
					onclick={() => activeTab = 'calendars'}
				>
					<i class="fas fa-calendar me-1"></i>
					Calendars
					{#if calendars.length > 0}
						<span class="badge bg-secondary ms-1">{calendars.length}</span>
					{/if}
				</button>
			</li>
		</ul>

		<div class="d-flex gap-2">
			<button class="btn btn-outline-primary btn-sm" onclick={loadData} disabled={loading} title="Refresh">
				<i class="fas fa-sync" class:fa-spin={loading}></i>
			</button>
			{#if activeTab === 'notifications'}
				<button class="btn btn-primary btn-sm" onclick={() => showCreateNotificationModal = true}>
					<i class="fas fa-plus me-1"></i>
					New Notification
				</button>
			{:else if activeTab === 'calendars'}
				<button class="btn btn-primary btn-sm" onclick={() => calendarListRef?.openCreateModal()}>
					<i class="fas fa-plus me-1"></i>
					New Calendar
				</button>
			{/if}
		</div>
	</div>

	{#if loading}
		<LoadingSpinner message="Loading notifications..." />
	{:else if loadError}
		<div class="alert alert-warning">
			<i class="fas fa-exclamation-triangle me-2"></i>
			<strong>Failed to load data:</strong> {loadError}
			<button class="btn btn-sm btn-outline-warning ms-3" onclick={loadData}>
				<i class="fas fa-sync me-1"></i>
				Retry
			</button>
		</div>
	{:else}
		<!-- Content -->
		{#if activeTab === 'notifications'}
			<NotificationList 
				{notifications}
				onDelete={handleNotificationDeleted}
				onUpdate={handleNotificationUpdated}
			/>
		{:else}
			<CalendarList 
				bind:this={calendarListRef}
				{calendars}
				onDelete={handleCalendarDeleted}
				onUpdate={handleCalendarUpdated}
				onCreate={handleCalendarCreated}
			/>
		{/if}
	{/if}
</div>

<!-- Create Notification Modal -->
<NotificationModal 
	bind:show={showCreateNotificationModal}
	notification={null}
	on:saved={handleNotificationCreated}
	on:cancel={() => showCreateNotificationModal = false}
/>

<style>
	.nav-pills .nav-link {
		color: #6c757d;
	}

	.nav-pills .nav-link.active {
		background-color: #df691a;
	}
</style>
