
function checkIfLoggedIn(...next) {
    chainFunctions(
        discovery,
        _checkIfLoggedIn,
        ...next,
    );
}

function _checkIfLoggedIn(...next) {
    console.log('_checkIfLoggedIn');
    let data = {
        'action':'introspect'
    };
    data = JSON.stringify(data);
    $.ajax({
        type: "POST",
        url: storageGet('tokeninfo_endpoint'),
        data: data,
        success: function(res){
            let iss = res['token']['oidc_iss'];
            if (iss) {
                storageSet('oidc_issuer', iss, true);
            }
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

