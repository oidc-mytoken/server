const mtConfig = $('#mt-config');
const mtResult = $('#mt-result');
const maxTokenLenDiv = $('#max_token_len_div');
const $mtOIDCIss = $('#mt-oidc-iss');

function mtResultColor(prefix = "") {
    return $(prefixId('mt-result-color', prefix));
}

function mtPendingHeading(prefix = "") {
    return $(prefixId('mt-result-heading-pending', prefix));
}

function mtPendingSpinner(prefix = "") {
    return $(prefixId('mt-pending-spinner', prefix));
}

function mtSuccessHeading(prefix = "") {
    return $(prefixId('mt-result-heading-success', prefix));
}

function mtErrorHeading(prefix = "") {
    return $(prefixId('mt-result-heading-error', prefix));
}

function mtResultMsg(prefix = "") {
    return $(prefixId('mt-result-msg', prefix));
}

function mtCopyButton(prefix = "") {
    return $(prefixId('mt-result-copy', prefix));
}

function authURL(prefix = "") {
    return $(prefixId('authorization-url', prefix));
}

function tokenTypeBadge(prefix = "") {
    return $(prefixId('token-badge', prefix));
}

function mtInstructions(prefix = "") {
    return $(prefixId('mt-instructions', prefix));
}

const mtPrefix = "createMT-";

function initCreateMT(...next) {
    if (loggedIn) {
        $mtOIDCIss.val(storageGet("oidc_issuer"));
    }
    initCapabilities(mtPrefix);
    checkCapability("tokeninfo", mtPrefix);
    checkCapability("AT", mtPrefix);
    initRestr(mtPrefix);
    updateRotationIcon(mtPrefix);
    initProfileSupport();
    fillPropertiesFromQuery();
    doNext(...next);
}

function fillGUIWithMaybeTemplate(data, template_type, set_in_gui, prefix = "") {
    if (!is_string(data)) {
        set_in_gui(data, prefix);
        return;
    }
    data = data.replace(/^@/, '');
    data = data.replace(/^_\//, '');
    let $select = $(prefixId(`${template_type}-template`, prefix));
    select_set_value_by_option_name($select, data);
    $select.trigger('change');
}


function initProfileSupport() {
    rot_enableProfileSupport(mtPrefix);
    cap_enableProfileSupport(mtPrefix);
    restr_enableProfileSupport(mtPrefix);

    let $template_select = $(prefixId(`profile-template`, mtPrefix));
    $template_select.on('change', function () {
        const v = $(this).val();
        if (v === null || v === "") {
            return;
        }
        const payload = JSON.parse(v);
        fillGUIFromRequestData(payload);
        $(prefixId('cap-template', mtPrefix)).val("");
        $(prefixId('rot-template', mtPrefix)).val("");
        $(prefixId('restr-template', mtPrefix)).val("");
    });
    let $checks = $(`.any-profile-input[instance-prefix=${mtPrefix}]`);
    $checks = $checks.add($(`.any-restr-input[instance-prefix=${mtPrefix}]`));
    $checks = $checks.add($(`.any-rot-input[instance-prefix=${mtPrefix}]`));
    $checks = $checks.add(capabilityChecks(mtPrefix));
    $checks.on('change change.datetimepicker', function (e) {
        if ($(e.currentTarget).hasClass('datetimepicker-input') && datetimepickerChangeTriggeredFromJS) {
            return;
        }
        $template_select.val(""); // reset template selection to custom if something is changed
    });
    select_set_value_by_option_name($template_select, "web-default");
    $template_select.trigger('change');
}

function fillGUIFromRequestData(req) {
    if (req === undefined || req === null) {
        return;
    }
    if (req.name !== undefined) {
        $('#tokenName').val(req.name);
    }
    if (req.oidc_issuer !== undefined) {
        $mtOIDCIss.val(req.oidc_issuer);
    }
    if (req.response_type !== undefined) {
        $('#select-token-type').val(req.response_type);
    }
    fillGUIWithMaybeTemplate(req.restrictions, "restr", set_restrictions_in_gui, mtPrefix);
    fillGUIWithMaybeTemplate(req.rotation, "rot", set_rotation_in_gui, mtPrefix);
    fillGUIWithMaybeTemplate(req.capabilities, "cap", set_capabilities_in_gui, mtPrefix);

}

function fillPropertiesFromQuery() {
    const params = new Proxy(new URLSearchParams(window.location.search), {
        get: (searchParams, prop) => searchParams.get(prop),
    });
    const base64 = params.r;
    if (base64 === null) {
        return;
    }
    const req_str = window.atob(base64);
    const req = JSON.parse(req_str);
    fillGUIFromRequestData(req);
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
            authURL(mtPrefix).attr("href", url);
            authURL(mtPrefix).text(url);
            mtInstructions(mtPrefix).showB();
            polling(code, interval);
            window.open(url, '_blank');
        },
        error: function (errRes) {
            let errMsg = getErrorMessage(errRes);
            mtShowError(errMsg, mtPrefix);
        },
        dataType: "json",
        contentType: "application/json"
    });
    mtShowPending(mtPrefix);
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

function mtShowPending(prefix = "") {
    mtPendingHeading(prefix).showB();
    mtPendingSpinner(prefix).showB();
    mtSuccessHeading(prefix).hideB();
    mtErrorHeading(prefix).hideB();
    mtCopyButton(prefix).hideB();
    mtResultMsg(prefix).text('');
    mtResultColor(prefix).addClass('alert-warning');
    mtResultColor(prefix).removeClass('alert-success');
    mtResultColor(prefix).removeClass('alert-danger');
}

function mtShowSuccess(msg, prefix = "") {
    mtPendingHeading(prefix).hideB();
    mtPendingSpinner(prefix).hideB();
    mtSuccessHeading(prefix).showB();
    mtErrorHeading(prefix).hideB();
    if (msg.length > 0) {
        mtCopyButton(prefix).showB();
    } else {
        mtCopyButton(prefix).hideB();
    }
    mtResultMsg(prefix).text(msg);
    mtResultColor(prefix).addClass('alert-success');
    mtResultColor(prefix).removeClass('alert-warning');
    mtResultColor(prefix).removeClass('alert-danger');
}

function mtShowError(msg, prefix = "") {
    mtPendingHeading(prefix).hideB();
    mtPendingSpinner(prefix).hideB();
    mtSuccessHeading(prefix).hideB();
    mtErrorHeading(prefix).showB();
    mtCopyButton(prefix).showB();
    mtResultMsg(prefix).text(msg);
    mtResultColor(prefix).addClass('alert-danger');
    mtResultColor(prefix).removeClass('alert-success');
    mtResultColor(prefix).removeClass('alert-warning');
}

let intervalID;

function polling_with_callback(code, interval, okCallback, errCallback) {
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
                okCallback(res);
                window.clearInterval(intervalID);
            },
            error: function (errRes) {
                if (!errCallback(errRes)) {
                    window.clearInterval(intervalID);
                }
            }
        });
    }, interval);
}


function polling(code, interval) {
    polling_with_callback(code, interval, function (res) {
        let token_type = res['mytoken_type'];
        let token = res['mytoken'];
        switch (token_type) {
            case "short_token":
                tokenTypeBadge(mtPrefix).text("Short Token");
                break;
            case "transfer_code":
                tokenTypeBadge(mtPrefix).text("Transfer Code");
                token = res['transfer_code'];
                break;
            case "token":
            default:
                tokenTypeBadge(mtPrefix).text("JWT");
        }
        storageSet("tokeninfo_token", token);
        mtResult.hideB();
        mtConfig.showB();
        $('#info-tab').click();
    }, function (errRes) {
        let error = errRes.responseJSON['error'];
        let message;
        switch (error) {
            case "authorization_pending":
                // message = "Authorization still pending.";
                mtShowPending(mtPrefix);
                return true;
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
        mtShowError(message, mtPrefix);
        return false;
    });
}

