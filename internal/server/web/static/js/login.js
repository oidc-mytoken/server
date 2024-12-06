$(function () {
    checkIfLoggedIn();
})

function login(issuer, new_tab = false) {
    let data = {
        "grant_Type": "oidc_flow",
        "oidc_flow": "authorization_code",
        "oidc_issuer": issuer,
        'restrictions': [
            {
                "exp": Math.floor(Date.now() / 1000) + cookieLifetime,
                "ip": ["this"]
            }
        ],
        'capabilities': [
            "tokeninfo",
            "AT",
            "settings",
            "list_mytokens",
            "manage_mytokens"
        ],
        'rotation': {
            "on_AT": true,
            "on_other": true,
            "auto_revoke": true,
            "lifetime": 3600 * 24,
        },
        'client_type': 'web',
        'redirect_uri': '/home',
        'name': "mytoken-web"
    };
    storageSet("oidc_issuer", data["oidc_issuer"]);
    data = JSON.stringify(data);
    $.ajax({
        type: "POST",
        url: storageGet("mytoken_endpoint"),
        data: data,
        success: function (res) {
            let consent_uri = res['consent_uri'];
            let auth_uri = res['authorization_uri'];
            let uri;
            if (consent_uri !== undefined) {
                uri = consent_uri;
            } else if (auth_uri !== undefined) {
                uri = auth_uri;
            } else {
                console.error("Unexpected response: ", res);
            }
            if (new_tab) {
                window.open(uri, '_blank').focus();
            } else {
                window.location.href = uri;
            }
        },
        dataType: "json",
        contentType: "application/json"
    });
    return false;
}

$('#login-op-selector').on('changed.bs.select', function (e, clickedIndex, isSelected, previousValue) {
    login($(this).val());
});
