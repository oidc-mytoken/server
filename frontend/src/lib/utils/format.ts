/**
 * Format a Unix timestamp to a human-readable date string
 */
export function formatDate(timestamp: number): string {
	const date = new Date(timestamp * 1000);
	return date.toLocaleDateString(undefined, {
		year: 'numeric',
		month: 'short',
		day: 'numeric'
	});
}

/**
 * Format a Unix timestamp to a human-readable datetime string
 */
export function formatDateTime(timestamp: number): string {
	const date = new Date(timestamp * 1000);
	return date.toLocaleString(undefined, {
		year: 'numeric',
		month: 'short',
		day: 'numeric',
		hour: '2-digit',
		minute: '2-digit'
	});
}

/**
 * Format a duration in seconds to a human-readable string
 */
export function formatDuration(seconds: number): string {
	if (seconds < 60) {
        return `${seconds} second${seconds === 1 ? '' : 's'}`;
	}

	const minutes = Math.floor(seconds / 60);
	if (minutes < 60) {
        return `${minutes} minute${minutes === 1 ? '' : 's'}`;
	}

	const hours = Math.floor(minutes / 60);
	if (hours < 24) {
		const remainingMinutes = minutes % 60;
		if (remainingMinutes === 0) {
            return `${hours} hour${hours === 1 ? '' : 's'}`;
		}
		return `${hours}h ${remainingMinutes}m`;
	}

	const days = Math.floor(hours / 24);
	const remainingHours = hours % 24;
	if (remainingHours === 0) {
        return `${days} day${days === 1 ? '' : 's'}`;
	}
	return `${days}d ${remainingHours}h`;
}

/**
 * Format relative time (e.g., "2 hours ago", "in 3 days")
 */
export function formatRelativeTime(timestamp: number): string {
	const now = Date.now() / 1000;
	const diff = timestamp - now;
	const absDiff = Math.abs(diff);

	const units: [number, string][] = [
		[60, 'second'],
		[3600, 'minute'],
		[86400, 'hour'],
		[604800, 'day'],
		[2592000, 'week'],
		[31536000, 'month'],
		[Infinity, 'year']
	];

    let value = 0;
    let unit = 'second';

	for (let i = 0; i < units.length; i++) {
		const [threshold, unitName] = units[i];
		if (absDiff < threshold) {
			const prevThreshold = i > 0 ? units[i - 1][0] : 1;
			value = Math.floor(absDiff / prevThreshold);
			unit = unitName;
			break;
		}
	}

    const suffix = value === 1 ? '' : 's';
	if (diff < 0) {
        return `${value} ${unit}${suffix} ago`;
	}
    return `in ${value} ${unit}${suffix}`;
}

/**
 * Truncate a string with ellipsis
 */
export function truncate(str: string, maxLength: number): string {
	if (str.length <= maxLength) {
		return str;
	}
	return str.slice(0, maxLength - 3) + '...';
}

/**
 * Format a token/ID for display (show first and last N characters)
 */
export function formatTokenPreview(token: string, chars: number = 8): string {
	if (token.length <= chars * 2 + 3) {
		return token;
	}
	return `${token.slice(0, chars)}...${token.slice(-chars)}`;
}
