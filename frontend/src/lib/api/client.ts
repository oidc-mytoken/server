import {get} from 'svelte/store';
import {dev} from '$app/environment';
import {discovery} from '$lib/stores/discovery';
import type {
	ApiError,
	Calendar,
	ConsentApprovalRequest,
	ConsentData,
	CreateAccessTokenRequest,
	CreateCalendarRequest,
	CreateMytokenRequest,
	CreateNotificationRequest,
	EmailSettings,
	EventHistoryEntry,
	Grant,
	Mytoken,
	MytokenEntry,
	MytokenEntryTree,
	Notification,
	PollingResponse,
	ServerProfile,
	SSHInfoResponse,
	SSHKeyInfo,
	TokenInfoResponse,
	TransferCodeRequest,
	TransferCodeResponse,
	WebCapability
} from '$lib/types';

class ApiClient {
	/**
	 * Request queue to serialize cookie-authenticated requests.
	 * This prevents race conditions when the session token has rotation enabled,
	 * where concurrent requests could cause "cipher: message authentication failed" errors
	 * because one request rotates the token while another is still in flight.
	 */
	private requestQueue: Promise<unknown> = Promise.resolve();

	/**
	 * Convert absolute URLs to relative URLs in development mode.
	 * This allows the Vite proxy to handle API requests properly.
	 */
	private toProxyUrl(url: string): string {
		if (!dev) return url;
		try {
			const parsed = new URL(url);
			// Return just the path + query string for the Vite proxy
			return parsed.pathname + parsed.search;
		} catch {
			// Already a relative URL
			return url;
		}
	}

	private getEndpoint(name: string): string {
		const $discovery = get(discovery);
		if (!$discovery.data) {
			throw new Error('Discovery document not loaded');
		}
		const endpoint = $discovery.data[name as keyof typeof $discovery.data];
		if (typeof endpoint !== 'string') {
            throw new TypeError(`Endpoint ${name} not found in discovery document`);
		}
		return this.toProxyUrl(endpoint);
	}

	/**
	 * Execute a request, optionally serializing it through the queue.
	 * Serialization is needed for cookie-authenticated requests to prevent
	 * rotation race conditions.
	 */
	private async request<T>(
		url: string,
		options: RequestInit = {},
		serialize: boolean = true
	): Promise<T> {
		const doRequest = async (): Promise<T> => {
			const response = await fetch(this.toProxyUrl(url), {
				...options,
				credentials: 'include', // Send cookies for authentication
				headers: {
					'Content-Type': 'application/json',
					...options.headers
				}
			});

			if (!response.ok) {
				let errorData: ApiError;
				try {
					errorData = await response.json();
				} catch {
					errorData = {
						error: 'request_failed',
						error_description: `Request failed with status ${response.status}`
					};
				}
				throw new ApiClientError(
					errorData.error,
					errorData.error_description,
					response.status
				);
			}

			// Handle empty responses
			const text = await response.text();
			if (!text) {
				return {} as T;
			}

			return JSON.parse(text);
		};

		if (serialize) {
			// Chain this request onto the queue to ensure serialization
			const result = this.requestQueue.then(
				() => doRequest(),
				() => doRequest() // Also execute after rejection to not block the queue
			);
			// Update the queue to wait for this request (but don't propagate errors to next request)
			this.requestQueue = result.catch(() => {});
			return result;
		} else {
			return doRequest();
		}
	}

	// ==================== Token Operations ====================

	/**
	 * Initiate login via OIDC flow (web client type)
	 * Returns consent_uri or authorization_uri to redirect to
	 * After OIDC flow completes, backend sets HttpOnly cookie and redirects to redirect_uri
	 */
	async login(issuer: string, cookieLifetime: number = 86400): Promise<{ authorization_uri?: string; consent_uri?: string }> {
		const endpoint = this.getEndpoint('mytoken_endpoint');

		const request = {
			grant_type: 'oidc_flow',
			oidc_flow: 'authorization_code',
			oidc_issuer: issuer,
			restrictions: [
				{
					exp: Math.floor(Date.now() / 1000) + cookieLifetime,
					ip: ['this']
				}
			],
			capabilities: [
				'tokeninfo',
				'AT',
				'settings',
				'list_mytokens',
				'manage_mytokens'
			],
			rotation: {
				on_AT: true,
				on_other: true,
				auto_revoke: true,
				lifetime: 86400
			},
			client_type: 'web',
			redirect_uri: '/',
			name: 'mytoken-web'
		};

		return this.request(endpoint, {
			method: 'POST',
			body: JSON.stringify(request)
		});
	}

	/**
	 * Check if logged in via cookie by calling introspect without a token
	 * Backend will use the HttpOnly cookie if present
	 * Returns token info if logged in, null otherwise
	 */
	async checkLogin(): Promise<TokenInfoResponse | null> {
		try {
			const endpoint = this.getEndpoint('tokeninfo_endpoint');
			return await this.request<TokenInfoResponse>(endpoint, {
				method: 'POST',
				body: JSON.stringify({
					action: 'introspect'
				})
			});
		} catch {
			return null;
		}
	}

	/**
	 * Create a new mytoken
	 */
	async createMytoken(request: CreateMytokenRequest): Promise<Mytoken> {
		const endpoint = this.getEndpoint('mytoken_endpoint');
		return this.request<Mytoken>(endpoint, {
			method: 'POST',
			body: JSON.stringify(request)
		});
	}

	/**
	 * Poll for mytoken creation (OIDC flow)
	 */
	async pollMytoken(pollingCode: string): Promise<PollingResponse> {
		const endpoint = this.getEndpoint('mytoken_endpoint');
		return this.request<PollingResponse>(endpoint, {
			method: 'POST',
			body: JSON.stringify({
				grant_type: 'polling_code',
				polling_code: pollingCode
			})
		});
	}

	/**
	 * Create an access token
	 */
	async createAccessToken(
		request: CreateAccessTokenRequest
	): Promise<{ access_token: string; token_type: string; expires_in?: number }> {
		const endpoint = this.getEndpoint('access_token_endpoint');
		return this.request(endpoint, {
			method: 'POST',
			body: JSON.stringify(request)
		});
	}

	/**
	 * Introspect a token
	 */
	async introspect(token: string): Promise<TokenInfoResponse> {
		const endpoint = this.getEndpoint('tokeninfo_endpoint');
		return this.request<TokenInfoResponse>(endpoint, {
			method: 'POST',
			body: JSON.stringify({
				action: 'introspect',
				mytoken: token
			})
		});
	}

	/**
	 * Get token event history by token string
	 */
	async getEventHistory(token: string): Promise<EventHistoryEntry[]> {
		const endpoint = this.getEndpoint('tokeninfo_endpoint');
		const response = await this.request<{ events: EventHistoryEntry[] }>(
			endpoint,
			{
				method: 'POST',
				body: JSON.stringify({
					action: 'event_history',
					mytoken: token
				})
			}
		);
		return response.events ?? [];
	}

	/**
	 * Get token event history by mom_id
	 */
	async getEventHistoryByMomId(momId: string): Promise<EventHistoryEntry[]> {
		const endpoint = this.getEndpoint('tokeninfo_endpoint');
		const response = await this.request<{ events: EventHistoryEntry[] }>(
			endpoint,
			{
				method: 'POST',
				body: JSON.stringify({
					action: 'event_history',
                    mom_ids: [momId]
				})
			}
		);
		return response.events ?? [];
	}

	/**
	 * Get subtokens as a tree structure
	 */
	async getSubtokens(token: string): Promise<MytokenEntryTree | null> {
		const endpoint = this.getEndpoint('tokeninfo_endpoint');
		const response = await this.request<{ mytokens: MytokenEntryTree }>(
			endpoint,
			{
				method: 'POST',
				body: JSON.stringify({
					action: 'subtokens',
					mytoken: token
				})
			}
		);
		return response.mytokens ?? null;
	}

	/**
	 * List all mytokens as a tree (uses cookie auth)
	 */
	async listMytokensTree(): Promise<MytokenEntryTree[]> {
		const endpoint = this.getEndpoint('tokeninfo_endpoint');
		const response = await this.request<{ mytokens: MytokenEntryTree[] }>(
			endpoint,
			{
				method: 'POST',
				body: JSON.stringify({
					action: 'list_mytokens'
				})
			}
		);
		return response.mytokens ?? [];
	}

	/**
	 * List all mytokens flattened (uses cookie auth)
	 */
	async listMytokens(): Promise<MytokenEntry[]> {
		const trees = await this.listMytokensTree();

		// Flatten the tree structure
		const flatten = (trees: MytokenEntryTree[]): MytokenEntry[] => {
			const result: MytokenEntry[] = [];
			for (const tree of trees) {
				if (tree.token) {
					result.push(tree.token);
				}
				if (tree.children && tree.children.length > 0) {
					result.push(...flatten(tree.children));
				}
			}
			return result;
		};

		return flatten(trees);
	}

	/**
	 * Revoke a token
	 */
	async revokeToken(
		token: string,
		recursive = false
	): Promise<void> {
		const endpoint = this.getEndpoint('revocation_endpoint');
		await this.request(endpoint, {
			method: 'POST',
			body: JSON.stringify({
				token,
				recursive
			})
		});
	}

	/**
	 * Revoke a token by its mom_id
	 */
	async revokeTokenByMomID(
		momId: string,
		recursive = false
	): Promise<void> {
		const endpoint = this.getEndpoint('revocation_endpoint');
		await this.request(endpoint, {
			method: 'POST',
			body: JSON.stringify({
				mom_id: momId,
				recursive
			})
		});
	}

	/**
	 * Revoke the current session (cookie-authenticated).
	 * Used for logout - revokes the token associated with the current session cookie.
	 */
	async revokeSession(recursive = true): Promise<void> {
		const endpoint = this.getEndpoint('revocation_endpoint');
		await this.request(endpoint, {
			method: 'POST',
			body: JSON.stringify({
				recursive
			})
		});
	}

	// ==================== Transfer Code ====================

	/**
	 * Create a transfer code
	 */
	async createTransferCode(
		request: TransferCodeRequest
	): Promise<TransferCodeResponse> {
		const endpoint = this.getEndpoint('token_transfer_endpoint');
		return this.request<TransferCodeResponse>(endpoint, {
			method: 'POST',
			body: JSON.stringify(request)
		});
	}

	/**
	 * Exchange a transfer code for a token
	 */
	async exchangeTransferCode(transferCode: string): Promise<Mytoken> {
		const endpoint = this.getEndpoint('mytoken_endpoint');
		return this.request<Mytoken>(endpoint, {
			method: 'POST',
			body: JSON.stringify({
				grant_type: 'transfer_code',
				transfer_code: transferCode
			})
		});
	}

	// ==================== User Settings ====================

	/**
	 * Get email settings
	 */
	async getEmailSettings(): Promise<EmailSettings> {
		const endpoint = this.getEndpoint('usersettings_endpoint');
		return this.request<EmailSettings>(`${endpoint}/email`);
	}

	/**
	 * Update email settings
	 */
    async updateEmailSettings(settings: { email_address?: string; prefer_html_mail?: boolean }): Promise<void> {
		const endpoint = this.getEndpoint('usersettings_endpoint');
		await this.request(`${endpoint}/email`, {
			method: 'PUT',
			body: JSON.stringify(settings)
		});
	}

	/**
	 * Get grants
	 */
	async getGrants(): Promise<Grant[]> {
		const endpoint = this.getEndpoint('usersettings_endpoint');
        // API returns { grant_types: [{ grant_type: string, enabled: boolean }] }
        const response = await this.request<{ grant_types: { grant_type: string; enabled: boolean }[] }>(
			`${endpoint}/grants`
		);
        // Map to frontend Grant type
        return (response.grant_types ?? []).map(g => ({
            type: g.grant_type,
            enabled: g.enabled
        }));
	}

	/**
	 * Enable a grant
	 */
	async enableGrant(grantType: string): Promise<void> {
		const endpoint = this.getEndpoint('usersettings_endpoint');
		await this.request(`${endpoint}/grants`, {
			method: 'POST',
			body: JSON.stringify({ grant_type: grantType })
		});
	}

	/**
	 * Disable a grant
	 */
	async disableGrant(grantType: string): Promise<void> {
		const endpoint = this.getEndpoint('usersettings_endpoint');
		await this.request(`${endpoint}/grants`, {
			method: 'DELETE',
			body: JSON.stringify({ grant_type: grantType })
		});
	}

	// ==================== SSH Keys ====================

	/**
     * Get SSH info including grant status and keys
	 */
    async getSSHInfo(): Promise<SSHInfoResponse> {
		const endpoint = this.getEndpoint('usersettings_endpoint');
        return this.request<SSHInfoResponse>(`${endpoint}/grants/ssh`);
    }

    /**
     * Get SSH keys (convenience method)
     */
    async getSSHKeys(): Promise<SSHKeyInfo[]> {
        const info = await this.getSSHInfo();
        return info.ssh_keys ?? [];
	}

	/**
	 * Add an SSH key
     * Returns authorization URL for OIDC flow completion
	 */
    async addSSHKey(key: {
        name?: string;
        ssh_key: string;
        restrictions?: unknown[];
        capabilities?: string[]
    }): Promise<{ consent_uri?: string; polling_code?: string; interval?: number }> {
		const endpoint = this.getEndpoint('usersettings_endpoint');
        return this.request(`${endpoint}/grants/ssh`, {
			method: 'POST',
            body: JSON.stringify({
                grant_type: 'mytoken',
                application_name: 'mytoken webinterface',
                ...key
            })
		});
	}

	/**
     * Poll for SSH key addition completion
	 */
    async pollSSHKey(pollingCode: string): Promise<{ status?: string; ssh_user?: string; ssh_host_config?: string }> {
        const endpoint = this.getEndpoint('usersettings_endpoint');
        return this.request(`${endpoint}/grants/ssh`, {
            method: 'POST',
            body: JSON.stringify({
                grant_type: 'polling_code',
                polling_code: pollingCode
            })
        });
    }

    /**
     * Delete an SSH key by fingerprint
     */
    async deleteSSHKey(sshKeyFP: string): Promise<void> {
		const endpoint = this.getEndpoint('usersettings_endpoint');
		await this.request(`${endpoint}/grants/ssh`, {
			method: 'DELETE',
            body: JSON.stringify({ssh_key_fp: sshKeyFP})
		});
	}

	// ==================== Notifications ====================

	/**
	 * Get notifications
	 */
	async getNotifications(): Promise<Notification[]> {
		const endpoint = this.getEndpoint('notifications_endpoint');
		const response = await this.request<{
			notifications: Notification[];
		}>(endpoint);
		return response.notifications ?? [];
	}

	/**
	 * Create a notification
	 */
	async createNotification(request: CreateNotificationRequest): Promise<Notification> {
		const endpoint = this.getEndpoint('notifications_endpoint');
		return this.request<Notification>(endpoint, {
			method: 'POST',
			body: JSON.stringify(request)
		});
	}

    /**
     * Get a specific notification by management code
     */
    async getNotificationByManagementCode(managementCode: string): Promise<Notification> {
        const endpoint = this.getEndpoint('notifications_endpoint');
        return this.request<Notification>(`${endpoint}/${managementCode}`);
    }

	/**
	 * Delete a notification
	 */
	async deleteNotification(managementCode: string): Promise<void> {
		const endpoint = this.getEndpoint('notifications_endpoint');
		await this.request(`${endpoint}/${managementCode}`, {
			method: 'DELETE'
		});
	}

	/**
	 * Update a notification (classes and/or tags)
	 */
	async updateNotification(
		managementCode: string,
		update: { notification_classes?: string[]; tags?: string[] }
	): Promise<void> {
		const endpoint = this.getEndpoint('notifications_endpoint');
		await this.request(`${endpoint}/${managementCode}/nc`, {
			method: 'PUT',
			body: JSON.stringify(update)
		});
	}

	/**
	 * Add a token to an existing notification
	 */
	async addTokenToNotification(
		managementCode: string,
		momId: string,
		includeChildren?: boolean
	): Promise<void> {
		const endpoint = this.getEndpoint('notifications_endpoint');
		await this.request(`${endpoint}/${managementCode}/token`, {
			method: 'POST',
			body: JSON.stringify({
				mom_id: momId,
				include_children: includeChildren ?? false
			})
		});
	}

	/**
	 * Remove a token from a notification
	 */
	async removeTokenFromNotification(
		managementCode: string,
		momId: string
	): Promise<void> {
		const endpoint = this.getEndpoint('notifications_endpoint');
		await this.request(`${endpoint}/${managementCode}/token`, {
			method: 'DELETE',
			body: JSON.stringify({
				mom_id: momId
			})
		});
	}

	// ==================== Calendars ====================

	/**
	 * Get calendars
	 */
	async getCalendars(): Promise<Calendar[]> {
		const endpoint = this.getEndpoint('notifications_endpoint');
		const response = await this.request<{ calendars: Calendar[] }>(
			`${endpoint}/calendars`
		);
		return response.calendars ?? [];
	}

	/**
	 * Create a calendar
	 */
	async createCalendar(
		request: CreateCalendarRequest
	): Promise<void> {
		const endpoint = this.getEndpoint('notifications_endpoint');
		await this.request(`${endpoint}/calendars`, {
			method: 'POST',
			body: JSON.stringify(request)
		});
	}

	/**
	 * Delete a calendar
	 */
	async deleteCalendar(calendarId: string): Promise<void> {
		const endpoint = this.getEndpoint('notifications_endpoint');
		await this.request(`${endpoint}/calendars/${calendarId}`, {
			method: 'DELETE'
		});
	}

	/**
	 * Update a calendar (description and/or tags)
	 */
	async updateCalendar(
		calendarId: string,
		update: { description?: string; tags?: string[] }
	): Promise<Calendar> {
		const endpoint = this.getEndpoint('notifications_endpoint');
		return this.request<Calendar>(`${endpoint}/calendars/${calendarId}`, {
			method: 'PUT',
			body: JSON.stringify(update)
		});
	}

	/**
	 * Add a token to an existing calendar
	 */
	async addTokenToCalendar(
		calendarId: string,
		momId: string,
		comment?: string
	): Promise<void> {
		const endpoint = this.getEndpoint('notifications_endpoint');
		await this.request(`${endpoint}/calendars/${calendarId}`, {
			method: 'POST',
			body: JSON.stringify({
				mom_id: momId,
				comment
			})
		});
	}

	/**
	 * Send a single calendar invitation email for a token
	 */
	async sendCalendarInvite(momId: string, comment?: string): Promise<void> {
		const endpoint = this.getEndpoint('notifications_endpoint');
		await this.request(endpoint, {
			method: 'POST',
			body: JSON.stringify({
				notification_type: 'ics_invite',
				mom_id: momId,
				comment
			})
		});
	}

	// ==================== Tags ====================

	/**
	 * Add a tag to a mytoken by mom_id
	 */
	async addTagToMytoken(momId: string, tag: string): Promise<void> {
		const endpoint = this.getEndpoint('mytoken_endpoint');
		await this.request(`${endpoint}/tags`, {
			method: 'POST',
			body: JSON.stringify({
				mom_id: momId,
				tag
			})
		});
	}

	/**
	 * Remove a tag from a mytoken by mom_id
	 */
	async removeTagFromMytoken(momId: string, tag: string): Promise<void> {
		const endpoint = this.getEndpoint('mytoken_endpoint');
		await this.request(`${endpoint}/tags`, {
			method: 'DELETE',
			body: JSON.stringify({
				mom_id: momId,
				tag
			})
		});
	}

	// ==================== Consent Operations ====================

	/**
	 * Get consent data for a consent code
	 */
	async getConsentData(consentCode: string): Promise<ConsentData> {
		// Use the API endpoint for consent data
		const url = `/api/v0/consent/${consentCode}`;
		return this.request<ConsentData>(url, {
			method: 'GET'
		});
	}

	/**
	 * Approve consent
	 */
	async approveConsent(consentCode: string, request: ConsentApprovalRequest): Promise<{ authorization_uri: string }> {
		const url = `/api/v0/consent/${consentCode}`;
		return this.request<{ authorization_uri: string }>(url, {
			method: 'POST',
			body: JSON.stringify(request)
		});
	}

	/**
	 * Decline consent
	 */
	async declineConsent(consentCode: string): Promise<{ url: string }> {
		const url = `/api/v0/consent/${consentCode}`;
		return this.request<{ url: string }>(url, {
			method: 'POST',
			body: '' // Empty body signals decline
		});
	}

	// ==================== Capabilities ====================

	/**
	 * Get all capabilities with full structure (tree, descriptions, read-only variants)
	 */
	async getAllCapabilities(): Promise<WebCapability[]> {
		const url = `/api/v0/capabilities`;
		return this.request<WebCapability[]>(url);
	}

	// ==================== Profiles/Templates ====================

	/**
	 * Get profile groups
	 */
	async getProfileGroups(): Promise<string[]> {
		const endpoint = this.getEndpoint('profiles_endpoint');
		return this.request<string[]>(endpoint);
	}

	/**
	 * Get full profiles (combines capabilities, restrictions, rotation, tags)
	 */
	async getProfiles(group?: string): Promise<ServerProfile[]> {
		const endpoint = this.getEndpoint('profiles_endpoint');
		const url = group ? `${endpoint}/${group}/profiles` : `${endpoint}/profiles`;
		return this.request<ServerProfile[]>(url);
	}

	/**
	 * Get capability templates
	 */
	async getCapabilityTemplates(group?: string): Promise<ServerProfile[]> {
		const endpoint = this.getEndpoint('profiles_endpoint');
		const url = group ? `${endpoint}/${group}/capabilities` : `${endpoint}/capabilities`;
		return this.request<ServerProfile[]>(url);
	}

	/**
	 * Get restriction templates
	 */
	async getRestrictionTemplates(group?: string): Promise<ServerProfile[]> {
		const endpoint = this.getEndpoint('profiles_endpoint');
		const url = group ? `${endpoint}/${group}/restrictions` : `${endpoint}/restrictions`;
		return this.request<ServerProfile[]>(url);
	}

	/**
	 * Get rotation templates
	 */
	async getRotationTemplates(group?: string): Promise<ServerProfile[]> {
		const endpoint = this.getEndpoint('profiles_endpoint');
		const url = group ? `${endpoint}/${group}/rotation` : `${endpoint}/rotation`;
		return this.request<ServerProfile[]>(url);
	}
}

/**
 * Custom error class for API errors
 */
export class ApiClientError extends Error {
	constructor(
		public code: string,
		public description: string | undefined,
		public status: number
	) {
		super(description ?? code);
		this.name = 'ApiClientError';
	}

	/**
	 * Check if this error indicates an authentication failure
	 * This includes session expiration, invalid cookies, cipher errors, etc.
	 */
	isAuthError(): boolean {
		// HTTP 401 Unauthorized
		if (this.status === 401) return true;
		
		// Common auth-related error codes
		const authErrorCodes = [
			'invalid_token',
			'invalid_grant',
			'access_denied',
			'unauthorized'
		];
		if (authErrorCodes.includes(this.code?.toLowerCase())) return true;
		
		// Check for cipher/authentication errors in the description
		const desc = this.description?.toLowerCase() ?? '';
		if (desc.includes('cipher') || 
		    desc.includes('authentication failed') ||
		    desc.includes('session') ||
		    desc.includes('expired') ||
		    desc.includes('invalid cookie')) {
			return true;
		}
		
		return false;
	}
}

// Export singleton instance
export const api = new ApiClient();
