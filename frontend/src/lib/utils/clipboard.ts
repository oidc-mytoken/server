/**
 * Copy text to clipboard using the Clipboard API
 * Falls back to execCommand for older browsers
 */
export async function copyToClipboard(text: string): Promise<boolean> {
	try {
        if (navigator.clipboard && globalThis.isSecureContext) {
			await navigator.clipboard.writeText(text);
			return true;
		}

		// Fallback for older browsers
		const textArea = document.createElement('textarea');
		textArea.value = text;
		textArea.style.position = 'fixed';
		textArea.style.left = '-999999px';
		textArea.style.top = '-999999px';
		document.body.appendChild(textArea);
		textArea.focus();
		textArea.select();

		const success = document.execCommand('copy');
        textArea.remove();
		return success;
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
