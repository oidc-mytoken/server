<script lang="ts">
	import type { Notification } from '$lib/types';
	import {
		getNotificationTypeIcon,
		getNotificationTypeShortLabel,
		isClassEnabled,
		getClassColor,
		getRootNotificationClasses
	} from '$lib/utils/notifications';
	import TagPill from '../TagPill.svelte';

	interface Props {
		notification: Notification;
		onAdd?: () => void;
		onRemove?: () => void;
		onEdit?: () => void;
		onDelete?: () => void;
		loading?: boolean;
		disabled?: boolean;
	}

	let {
		notification,
		onAdd,
		onRemove,
		onEdit,
		onDelete,
		loading = false,
		disabled = false
	}: Props = $props();

	const rootClasses = getRootNotificationClasses();

	function getTokenCount(): number {
		return notification.subscribed_tokens?.length ?? notification.mom_ids?.length ?? 0;
	}
</script>

<div class="notification-item list-group-item d-flex align-items-center gap-3">
	<!-- Type Badge -->
	<span class="badge bg-secondary type-badge" title={notification.notification_type === 'mail' ? 'Email' : 'WebSocket'}>
		<i class="fas {getNotificationTypeIcon(notification.notification_type)} me-1"></i>
		{getNotificationTypeShortLabel(notification.notification_type)}
	</span>

	<!-- Notification Classes -->
	<div class="notification-classes">
		{#each rootClasses as cls}
			{@const status = isClassEnabled(notification, cls.id)}
			<span
				class="me-2 {getClassColor(status)}"
				title="{cls.label}: {status === 'full' ? 'Enabled' : status === 'partial' ? 'Partially enabled' : 'Disabled'}"
			>
				<i class="fas {cls.icon}"></i>
			</span>
		{/each}
	</div>

	<!-- Tags -->
	<div class="tags">
		{#if notification.tags && notification.tags.length > 0}
			{#each notification.tags as tag}
				<TagPill name={tag.tag} color={tag.color} small />
			{/each}
		{:else}
			<span class="text-muted">-</span>
		{/if}
	</div>

	<!-- Tokens -->
	<div class="tokens">
		{#if notification.user_wide}
			<span class="badge bg-primary" title="Applies to all tokens">
				<i class="fas fa-user me-1"></i>
				All
			</span>
		{:else}
			<span class="badge bg-info">
				{getTokenCount()} tokens
			</span>
		{/if}
	</div>

	<!-- Actions -->
	<div class="actions d-flex gap-1">
		{#if onAdd}
			<button
				type="button"
				class="btn btn-sm btn-outline-success"
				title="Add token to notification"
				onclick={onAdd}
				disabled={loading || disabled}
			>
				{#if loading}
					<i class="fas fa-spinner fa-spin"></i>
				{:else}
					<i class="fas fa-plus"></i>
				{/if}
			</button>
		{/if}
		{#if onRemove}
			<button
				type="button"
				class="btn btn-sm btn-outline-danger"
				title="Remove token from notification"
				onclick={onRemove}
				disabled={loading || disabled}
			>
				{#if loading}
					<i class="fas fa-spinner fa-spin"></i>
				{:else}
					<i class="fas fa-times"></i>
				{/if}
			</button>
		{/if}
		{#if onEdit}
			<button
				type="button"
				class="btn btn-sm btn-outline-secondary"
				title="Edit"
				onclick={onEdit}
				disabled={loading || disabled}
			>
				<i class="fas fa-pencil-alt"></i>
			</button>
		{/if}
		{#if onDelete}
			<button
				type="button"
				class="btn btn-sm btn-outline-danger"
				title="Delete"
				onclick={onDelete}
				disabled={loading || disabled}
			>
				<i class="fas fa-trash"></i>
			</button>
		{/if}
	</div>
</div>

<style>
	.notification-item {
		padding: 0.5rem 0.75rem;
	}

	.type-badge {
		width: 65px;
		flex-shrink: 0;
		text-align: center;
	}

	.notification-classes {
		width: 130px;
		flex-shrink: 0;
		font-size: 1.1em;
	}

	.tags {
		flex: 1;
		min-width: 0;
	}

	.tokens {
		width: 70px;
		flex-shrink: 0;
	}

	.actions {
		width: 70px;
		flex-shrink: 0;
	}
</style>
