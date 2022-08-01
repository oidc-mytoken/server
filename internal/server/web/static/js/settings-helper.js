const $errorModal = $('#error-modal')
const $errorModalMsg = $('#error-modal-msg')

function useSettingsToken(callback) {
    let tok = storageGet("settings_mytoken")
    _tokeninfo(
        'introspect',
        function () {
            callback(tok);
        },
        function () {
            requestMT(
                {
                    "name": "mytoken-web MT for settings",
                    "grant_type": "mytoken",
                    "capabilities": ["tokeninfo:introspect", "settings"],
                    "restrictions": [
                        {
                            "exp": Math.floor(Date.now() / 1000) + 300,
                            "ip": ["this"],
                            "usages_AT": 0,
                            "usages_other": 50,
                        }
                    ]
                },
                function (res) {
                    let token = res['mytoken'];
                    storageSet('settings_mytoken', token, true);
                    callback(token);
                },
                function (errRes) {
                    $errorModalMsg.text(getErrorMessage(errRes));
                    $errorModal.modal();
                }
            );
        },
        tok);
}

function sendGrantRequest(grant, enable, okCallback) {
    useSettingsToken(function (token) {
        let data = {
            "grant_type": grant,
            "mytoken": token,
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
                $errorModalMsg.text(getErrorMessage(errRes));
                $errorModal.modal();
            },
        });
    });
}
