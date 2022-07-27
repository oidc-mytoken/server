$(function (){
    $('#tokenName').val(tokenName);
    if (!rot_null) {
        rotationAT.prop("checked", rot_onAT);
        rotationOther.prop("checked", rot_onOther);
        rotationLifetime.val(rot_lifetime);
        rotationLifetime.prop("disabled", !rot_onAT && !rot_onOther);
        rotationAutoRevoke.prop("checked", rot_autoRevoke);
        rotationAutoRevoke.prop("disabled", !rot_onAT && !rot_onOther);
    }
    updateRotationIcon();
    checkedCapabilities.forEach(function (value) {
        $('#cp-' + escapeSelector(value)).prop("checked", true)
    })
    checkedSubtokenCapabilities.forEach(function (value) {
        $('#sub-cp-' + escapeSelector(value)).prop("checked", true)
    })
    chainFunctions(
        discovery,
        initRestrGUI,
    );
})

function approve() {
    let data = {
        "oidc_iss": issuer,
        "restrictions": restrictions,
        "capabilities": getCheckedCapabilities(),
        "subtoken_capabilities": getCheckedSubtokenCapabilities(),
        "name": $('#tokenName').val(),
        "rotation": getRotationFromForm()
    };
    data = JSON.stringify(data);
    $.ajax({
        type: "POST",
        url: window.location.href,
        data: data,
        headers: {
            Accept: "application/json"
        },
        success: function (res) {
            window.location.href = res['authorization_uri'];
        },
        error: function (errRes) {
            let errMsg = getErrorMessage(errRes);
            if (errRes.status === 404) {
                errMsg = "Expired. Please start the flow again.";
            }
            $('#error-modal-msg').text(errMsg);
            $('#error-modal').modal();
        },
        dataType: "json",
        contentType : "application/json"
    });
}

function cancel() {
    $.ajax({
        type: "POST",
        url: window.location.href,
        success: function (data){
            window.location.href = data['url'];
        },
        error: function(errRes){
            let errMsg = getErrorMessage(errRes);
            console.error(errMsg);
            window.location.href = errRes.responseJSON['url'];
        },
        dataType: "json",
        contentType : "application/json"
    });
}
