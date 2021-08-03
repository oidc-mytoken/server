$(function (){
    $('.capability-check').click();
    $('#tokenName').val(tokenName);
    if (!rot_null) {
        rotationAT.prop("checked", rot_onAT);
        rotationOther.prop("checked", rot_onOther);
        rotationLifetime.val(rot_lifetime);
        rotationLifetime.prop("disabled", !rot_onAT && !rot_onOther);
        rotationAutoRevoke.prop("checked", rot_autoRevoke);
        rotationAutoRevoke.prop("disabled", !rot_onAT && !rot_onOther);
    }
    chainFunctions(
        discovery,
        initRestrGUI,
    );
})

function approve() {
    let data = {
        "oidc_iss": issuer,
        "restrictions": restrictions,
        "capabilities": $('.capability-check:checked').map(function(_, el) {
            return $(el).val();
        }).get(),
        "subtoken_capabilities": $('.subtoken-capability-check:checked').map(function(_, el) {
            return $(el).val();
        }).get(),
        "name": $('#tokenName').val(),
        "rotation": getRotationFromForm()
    };
    data = JSON.stringify(data);
    $.ajax({
        type: "POST",
        url: window.location.href,
        data: data,
        success: function (res){
            window.location.href = res['authorization_url'];
        },
        error: function(errRes){
            let errMsg = getErrorMessage(errRes);
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
            console.log(errMsg);
            window.location.href = errRes.responseJSON['url'];
        },
        dataType: "json",
        contentType : "application/json"
    });
}
