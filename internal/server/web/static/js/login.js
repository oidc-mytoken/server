$(function () {
    checkIfLoggedIn();
})

$('#login-form').on('submit', function (e) {
    e.preventDefault();
    let data = $(this).serializeObject()
    data['restrictions'] = [
        {
            "exp": Math.floor(Date.now() / 1000) + cookieLifetime,
            "ip": ["this"]
        }
    ]
    data['capabilities'] = [
        "tokeninfo",
        "AT",
        "settings",
        "list_mytokens",
        "revoke_any_token"
    ]
    data['rotation'] = {
        "on_other": true,
        "lifetime": 3600 * 24,
    }
    data['client_type'] = 'web';
    data['redirect_uri'] = '/home';
    data['name'] = "mytoken-web";
    storageSet("oidc_issuer", data["oidc_issuer"], true);
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
            window.location.href = uri;
        },
        dataType: "json",
        contentType: "application/json"
    });
    return false;
});
