import { NOTIFICATION_CLASSES, type Notification } from '$lib/types';

/**
 * Get the icon class for a notification type
 */
export function getNotificationTypeIcon(type: string): string {
	return type === 'mail' ? 'fa-envelope' : 'fa-rss';
}

/**
 * Get the display label for a notification type
 */
export function getNotificationTypeLabel(type: string): string {
	return type === 'mail' ? 'Email' : 'WebSocket';
}

/**
 * Get the short label for a notification type
 */
export function getNotificationTypeShortLabel(type: string): string {
	return type === 'mail' ? 'Email' : 'WS';
}

/**
 * Check if a notification class is enabled for a notification
 * Returns 'full' if the class is directly enabled,
 * 'partial' if only child classes are enabled,
 * 'none' if the class is not enabled
 */
export function isClassEnabled(
	notification: Notification,
	classId: string
): 'full' | 'partial' | 'none' {
	if (notification.notification_classes.includes(classId)) {
		return 'full';
	}
	// Check for partial match (child class enabled)
	const hasChildEnabled = notification.notification_classes.some((c) =>
		c.startsWith(classId + ':')
	);
	if (hasChildEnabled) {
		return 'partial';
	}
	// Check if this is enabled via parent
	const parts = classId.split(':');
	for (let i = 1; i < parts.length; i++) {
		const parent = parts.slice(0, i).join(':');
		if (notification.notification_classes.includes(parent)) {
			return 'full';
		}
	}
	return 'none';
}

/**
 * Get the CSS color class for a notification class status
 */
export function getClassColor(status: 'full' | 'partial' | 'none'): string {
	switch (status) {
		case 'full':
			return 'text-success';
		case 'partial':
			return 'text-info';
		default:
			return 'text-muted';
	}
}

/**
 * Get root notification classes (classes without a parent)
 */
export function getRootNotificationClasses() {
	return NOTIFICATION_CLASSES.filter((c) => !c.parent);
}
