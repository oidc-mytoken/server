const $login_hint_issuer = $('#login-hint-issuer');
const $login_hint = $('#token-list-login-hint');

function loadNotificationManagement(...next) {
    $login_hint.hideB();
    $('#btn-add-token-to-notification').showB();
    let mc = window.location.pathname.split('/').slice(-1)[0];
    $managementCodeInput.val(mc);
    $.ajax({
        type: "GET",
        url: `${storageGet('notifications_endpoint')}/${mc}`,
        success: function (res) {
            $('#notifications-msg').html(notificationsToTable([res], false, n => `<button class="btn" type="button" onclick="showDeleteNotificationModal('${mc}')" data-toggle="tooltip" data-placement="right" data-original-title="Delete Notification"><i class="fas fa-trash"></i></button>`));
            let ncs = res["notification_classes"];
            capabilityChecks().prop("checked", false);
            ncs.forEach(function (nc) {
                checkCapability(nc);
            });
            if (res["user_wide"]) {
                $notificationSubscribedTokensDetailsUserWide.showB();
                $notificationSubscribedTokensDetails.hideB();
                $('[data-toggle="tooltip"]').tooltip();
                doNext(...next);
            } else {
                $login_hint_issuer.val(res["oidc_iss"]);
                listMytokensForManagementNotification(res["subscribed_tokens"] || [], function () {
                    $('[data-toggle="tooltip"]').tooltip();
                    doNext(...next);
                });
            }
        },
        error: standardErrorHandler,
    });
}

// nop define
function checkIfLoggedIn() {
}

// this is only used when creating the token table for determining the action btns
let loggedIn = true;

function listMytokensForManagementNotification(tokens, ...next) {
    let data = {
        'action': 'list_mytokens'
    };
    data = JSON.stringify(data);
    $.ajax({
        type: "POST",
        url: storageGet("tokeninfo_endpoint"),
        data: data,
        success: function (res) {
            let tableEntries = "";
            let all_mytokens = res["mytokens"];
            all_mytokens.forEach(function (tokenTree) {
                tableEntries = _tokenTreeToHTML(tokenTree, "", 0, 0, false, tokens) + tableEntries;
            });
            if (tableEntries === "") {
                tableEntries = `<tr><td colSpan="4" class="text-muted text-center">No tokens subscribed</td></tr>`;
            }
            $notificationsTokenTable.html(tableEntries);
            $notificationsTokenTable.find('td.actions-td').html(`<button class="btn" onclick="removeTokenFromNotificationFromSingleManagement.call(this)"><i class="fas fa-minus"</button>`);

            if (all_mytokens.length === tokens.length) {
                $('#btn-add-token-to-notification').hideB();
            } else {
                tableEntries = "";
                all_mytokens.forEach(function (tokenTree) {
                    tableEntries = _tokenTreeToHTML(tokenTree, "", 0, 0, false, tokens, true) + tableEntries;
                });
                $('#notifications-all-tokens-to-subscribe-table').html(tableEntries);
                $('#notifications-all-tokens-to-subscribe-table').find('tr').each((_, tr) => {
                    $(tr).prepend(`
            <td class="d-inline-flex">
                <input type="checkbox" class="notification-token-select" onclick="toggleNotificationTokenSelect.call(this)" value="${$(tr).attr("mom-id")}" data-toggle="tooltip" data-original-title="Select Token">
                <div class="ml-1 custom-control custom-switch" data-toggle="tooltip" data-original-title="Include Children" onclick="toggleIncludeChildren.call(this)">
                  <input type="checkbox" class="custom-control-input include-children-switch" disabled checked>
                  <label class="custom-control-label"></label>
                </div>
            </td>`);
                });
                $('.notification-token-select').prop("checked", false);
                $('.include-children-switch').prop("disabled", true).prop("checked", true);
            }
            activateTokenList();
            doNext(...next);
        },
        error: function (errRes) {
            $login_hint.showB();
            $('#token-list-login-hint-hide').hideB();
        },
        dataType: "json",
        contentType: "application/json"
    });
}

if (typeof cookieLifetime === 'undefined') {
    cookieLifetime = 3600 * 24;
}

$(function () {
    chainFunctions(
        discovery,
        loadNotificationManagement,
    );
});


$('#btn-save-notification-classes').off('click').on('click', function () {
    let mc = $managementCodeInput.val();
    let data = {"notification_classes": getCheckedCapabilities()};
    data = JSON.stringify(data);
    $.ajax({
        type: "PUT",
        data: data,
        dataType: "json",
        contentType: "application/json",
        url: `${storageGet('notifications_endpoint')}/${mc}/nc`,
        success: function () {
            chainFunctions(
                loadNotificationManagement,
                function () {
                    $('.collapse').collapse('hide');
                    $('.my-expand').text('Expand');
                }
            );
        },
        error: standardErrorHandler
    });
})

//overwriting
function deleteNotification() {
    $.ajax({
        type: "DELETE",
        dataType: "json",
        contentType: "application/json",
        url: `${storageGet('notifications_endpoint')}/${$managementCodeInput.val()}`,
        success: function () {
            window.location.href = "/";
        },
        error: standardErrorHandler
    });
}

$('#token-list-login-btn').on('click', function () {
    login($login_hint_issuer.val(), true);
})

function removeTokenFromNotificationFromSingleManagement() {
    let mc = $managementCodeInput.val();
    if ($notificationsTokenTable.find('tr').length === 1) {
        $lastTokenInNotificationHint.showB();
        $deleteNotificationModal.modal();
        return;
    }
    let data = {
        "mom_id": $(this).parent().parent().attr("mom-id")
    };
    console.log(data);
    $.ajax({
        type: "DELETE",
        data: data,
        dataType: "json",
        contentType: "application/json",
        url: `${storageGet('notifications_endpoint')}/${mc}/token`,
        success: function () {
            loadNotificationManagement();
        },
        error: standardErrorHandler
    });
}

// overwriting from notifications.js
$('#btn-add-token-to-notification').off('click').on('click', function () {
    $onlyAddTokensContent.showB();
    $newNotificationContent.hideB();
    $newNotificationModal.modal()
});

// overwriting from notifications.js
$('.new-notification-save-btn').off('click').on('click', function () {
    addTokensToNotification(function () {
        loadNotificationManagement(function () {
            $newNotificationModal.modal('hide');
        });
    });
})
