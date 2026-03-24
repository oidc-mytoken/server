<script lang="ts">
	import type { Notification, Calendar, MytokenEntry } from '$lib/types';
	import { NOTIFICATION_CLASSES } from '$lib/types';
	import { api, ApiClientError } from '$lib/api/client';
	import { ui } from '$lib/stores/ui';
	import { discovery } from '$lib/stores/discovery';
	import TagPill from '../TagPill.svelte';
	import LoadingSpinner from '../LoadingSpinner.svelte';
	import NotificationItem from '../notifications/NotificationItem.svelte';

	// Props
	let {
		show = $bindable(false),
		token = null as MytokenEntry | null,
		onChanged = () => {}
	}: {
		show: boolean;
		token: MytokenEntry | null;
		onChanged?: () => void;
	} = $props();

	// Check if notifications are supported
	let notificationsSupported = $derived(!!$discovery.data?.notifications_endpoint);

	// Data
	let notifications: Notification[] = $state([]);
	let calendars: Calendar[] = $state([]);
	let loading = $state(false);
	let error = $state('');

	// For creating new notification
	let showNewNotificationForm = $state(false);
	let newNotificationClasses: string[] = $state([]);
	let includeChildrenNewNotif = $state(false);
	
	// For adding to existing notification
	let includeChildrenForAdd = $state(false);
	
	// For calendar actions
	let showAddToCalendarForm = $state(false);
	let showSendInviteForm = $state(false);
	let calendarComment = $state('');
	let selectedCalendarIds: string[] = $state([]);
	let inviteSent = $state(false);
	
	let saving = $state(false);

	// Directly subscribed notifications (can be removed)
	let directlySubscribedNotifications = $derived(
		notifications.filter(n => 
			!n.user_wide &&
			n.subscribed_tokens?.includes(token?.mom_id ?? '')
		)
	);

	// Tag-based subscriptions (read-only)
	let tagSubscribedNotifications = $derived(
		notifications.filter(n =>
			!n.user_wide &&
			!n.subscribed_tokens?.includes(token?.mom_id ?? '') &&
			token?.tags && n.tags?.some(nt => token.tags?.some(tt => tt.tag === nt.tag))
		)
	);

	// User-wide notifications (read-only)
	let userWideNotifications = $derived(
		notifications.filter(n => n.user_wide)
	);

	// Available to add (exclude user_wide and already subscribed via direct or tags)
	let availableNotifications = $derived(
		notifications.filter(n => 
			!n.user_wide &&
			!n.subscribed_tokens?.includes(token?.mom_id ?? '') &&
			!(token?.tags && n.tags?.some(nt => token.tags?.some(tt => tt.tag === nt.tag)))
		)
	);

	// Check if there are any subscribed notifications
	let hasSubscribedNotifications = $derived(
		directlySubscribedNotifications.length > 0 ||
		tagSubscribedNotifications.length > 0 ||
		userWideNotifications.length > 0
	);

	// Calendars the token is directly subscribed to (by mom_id)
	let directlySubscribedCalendars = $derived(
		calendars.filter(c => 
			c.subscribed_tokens?.includes(token?.mom_id ?? '')
		)
	);

	// Calendars subscribed via tags (not directly subscribed)
	let tagSubscribedCalendars = $derived(
		calendars.filter(c =>
			!c.subscribed_tokens?.includes(token?.mom_id ?? '') &&
			token?.tags && c.tags?.some(ct => token.tags?.some(tt => tt.tag === ct.tag))
		)
	);

	// Calendars this token can be added to (not already directly subscribed and not via tags)
	let availableCalendars = $derived(
		calendars.filter(c => 
			!c.subscribed_tokens?.includes(token?.mom_id ?? '') &&
			!(token?.tags && c.tags?.some(ct => token.tags?.some(tt => tt.tag === ct.tag)))
		)
	);

	$effect(() => {
		if (show && token) {
			loadData();
			// Reset forms
			showNewNotificationForm = false;
			showAddToCalendarForm = false;
			showSendInviteForm = false;
			newNotificationClasses = [];
			calendarComment = '';
			selectedCalendarIds = [];
			includeChildrenNewNotif = false;
			includeChildrenForAdd = false;
			inviteSent = false;
		}
	});

	async function loadData() {
		if (!notificationsSupported) return;
		
		loading = true;
		error = '';
		
		try {
			const [notifs, cals] = await Promise.all([
				api.getNotifications(),
				api.getCalendars()
			]);
			notifications = notifs;
			calendars = cals;
		} catch (err) {
			if (err instanceof ApiClientError) {
				error = err.description ?? err.code;
			} else {
				error = (err as Error).message;
			}
		} finally {
			loading = false;
		}
	}

	function close() {
		show = false;
	}

	async function addToNotification(notification: Notification) {
		if (!token) return;
		
		saving = true;
		try {
			await api.addTokenToNotification(
				notification.management_code,
				token.mom_id,
				includeChildrenForAdd
			);
			
			ui.success('Token added to notification');
			await loadData();
			onChanged();
		} catch (err) {
			if (err instanceof ApiClientError) {
				ui.showError('Failed to add to notification', err.description ?? err.code);
			}
		} finally {
			saving = false;
		}
	}

	async function removeFromNotification(notification: Notification) {
		if (!token) return;
		
		const confirmed = await ui.confirm({
			title: 'Remove from Notification',
			message: 'Remove this token from the notification?',
			confirmText: 'Remove',
			confirmVariant: 'danger'
		});
		
		if (!confirmed) return;
		
		saving = true;
		try {
			await api.removeTokenFromNotification(
				notification.management_code,
				token.mom_id
			);
			
			ui.success('Token removed from notification');
			await loadData();
			onChanged();
		} catch (err) {
			if (err instanceof ApiClientError) {
				ui.showError('Failed to remove from notification', err.description ?? err.code);
			}
		} finally {
			saving = false;
		}
	}

	async function createNewNotification() {
		if (!token || newNotificationClasses.length === 0) return;
		
		saving = true;
		try {
			await api.createNotification({
				notification_type: 'mail',
				notification_classes: newNotificationClasses,
				mom_id: token.mom_id,
				include_children: includeChildrenNewNotif
			});
			
			ui.success('Notification created');
			showNewNotificationForm = false;
			newNotificationClasses = [];
			includeChildrenNewNotif = false;
			await loadData();
			onChanged();
		} catch (err) {
			if (err instanceof ApiClientError) {
				ui.showError('Failed to create notification', err.description ?? err.code);
			}
		} finally {
			saving = false;
		}
	}

	async function addToCalendars() {
		if (!token || selectedCalendarIds.length === 0) return;
		
		saving = true;
		try {
			// Add to all selected calendars
			await Promise.all(
				selectedCalendarIds.map(calendarId =>
					api.addTokenToCalendar(
						calendarId,
						token.mom_id,
						calendarComment || undefined
					)
				)
			);
			
			const count = selectedCalendarIds.length;
			ui.success(`Token added to ${count} calendar${count > 1 ? 's' : ''}`);
			showAddToCalendarForm = false;
			selectedCalendarIds = [];
			calendarComment = '';
			await loadData();
			onChanged();
		} catch (err) {
			if (err instanceof ApiClientError) {
				ui.showError('Failed to add to calendar', err.description ?? err.code);
			}
		} finally {
			saving = false;
		}
	}

	function toggleCalendarSelection(calendarId: string) {
		if (selectedCalendarIds.includes(calendarId)) {
			selectedCalendarIds = selectedCalendarIds.filter(id => id !== calendarId);
		} else {
			selectedCalendarIds = [...selectedCalendarIds, calendarId];
		}
	}

	async function sendCalendarInvite() {
		if (!token) return;
		
		saving = true;
		try {
			await api.sendCalendarInvite(
				token.mom_id,
				calendarComment || undefined
			);
			
			inviteSent = true;
			ui.success('Calendar invitation sent');
		} catch (err) {
			if (err instanceof ApiClientError) {
				ui.showError('Failed to send invitation', err.description ?? err.code);
			}
		} finally {
			saving = false;
		}
	}

	function toggleNotificationClass(classId: string) {
		if (newNotificationClasses.includes(classId)) {
			newNotificationClasses = newNotificationClasses.filter(c => c !== classId);
		} else {
			newNotificationClasses = [...newNotificationClasses, classId];
		}
	}

	function copyToClipboard(text: string) {
		navigator.clipboard.writeText(text);
		ui.success('Copied to clipboard');
	}
</script>

{#if show}
	<div class="modal show d-block" tabindex="-1" role="dialog">
		<div class="modal-dialog modal-lg modal-dialog-centered modal-dialog-scrollable">
			<div class="modal-content">
				<div class="modal-header">
					<h5 class="modal-title">
						<i class="fas fa-bell me-2"></i>
						Notifications for "{token?.name || 'Unnamed Token'}"
					</h5>
					<button type="button" class="btn-close" onclick={close} aria-label="Close"></button>
				</div>
				
				<div class="modal-body">
					{#if !notificationsSupported}
						<div class="alert alert-warning">
							<i class="fas fa-exclamation-triangle me-2"></i>
							Notifications are not supported by this server.
						</div>
					{:else if loading}
						<LoadingSpinner message="Loading notifications..." />
					{:else if error}
						<div class="alert alert-danger">
							<i class="fas fa-exclamation-triangle me-2"></i>
							{error}
						</div>
					{:else}
						<!-- Email Notifications Section -->
						<div class="mb-4">
							<h6 class="border-bottom pb-2 mb-3">
								<i class="fas fa-envelope me-2"></i>
								Email Notifications
							</h6>
							
							<!-- User-wide notifications -->
							{#if userWideNotifications.length > 0}
								<div class="mb-3">
									<small class="text-muted d-block mb-2">Applies to all tokens:</small>
									<div class="list-group">
										{#each userWideNotifications as notification}
											<NotificationItem {notification} />
										{/each}
									</div>
								</div>
							{/if}

							<!-- Directly subscribed (can remove) -->
							{#if directlySubscribedNotifications.length > 0}
								<div class="mb-3">
									<small class="text-muted d-block mb-2">Directly subscribed:</small>
									<div class="list-group">
										{#each directlySubscribedNotifications as notification}
											<NotificationItem 
												{notification} 
												onRemove={() => removeFromNotification(notification)}
												loading={saving}
											/>
										{/each}
									</div>
								</div>
							{/if}

							<!-- Tag-based (read-only) -->
							{#if tagSubscribedNotifications.length > 0}
								<div class="mb-3">
									<small class="text-muted d-block mb-2">Subscribed via tags:</small>
									<div class="list-group">
										{#each tagSubscribedNotifications as notification}
											<NotificationItem {notification} />
										{/each}
									</div>
								</div>
							{/if}

							<!-- Empty state -->
							{#if !hasSubscribedNotifications}
								<p class="text-muted">No email notifications for this token.</p>
							{/if}

							<!-- Add to existing notification -->
							{#if availableNotifications.length > 0 && !showNewNotificationForm}
								<div class="mb-3">
									<small class="text-muted d-block mb-2">Add to existing notification:</small>
									
									<!-- Include children checkbox -->
									<div class="form-check mb-2">
										<input 
											type="checkbox" 
											class="form-check-input" 
											id="includeChildrenAdd"
											bind:checked={includeChildrenForAdd}
										/>
										<label class="form-check-label" for="includeChildrenAdd">
											Include child tokens
										</label>
									</div>
									
									<div class="list-group">
										{#each availableNotifications as notification}
											<NotificationItem 
												{notification} 
												onAdd={() => addToNotification(notification)}
												loading={saving}
											/>
										{/each}
									</div>
								</div>
							{/if}

							<!-- New notification form -->
							{#if showNewNotificationForm}
								<div class="card mb-3">
									<div class="card-body">
										<h6 class="card-title">Create New Notification</h6>
										<div class="mb-3">
											<span class="form-label small d-block">Notification Classes:</span>
											<div class="d-flex flex-wrap gap-2">
												{#each NOTIFICATION_CLASSES.filter(c => !c.parent) as cls}
													<button
														type="button"
														class="btn btn-sm"
														class:btn-primary={newNotificationClasses.includes(cls.id)}
														class:btn-outline-secondary={!newNotificationClasses.includes(cls.id)}
														onclick={() => toggleNotificationClass(cls.id)}
													>
														<i class="fas {cls.icon} me-1"></i>
														{cls.label}
													</button>
												{/each}
											</div>
										</div>
										<div class="form-check mb-3">
											<input
												type="checkbox"
												class="form-check-input"
												id="includeChildrenNewNotif"
												bind:checked={includeChildrenNewNotif}
											/>
											<label class="form-check-label" for="includeChildrenNewNotif">
												Include child tokens
											</label>
										</div>
										<div class="d-flex gap-2">
											<button
												type="button"
												class="btn btn-primary btn-sm"
												onclick={createNewNotification}
												disabled={saving || newNotificationClasses.length === 0}
											>
												{#if saving}
													<i class="fas fa-spinner fa-spin me-1"></i>
												{/if}
												Create
											</button>
											<button
												type="button"
												class="btn btn-secondary btn-sm"
												onclick={() => showNewNotificationForm = false}
											>
												Cancel
											</button>
										</div>
									</div>
								</div>
							{:else}
								<button
									type="button"
									class="btn btn-outline-primary btn-sm"
									onclick={() => showNewNotificationForm = true}
								>
									<i class="fas fa-plus me-1"></i>
									New Email Notification
								</button>
							{/if}
						</div>

						<!-- Calendars Section -->
						<div class="mb-4">
							<h6 class="border-bottom pb-2 mb-3">
								<i class="fas fa-calendar me-2"></i>
								Calendar Subscriptions
							</h6>
							
							<!-- Calendars directly subscribed -->
							{#if directlySubscribedCalendars.length > 0}
								<div class="mb-3">
									<small class="text-muted d-block mb-2">Directly subscribed:</small>
									<div class="list-group">
										{#each directlySubscribedCalendars as calendar}
											<div class="list-group-item">
												<div class="d-flex justify-content-between align-items-start">
													<div class="flex-grow-1">
														<div class="fw-medium">{calendar.description || 'Calendar'}</div>
														{#if calendar.tags && calendar.tags.length > 0}
															<small class="text-muted">
																Tags: 
																{#each calendar.tags as tag}
																	<TagPill name={tag.tag} color={tag.color} small />
																{/each}
															</small>
														{/if}
														{#if calendar.ics_url}
															<div class="mt-1">
																<small class="text-muted font-monospace text-break">{calendar.ics_url}</small>
															</div>
														{/if}
													</div>
													{#if calendar.ics_url}
														<button
															type="button"
															class="btn btn-sm btn-outline-secondary ms-2"
															onclick={() => copyToClipboard(calendar.ics_url ?? '')}
															title="Copy URL"
														>
															<i class="fas fa-copy"></i>
														</button>
													{/if}
												</div>
											</div>
										{/each}
									</div>
								</div>
							{/if}

							<!-- Calendars subscribed via tags -->
							{#if tagSubscribedCalendars.length > 0}
								<div class="mb-3">
									<small class="text-muted d-block mb-2">Subscribed via tags:</small>
									<div class="list-group">
										{#each tagSubscribedCalendars as calendar}
											<div class="list-group-item">
												<div class="d-flex justify-content-between align-items-start">
													<div class="flex-grow-1">
														<div class="fw-medium">{calendar.description || 'Calendar'}</div>
														{#if calendar.tags && calendar.tags.length > 0}
															<small class="text-muted">
																Tags: 
																{#each calendar.tags as tag}
																	<TagPill name={tag.tag} color={tag.color} small />
																{/each}
															</small>
														{/if}
														{#if calendar.ics_url}
															<div class="mt-1">
																<small class="text-muted font-monospace text-break">{calendar.ics_url}</small>
															</div>
														{/if}
													</div>
													{#if calendar.ics_url}
														<button
															type="button"
															class="btn btn-sm btn-outline-secondary ms-2"
															onclick={() => copyToClipboard(calendar.ics_url ?? '')}
															title="Copy URL"
														>
															<i class="fas fa-copy"></i>
														</button>
													{/if}
												</div>
											</div>
										{/each}
									</div>
								</div>
							{/if}

							{#if tagSubscribedCalendars.length === 0 && directlySubscribedCalendars.length === 0}
								<p class="text-muted mb-3">No calendar subscriptions for this token.</p>
							{/if}

							<!-- Add to existing calendar -->
							{#if showAddToCalendarForm}
								<div class="card mb-3">
									<div class="card-body">
										<h6 class="card-title">Add to Calendar</h6>
										<div class="mb-3">
											<small class="text-muted d-block mb-2">Select calendars:</small>
											<div class="list-group">
												{#each availableCalendars as calendar}
													<button
														type="button"
														class="list-group-item list-group-item-action"
														class:active={selectedCalendarIds.includes(calendar.id)}
														onclick={() => toggleCalendarSelection(calendar.id)}
													>
														<div class="d-flex justify-content-between align-items-start">
															<div class="flex-grow-1">
																<div class="fw-medium">{calendar.description || 'Calendar'}</div>
																{#if calendar.tags && calendar.tags.length > 0}
																	<small class:text-muted={!selectedCalendarIds.includes(calendar.id)}>
																		Tags: 
																		{#each calendar.tags as tag}
																			<TagPill name={tag.tag} color={tag.color} small />
																		{/each}
																	</small>
																{/if}
															</div>
															{#if selectedCalendarIds.includes(calendar.id)}
																<i class="fas fa-check"></i>
															{/if}
														</div>
													</button>
												{/each}
											</div>
										</div>
										{#if selectedCalendarIds.length > 0}
											<div class="mb-3">
												<label class="form-label small" for="calComment">Comment (optional):</label>
												<input
													type="text"
													class="form-control form-control-sm"
													id="calComment"
													bind:value={calendarComment}
													placeholder="Add a comment to the calendar entries"
												/>
											</div>
										{/if}
										<div class="d-flex gap-2">
											<button
												type="button"
												class="btn btn-primary btn-sm"
												onclick={addToCalendars}
												disabled={saving || selectedCalendarIds.length === 0}
											>
												{#if saving}
													<i class="fas fa-spinner fa-spin me-1"></i>
												{/if}
												Add to {selectedCalendarIds.length === 1 ? 'Calendar' : `${selectedCalendarIds.length} Calendars`}
											</button>
											<button
												type="button"
												class="btn btn-secondary btn-sm"
												onclick={() => { showAddToCalendarForm = false; selectedCalendarIds = []; calendarComment = ''; }}
											>
												Cancel
											</button>
										</div>
									</div>
								</div>
							{/if}

							<!-- Send calendar invite form -->
							{#if showSendInviteForm}
								<div class="card mb-3">
									<div class="card-body">
										{#if inviteSent}
											<div class="text-center text-success">
												<i class="fas fa-check-circle fa-2x mb-2"></i>
												<p>Calendar invitation sent successfully!</p>
												<button
													type="button"
													class="btn btn-secondary btn-sm"
													onclick={() => { showSendInviteForm = false; inviteSent = false; calendarComment = ''; }}
												>
													Close
												</button>
											</div>
										{:else}
											<h6 class="card-title">Send Calendar Invitation</h6>
											<p class="text-muted small">
												Send a single calendar invitation email with the token expiration date.
											</p>
											<div class="mb-3">
												<label class="form-label small" for="inviteComment">Comment (optional):</label>
												<input
													type="text"
													class="form-control form-control-sm"
													id="inviteComment"
													bind:value={calendarComment}
													placeholder="Add a comment to the invitation"
												/>
											</div>
											<div class="d-flex gap-2">
												<button
													type="button"
													class="btn btn-primary btn-sm"
													onclick={sendCalendarInvite}
													disabled={saving}
												>
													{#if saving}
														<i class="fas fa-spinner fa-spin me-1"></i>
													{/if}
													<i class="fas fa-envelope me-1"></i>
													Send Invitation
												</button>
												<button
													type="button"
													class="btn btn-secondary btn-sm"
													onclick={() => { showSendInviteForm = false; calendarComment = ''; }}
												>
													Cancel
												</button>
											</div>
										{/if}
									</div>
								</div>
							{/if}

							<!-- Action buttons -->
							{#if !showAddToCalendarForm && !showSendInviteForm}
								<div class="d-flex flex-wrap gap-2">
									{#if availableCalendars.length > 0}
										<button
											type="button"
											class="btn btn-outline-primary btn-sm"
											onclick={() => showAddToCalendarForm = true}
										>
											<i class="fas fa-calendar-plus me-1"></i>
											Add to Calendar
										</button>
									{/if}
									<button
										type="button"
										class="btn btn-outline-secondary btn-sm"
										onclick={() => showSendInviteForm = true}
									>
										<i class="fas fa-envelope me-1"></i>
										Send Invite
									</button>
								</div>
							{/if}
						</div>
					{/if}
				</div>
				
				<div class="modal-footer">
					<button type="button" class="btn btn-secondary" onclick={close}>
						Close
					</button>
				</div>
			</div>
		</div>
	</div>
	<div class="modal-backdrop show"></div>
{/if}

<style>
	.modal {
		background-color: rgba(0, 0, 0, 0.5);
	}
	
	.list-group-item {
		border-left: 3px solid transparent;
	}
	
	.list-group-item:hover {
		border-left-color: var(--bs-primary);
	}
</style>
