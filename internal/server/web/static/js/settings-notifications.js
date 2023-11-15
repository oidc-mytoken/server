const $emailInput = $('#email-input');
const $emailVerifiedIcon = $('#email-verified-icon');
const $emailUnverifiedIcon = $('#email-unverified-icon');
const $preferredMimeTypeHTML = $('#preferred_mimetype_html');
const $preferredMimeTypeText = $('#preferred_mimetype_txt');
const $editMailBtn = $('#edit-mail-btn');
const $saveMailBtn = $('#save-mail-btn');

let settingsStatus = {"email_data_obtained": false};

$('#email-trigger-btn').on('click', function () {
    getEmailInfo();
})

function getEmailInfo() {
    if (settingsStatus["email_data_obtained"]) {
        return;
    }
    $.ajax({
        type: "GET",
        url: storageGet('usersettings_endpoint') + "/email",
        success: function (res) {
            console.log(res);
            $emailInput.val(res["email_address"]);
            if (res["email_verified"]) {
                $emailUnverifiedIcon.hideB();
                $emailVerifiedIcon.showB();
            } else {
                $emailVerifiedIcon.hideB();
                $emailUnverifiedIcon.showB();
            }
            if (res["prefer_html_mail"]) {
                $preferredMimeTypeHTML.prop("checked", true);
            } else {
                $preferredMimeTypeText.prop("checked", true);
            }
            settingsStatus["email_data_obtained"] = true;
        },
        error: function (errRes) {
            $settingsErrorModalMsg.text(getErrorMessage(errRes));
            $settingsErrorModal.modal();
        },
    });
}

$('input[type=radio][name=preferred_mimetype]').on('change', function () {
    let data = {
        "prefer_html_mail": $preferredMimeTypeHTML.prop("checked")
    }
    data = JSON.stringify(data);
    $.ajax({
        type: "PUT",
        data: data,
        dataType: "json",
        contentType: "application/json",
        url: storageGet('usersettings_endpoint') + "/email",
        error: function (errRes) {
            $settingsErrorModalMsg.text(getErrorMessage(errRes));
            $settingsErrorModal.modal();
        },
    });
})

$editMailBtn.on('click', function () {
    $editMailBtn.hideB();
    $saveMailBtn.showB();
    $emailInput.prop("disabled", false);
    $emailInput.select();
})

function saveMail() {
    let data = {
        "email_address": $emailInput.val()
    }
    data = JSON.stringify(data);
    $.ajax({
        type: "PUT",
        data: data,
        dataType: "json",
        contentType: "application/json",
        url: storageGet('usersettings_endpoint') + "/email",
        success: function (res) {
            $saveMailBtn.hideB();
            $editMailBtn.showB();
            $emailVerifiedIcon.hideB();
            $emailUnverifiedIcon.showB();
            $emailInput.prop("disabled", true);
        },
        error: function (errRes) {
            $settingsErrorModalMsg.text(getErrorMessage(errRes));
            $settingsErrorModal.modal();
        },
    });
}

$saveMailBtn.on('click', saveMail)

$emailInput.on('keyup', function (e) {
    if (e.keyCode === 13) { // Enter
        e.preventDefault();
        saveMail();
    }
})
