function checkIfLoggedIn(...next) {
    chainFunctions(
        discovery,
        _checkIfLoggedIn,
        ...next,
    );
}

function _checkIfLoggedIn(...next) {
    let data = {
        'action': 'introspect'
    };
    data = JSON.stringify(data);
    $.ajax({
        type: "POST",
        url: storageGet('tokeninfo_endpoint'),
        data: data,
        success: function (res) {
            let token = res['token'];
            let iss = token['oidc_iss'];
            if (iss) {
                storageSet('oidc_issuer', iss);
            }
            let scopes = extractMaxScopesFromToken(token);
            storageSet('token_scopes', scopes);
            if (window.location.pathname === "/") {
                window.location.href = "/home" + window.location.search + window.location.hash;
            }
            doNext(...next);
        },
        error: function (res) {
            if (window.location.pathname !== "/") {
                window.location.href = "/" + window.location.search + window.location.hash;
            }
        },
        dataType: "json",
        contentType: "application/json"
    });
}

