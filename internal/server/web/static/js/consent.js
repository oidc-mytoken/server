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
    initCapabilities();
    checkedCapabilities.forEach(function (value) {
        checkCapability(value);
    })
    // Display tags if any
    initConsentTags();
    chainFunctions(
        discovery,
        function (...next) {
            initRestr("", ...next);
        },
    );
})

function initConsentTags() {
    const $tagsSection = $('#consent-tags-section');
    const $tagsContainer = $('#consent-tags-container');
    const $noTags = $('#consent-no-tags');

    if (typeof consentTags === 'undefined' || consentTags.length === 0) {
        $noTags.show();
        return;
    }

    $noTags.hide();
    displayTagsInContainer($tagsContainer, consentTags, true);
}

function _approve() {
    let data = {
        "oidc_issuer": issuer,
        "restrictions": getRestrictionsData(),
        "capabilities": getCheckedCapabilities(),
        "name": $('#tokenName').val(),
        "rotation": getRotationFromForm()
    };
    // Include tags if any were specified
    if (typeof consentTags !== 'undefined' && consentTags.length > 0) {
        data["tags"] = consentTags;
    }
    approve(data);

}