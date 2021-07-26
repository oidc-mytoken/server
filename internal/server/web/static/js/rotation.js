const rotationAT = $('#rotationAT');
const rotationOther = $('#rotationOther');
const rotationLifetime = $('#rotationLifetime');
const rotationAutoRevoke = $('#rotationAutoRevoke');

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