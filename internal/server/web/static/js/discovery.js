const configElements = [
    "access_token_endpoint",
    "mytoken_endpoint",
    "usersettings_endpoint",
    "revocation_endpoint",
    "tokeninfo_endpoint",
    "token_transfer_endpoint",
    "providers_supported",
    "jwks_uri"
]

function discovery(...next) {
    try {
        if (storageGet('discovery') !== null) {
            doNext(...next);
            return;
        }
        $.ajax({
            type: "Get",
            url: instanceURL + (instanceURL.endsWith("/") ? "" : "/") + ".well-known/mytoken-configuration",
            success: function (res) {
                configElements.forEach(function (el) {
                    storageSet(el, res[el]);
                })
                storageSet('discovery', Date.now())
                doNext(...next);
            }
        });
    } catch (e) {
        console.error(e);
        doNext(...next);
    }
}
