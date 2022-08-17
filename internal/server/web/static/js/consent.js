$(document).ready(function () {
    $('#tokenName').val(tokenName);
    if (!rot_null) {
        rotationAT().prop("checked", rot_onAT);
        rotationOther().prop("checked", rot_onOther);
        rotationLifetime().val(rot_lifetime);
        rotationLifetime().prop("disabled", !rot_onAT && !rot_onOther);
        rotationAutoRevoke().prop("checked", rot_autoRevoke);
        rotationAutoRevoke().prop("disabled", !rot_onAT && !rot_onOther);
    }
    updateRotationIcon();
    checkedCapabilities.forEach(function (value) {
        checkCapability(value, 'cp');
    })
    checkedSubtokenCapabilities.forEach(function (value) {
        checkCapability(value, 'sub-cp');
    })
    initCapabilities();
    chainFunctions(
        discovery,
        function (...next) {
            initRestr("", ...next);
        },
    );
})

function _approve() {
    let data = {
        "oidc_issuer": issuer,
        "restrictions": getRestrictionsData(),
        "capabilities": getCheckedCapabilities(),
        "subtoken_capabilities": getCheckedSubtokenCapabilities(),
        "name": $('#tokenName').val(),
        "rotation": getRotationFromForm()
    };
    approve(data);

}