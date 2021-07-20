
const subtokenCapabilityChecks = $('.subtoken-capability-check');
const mtResult = $('#mt-result');
const mtResultColor = $('#mt-result-color');
const mtConfig = $('#mt-config');
const pendingHeading = $('#mt-result-heading-pending');
const pendingSpinner = $('#pending-spinner');
const successHeading = $('#mt-result-heading-success');
const errorHeading = $('#mt-result-heading-error');
const mtResultMsg = $('#mt-result-msg');
const copyButton = $('#mt-result-copy');
const rotationAT = $('#rotationAT');
const rotationOther = $('#rotationOther');
const rotationLifetime = $('#rotationLifetime');
const rotationAutoRevoke = $('#rotationAutoRevoke');
const authURL = $('#authorization-url');
const capabilityCreateMytoken = $('#cp-create_mytoken');


$(function () {
    $('#cp-AT').prop('checked', true)
    $('#cp-tokeninfo_introspect').prop('checked', true)
    if (capabilityCreateMytoken.prop("checked")) {
        $('#subtokenCapabilities').showB();
    }

})

capabilityCreateMytoken.on("click", function() {
    let enabled = $(this).prop("checked");
    subtokenCapabilityChecks.prop("disabled", !enabled);
    if (!enabled) {
        subtokenCapabilityChecks.prop("checked", false);
    }
    $('#subtokenCapabilities').toggleClass('d-none');
});


$('#next-mt').on('click', function(e){
    window.clearInterval(intervalID);
    mtResult.hideB();
    mtConfig.showB();
})

$('#get-mt').on('click', function(e){
    let data = {
        "name": $('#tokenName').val(),
        "response_type": $('#shortToken').prop("checked") ? "short_token":"token",
        "oidc_issuer": storageGet("oidc_issuer"),
        "grant_type": "oidc_flow",
        "oidc_flow": "authorization_code",
        "redirect_type": "native",
        "restrictions": restrictions,
        "capabilities": $('.capability-check:checked').map(function(_, el) {
            return $(el).val();
        }).get(),
        "subtoken_capabilities": $('.subtoken-capability-check:checked').map(function(_, el) {
            return $(el).val();
        }).get()
    };
    if (rotationAT.prop("checked")||rotationOther.prop("checked")) {
       data["rotation"] = {
           "on_AT": rotationAT.prop("checked"),
           "on_other": rotationOther.prop("checked"),
           "lifetime": Number(rotationLifetime.val()),
           "auto_revoke": rotationAutoRevoke.prop("checked")
       };
    }
    data = JSON.stringify(data);
    console.log(data);
    $.ajax({
        type: "POST",
        url: storageGet('mytoken_endpoint'),
        data: data,
        success: function (res){
            let url = res['authorization_url'];
            let code = res['polling_code'];
            let interval = res['polling_code_interval'];
            authURL.attr("href", url);
            authURL.text(url);
            polling(code, interval)
        },
        error: function(errRes){
            let errMsg = getErrorMessage(errRes);
            showError(errMsg);
        },
        dataType: "json",
        contentType : "application/json"
    });
    showPending();
    mtResult.showB();
    mtConfig.hideB();
})

function showPending() {
    pendingHeading.showB();
    pendingSpinner.showB();
    successHeading.hideB();
    errorHeading.hideB();
    copyButton.hideB();
    mtResultMsg.text('');
    mtResultColor.addClass('alert-warning');
    mtResultColor.removeClass('alert-success', 'alert-danger');
}
function showSuccess(msg) {
    pendingHeading.hideB();
    pendingSpinner.hideB();
    successHeading.showB();
    errorHeading.hideB();
    copyButton.showB();
    mtResultMsg.text(msg);
    mtResultColor.addClass('alert-success');
    mtResultColor.removeClass('alert-warning', 'alert-danger');
}
function showError(msg) {
    pendingHeading.hideB();
    pendingSpinner.hideB();
    successHeading.hideB();
    errorHeading.showB();
    copyButton.showB();
    mtResultMsg.text(msg);
    mtResultColor.addClass('alert-danger');
    mtResultColor.removeClass('alert-success', 'alert-warning');
}

var intervalID;

function polling(code, interval) {
    interval = interval ? interval*1000 : 5000;
    let data = {
        "grant_type": "polling_code",
        "polling_code": code,
    }
    data = JSON.stringify(data);
   intervalID = window.setInterval(function (){
        $.ajax({
            type: "POST",
            url: storageGet("mytoken_endpoint"),
            data: data,
            success: function(res) {
                showSuccess(res['mytoken']);
                window.clearInterval(intervalID);
            },
            error: function(errRes) {
                let error = errRes.responseJSON['error'];
                switch (error) {
                    case "authorization_pending":
                        message = "Authorization still pending.";
                        showPending();
                        return;
                    case "access_denied":
                        message = "You denied the authorization request.";
                        window.clearInterval(intervalID);
                        break;
                    case "expired_token":
                        message = "Code expired. You might want to restart the flow.";
                        window.clearInterval(intervalID);
                        break;
                    case "invalid_grant":
                    case "invalid_token":
                        message = "Code already used.";
                        window.clearInterval(intervalID);
                        break;
                    case "undefined":
                        message = "No response from server";
                        window.clearInterval(intervalID);
                        break;
                    default:
                        message = getErrorMessage(errRes);
                        window.clearInterval(intervalID);
                        break;
                }
                showError(message)
            }
        });
    }, interval);
}


function checkRotation() {
    let atEnabled = rotationAT.prop("checked");
    let otherEnabled = rotationOther.prop("checked");
    let rotationEnabled = atEnabled||otherEnabled;
    rotationLifetime.prop("disabled", !rotationEnabled);
    rotationAutoRevoke.prop("disabled", !rotationEnabled);
    return [atEnabled, otherEnabled];
}

rotationAT.on("click", function() {
    let en = checkRotation();
    if (en[0]&&!en[1]) {
        rotationAutoRevoke.prop("checked", true);
    }
});

rotationOther.on("click", function() {
    let en = checkRotation();
    if (en[1]&&!en[0]) {
        rotationAutoRevoke.prop("checked", true);
    }
});
