package templating

// LayoutMain is the main layout
const LayoutMain = "layouts/main"

// Keys used in the mustache templates
const (
	MustacheKeyHome                         = "home"
	MustacheKeyCapabilities                 = "capabilities"
	MustacheKeyCheckedCapabilities          = "checked-capabilities"
	MustacheKeyEmptyNavbar                  = "empty-navbar"
	MustacheKeyLoggedIn                     = "logged-in"
	MustacheKeyRestrictionsGUI              = "restr-gui"
	MustacheKeyRestrictions                 = "restrictions"
	MustacheKeyRotation                     = "rotation"
	MustacheKeyCookieLifetime               = "cookie-lifetime"
	MustacheKeyPrefix                       = "prefix"
	MustacheKeyReadOnly                     = "read-only"
	MustacheKeyCollapse                     = "collapse"
	MustacheKeySettings                     = "settings"
	MustacheKeySettingsSSH                  = "settings-ssh"
	MustacheKeyGrants                       = "grants"
	MustacheKeyApplication                  = "application"
	MustacheKeyName                         = "name"
	MustacheKeyHomepage                     = "homepage"
	MustacheKeyContact                      = "contact"
	MustacheKeyPrivacyContact               = "privacy-contact"
	MustacheKeyConsent                      = "consent"
	MustacheKeyConsentSend                  = "consent-send"
	MustacheKeyTokenName                    = "token-name"
	MustacheKeyIss                          = "iss"
	MustacheKeySupportedScopes              = "supported_scopes"
	MustacheKeyInstanceURL                  = "instance-url"
	MustacheKeyCreateWithProfiles           = "create-with-profiles"
	MustacheKeyProfiles                     = "profiles"
	MustacheKeyProfilesProfile              = "profile"
	MustacheKeyProfilesCapabilities         = "capabilities"
	MustacheKeyProfilesRestrictions         = "restrictions"
	MustacheKeyProfilesRotation             = "rotation"
	MustacheKeyNotificationClasses          = "notification-classes"
	MustacheKeyCalendarsEditable            = "calendar-editable"
	MustacheKeySubscribeNotifications       = "subscribe-notifications"
	MustacheKeyNotificationsMailEnabled     = "notifications-mail-enabled"
	MustacheKeyNotificationsCalendarEnabled = "notifications-calendar-enabled"
)

// Keys for sub configs
const (
	MustacheSubTokeninfo            = "tokeninfo"
	MustacheSubCreateMT             = "create-mt"
	MustacheSubNotifications        = "notifications"
	MustacheSubNotificationListing  = "notification-listing"
	MustacheSubNewNotificationModal = "new-notification-modal"
	MustacheSubMTListing            = "mt-listing"
)
