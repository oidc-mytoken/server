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

## mytoken 0.3.3

### API

- Don't redirect from `/.well-known/openid-configuration` to `/.well-known/mytoken-configuration`. Instead returning the
  same content on both endpoints.

### Enhancements

- Removed buttons from webinterface in the tokeninfo tabs. The content now loads directly when switching the tab.

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