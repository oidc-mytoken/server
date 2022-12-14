const mtResult = $('#mt-result');
const mtResultColor = $('#mt-result-color');
const mtConfig = $('#mt-config');
const mtPendingHeading = $('#mt-result-heading-pending');
const mtPendingSpinner = $('#mt-pending-spinner');
const mtSuccessHeading = $('#mt-result-heading-success');
const mtErrorHeading = $('#mt-result-heading-error');
const mtResultMsg = $('#mt-result-msg');
const mtCopyButton = $('#mt-result-copy');
const authURL = $('#authorization-url');
const maxTokenLenDiv = $('#max_token_len_div');
const tokenTypeBadge = $('#token-badge');
const $mtInstructions = $('#mt-instructions');
const $mtOIDCIss = $('#mt-oidc-iss');

const mtPrefix = "createMT-";

function initCreateMT(...next) {
    if (loggedIn) {
        $mtOIDCIss.val(storageGet("oidc_issuer"));
    }
    initCapabilities(mtPrefix);
    checkCapability("tokeninfo", mtPrefix);
    checkCapability("AT", mtPrefix);
    initRestr(mtPrefix);
    fillPropertiesFromQuery();
    updateRotationIcon(mtPrefix);
    doNext(...next);
}

function fillPropertiesFromQuery() {
    const params = new Proxy(new URLSearchParams(window.location.search), {
        get: (searchParams, prop) => searchParams.get(prop),
    });
    const base64 = params.r;
    if (base64 == undefined) {
        return;
    }
    const req_str = window.atob(base64);
    console.log(req_str);
    const req = JSON.parse(req_str);

    if (req.name !== undefined) {
        $('#tokenName').val(req.name);
    }
    if (req.oidc_issuer !== undefined) {
        $mtOIDCIss.val(req.oidc_issuer);
    }
    if (req.restrictions !== undefined) {
        setRestrictionsData(req.restrictions, mtPrefix);
        RestrToGUI(mtPrefix);
    }
    if (req.response_type !== undefined) {
        $('#select-token-type').val(req.response_type);
    }
    if (req.rotation !== undefined) {
        rotationAT(mtPrefix).prop("checked", req.rotation.on_AT || false);
        rotationOther(mtPrefix).prop("checked", req.rotation.on_other || false);
        rotationLifetime(mtPrefix).val(req.rotation.lifetime);
        rotationAutoRevoke(mtPrefix).prop("checked", req.rotation.auto_revoke || false);
    }
    if (req.capabilities !== undefined) {
        capabilityChecks(mtPrefix).prop("checked", false);
        req.capabilities.forEach(function (c) {
            checkCapability(c, mtPrefix);
        });
    }
}

$mtOIDCIss.on('change', function () {
    initRestrGUI(mtPrefix);
});

$('#next-mt').on('click', function () {
    window.clearInterval(intervalID);
    mtResult.hideB();
    mtConfig.showB();
})

$('#select-token-type').on('change', function () {
    if ($(this).val() === "auto") {
        maxTokenLenDiv.showB();
    } else {
        maxTokenLenDiv.hideB();
    }
})

function sendCreateMTReq() {
    let data = {
        "name": $('#tokenName').val(),
        "oidc_issuer": $mtOIDCIss.val(),
        "grant_type": "oidc_flow",
        "oidc_flow": "authorization_code",
        "redirect_type": "native",
        "restrictions": getRestrictionsData(mtPrefix),
        "capabilities": getCheckedCapabilities(mtPrefix),
        "application_name": "mytoken webinterface"
    };
    let token_type = $('#select-token-type').val();
    if (token_type === "auto") {
        data['max_token_len'] = Number($('#max_token_len').val());
    } else {
        data['response_type'] = token_type;
    }
    let rot = getRotationFromForm(mtPrefix);
    if (rot) {
        data["rotation"] = rot;
    }
    data = JSON.stringify(data);
    $.ajax({
        type: "POST",
        url: storageGet('mytoken_endpoint'),
        data: data,
        success: function (res) {
            let url = res['consent_uri'];
            let code = res['polling_code'];
            let interval = res['interval'];
            authURL.attr("href", url);
            authURL.text(url);
            $mtInstructions.showB();
            polling(code, interval);
            window.open(url, '_blank');
        },
        error: function (errRes) {
            let errMsg = getErrorMessage(errRes);
            mtShowError(errMsg);
        },
        dataType: "json",
        contentType: "application/json"
    });
    mtShowPending();
    mtResult.showB();
    mtConfig.hideB();
}

function checkRestrEmpty() {
    let restr = getRestrictionsData(mtPrefix);
    if (restr.length === 0) {
        return true;
    }
    let found = false;
    restr.forEach(function (r) {
        if (Object.keys(r).length > 0) {
            found = true;
        }
    })
    return !found;
}

$('#get-mt').on('click', function () {
    if (checkRestrEmpty() && !confirm("You do not have any restrictions defined for this mytoken. Do you really want" +
        " to create an unrestricted mytoken?")) {
        return;
    }
    sendCreateMTReq();
})

function mtShowPending() {
    mtPendingHeading.showB();
    mtPendingSpinner.showB();
    mtSuccessHeading.hideB();
    mtErrorHeading.hideB();
    mtCopyButton.hideB();
    mtResultMsg.text('');
    mtResultColor.addClass('alert-warning');
    mtResultColor.removeClass('alert-success');
    mtResultColor.removeClass('alert-danger');
}

function mtShowSuccess(msg) {
    mtPendingHeading.hideB();
    mtPendingSpinner.hideB();
    mtSuccessHeading.showB();
    mtErrorHeading.hideB();
    mtCopyButton.showB();
    mtResultMsg.text(msg);
    mtResultColor.addClass('alert-success');
    mtResultColor.removeClass('alert-warning');
    mtResultColor.removeClass('alert-danger');
}

function mtShowError(msg) {
    mtPendingHeading.hideB();
    mtPendingSpinner.hideB();
    mtSuccessHeading.hideB();
    mtErrorHeading.showB();
    mtCopyButton.showB();
    mtResultMsg.text(msg);
    mtResultColor.addClass('alert-danger');
    mtResultColor.removeClass('alert-success');
    mtResultColor.removeClass('alert-warning');
}

let intervalID;

function polling(code, interval) {
    interval = interval ? interval * 1000 : 5000;
    let data = {
        "grant_type": "polling_code",
        "polling_code": code,
    }
    data = JSON.stringify(data);
    intervalID = window.setInterval(function () {
        $.ajax({
            type: "POST",
            url: storageGet("mytoken_endpoint"),
            data: data,
            success: function (res) {
                let token_type = res['mytoken_type'];
                let token = res['mytoken'];
                switch (token_type) {
                    case "short_token":
                        tokenTypeBadge.text("Short Token");
                        break;
                    case "transfer_code":
                        tokenTypeBadge.text("Transfer Code");
                        token = res['transfer_code'];
                        break;
                    case "token":
                    default:
                        tokenTypeBadge.text("JWT");
                }
                storageSet("tokeninfo_token", token, true);
                window.clearInterval(intervalID);
                mtResult.hideB();
                mtConfig.showB();
                $('#info-tab').click();
            },
            error: function (errRes) {
                let error = errRes.responseJSON['error'];
                let message;
                switch (error) {
                    case "authorization_pending":
                        // message = "Authorization still pending.";
                        mtShowPending();
                        return;
                    case "access_denied":
                        message = "You denied the authorization request.";
                        break;
                    case "expired_token":
                        message = "Code expired. You might want to restart the flow.";
                        break;
                    case "invalid_grant":
                    case "invalid_token":
                        message = "Code already used.";
                        break;
                    case "undefined":
                        message = "No response from server";
                        break;
                    default:
                        message = getErrorMessage(errRes);
                        break;
                }
                window.clearInterval(intervalID);
                mtShowError(message)
            }
        });
    }, interval);
}

