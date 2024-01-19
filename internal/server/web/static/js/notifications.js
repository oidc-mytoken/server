let notificationsMap = {};

function listNotifications() {
    $.ajax({
        type: "GET",
        url: storageGet('notifications_endpoint'),
        success: function (res) {
            $('#notifications-msg').html(notificationsToTable(res["notifications"]));
            $('[data-toggle="tooltip"]').tooltip();
        },
        error: function (errRes) {
            $errorModalMsg.text(getErrorMessage(errRes));
            $errorModal.modal();
        },
    });
}

$('#notifications-tab').on('shown.bs.tab', listNotifications)
$('#notifications-reload').on('click', listNotifications)

function getNotificationClassIcon(notificationInfo, class_name) {
    const icons = {
        "AT_creations": {
            "green": `<i class="fab fa-openid text-success"></i>`,
            "yellow": `<i class="fab fa-openid text-info"></i>`,
            "grey": `<i class="fab fa-openid text-muted"></i>`
        },
        "subtoken_creations": {
            "green": `<i class="fas fa-key text-success"></i>`,
            "yellow": `<i class="fas fa-key text-info"></i>`,
            "grey": `<i class="fas fa-key text-muted"></i>`
        },
        "setting_changes": {
            "green": `<i class="fas fa-cog text-success"></i>`,
            "yellow": `<i class="fas fa-cog text-info"></i>`,
            "grey": `<i class="fas fa-cog text-muted"></i>`
        },
        "security": {
            "green": `<i class="fas fa-shield-alt text-success"></i>`,
            "yellow": `<i class="fas fa-shield-alt text-info"></i>`,
            "grey": `<i class="fas fa-shield-alt text-muted"></i>`
        }
    };
    if (notificationInfo["notification_classes"].includes(class_name)) {
        return icons[class_name]["green"];
    } else if (notificationInfo["notification_classes"].includesPrefix(class_name)) {
        return icons[class_name]["yellow"];
    } else {
        return icons[class_name]["grey"];
    }
}

function notificationsToTable(notifications) {
    let tableEntries = "";
    const websocketIcon = `<i class="fas fa-rss"></i>`;
    const mailIcon = `<i class="fas fa-envelope"></i>`;
    const userwideIcon = `<i class="fas fa-user"></i>`;

    notificationsMap = {};

    notifications.forEach(function (n) {
        notificationsMap[n["management_code"]] = n;
        let typeIcon;
        switch (n["notification_type"]) {
            case "mail":
                typeIcon = mailIcon;
                break;
            case "ws":
                typeIcon = websocketIcon;
                break;
        }
        let tokens = userwideIcon;
        if (!n["user_wide"]) {
            tokens = `<span class="badge badge-pill badge-primary">${n["subscribed_tokens"].length}</span>`;
        }
        let management_code = n["management_code"];
        let notification_classes_icons = getNotificationClassIcon(n, "AT_creations") +
            getNotificationClassIcon(n, "subtoken_creations") +
            getNotificationClassIcon(n, "setting_changes") +
            getNotificationClassIcon(n, "security");
        let notification_classes_str = n["notification_classes"].join(", ");
        let notification_classes_html = `<span data-toggle="tooltip"
              data-placement="bottom" title=""
              data-original-title="${notification_classes_str}">${notification_classes_icons}</span>`
        tableEntries += `<tr management-code="${management_code}" role="button" onclick="toggleSubscriptionDetails('${management_code}')"><td>${typeIcon}</td><td>${notification_classes_html}</td><td>${tokens}</td></tr>`;
        tableEntries += `<tr><td colspan="4" management-code="${management_code}" class="notification-details-container d-none pl-5"></td></tr>`
    });
    if (tableEntries === "") {
        tableEntries = `<tr><td colSpan="4" class="text-muted text-center">No notifications created yet</td></tr>`;
    }
    return '<table class="table table-hover table-grey">' +
        '<thead><tr>' +
        '<th>Type</th>' +
        '<th>Notification Classes</th>' +
        '<th>Subscribed Tokens</th>' +
        '<th></th>' +
        '</tr></thead>' +
        '<tbody>' +
        tableEntries +
        '</tbody></table>';
}

const $managementCodeInput = $('#management-code');
const $notificationsTokenTable = $('#notifications-tokens-table');
const $notificationsModifyDiv = $('#notifications-modify');
const $notificationSubscribedTokensDetails = $('#subscribed-tokens-details');
const $notificationSubscribedTokensDetailsUserWide = $('#subscribed-tokens-details-user-wide');

function toggleSubscriptionDetails(managementCode) {
    let $container = $(`td.notification-details-container[management-code="${managementCode}"`);
    let washidden = $container.hasClass("d-none");
    $('td.notification-details-container').hideB();
    if (!washidden) {
        return;
    }
    fillModifyNotification(managementCode);
    $container.append($notificationsModifyDiv);
    $notificationsModifyDiv.showB();
    $container.showB();
}

function fillModifyNotification(managementCode) {
    if (loadedTokenList) {
        _fillModifyNotification(managementCode);
        return;
    }
    _getListTokenInfo(undefined, function () {
        loadedTokenList = true;
        _fillModifyNotification(managementCode);
    });
}

function _fillModifyNotification(managementCode) {
    $managementCodeInput.val(managementCode);
    $notificationsTokenTable.html("");
    let n = notificationsMap[managementCode];
    capabilityChecks("notifications-modify-").prop("checked", false);
    n["notification_classes"].forEach(function (nc) {
        checkCapability(nc, "notifications-modify-");
    });
    let tokens = n["subscribed_tokens"];
    if (tokens === undefined) {
        $notificationSubscribedTokensDetailsUserWide.showB();
        $notificationSubscribedTokensDetails.hideB();
    } else {
        tokens.forEach(function (momid) {
            copyTokenTr($(`tr[mom-id="${momid}"]`).clone(true), false, tokens);
        });
        $notificationSubscribedTokensDetailsUserWide.hideB();
        $notificationSubscribedTokensDetails.showB();
    }
}

function copyTokenTr($tr, force, tokens) {
    let old_id = $tr.attr("id");
    let new_id = 'notifications-' + old_id;
    $tr.attr("id", new_id);
    let old_parent = $tr.attr("parent-id");
    if (old_parent !== undefined && old_parent !== null && old_parent !== "" && old_parent !== "0") {
        let parent_momid = $('#' + old_parent).attr("mom-id");
        if (!force && tokens.includes(parent_momid)) {
            // if we are a child of a token that is also in the list of subscribed tokens we don't add the tr here,
            // it will be added later or was already added
            // Reasoning for this overly-complicated approach is to ensure that we have a correct order in the table,
            // i.e. childs are actually under the parent
            return;
        }
    }
    $tr.find('i.fa-bell').parent().remove();
    $tr.find('i.fa-trash').parent().remove();
    $notificationsTokenTable.append($tr);
    let $childs = $(`tr[parent-id="${old_id}"`);
    $childs.each(function () {
        if (!tokens.includes($(this).attr("mom-id"))) {
            // only add childs that are also subscribed
            return;
        }
        let $child = $(this).clone(true);
        $child.attr("parent-id", new_id);
        copyTokenTr($child, true, tokens);
    })

}