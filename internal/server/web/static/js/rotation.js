function rotationAT(prefix = "") {
    return $('#' + prefix + 'rotationAT');
}

function rotationOther(prefix = "") {
    return $('#' + prefix + 'rotationOther');
}

function rotationLifetime(prefix = "") {
    return $('#' + prefix + 'rotationLifetime');
}

function rotationAutoRevoke(prefix = "") {
    return $('#' + prefix + 'rotationAutoRevoke');
}

function rotIcon(prefix = "") {
    return $('#' + prefix + 'rot-icon');
}

function checkRotation(prefix = "") {
    let atEnabled = rotationAT(prefix).prop("checked");
    let otherEnabled = rotationOther(prefix).prop("checked");
    let rotationEnabled = atEnabled || otherEnabled;
    let readOnlyMode = getPrefixData(prefix)['rotation']['read-only'] || false;
    rotationLifetime(prefix).prop("disabled", !rotationEnabled || readOnlyMode);
    rotationAutoRevoke(prefix).prop("disabled", !rotationEnabled || readOnlyMode);
    return [atEnabled, otherEnabled];
}

function updateRotationIcon(prefix = "") {
    let en = checkRotation(prefix);
    if (en[0] && en[1]) {
        rotIcon(prefix).addClass("text-success");
        rotIcon(prefix).removeClass("text-info");
        rotIcon(prefix).removeClass("text-primary");
        rotIcon(prefix).attr('data-original-title', "This token is rotated whenever it is used.");
    } else if (en[0]) {
        rotIcon(prefix).addClass("text-primary");
        rotIcon(prefix).removeClass("text-info");
        rotIcon(prefix).removeClass("text-success");
        rotIcon(prefix).attr('data-original-title', "This token is rotated on access token requests.");
    } else if (en[1]) {
        rotIcon(prefix).addClass("text-info");
        rotIcon(prefix).removeClass("text-success");
        rotIcon(prefix).removeClass("text-primary");
        rotIcon(prefix).attr('data-original-title', "This token is rotated on other requests than access token requests.");
    } else {
        rotIcon(prefix).removeClass("text-success");
        rotIcon(prefix).removeClass("text-info");
        rotIcon(prefix).removeClass("text-primary");
        rotIcon(prefix).attr('data-original-title', "This token is never rotated.");
    }
}

$('.rotationAT').on("click", function () {
    let prefix = extractPrefix("rotationAT", $(this).prop("id"));
    let en = checkRotation(prefix);
    if (en[0] && !en[1]) {
        rotationAutoRevoke(prefix).prop("checked", true);
    }
    updateRotationIcon(prefix);
});

$('.rotationOther').on("click", function () {
    let prefix = extractPrefix("rotationOther", $(this).prop("id"));
    let en = checkRotation(prefix);
    if (en[1] && !en[0]) {
        rotationAutoRevoke(prefix).prop("checked", true);
    }
    updateRotationIcon(prefix);
});

function getRotationFromForm(prefix = "") {
    if (!rotationAT(prefix).prop("checked") && !rotationOther(prefix).prop("checked")) {
        return null;
    }
    return {
        "on_AT": rotationAT(prefix).prop("checked"),
        "on_other": rotationOther(prefix).prop("checked"),
        "lifetime": Number(rotationLifetime(prefix).val()),
        "auto_revoke": rotationAutoRevoke(prefix).prop("checked")
    };
}