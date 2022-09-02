const $settingsErrorModal = $('#settings-error-modal')
const $settingsErrorModalMsg = $('#settings-error-modal-msg')

function sendGrantRequest(grant, enable, okCallback) {
    let data = {
        "grant_type": grant,
    }
    data = JSON.stringify(data);
    $.ajax({
        type: enable ? "POST" : "DELETE",
        data: data,
        dataType: "json",
        contentType: "application/json",
        url: storageGet('usersettings_endpoint') + "/grants",
        success: function () {
            okCallback();
        },
        error: function (errRes) {
            $settingsErrorModalMsg.text(getErrorMessage(errRes));
            $settingsErrorModal.modal();
        },
    });
}
