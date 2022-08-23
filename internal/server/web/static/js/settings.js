$(function () {
    chainFunctions(
        checkIfLoggedIn,
        initGrants,
        initSSH,
    );
    // https://stackoverflow.com/a/17552459
    // Javascript to enable link to tab
    let url = document.location.toString();
    if (url.match('#')) {
        let hash = url.split('#')[1];
        $('.nav-tabs a[href="#' + hash + '"]').tab('show');
    }
})

function initGrants(...next) {
    useSettingsToken(function (token) {
        $.ajax({
            type: "GET",
            headers: {
                "Authorization": "Bearer " + token,
            },
            url: storageGet('usersettings_endpoint') + "/grants",
            success: function (res) {
                let grants = res['grant_types'] || [];
                grants.forEach(function (grant) {
                    $('#' + grant['grant_type'] + '-GrantEnable').prop('checked', grant['enabled']);
                })
                doNext(...next);
            },
            error: function (errRes) {
                $settingsErrorModalMsg.text(getErrorMessage(errRes));
                $settingsErrorModal.modal();
            },
        });
    });
}

$('.grant-enable').click(function () {
    let enable = $(this).prop('checked');
    $(this).prop('checked', !enable);
    let name = $(this).prop('name');
    if (enable) {
        $('#' + name + '-grantEnableModal').modal();
    } else {
        $('#' + name + '-grantDisableModal').modal();
    }
})

function enableGrant(grant) {
    sendGrantRequest(grant, true, function () {
        $('#' + grant + '-GrantEnable').prop('checked', true);
        enableGrantCallbacks[grant]();
    });
}

function disableGrant(grant) {
    sendGrantRequest(grant, false, function () {
        $('#' + grant + '-GrantEnable').prop('checked', false);
        disableGrantCallbacks[grant]();
    });
}

let enableGrantCallbacks = {};
let disableGrantCallbacks = {};
