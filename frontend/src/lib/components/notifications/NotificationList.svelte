<script lang="ts">
	import type { Notification, MytokenEntry } from '$lib/types';
	import { api, ApiClientError } from '$lib/api/client';
	import { ui } from '$lib/stores/ui';
	import LoadingSpinner from '../LoadingSpinner.svelte';
	import NotificationItem from './NotificationItem.svelte';
	import NotificationModal from './NotificationModal.svelte';

	export let notifications: Notification[] = [];
	export let tokens: MytokenEntry[] = [];
	export let loading = false;
	export let onDelete: ((managementCode: string) => void) | undefined = undefined;
	export let onUpdate: ((notification: Notification | void) => void) | undefined = undefined;

	// Edit modal state
	let showEditModal = false;
	let editingNotification: Notification | null = null;

	function openEditModal(notification: Notification) {
		editingNotification = notification;
		showEditModal = true;
	}

	function handleEditSaved(event: CustomEvent<Notification | void>) {
		const updated = event.detail;
		onUpdate?.(updated);
		showEditModal = false;
		editingNotification = null;
	}

	function handleEditCancel() {
		showEditModal = false;
		editingNotification = null;
	}

	async function handleDelete(managementCode: string) {
		const confirmed = await ui.confirm({
			title: 'Delete Notification',
			message: 'Are you sure you want to delete this notification subscription?',
			confirmText: 'Delete',
			confirmVariant: 'danger'
		});
		
		if (!confirmed) return;

		try {
			await api.deleteNotification(managementCode);
			ui.success('Notification deleted');
			onDelete?.(managementCode);
		} catch (error) {
			if (error instanceof ApiClientError) {
				ui.showError('Failed to delete notification', error.description ?? error.code);
			}
		}
	}
</script>

<div class="notification-list">
	{#if loading}
		<LoadingSpinner message="Loading notifications..." />
	{:else if notifications.length === 0}
		<div class="text-center py-4 text-muted">
			<i class="fas fa-bell-slash fa-3x mb-3"></i>
			<p>No notification subscriptions found</p>
		</div>
	{:else}
		<!-- Header row -->
		<div class="notification-header d-flex align-items-center gap-3 px-3 py-2 border-bottom">
			<span class="header-type">Type</span>
			<span class="header-classes">Notification Classes</span>
			<span class="header-tags">Tags</span>
			<span class="header-tokens">Tokens</span>
			<span class="header-actions">Actions</span>
		</div>

		<!-- Notification items -->
		<div class="list-group list-group-flush">
			{#each notifications as notification}
				<NotificationItem
					{notification}
					{tokens}
					onEdit={() => openEditModal(notification)}
					onDelete={() => handleDelete(notification.management_code)}
				/>
			{/each}
		</div>
	{/if}
</div>

<!-- Edit Modal -->
<NotificationModal 
	bind:show={showEditModal}
	notification={editingNotification}
	on:saved={handleEditSaved}
	on:cancel={handleEditCancel}
/>

<style>
	.notification-header {
		font-weight: 500;
		font-size: 0.875rem;
		color: var(--bs-secondary);
		background: var(--bs-tertiary-bg, #f8f9fa);
	}

	.header-type {
		width: 65px;
		flex-shrink: 0;
	}

	.header-classes {
		width: 130px;
		flex-shrink: 0;
	}

	.header-tags {
		flex: 1;
		min-width: 0;
	}

	.header-tokens {
		width: 70px;
		flex-shrink: 0;
	}

	.header-actions {
		width: 70px;
		flex-shrink: 0;
	}
</style>
