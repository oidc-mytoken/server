/**
 * Color utilities for tag color generation.
 *
 * The color generation algorithm matches the backend's MySQL implementation:
 * LPAD(HEX(CRC32(tag_name)), 6, '0')
 */

// CRC32 lookup table (standard polynomial 0xEDB88320)
const crc32Table = new Uint32Array(256);
for (let i = 0; i < 256; i++) {
    let c = i;
    for (let j = 0; j < 8; j++) {
        c = (c & 1) ? (0xEDB88320 ^ (c >>> 1)) : (c >>> 1);
    }
    crc32Table[i] = c;
}

/**
 * Calculate CRC32 hash of a string.
 * This matches MySQL's CRC32() function.
 */
export function crc32(str: string): number {
    let crc = 0xFFFFFFFF;
    for (let i = 0; i < str.length; i++) {
        crc = crc32Table[(crc ^ str.charCodeAt(i)) & 0xFF] ^ (crc >>> 8);
    }
    return (crc ^ 0xFFFFFFFF) >>> 0;
}

/**
 * Generate a color for a tag name using the same algorithm as the backend.
 * The algorithm calculates a CRC32 hash of the tag name and converts it to a hex color.
 *
 * @param tagName - The tag name to generate a color for
 * @returns A hex color string with # prefix (e.g., "#a1b2c3")
 */
export function generateTagColor(tagName: string): string {
    if (!tagName || !tagName.trim()) {
        return '#6c757d'; // Default gray for empty names
    }
    const hash = crc32(tagName.trim());
    // Convert to hex, take last 6 chars (in case of overflow), pad with zeros
    const hex = (hash & 0xFFFFFF).toString(16).padStart(6, '0');
    return `#${hex}`;
}

/**
 * Normalize a color value to ensure it has a # prefix and is a valid 6-digit hex color.
 * Handles colors from the API (which don't have #) and shorthand hex colors.
 *
 * @param color - The color to normalize (may or may not have # prefix)
 * @returns A normalized hex color with # prefix, or default gray if invalid
 */
export function normalizeColor(color: string | undefined | null): string {
    if (!color || !color.trim()) {
        return '#6c757d'; // Default gray
    }
    const trimmed = color.trim();
    // Check if it's a valid hex color (with or without #)
    const hexPattern = /^#?([0-9A-Fa-f]{3}|[0-9A-Fa-f]{6})$/;
    const match = trimmed.match(hexPattern);
    if (match) {
        // Ensure it has # prefix
        let hex = match[1];
        // Expand shorthand (e.g., "abc" -> "aabbcc")
        if (hex.length === 3) {
            hex = hex[0] + hex[0] + hex[1] + hex[1] + hex[2] + hex[2];
        }
        return `#${hex.toLowerCase()}`;
    }
    return '#6c757d'; // Default if not valid hex
}

/**
 * Strip the # prefix from a color for API calls.
 * The backend API expects colors without the # prefix.
 *
 * @param color - The color with # prefix
 * @returns The color without # prefix
 */
export function stripHashFromColor(color: string): string {
    return color.replace(/^#/, '');
}
