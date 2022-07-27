const rotationAT = $('#rotationAT');
const rotationOther = $('#rotationOther');
const rotationLifetime = $('#rotationLifetime');
const rotationAutoRevoke = $('#rotationAutoRevoke');
const $rotIcon = $('#rot-icon');

function checkRotation() {
    let atEnabled = rotationAT.prop("checked");
    let otherEnabled = rotationOther.prop("checked");
    let rotationEnabled = atEnabled || otherEnabled;
    rotationLifetime.prop("disabled", !rotationEnabled);
    rotationAutoRevoke.prop("disabled", !rotationEnabled);
    return [atEnabled, otherEnabled];
}

function updateRotationIcon() {
    let en = checkRotation();
    if (en[0] && en[1]) {
        $rotIcon.addClass("text-success");
        $rotIcon.removeClass("text-info");
        $rotIcon.removeClass("text-primary");
        $rotIcon.attr('data-original-title', "This token is rotated whenever it is used.");
    } else if (en[0]) {
        $rotIcon.addClass("text-primary");
        $rotIcon.removeClass("text-info");
        $rotIcon.removeClass("text-success");
        $rotIcon.attr('data-original-title', "This token is rotated on access token requests.");
    } else if (en[1]) {
        $rotIcon.addClass("text-info");
        $rotIcon.removeClass("text-success");
        $rotIcon.removeClass("text-primary");
        $rotIcon.attr('data-original-title', "This token is rotated on other requests than access token requests.");
    } else {
        $rotIcon.removeClass("text-success");
        $rotIcon.removeClass("text-info");
        $rotIcon.removeClass("text-primary");
        $rotIcon.attr('data-original-title', "This token is never rotated.");
    }
}

rotationAT.on("click", function () {
    let en = checkRotation();
    if (en[0] && !en[1]) {
        rotationAutoRevoke.prop("checked", true);
    }
    updateRotationIcon();
});

rotationOther.on("click", function () {
    let en = checkRotation();
    if (en[1] && !en[0]) {
        rotationAutoRevoke.prop("checked", true);
    }
    updateRotationIcon();
});

function getRotationFromForm() {
    if (!rotationAT.prop("checked") && !rotationOther.prop("checked")) {
        return null;
    }
    return {
        "on_AT": rotationAT.prop("checked"),
        "on_other": rotationOther.prop("checked"),
        "lifetime": Number(rotationLifetime.val()),
        "auto_revoke": rotationAutoRevoke.prop("checked")
    };
}

$(function () {
    updateRotationIcon();
});