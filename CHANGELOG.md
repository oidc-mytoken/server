<!-- Template: -->
<!-- ### Features -->
<!--  -->
<!-- ### API -->
<!--  -->
<!-- ### Enhancements -->
<!--  -->
<!-- ### Bugfixes -->
<!--  -->
<!-- ### OpenID Provider -->
<!--  -->
<!-- ### Dependencies -->
<!--  -->


## mytoken 0.10.0

### Features

- Add support for notifications:
  - Allows to create email notifications for various things
  - Allows to calendar invites for token expirations
  - Allows to create calendars and add token expirations to it; the ics feed can be subscribed to
  - Allows to manage notifications on the web-interface
- Add "Enforceable Restrictions"
  - Depending on a user attribute different restriction templates can be
    enforced
- Add possibility to have an healthcheck endpoint

### Enhancements

- In the tokeninfo pane in the webinterface expired JWTs now get a more precise badge.
- Improved on returning json errors instead of html on api paths
- When not being logged in and no OP was selected now the 'Create new Mytoken' button in the webinterface is disabled.

### Bugfixes

- Fixed an issue with parallel access to refresh tokens if token rotation is used; this problem could for example
  occur with EGI-checkin.
- Fixed unwanted behavior: If a profile was used and changes to the mytoken
  spec would be made in the consent screen that would narrow it down, the
  profile would still be applied.
- Fixed problems with the caching implementation.

### Other

- Changed CORP settings for `/api` and `/static` as this lead to problems with oidc-agent.

### Dependencies

- Bump go version from 1.19 to 1.22
- Bump github.com/coreos/go-oidc/v3 from 3.9.0 to 3.11.0
- Bump github.com/gliderlabs/ssh from 0.3.6 to 0.3.7
- Bump github.com/go-resty/resty/v2 from 2.11.0 to 2.16.2
- Bump github.com/go-sql-driver/mysql from 1.8.0 to 1.8.1
- Bump github.com/gofiber/fiber/v2 from 2.52.2 to 2.52.5
- Bump github.com/gofiber/template/mustache/v2 from 2.0.9 to 2.0.12
- Bump github.com/jmoiron/sqlx from 1.3.5 to 1.4.0
- Bump github.com/lestrrat-go/jwx from 1.2.29 to 1.2.30
- Bump github.com/pires/go-proxyproto from 0.7.0 to 0.8.0
- Bump github.com/redis/go-redis/v9 from 9.5.1 to 9.7.0
- Bump github.com/valyala/fasthttp from 1.52.0 to 1.57.0
- Bump golang.org/x/crypto from 0.21.0 to 0.30.0
- Bump golang.org/x/mod from 0.16.0 to 0.22.0
- Bump golang.org/x/oauth2 from 0.18.0 to 0.24.0
- Bump golang.org/x/term from 0.18.0 to 0.27.0

## mytoken 0.9.2

### Packaging

- Fixed `mariadb-client` dependecy for `mytoken-server-migratedb` on rpm based distros

### Dependencies
- Bump github.com/go-sql-driver/mysql from 1.7.1 to 1.8.0
- Bump github.com/gofiber/fiber/v2 from to 2.52.0 2.52.2
- Bump github.com/gofiber/template/mustache/v2 from 2.0.8 to 2.0.9
- Bump github.com/lestrrat-go/jwx from 1.2.28 to 1.2.29
- Bump github.com/redis/go-redis/v9 from 9.4.0 to 9.5.1
- Bump golang.org/x/crypto from 0.19.0 to 0.21.0
- Bump golang.org/x/mod from 0.15.0 to 0.16.0
- Bump golang.org/x/oauth2 from 0.17.0 to 0.18.0
- Bump golang.org/x/term from 0.17.0 to 0.18.0
- 
## mytoken 0.9.1

### Enhancements

- Improve includes handling in the webinterface restrictions editor.

### Dependencies

- Bump golang.org/x/oauth2 from 0.15.0 to 0.17.0
- Bump golang.org/x/crypto from 0.17.0 to 0.19.0
- Bump golang.org/x/mod from 0.14.0 to 0.15.0
- Bump github.com/evanphx/json-patch/v5 from 5.7.0 to 5.9.0
- Bump github.com/gofiber/template/mustache/v2 from 2.0.7 to 2.0.8
- Bump github.com/lestrrat-go/jwx from 1.2.27 to 1.2.28
- Bump github.com/gofiber/fiber/v2 from 2.51.0 to 2.52.0
- Bump github.com/redis/go-redis/v9 from 9.3.1 to 9.4.0
- Bump github.com/valyala/fasthttp from 1.51.0 to 1.52.0
- Bump github.com/coreos/go-oidc/v3 from 3.8.0 to 3.9.0
- Bump github.com/gliderlabs/ssh from 0.3.5 to 0.3.6
- Bump github.com/go-resty/resty/v2 from 2.10.0 to 2.11.0
- Bump golang.org/x/term from 0.15.0 to 0.17.0

## mytoken 0.9.0

### Changes

- Changed the tokeninfo history api when used with a `mom_id`, now multiple `mom_ids` can be passed in a single
  request. Also, the response now contains the `mom_id` in the entry object.

### Features

- Added experimental support for OpenID Connect federations
- Added "Guest mode" to try mytoken out without using a real OP

### API

- Added `mom_id` parameter to tokeninfo introspection response
- Added `mom_id` parameter to mytoken responses

### Enhancements

- Webinterface: Improved the title / placeholder for the `hosts` restrictions key in the GUI editor to make it more
  clear that also subnets can be used.
- Webinterface: Changed the login provider selector and added search functionality
- Webinterface: Improved (re-)discovery of mytoken configuration
- Webinterface: Fixed a problem with scope discovery if there was no OP selected.
- Profiles: Improved / Fixed includes in especially restrictions when includes involve arrays.

### Bugfixes

- Finally fixed a problem with database times when the database was not set to UTC.
- Fixed a bug where sometimes a 'state mismatch' occured

### Dependencies

- Bump golang.org/x/mod from 0.11.0 to 0.14.0
- Bump golang.org/x/oauth2 from 0.9.0 to 0.15.0
- Bump golang.org/x/term from 0.9.0 to 0.15.0
- Bump golang.org/x/crypto from 0.10.0 to 0.16.0
- Bump golang.org/x/net from 0.14.0 to 0.17.0
- Bump github.com/valyala/fasthttp from 1.47.0 to 1.51.0
- Bump github.com/gofiber/fiber/v2 from 2.49.1 to 2.51.0
- Bump github.com/gofiber/template/mustache/v2 from 2.0.4 to 2.0.7
- Bump github.com/lestrrat-go/jwx from 1.2.26 to 1.2.27
- Bump github.com/redis/go-redis/v9 from 9.1.0 to 9.3.0
- Bump github.com/evanphx/json-patch/v5 from 5.6.0 to 5.7.0
- Bump github.com/go-resty/resty/v2 from 2.7.0 to 2.10.0
- Bump github.com/go-jose/go-jose/v3 from 3.0.0 to 3.0.1
- Bump github.com/coreos/go-oidc/v3 from 3.6.0 to 3.8.0

## mytoken 0.8.1

### Enhancements

- Improved returned transfercodes (do not include `l` and `I`)

### Bugfixes

- Fixed wrong (negative) `expires_at` time returned in tokeninfo for tokens without expiration
- Fixed response if token revocation call does not contain token

### Dependencies

- Bump github.com/sirupsen/logrus from 1.9.2 to 1.9.3
- Bump golang.org/x/term from 0.8.0 to 0.9.0
- Bump github.com/lestrrat-go/jwx from 1.2.25 to 1.2.26
- Bump golang.org/x/crypto from 0.9.0 to 0.10.0
- Bump golang.org/x/mod from 0.10.0 to 0.11.0
- Bump github.com/gofiber/template from 1.8.1 to 1.8.2
- Bump golang.org/x/oauth2 from 0.8.0 to 0.9.0
- Bump github.com/gofiber/fiber/v2 from 2.46.0 to 2.47.0

## mytoken 0.8.0

### Features

- Added support for RFC8707 for requesting audience restricted ATs

### Changes

- Default behavior for requesting audience restricted ATs is now according to RFC8707; the previous behavor can be
  configured with these options:
  ```yaml
  audience:
    use_rfc8707: false
    request_parameter: "audience"
    space_separate_auds: true
  ```

### API

- When creating a mytoken from a mytoken and it is returned as a transfer code the response now contains the
  `mom_id` of the created mytoken.

### Bugfixes

- Fixed a bug where wrong dates where returned if the database used a different timezone than UTC.
- Fixed a bug in `mytoken-migratedb` were empty databases could not be setup.

### Security Fixes

- Replaced the uuid library; the old library had a security flaw CVE-2021-3538

### Dependencies

- Bump golang.org/x/term from 0.5.0 to 0.8.0
- Bump github.com/valyala/fasthttp from 1.44.0 to 1.47.0
- Bump golang.org/x/net from 0.6.0 to 0.7.0
- Bump golang.org/x/crypto from 0.6.0 to 0.9.0
- Bump golang.org/x/oauth2 from 0.5.0 to 0.8.0
- Bump golang.org/x/mod from 0.8.0 to 0.9.0
- Bump github.com/gofiber/helmet/v2 from 2.2.24 to 2.2.25
- Bump github.com/gofiber/template from 1.7.5 to 1.8.0
- Bump github.com/gofiber/fiber/v2 from 2.42.0 to 2.46.0
- Bump github.com/pires/go-proxyproto from 0.6.2 to 0.7.0
- Bump github.com/go-sql-driver/mysql from 1.7.0 to 1.7.1
- Bump github.com/sirupsen/logrus from 1.9.0 to 1.9.2
- Bump github.com/coreos/go-oidc/v3 from 3.5.0 to 3.6.0
- Replaced github.com/satori/go.uuid with github.com/gofrs/uuid

## mytoken 0.7.2

### Bugfixes

- Fixed a bug in the webinterface where the metadata discovery was broken.

## mytoken 0.7.1

### Bugfixes

- Fixed a bug in the webinterface with the local storage that caused problems with outdated discovery information
- Fixed a bug in the webinterface where the `Expand` `Collapse` buttons (e.g. in the consent screen) got the wrong text.

## mytoken 0.7.0

### Features

- Webinterface has option to show event history for other mytokens in mytoken list.
- Webinterface has a new option in the tokeninfo pane to create a new mytoken with the same properties.
- Added server side `profiles` and `templates`
  - Can be used in the API, i.e. mytoken requests can include profiles, the capability, restrictions, and rotation
    claims can use templates
  - Can be used in the webinterface

### Enhancements

- Improved responsiveness of webinterface
- Expired mytokens are now greyed-out in webinterface mytoken list
- The database auto-cleanup now only removes mytokens expired more than a month ago.
  - This allows expired tokens to be shown in a mytoken list for extended periods.
  - This also allows to obtain history for expired tokens (by using a mytoken with the `manage_mytokens:list`
    capability) for a longer time.
  - Mytokens are still directly deleted when revoked.
- Requests from private IPs (e.g. from within the same entwork where the server is located) are now geolocated to
  the country where the server stands.
- The 'Create Mytoken' tab in the webitnerface now supports an `r` query parameter that takes a base64 encoded
  request from which the form is prefilled.
  - This allows 'create-a-mytoken-with-these-properties' links.

### API

- Added profile endpoint:
  - Any user can get list of groups
  - Any user can get profiles, and templates (capabilities, restrictions, rotation) for all the groups
  - Groups credentials are defined in the config file
    - With Basic authentication profiles and templates for the authenticated group can be created, updated, and deleted.
- Renamed `revocation_id` to `mom_id`
- Restructured capabilities related to other mytokens
- Added possibility to obtain history information for children and other tokens (capability)
- Added a name for OPs in the `supported_providers` of the mytoken configuration endpoint

### Bugfixes

- Fixed a bug where transfer codes could be used just like a short token (but only while the transfer code did not
  expire)

## mytoken 0.6.1

### API

- Changed the restriction `ip` key to `hosts`:
  - Backward compatibility is preserved. The legacy key `ip` is still accepted.
  - The `hosts` entry can contain:
    - Single ip address
    - Subnet address
    - Host name (with or without wildcard)
      - To compare against this, on request a reverse dns lookup is done for the request's ip address

### Enhancements

- Location restriction can now be done with host names, not only plain ip addresses, see above for more details.
- Webinterface: Added message to tokeninfo after MT creation and TC exchange to indicate that users must copy the
  mytoken to persist it.
- Improved code quality

### Bugfixes

- Fixed a bug in the web interface where the scope selection indicator for access tokens where not updated.

### Dependencies

- Bump go version to 1.19
- Bump golang.org/x/mod from 0.5.1 to 0.7.0
- Bump golang.org/x/crypto to 0.2.0
- Bump golang.org/x/term to 0.2.0
- Bump github.com/gofiber/fiber/v2 from 2.37.1 to 2.39.0
- Bump github.com/gofiber/helmet/v2 from 2.2.16 to 2.2.18

## mytoken 0.6.0

### API

- Dropped `subtoken_capabilities`, since the benefit was minimal, but made things more complex
  - Removed `subtoken_capabilities` from all API requests and responses
  - Removed `subtoken_capabilities` from the mytoken

### Enhancements

- Added introduction text in the web interface
- Session mytoken in web interface no longer uses `subtoken_capabilities` due to the drop, moved subtoken
  capabilities to the session mytoken as capabilities; added rotation on AT requests, added auto revocation

### Bugfixes

- Fixed a bug where mytokens with the `revoke_any_token` capabilities could revoke mytokens of other users if they
  can get possesion of the `revocation_id`
- Fixed problems in the web interface with restrictions / issuer selection when not logged in.

### Dependencies

- Bump github.com/coreos/go-oidc/v3 from 3.2.0 to 3.4.0
- Bump github.com/gofiber/template from 1.6.30 to 1.7.1
- Bump github.com/gofiber/fiber/v2 from 2.36.0 to 2.37.1
- Bump github.com/valyala/fasthttp from 1.39.0 to 1.40.0
- Bump github.com/gliderlabs/ssh from 0.3.4 to 0.3.5
- Bump github.com/gofiber/helmet/v2 from 2.2.15 to 2.2.16

## mytoken 0.5.4

### Bugfixes

- Fixed a bug in the webinterface where scope restrictions did not update correctly when not logged in and issuer
  changed

## mytoken 0.5.3

### Bugfixes

- Fixed a bug in the webinterface where mytokens could not be created when not logged-in

## mytoken 0.5.2

### Bugfixes

- Fixed a bug with requesting a consent screen for mytoken requests

## mytoken 0.5.1

### Enhancements

- In the tokeninfo - subtokens pane of the webinterface now only show the subtokens of the token in question,
  leaving out the actual token as their parent

### Bugfixes

- Fixed two bugs in the tokeninfo webinterface when introspecting mytokens issued by another server
- Fixed CORS of jwks

## mytoken 0.5.0

### Features

- Trusted web applications can skip the consent screen
- Reworked and improved major parts of the web interface:
  - Consent Screen:
    - On default a more compressed view is shown, where sections can be expanded if needed.
    - Displays the content of the `application_name` parameter if given.
    - Added possibility for clients to create a consent screen for mytoken-from-mytoken requests
  - Home Screen:
    - Replaced the tokeninfo pane with a new one
      - Removed tokeninfo about the session's mytoken
      - Added a tokeninfo pane to display tokeninfo for arbitrary mytokens (incl. introspection, history, subtokens)
      - Added possibility to create a transfer code
      - Moved the list of mytokens to a separate pane
      - Improved displaying the tree structure of mytokens
      - Reversed the token history order
    - Added "Exchange transfercode" pane, where a transfercode can be exchanged into a mytoken
    - Some parts can be used without being logged-in
  - Token Revocation:
    - Added possibility to revoke a mytoken in the tokeninfo pane
    - Added possibility to revoke listed tokens in the "My Mytokens" pane and in the "Tokeninfo - Subtokens" pane.
  - Capabilities:
    - Simplified the checking of capabilities
    - Read/Write capabilities are now not split but can be toggled
  - Create Mytoken:
    - After creation the mytoken is displayed in the tokeninfo pane, where it can be copied and of course
      information about the token is displayed
  - Settings:
    - Grant Types:
      - Include pages of different grant types in this view.
      - Grant Types can be expanded (collapsed on default).
      - Link to grant type page that was not clear enough is no longer needed.

### API

- Added `application_name` to mytoken requests.
- Added `token_type` to token introspection response.
- Added possibility to revoke tokens by `revocation_id`:
  - Added new `revoke_any_token` capability.
  - Added `revocation_id` parameter to responses that list tokens.

### Enhancements

- Admins can adapt the webinterface, i.e. for a custom style

### Bugfixes

- Fixed a bug in the mytoken webinterface where token introspection did not work on the settings page
- Fixed a bug in the mytoken webinterface restrictions editor, where audiences would always be set to zero when
  switching from the JSON editor to the GUI editor
- Fixed a bug where non-expiring mytokens would be revoked when database cleanup was enabled.
- Fixed a bug where the server could potentially crash

### Dependencies

- Bump github.com/valyala/fasthttp from 1.37.0 to 1.39.0
- Bump github.com/gofiber/fiber/v2 from 2.34.0 to 2.35.0
- Bump github.com/sirupsen/logrus from 1.8.1 to 1.9.0
- Bump github.com/gofiber/template from 1.6.28 to 1.6.30
- Bump github.com/gofiber/helmet/v2 from 2.2.13 to 2.2.15

## mytoken 0.4.3

### Bugfixes

- Fixed a bug where mytokens could not be used with x-www-form-urlencoding
- Fixed a bug where `x-www-form-urlencoding` was not accepted on token revocation endpoint

### Dependencies

- Bumped github.com/jmoiron/sqlx from 1.3.4 to 1.3.5
- Bumped github.com/lestrrat-go/jwx from 1.2.18 to 1.2.23
- Bumped github.com/gofiber/template from 1.6.22 to 1.6.27
- Bumped github.com/gofiber/helmet/v2 from 2.2.6 to 2.2.12
- Bumped github.com/pires/go-proxyproto from 0.6.1 to 0.6.2
- Bumped github.com/gofiber/fiber/v2 from 2.26.0 to 2.32.0
- Bumped github.com/valyala/fasthttp from 1.33.0 to 1.36.0

## mytoken 0.4.2

### Bugfixes

- Fixed a bug where the webinterface was not updated to use the renamed tokeninfo subtokens action

## mytoken 0.4.1

### API

- Changed tokeninfo subtokens action name
- Added the `tokeninfo` capability to the default capabilities of a mytoken

### Enhancements

- The `tokeninfo` capability is now checked by default when creating a mytoken
- Improved the output in the ssh protocol on bad requests

### Bugfixes

- Fixed tooltip text in webinterface on the book icon of read-only capabilities
- Fixed a bug where in the webinterface when creating a new mytoken the instructions to go to the consent screen, where
  still visible after the mytoken was obtained
- Fixed a bug where the consent screen stopped working after a timeout without displaying any error message
- Fixed a bug where 404 and other status codes where logged as errors

### Dependencies

- Bumped github.com/gofiber/fiber/v2 from 2.25.0 to 2.26.0
- Bumped github.com/gofiber/template from 1.6.21 to 1.6.22
- Bumped github.com/gofiber/helmet/v2 from 2.2.5 to 2.2.6

## mytoken 0.4.0

### Features

- Smart Logging: Only log up to a certain log level on default, but on error log everything
- Added User Settings endpoint
- Added possibility for user grants: grants that are not enabled on default, but can be enabled / disabled by a user
  and (might) require additional setup
- Added `ssh` user grant:
  - Can be enabled / disabled at the grants endpoint
  - SSH keys can be added, removed, listed at ssh grant endpoint
  - Added ssh keys can be used to obtain ATs, MTs, and other information (e.g. tokeninfo) through the ssh protocol at
    port `2222`
- Extended capabilities:
  - Some capabilities now have a "path" and "sub"-capabilities, e.g. (`tokeninfo` includes `tokeninfo:introspect`
    and more).
  - Some capabilities have a read only version, e.g. `read@settings`
  - Some capabilities have been renamed, e.g. (`tokeninfo_introspect` -> `tokeninfo:introspect`)

### API

- Changed default redirect type in auth code grant to `native`

### Mytoken

- Added `auth_time` to mytoken

### Enhancements

- Added request ids to response header and logging
- Refactored database; now using stored procedures which should ease database migration
- Moved automatic cleanup of expired database entries to the database
- Support symlinks when reading files

### Security Fixes

- Fixed a bug, where mytokens could be created from any mytoken not only from mytokens with the `create_mytoken`
  capability.

### Bugfixes

- Fixed a bug where restrictions did not behave correctly when multiple subnets were used
- Fixed response type on oidc errors on redirect in the authorization code flow
- Fixed `404` on api paths returning `html` instead of `json`

### Dependencies

- Updated various dependencies to the newest version

### Other

- Dropped the `mytoken-dbgc` tool, now moved to the database

## mytoken 0.3.3

### Mytoken

- Added the name of a mytoken to the JWT.

### API

- Don't redirect from `/.well-known/openid-configuration` to `/.well-known/mytoken-configuration`. Instead,
  returning the same content on both endpoints.

### Enhancements

- Removed buttons from webinterface in the tokeninfo tabs. The content now loads directly when switching the tab.
- Removed most need for CDNs; now self-hosting resources.
- Added setup of db database and db user to the setup utility.
- Made Link in the web interface on the create-mytoken page better visible.

### Bugfixes

- Fixed the error returned from the server if no capability for a mytoken was provided.
- Fixed PKCE code verifier length.
- Fixed Datetimepicker issues on consent page.
- Fixed response type if an (oidc) error occures on the redirect step of the authorization code flow.
- Fixed a bug where mytokens that are not yet valid could not be created

## mytoken 0.3.2

- Fixed password prompt for migratedb

## mytoken 0.3.1

- Improved helper tools

## mytoken 0.3.0

### Features

- Changes to the mytoken
  - Added a version to the mytoken token
  - Added token type 'mytoken'
  - Now using a hash value as the subject
- Added Dockerfiles; mytoken can easily run with swarm
- Added OIDC-compatibility for requesting ATs
  - ATs can be requested using the mytoken as the refresh token in a OIDC refresh flow
- Deployment Configuration
  - Added option to set maximum lifetime of mytokens
  - Added option to disable restriction keys 
  - Made request limits configurable
- Changed setup db to new db migration tool
- Added support for token rotation, incl. optional auto revocation
- Added option to set maximum token length when requesting a mytoken

### Webinterface
- Added option to create mytoken in the web interface
- Reworked consent screen
- Added possibility to set scopes and audiences when requesting an AT
- Improvements

### Enhancements
- Using better cryptographic functions
- Set cookie as secure if issuer uses https, indepent of a potential proxy
- Improved packaging
- Improved code base
- Improved error tracebility

### Bugfixes
- Fixed bugs in the webinterface
- Fixed other bugs

### OIDC
- Add PKCE support

### Dependencies
- Bumped several dependencies
