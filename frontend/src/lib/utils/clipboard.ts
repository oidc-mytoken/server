/**
 * Copy text to clipboard using the Clipboard API
 */
export async function copyToClipboard(text: string): Promise<boolean> {
	try {
        if (navigator.clipboard && globalThis.isSecureContext) {
			await navigator.clipboard.writeText(text);
			return true;
		}
        return false;
	} catch (err) {
		console.error('Failed to copy to clipboard:', err);
		return false;
	}
}

/**
 * Check if clipboard API is available
 */
export function isClipboardAvailable(): boolean {
    return !!(navigator.clipboard && globalThis.isSecureContext);
}
