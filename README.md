![mytoken logo](mytoken.png)

[![License](https://img.shields.io/github/license/oidc-mytoken/server.svg)](https://github.com/oidc-mytoken/server/blob/master/LICENSE)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/oidc-mytoken/server)
![GitHub Workflow Status](https://img.shields.io/github/workflow/status/oidc-mytoken/server/Go)
[![Go Report](https://goreportcard.com/badge/github.com/oidc-mytoken/server)](https://goreportcard.com/report/github.com/oidc-mytoken/server)
[![DeepSource](https://deepsource.io/gh/oidc-mytoken/server.svg/?label=active+issues&show_trend=true)](https://deepsource.io/gh/oidc-mytoken/server/?ref=repository-badge)
[![Release date](https://img.shields.io/github/release-date/oidc-mytoken/server.svg)](https://github.com/oidc-mytoken/server/releases/latest)
[![Release version](https://img.shields.io/github/release/oidc-mytoken/server.svg)](https://github.com/oidc-mytoken/server/releases/latest)

<!-- [![Code size](https://img.shields.io/github/languages/code-size/oidc-mytoken/server.svg)](https://github.com/oidc-mytoken/server/tree/master) -->

# mytoken

`mytoken` is a central web service with the goal to easily obtain OpenID Connect access tokens across devices.

A user can create a special string called `mytoken`. This mytoken then can be used to obtain OpenID Connect access
tokens from any device. The power of a mytoken can be restricted by the user, so they can create exactly the token they
need for a certain use case.

The mytoken command line client can be found at [https://github.com/oidc-mytoken/client](https://github.com/oidc-mytoken/client).

A go library for interacting with the mytoken server can be found at [https://github.com/oidc-mytoken/lib](https://github.com/oidc-mytoken/lib).

Documentation is available at [https://mytoken-docs.data.kit.edu/](https://mytoken-docs.data.kit.edu/).

A demo instance of mytoken is running at [https://mytoken.data.kit.edu/](https://mytoken.data.kit.edu/).
