![mytoken logo](https://git.scc.kit.edu/oidc/mytoken/-/raw/master/docs/img/mytoken.png)

[![License](https://img.shields.io/github/license/zachmann/mytoken.svg)](https://github.com/zachmann/mytoken/blob/master/LICENSE)
[![Release date](https://img.shields.io/github/release-date/zachmann/mytoken.svg)](https://github.com/zachmann/mytoken/releases/latest)
[![Release version](https://img.shields.io/github/release/zachmann/mytoken.svg)](https://github.com/zachmann/mytoken/releases/latest)

<!-- [![Code size](https://img.shields.io/github/languages/code-size/zachmann/mytoken.svg)](https://github.com/zachmann/mytoken/tree/master) -->

# mytoken

`mytoken` is a central web service with the goal to easily obtain OpenID Connect access tokens across devices.

A user can create a special string called `super token`. This super token then can be used to obtain OpenID Connect access tokens from any device.
The power of a super token can be restricted by the user, so he can create exactly the token he needs for a certain use case.

The mytoken command line client can be found at [https://github.com/zachmann/mytoken-client](https://github.com/zachmann/mytoken-client).

Documentation is available at [https://docs-sdm.scc.kit.edu/mytoken](https://docs-sdm.scc.kit.edu/mytoken) (currently no public access).

A demo instance of mytoken is running at [https://mytoken.data.kit.edu/](https://mytoken.data.kit.edu/).
