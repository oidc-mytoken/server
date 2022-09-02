const $sshGrantStatusEnabled = $('#sshGrantStatusEnabled');
const $sshGrantStatusDisabled = $('#sshGrantStatusDisabled');
const $addSSHKeyBtn = $('#addSSHKeyBtn');
const $sshKeyTable = $('#sshKeys');
const $sshKeyFile = $('#ssh_key_file');
const $sshKeyInput = $('#ssh_key');
const $sshResult = $('#sshResult');
const $sshResultColor = $('#sshResult-color');
const $sshForm = $('#sshForm');
const $sshPendingHeading = $('#sshResult-heading-pending');
const $sshPendingSpinner = $('#pending-spinner');
const $sshSuccessHeading = $('#sshResult-heading-success');
const $sshErrorHeading = $('#sshResult-heading-error');
const $authURL = $('#authorization-url');
const $sshSuccessContent = $('#sshSuccessContent');
const $sshErrorContent = $('#sshErrorContent');
const $sshErrorPre = $('#sshErrorPre');
const $hostConfigDiv = $('#sshHostConfigDiv');
const $sshFollowInstructions = $('#follow-instructions');
const $noSSHKeyEntry = $('#noSSHKeyEntry')

$('#addModal').on('hidden.bs.modal', function () {
    window.clearInterval(intervalID);
    $sshResult.hideB();
    $sshForm.showB();
    initSSH();
})

enableGrantCallbacks['ssh'] = function enableSSHCallback() {
    $sshGrantStatusEnabled.showB();
    $sshGrantStatusDisabled.hideB();
    $addSSHKeyBtn.prop('disabled', false);
};

disableGrantCallbacks['ssh'] = function disableSSHCallback() {
    $sshGrantStatusEnabled.hideB();
    $sshGrantStatusDisabled.showB();
    $addSSHKeyBtn.prop('disabled', true);
};

function initSSH(...next) {
    initRestr();
    initCapabilities();
    checkCapability("tokeninfo", mtPrefix);
    checkCapability("AT", mtPrefix);
    clearSSHKeyTable();
    $.ajax({
        type: "GET",
        url: storageGet('usersettings_endpoint') + "/grants/ssh",
        success: function (res) {
            if (res['grant_enabled']) {
                $sshGrantStatusEnabled.showB();
            } else {
                $sshGrantStatusDisabled.showB();
            }
            let sshKeys = res['ssh_keys'];
            if (sshKeys === undefined || sshKeys === null) {
                $noSSHKeyEntry.showB();
            } else {
                sshKeys.forEach(function (key) {
                    addSSHKeyToTable(key);
                })
            }
            doNext(...next);
        },
        error: function (errRes) {
            $settingsErrorModalMsg.text(getErrorMessage(errRes));
            $settingsErrorModal.modal();
        },
    });
}

const deleteKeyHtml = `<td><a href="#" role="button" onclick="deleteKey(this)"><i class="fas fa-trash-alt text-danger"></i></a></td>`;

function clearSSHKeyTable() {
    $sshKeyTable.find('tr.key-entry').remove();
}

function addSSHKeyToTable(key) {
    $noSSHKeyEntry.hideB();
    let created = new Date(key['created'] * 1000).toLocaleString();
    let last_used = key['last_used'];
    if (last_used === undefined || last_used === null || last_used === 0) {
        last_used = "Never";
    } else {
        last_used = new Date(last_used * 1000).toLocaleString();
    }
    let name = key['name'];
    if (name === undefined) {
        name = '';
    }
    let keyFP = key['ssh_key_fp'];
    const html = `<tr id="${keyFP}" class="key-entry"><td>${name}</td><td>${keyFP}</td><td>${created}</td><td>${last_used}</td>${deleteKeyHtml}</tr>`;
    $sshKeyTable.append(html);
}


$sshKeyFile.on('change', function () {
    let file = $(this).prop('files')[0];
    let fileReader = new FileReader();
    fileReader.onload = function () {
        $sshKeyInput.val(fileReader.result);
    };
    fileReader.readAsText(file);
})

function addSSHKey() {
    let data = {
        "grant_type": "mytoken",
        "ssh_key": $sshKeyInput.val(),
        "name": $('#keyName').val(),
        "restrictions": getRestrictionsData(),
        "capabilities": getCheckedCapabilities(),
        "application_name": "mytoken webinterface"
    };
    data = JSON.stringify(data);
    $.ajax({
        type: "POST",
        data: data,
        dataType: "json",
        contentType: "application/json",
        url: storageGet('usersettings_endpoint') + "/grants/ssh",
        success: function (res) {
            let url = res['consent_uri'];
            let code = res['polling_code'];
            let interval = res['interval'];
            $authURL.attr("href", url);
            $authURL.text(url);
            polling(code, interval)
        },
        error: function (errRes) {
            let errMsg = getErrorMessage(errRes);
            sshShowError(errMsg);
        },
    });
    sshShowPending();
    $sshResult.showB();
    $sshForm.hideB();
}

function sshShowPending() {
    $sshPendingHeading.showB();
    $sshPendingSpinner.showB();
    $sshSuccessHeading.hideB();
    $sshErrorHeading.hideB();
    $sshFollowInstructions.showB();
    $sshErrorContent.hideB();
    $sshSuccessContent.hideB();
    $sshResultColor.addClass('alert-warning');
    $sshResultColor.removeClass('alert-success');
    $sshResultColor.removeClass('alert-danger');
}

function sshShowSuccess(user, hostConfig) {
    $sshPendingHeading.hideB();
    $sshPendingSpinner.hideB();
    $sshSuccessHeading.showB();
    $sshErrorHeading.hideB();
    $sshFollowInstructions.hideB();
    $('.sshUserName').text(user);
    if (hostConfig !== undefined) {
        $('#sshHostConfig').text(hostConfig);
        $hostConfigDiv.showB();
    } else {
        $hostConfigDiv.hideB();
    }
    $sshErrorContent.hideB();
    $sshSuccessContent.showB();
    $sshResultColor.addClass('alert-success');
    $sshResultColor.removeClass('alert-warning');
    $sshResultColor.removeClass('alert-danger');
}

function sshShowError(msg) {
    $sshPendingHeading.hideB();
    $sshPendingSpinner.hideB();
    $sshSuccessHeading.hideB();
    $sshErrorHeading.showB();
    $sshFollowInstructions.hideB();
    $sshErrorPre.text(msg);
    $sshErrorContent.showB();
    $sshSuccessContent.hideB();
    $sshResultColor.addClass('alert-danger');
    $sshResultColor.removeClass('alert-success');
    $sshResultColor.removeClass('alert-warning');
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
            url: storageGet('usersettings_endpoint') + "/grants/ssh",
            data: data,
            success: function (res) {
                let user = res['ssh_user'];
                let hostConfig = res['ssh_host_config'];
                sshShowSuccess(user, hostConfig);
                window.clearInterval(intervalID);
            },
            error: function (errRes) {
                let error = errRes.responseJSON['error'];
                let message;
                switch (error) {
                    case "authorization_pending":
                        sshShowPending();
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
                sshShowError(message)
            }
        });
    }, interval);
}


function deleteKey(el) {
    const name = $(el).parent().siblings()[0].innerHTML;
    const fp = $(el).parent().siblings()[1].innerHTML;
    $(`
        <div class="modal fade" tabindex="-1" role="dialog" id="{{Name}}-grantEnableModal">
            <div class="modal-dialog modal-dialog-centered modal-lg" role="document">
                <div class="modal-content">
                    <div class="modal-header">
                        <h5 class="modal-title">Delete SSH Key</h5>
                        <button type="button" class="close" data-dismiss="modal" aria-label="Close">
                            <span aria-hidden="true">&times;</span>
                        </button>
                    </div>
                    <div class="modal-body">Confirm to delete the SSH Key with Name '${name}' and fingerprint '${fp}'.</div>
                    <div class="modal-footer">
                        <button type="button" class="btn btn-secondary" data-dismiss="modal">Cancel</button>
                        <button type="button" class="btn btn-danger" data-dismiss="modal" onclick="sendDeleteKeyRequest('${fp}')">Delete</button>
                    </div>
                </div>
            </div>
        </div>
   `).modal();
}

function sendDeleteKeyRequest(keyFP) {
    let data = {
        "ssh_key_fp": keyFP,
    }
    data = JSON.stringify(data);
    $.ajax({
        type: "DELETE",
        data: data,
        dataType: "json",
        contentType: "application/json",
        url: storageGet('usersettings_endpoint') + "/grants/ssh",
        success: function () {
            initSSH();
        },
        error: function (errRes) {
            $settingsErrorModalMsg.text(getErrorMessage(errRes));
            $settingsErrorModal.modal();
        },
    });
}