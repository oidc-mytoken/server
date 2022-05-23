$(function (){
    checkIfLoggedIn();
})

$('#login-form').on('submit', function(e){
    e.preventDefault();
    let data = $(this).serializeObject()
    data['restrictions'] = [
        {
            "exp": Math.floor(Date.now() / 1000) + 3600 * 24 * 7, // TODO configurable
            "ip": ["this"],
            "usages_AT": 1,
            "usages_other": 100,
        }
    ]
    data['capabilities'] = [
        "create_mytoken",
        "tokeninfo",
    ]
    data['subtoken_capabilities'] = [
        "AT",
        "settings",
        "list_mytokens",
        "tokeninfo:introspect"
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
            window.location.href = res['consent_uri'];
        },
        dataType: "json",
        contentType : "application/json"
    });
    return false;
});
