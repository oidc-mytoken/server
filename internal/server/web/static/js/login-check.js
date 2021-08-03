
function checkIfLoggedIn(...next) {
    chainFunctions(
        discovery,
        _checkIfLoggedIn,
        ...next,
    );
}

function _checkIfLoggedIn(...next) {
    let data = {
        'action':'introspect'
    };
    data = JSON.stringify(data);
    $.ajax({
        type: "POST",
        url: storageGet('tokeninfo_endpoint'),
        data: data,
        success: function(res){
            let token = res['token'];
            let iss = token['oidc_iss'];
            if (iss) {
                storageSet('oidc_issuer', iss, true);
            }
            let scopes = extractMaxScopesFromToken(token);
            storageSet('token_scopes', scopes, true);
            if (window.location.pathname === "/") {
                window.location.href = "/home";
            }
            doNext(...next);
        },
        error: function (res) {
            if (window.location.pathname !== "/") {
                window.location.href = "/";
            }
        },
        dataType: "json",
        contentType : "application/json"
    });
}

