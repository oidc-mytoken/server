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
