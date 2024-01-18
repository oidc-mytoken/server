$(function () {
    chainFunctions(
        checkIfLoggedIn,
        initGrants,
        initSSH,
    );
    openCorrectTab();
})

function initGrants(...next) {
    $.ajax({
        type: "GET",
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

function openCorrectTab() {
    // https://stackoverflow.com/a/17552459
    // Javascript to enable link to tab
    let url = document.location.toString();
    if (url.match('#')) {
        let hash = url.split('#')[1];
        $('.nav-tabs a[href="#' + hash + '"]').tab('show');
    }
}


// With HTML5 history API, we can easily prevent scrolling!
$('.nav-tabs a').on('shown.bs.tab', function (e) {
    if (history.pushState) {
        history.pushState(null, null, e.target.hash);
    } else {
        window.location.hash = e.target.hash; //Polyfill for old browsers
    }
    let $found = $(this).parents('.card').find('.tab-pane.active .nav-tabs a.active');
    if ($found.attr('id') !== $(this).attr('id')) {
        $found.triggerHandler('shown.bs.tab');
    }
})
