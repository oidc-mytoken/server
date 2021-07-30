
const configElements = [
    "access_token_endpoint",
    "mytoken_endpoint",
    "usersettings_endpoint",
    "revocation_endpoint",
    "tokeninfo_endpoint",
    "providers_supported"
]

function discovery(...next) {
    if (storageGet('discovery') !== null) {
        doNext(...next);
        return;
    }
    $.ajax({
        type: "Get",
        url: "/.well-known/mytoken-configuration",
        success: function(res){
            configElements.forEach(function (el){
                storageSet(el, res[el]);
            })
            storageSet('discovery', Date.now())
            doNext(...next);
        }
    });
}
