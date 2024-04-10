const configElements = [
    "access_token_endpoint",
    "mytoken_endpoint",
    "usersettings_endpoint",
    "notifications_endpoint",
    "revocation_endpoint",
    "tokeninfo_endpoint",
    "token_transfer_endpoint",
    "providers_supported",
    "jwks_uri"
]

function discovery(...next) {
    try {
        const discovery = storageGet('discovery');
        if (discovery !== null && discovery !== undefined && discovery > (Date.now() / 1000 - 3600)) {
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
                storageSet('discovery', Date.now() / 1000)
                doNext(...next);
            }
        });
    } catch (e) {
        console.error(e);
        doNext(...next);
    }
}
