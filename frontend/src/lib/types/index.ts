// Discovery response from /.well-known/mytoken-configuration
export interface DiscoveryDocument {
	issuer: string;
	mytoken_endpoint: string;
	token_endpoint?: string;
	access_token_endpoint: string;
	tokeninfo_endpoint: string;
	revocation_endpoint: string;
	token_transfer_endpoint: string;
	usersettings_endpoint: string;
	notifications_endpoint?: string;
	jwks_uri: string;
	providers_supported: Provider[];
	token_signing_alg_values_supported: string[];
	response_types_supported: string[];
	grant_types_supported: string[];
	capabilities_supported: string[];
	scopes_supported?: string[];
}

export interface Provider {
	issuer: string;
	name?: string;
	scopes_supported?: string[];
	audiences_supported?: string[];
    oidfed?: boolean;
}

// Capability types
export interface Capability {
	name: string;
	description?: string;
	children?: Capability[];
	hasReadOnlyOption?: boolean;  // true if capability has a read-only alternative (can toggle between modes)
	enabled?: boolean;
	readOnlyMode?: boolean;       // true if currently in read-only mode (only valid if hasReadOnlyOption is true)
	colorClass?: string;          // "text-success", "text-warning", "text-danger"
	capabilityLevel?: string;     // Description of power level
}

// Templates - individual presets for capabilities, restrictions, or rotation
export interface CapabilityTemplate {
	name: string;
	capabilities: string[];
}

// Restriction types
export interface Restriction {
	nbf?: number;
	exp?: number;
	scope?: string;
	audience?: string[];
	hosts?: string[];
	geoip_allow?: string[];
	geoip_disallow?: string[];
	usages_AT?: number;
	usages_other?: number;
}

// UsedRestriction extends Restriction with usage tracking
export interface UsedRestriction extends Restriction {
	usages_AT_done?: number;
	usages_other_done?: number;
}

export interface RestrictionTemplate {
	name: string;
	restrictions: Restriction[];
}

// Rotation types
export interface Rotation {
	on_AT?: boolean;
	on_other?: boolean;
	lifetime?: number;
	auto_revoke?: boolean;
}

export interface RotationTemplate {
	name: string;
	rotation: Rotation;
}

// Profile - top-level preset that combines capabilities, restrictions, rotation, and tags
export interface MytokenProfile {
	name: string;
	capabilities?: string[];
	restrictions?: Restriction[];
	rotation?: Rotation;
	tags?: string[];
}

// Generic profile/template type (from server API)
export interface ServerProfile {
	id?: string;
	name: string;
	payload: unknown; // JSON payload varies by profile type
}

// Token types
export interface Mytoken {
	mytoken: string;
	token_type: string;
	expires_in?: number;
	mytoken_type?: string;
	mom_id?: string;
	restrictions?: Restriction[];
	capabilities?: string[];
	rotation?: Rotation;
}

// Tag info attached to a mytoken
export interface MTTagInfo {
	tag: string;
	color: string;
	include_children?: boolean;
}

export interface MytokenEntry {
	mom_id: string;
	name?: string;
	created: number;
	expires_at?: number;
	ip?: string;
	user_agent?: string;
	tags?: MTTagInfo[];
	restrictions?: Restriction[];
	capabilities?: string[];
	rotation?: Rotation;
}

// Tree structure for token listing
export interface MytokenEntryTree {
	token: MytokenEntry;
	children?: MytokenEntryTree[];
}

export interface NetworkData {
	ip: string;
	user_agent?: string;
}

// Token info response - matches backend TokeninfoIntrospectResponse
export interface TokenInfoResponse {
	valid: boolean;
	token_type: string;
	mom_id: string;
	token: UsedMytokenInfo;
	tags?: MTTagInfo[];
}

// UsedMytokenInfo represents the token field in introspection response
export interface UsedMytokenInfo {
	ver: string;
	token_type: string;
	iss: string;
	sub: string;
	exp?: number;
	nbf: number;
	iat: number;
	auth_time?: number;
	jti: string;
	seq_no: number;
	name?: string;
	aud: string;
	oidc_sub: string;
	oidc_iss: string;
	restrictions?: UsedRestriction[];
	capabilities: string[];
	rotation?: Rotation;
}

// Event history
export interface EventHistoryEntry {
	time: number;
	event: string;
	comment?: string;
	ip?: string;
	user_agent?: string;
}

// Create mytoken request
export interface CreateMytokenRequest {
	grant_type: string;
	mytoken?: string;
	oidc_issuer?: string;
	oidc_flow?: string;
	redirect_type?: string;
	application_name?: string;
	restrictions?: Restriction[];
	capabilities?: string[];
	rotation?: Rotation;
	name?: string;
	response_type?: string;
	max_token_len?: number;
	tags?: CreateMytokenTag[];
}

// Create access token request
export interface CreateAccessTokenRequest {
	grant_type: string;
	mytoken?: string; // Optional - if not provided, cookie auth is used
	scope?: string;
	audience?: string;
}

// Transfer code
export interface TransferCodeRequest {
	grant_type: string;
	mytoken?: string; // Optional - if not provided, cookie auth is used
}

export interface TransferCodeResponse {
	transfer_code: string;
	expires_in: number;
}

// Tags
export interface Tag {
	tag: string;
	color?: string;
}

// Tag info for notifications/calendars (different from MTTagInfo)
export interface TagInfo {
	tag: string;
	color: string;
}

// Notifications
export interface Notification {
	notification_id?: string;
	management_code: string;
	notification_type: 'mail' | 'ws';
	notification_classes: string[];
	user_wide?: boolean;
	subscribed_tokens?: string[];
	tags?: TagInfo[];
	oidc_iss?: string;
    oidc_sub?: string;
}

export interface CreateNotificationRequest {
	notification_type: 'mail' | 'ws';
	notification_classes: string[];
	user_wide?: boolean;
	tags?: string[];
	mom_id?: string;
	include_children?: boolean;
}

export interface UpdateNotificationRequest {
	notification_classes?: string[];
	tags?: string[];
}

export interface Calendar {
	id: string;
	description?: string;
	ics_path?: string;
	ics_url?: string;
	tags?: TagInfo[];
	subscribed_tokens?: string[];
}

export interface CreateCalendarRequest {
	description?: string;
	tags?: string[];
	mom_id?: string;
	include_children?: boolean;
}

export interface UpdateCalendarRequest {
	description?: string;
	tags?: string[];
}

// Notification class definition type
export interface NotificationClass {
	id: string;
	label: string;
	icon: string;
	description: string;
	parent?: string;
}

// Notification class definitions
export const NOTIFICATION_CLASSES: NotificationClass[] = [
	{ id: 'AT_creations', label: 'Access Token Creations', icon: 'fa-key', description: 'When access tokens are created' },
	{
		id: 'rt_failure',
		label: 'Refresh Token Failures',
		icon: 'fa-unlink',
		description: 'When a refresh token could not be used'
	},
	{ id: 'subtoken_creations', label: 'Subtoken Creations', icon: 'fa-sitemap', description: 'When subtokens are created' },
	{ id: 'setting_changes', label: 'Setting Changes', icon: 'fa-cog', description: 'When settings are modified' },
	{ id: 'security', label: 'Security Events', icon: 'fa-shield-alt', description: 'All security-related events' },
	{ id: 'security:blocked_usages', label: 'Blocked Usages', icon: 'fa-hand-paper', description: 'When usage is blocked', parent: 'security' },
	{ id: 'security:blocked_usages:capabilities', label: 'Blocked by Capabilities', icon: 'fa-ban', description: 'Blocked due to insufficient capabilities', parent: 'security:blocked_usages' },
	{ id: 'security:blocked_usages:restrictions', label: 'Blocked by Restrictions', icon: 'fa-lock', description: 'Blocked due to restrictions', parent: 'security:blocked_usages' },
	{ id: 'security:revoked', label: 'Revoked Token Usage', icon: 'fa-times-circle', description: 'Attempted use of revoked token', parent: 'security' },
	{ id: 'security:ips', label: 'Unknown IP Addresses', icon: 'fa-network-wired', description: 'Usage from unknown IPs', parent: 'security' },
	{ id: 'expiration', label: 'Expiration Warnings', icon: 'fa-clock', description: 'Before tokens expire' }
];

// User settings
export interface EmailSettings {
    email_address?: string;
    email_verified?: boolean;
    prefer_html_mail?: boolean;
}

export interface Grant {
	type: string;
	enabled: boolean;
}

export interface SSHKey {
	name: string;
	key: string;
	restrictions?: Restriction[];
	capabilities?: string[];
}

// SSH key info from the server (read operations)
export interface SSHKeyInfo {
    name?: string;
    ssh_key?: string;
    ssh_key_fp?: string;
    created: number;
    last_used?: number;
}

// SSH info response from server
export interface SSHInfoResponse {
    grant_enabled: boolean;
    ssh_keys: SSHKeyInfo[];
}

// API error response
export interface ApiError {
	error: string;
	error_description?: string;
}

// Consent data from API
export interface ConsentData {
	oidc_issuer: string;
	application_name?: string;
	token_name?: string;
	capabilities: string[];
	restrictions?: Restriction[];
	rotation?: Rotation;
	tags?: CreateMytokenTag[];
	supported_scopes?: string[];
	all_capabilities: WebCapability[];
}

// Web capability details (from server)
export interface WebCapabilityDetails {
    name: string;
    description?: string;
    is_read_only?: boolean;
    color_class?: string;      // "text-success", "text-warning", "text-danger"
    capability_level?: string; // Description of power level
}

// Web capability for consent display
// read_write_capability/read_only_capability can be either:
// - A string (capability name like "AT", "tokeninfo")
// - An object with name, description, color_class, capability_level fields
export interface WebCapability {
    read_write_capability: string | WebCapabilityDetails;
    read_only_capability?: string | WebCapabilityDetails;
    children?: WebCapability[];
}

// Web capability for consent display
// read_write_capability/read_only_capability can be either:
// - A string (capability name like "AT", "tokeninfo")
// - An object with name, description, color_class, capability_level fields
export interface WebCapability {
    read_write_capability: string | WebCapabilityDetails;
    read_only_capability?: string | WebCapabilityDetails;
    children?: WebCapability[];
}

// Tag for creating mytoken
export interface CreateMytokenTag {
	tag: string;
	include_children?: boolean;
}

// Consent approval request
export interface ConsentApprovalRequest {
	oidc_issuer: string;
	restrictions?: Restriction[];
	capabilities: string[];
	name?: string;
	rotation?: Rotation;
	tags?: CreateMytokenTag[];
}

// Polling response for OIDC flow
export interface PollingResponse {
	status: 'pending' | 'completed' | 'failed';
	mytoken?: string;
	token_type?: string;
	expires_in?: number;
	error?: string;
	error_description?: string;
}

/**
 * Request data structure for pre-populating Create Mytoken form via URL parameter (?r=<base64>)
 * Used by token recreation and external tools
 */
export interface InitialMytokenRequest {
    name?: string;
    oidc_issuer?: string;
    response_type?: 'token' | 'short_token' | 'auto';
    capabilities?: string[];
    restrictions?: Restriction[];
    rotation?: Rotation;
    tags?: (string | { tag: string; include_children?: boolean })[];
}
