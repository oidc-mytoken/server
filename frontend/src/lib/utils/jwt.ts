/**
 * JWT decoding utilities for extracting claims from mytoken JWTs.
 * Note: This only decodes the JWT, it does NOT verify the signature.
 */

/**
 * Decode a JWT payload without verification.
 * Returns null if the token is not a valid JWT format.
 */
export function decodeJwtPayload(token: string): Record<string, unknown> | null {
	const parts = token.split('.');
	if (parts.length !== 3) return null;

	try {
		const payload = parts[1];
		// Handle base64url encoding: replace URL-safe chars and add padding
        const base64 = payload.replaceAll('-', '+').replaceAll('_', '/');
		const padded = base64.padEnd(base64.length + ((4 - (base64.length % 4)) % 4), '=');
		const decoded = atob(padded);
		return JSON.parse(decoded);
	} catch {
		return null;
	}
}

/**
 * Extract the OIDC issuer (oidc_iss) from a mytoken JWT.
 * Returns null if the token is not a JWT or doesn't contain oidc_iss.
 */
export function extractOidcIssuer(mytoken: string): string | null {
	const payload = decodeJwtPayload(mytoken);
	if (!payload) return null;
	return typeof payload.oidc_iss === 'string' ? payload.oidc_iss : null;
}
