// noinspection GrazieInspection

let notificationsMap = {};

let momIDNotificationsMap = {};
let momIDCalendarsMap = {};

const notificationPrefix = "notifications-";
const notificationListPrefix = "notification-listing-";

const $deleteNotificationModal = $('#delete-notification-modal');
const $lastTokenInNotificationHint = $('#last-token-hint');

const $newNotificationModal = $('#new-notification-modal');
const $newNotificationContent = $('.new-notification-content');
const $onlyAddTokensContent = $('.only-add-tokens-content');
const $newNotificationUserWideInput = $('#new-notification-user-wide-input');

$(function () {
    $newNotificationUserWideInput.prop("checked", false);
})

function listNotifications(...next) {
    if (!email_notifications_supported) {
        doNext(...next);
        return;
    }
    let notificationIterator = function (n) {
        notificationsMap[n["management_code"]] = n;
        if (n["user_wide"]) {
            addToArrayMap(momIDNotificationsMap, "user", n, (a, b) => a["notification_id"] === b["notification_id"])
        } else {
            let tokens = n["subscribed_tokens"];
            if (tokens !== undefined) {
                tokens.forEach(momid => {
                    addToArrayMap(momIDNotificationsMap, momid, n, (a, b) => a["notification_id"] === b["notification_id"])
                });
            }
        }
    };

    $.ajax({
        type: "GET",
        url: storageGet('notifications_endpoint'),
        success: function (res) {
            let notifications = res["notifications"] || [];
            notificationsMap = {};
            momIDNotificationsMap = {};
            notifications.forEach(notificationIterator);
            $('#notifications-msg').html(notificationsToTable(notifications, true, n => `<button class="btn" type="button" onclick="showDeleteNotificationModal('${n["management_code"]}')" data-toggle="tooltip" data-placement="right" data-original-title="Delete Notification"><i class="fas fa-trash"></i></button>`));
            $('[data-toggle="tooltip"]').tooltip();
            doNext(...next);
        },
        error: standardErrorHandler,
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
        },
        "expiration": {
            "green": `<i class="fas fa-clock text-success"></i>`,
            "yellow": `<i class="fas fa-clock text-info"></i>`,
            "grey": `<i class="fas fa-clock text-muted"></i>`
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

function notificationsToTable(notifications, details = true, last_td = (n => "")) {
    let tableEntries = "";
    const websocketIcon = `<i class="fas fa-rss"></i>`;
    const mailIcon = `<i class="fas fa-envelope"></i>`;
    const userwideIcon = `<i class="fas fa-user"></i>`;


    notifications.forEach(function (n) {
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
            tokens = `<span class="badge badge-pill badge-primary">${(n["subscribed_tokens"] || []).length}</span>`;
        }
        let management_code = n["management_code"];
        let notification_classes_icons = getNotificationClassIcon(n, "AT_creations") +
            getNotificationClassIcon(n, "subtoken_creations") +
            getNotificationClassIcon(n, "setting_changes") +
            getNotificationClassIcon(n, "security") +
            getNotificationClassIcon(n, "expiration");
        let notification_classes_str = n["notification_classes"].join(", ");
        let notification_classes_html = `<span data-toggle="tooltip"
              data-placement="bottom" title=""
              data-original-title="${notification_classes_str}">${notification_classes_icons}</span>`
        let details_html = details ? `role="button" onclick="toggleSubscriptionDetails('${management_code}')"` : '';
        tableEntries += `<tr management-code="${management_code}"><td ${details_html}>${typeIcon}</td><td ${details_html}>${notification_classes_html}</td><td ${details_html}>${tokens}</td><td>${last_td(n)}</td></tr>`;
        if (details) {
            tableEntries += `<tr><td colspan="4" management-code="${management_code}" class="notification-details-container d-none pl-5"></td></tr>`;
        }
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
    enableSaveNotificationClassesButton();
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
    capabilityChecks(notificationListPrefix).prop("checked", false);
    n["notification_classes"].forEach(function (nc) {
        checkCapability(nc, notificationListPrefix);
    });
    if (n["user_wide"]) {
        $notificationSubscribedTokensDetailsUserWide.showB();
        $notificationSubscribedTokensDetails.hideB();
    } else {
        let tokens = n["subscribed_tokens"] || [];
        tokens.forEach(function (momid) {
            copyTokenTr($('#token-list-table').find(`tr[mom-id="${momid}"]`).clone(true), false, tokens);
        });
        $notificationSubscribedTokensDetailsUserWide.hideB();
        $notificationSubscribedTokensDetails.showB();
    }
}


function notificationAddAllTokenList($container, hide_tokens = undefined) {
    if (loadedTokenList) {
        _notificationAddAllTokenList($container, hide_tokens);
    } else {
        _getListTokenInfo(undefined, function () {
            loadedTokenList = true;
            _notificationAddAllTokenList($container, hide_tokens);
        });
    }
}

function _notificationAddAllTokenList($container, hide_tokens = undefined) {
    let $original_trs = $('#token-list-table').find('tr');
    if ($original_trs.length === $container.find('tr').length) {
        $container.find('tr').showB();
        if (hide_tokens !== undefined) {
            hide_tokens.forEach(function (mom_id) {
                $container.find(`tr[mom-id=${escapeSelector(mom_id)}]`).hideB();
            });
        }
        tokenFoldCollapse();
        $('.notification-token-select').prop("checked", false);
        $('.include-children-switch').prop("disabled", true).prop("checked", true);
        return;
    }
    let $all_trs = $original_trs.clone(true);
    $container.html("");
    $all_trs.each(function (_, tr) {
        let $tr = $(tr);
        let old_id = $tr.attr("id");
        let new_id = 'notifications-all-token-list-to-add-' + old_id;
        $tr.attr("id", new_id);
        let old_parent = $tr.attr("parent-id");
        if (old_parent !== undefined && old_parent !== null && old_parent !== "" && old_parent !== "0") {
            let new_parent = 'notifications-all-token-list-to-add-' + old_parent;
            $tr.attr("parent-id", new_parent);
        }
        let mom_id = $tr.find('i.fa-trash').parent().attr("id").replace("revoke-", "");
        if (hide_tokens !== undefined && hide_tokens.includes(mom_id)) {
            $tr.hideB();
        }
        $tr.find('i.fa-bell').parent().remove();
        $tr.find('i.fa-trash').parent().remove();
        $tr = $tr.prepend(`
            <td class="d-inline-flex">
                <input type="checkbox" class="notification-token-select" onclick="toggleNotificationTokenSelect.call(this)" value="${mom_id}" data-toggle="tooltip" data-original-title="Select Token">
                <div class="ml-1 custom-control custom-switch" data-toggle="tooltip" data-original-title="Include Children" onclick="toggleIncludeChildren.call(this)">
                  <input type="checkbox" class="custom-control-input include-children-switch" disabled checked>
                  <label class="custom-control-label"></label>
                </div>
            </td>`);
        $container.append($tr);
    });
    tokenFoldCollapse();
    $('.notification-token-select').prop("checked", false);
    $('.include-children-switch').prop("disabled", true).prop("checked", true);
    $('[data-toggle="tooltip"]').tooltip();
}

function toggleIncludeChildren() {
    let $input = $(this).find('input.include-children-switch');
    if ($input.prop("disabled")) {
        return;
    }
    $input.prop("checked", !$input.prop("checked"));
    let $tr = $(this).parent().parent();

    let $child_trs = findTrTokenChilds($tr.prop("id"));
    $child_trs.find('.notification-token-select').prop("checked", $input.prop("checked"));
    $child_trs.find('.include-children-switch').prop("disabled", !$input.prop("checked"));
}

function findTrTokenChilds(parent_id) {
    let $all_childs = $();
    let $childs = $('#notifications-all-tokens-to-subscribe-table').find(`tr[parent-id="${parent_id}"`);
    $childs.each(function (_, c) {
        let $c = $(c);
        $all_childs = $all_childs.add($c)
        $all_childs = $all_childs.add(findTrTokenChilds($c.prop("id")));
    });
    return $all_childs;
}

function toggleNotificationTokenSelect() {
    let checked = $(this).prop("checked");
    let $childSwitch = $(this).siblings().find('.include-children-switch');
    $childSwitch.prop('disabled', !checked);
    if ($childSwitch.prop("checked")) {
        let $child_trs = findTrTokenChilds($(this).parent().parent().prop("id"));
        $child_trs.find('.notification-token-select').prop("checked", checked);
        $child_trs.find('.include-children-switch').prop("disabled", !checked);
    }
}

function getSelectedTokensForNotification() {
    return $('.notification-token-select:checked').map((_, v) => {
        return {
            "mom_id": $(v).val(),
            "include_children": $(v).siblings().find('.include-children-switch').prop("checked")
        }
    }).get()
}

function copyTokenTr($tr, force, tokens) {
    let old_id = $tr.attr("id");
    let new_id = notificationPrefix + old_id;
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
    let $btntd = $tr.find('td.actions-td');
    let mom_id = $tr.find('i.fa-trash').parent().attr("id").replace("revoke-", "");
    $tr.find('i.fa-bell').parent().remove();
    $tr.find('i.fa-trash').parent().remove();
    $btntd.append(`<button class="btn" onclick="removeTokenFromNotificationFromMangement('${mom_id}')" data-toggle="tooltip" data-placement="right" data-original-title="Remove token from notification"><i class="fas fa-minus"</button>`);
    if (!force) {
        $tr.showB();
    }
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
    $('[data-toggle="tooltip"]').tooltip();
}

function showDeleteNotificationModal(management_code) {
    $managementCodeInput.val(management_code);
    $lastTokenInNotificationHint.hideB();
    $deleteNotificationModal.modal();
}

function deleteNotification() {
    $.ajax({
        type: "DELETE",
        dataType: "json",
        contentType: "application/json",
        url: `${storageGet('notifications_endpoint')}/${$managementCodeInput.val()}`,
        success: function () {
            listNotifications();
        },
        error: standardErrorHandler
    });
}

function removeTokenFromNotificationFromMangement(mom_id) {
    let mc = $managementCodeInput.val();
    if ($notificationsTokenTable.find('tr').length === 1) {
        $lastTokenInNotificationHint.showB();
        $deleteNotificationModal.modal();
        return;
    }
    let data = {
        "mom_id": mom_id
    };
    data = JSON.stringify(data);
    $.ajax({
        type: "DELETE",
        data: data,
        dataType: "json",
        contentType: "application/json",
        url: `${storageGet('notifications_endpoint')}/${mc}/token`,
        success: function () {
            listNotifications(function () {
                $('#notifications-msg').find(`tr[management-code=${mc}] td[role=button]`)[0].click();
            });
        },
        error: standardErrorHandler
    });
}

function $tokeninfoCalendarListing(prefix = "") {
    return $(prefixId('tokeninfo-calendar-listing', prefix));
}

function $tokeninfoNoCalendars(prefix = "") {
    return $(prefixId('tokeninfo-calendar-no', prefix));
}

function $tokeninfoNotificationsListing(prefix = "") {
    return $(prefixId('tokeninfo-notifications-listing', prefix));
}

function $tokeninfoNoNotifications(prefix = "") {
    return $(prefixId('tokeninfo-notifications-no', prefix));
}

function $tokeninfoNotificationsListingTableContainer(prefix = "") {
    return $(prefixId('tokeninfo-notifications-listing-table-container', prefix));
}

function $calendarURLContainer(prefix = "") {
    return $(prefixId('calendar-url-container', prefix));
}

function $newCalendarContent(prefix = "") {
    return $(prefixId('new-calendar-content', prefix));
}

function $newCalendarInput(prefix = "") {
    return $(prefixId('new-calendar-input', prefix));
}

function fillCalendarInfo(cals, calsSet, prefix = "") {
    if (!calendar_notifications_supported) {
        return;
    }
    if (!calsSet) {
        $tokeninfoCalendarListing(prefix).hideB();
        $tokeninfoNoCalendars(prefix).showB();
    } else {
        $tokeninfoNoCalendars(prefix).hideB();
        $tokeninfoCalendarListing(prefix).showB();
        $calendarTable(prefix).html("");
        cals.forEach(function (cal) {
            addCalendarToTable(cal, prefix, false);
        })
    }
}

function fillNotificationInfo(notifications, notificationsSet, prefix = "") {
    if (!email_notifications_supported) {
        return;
    }
    if (!notificationsSet || notifications.length === 0) {
        $tokeninfoNotificationsListing(prefix).hideB();
        $tokeninfoNoNotifications(prefix).showB();
        notifications = [];
    } else {
        $tokeninfoNoNotifications(prefix).hideB();
        $tokeninfoNotificationsListing(prefix).showB();
        $tokeninfoNotificationsListingTableContainer(prefix).html(notificationsToTable(notifications, false, n => n["user_wide"] ? `` : `<button class="btn" onclick="removeTokenFromNotificationFromModal('${n['management_code']}', '${prefix}')"><i class="fas fa-minus"></i></button>`));
    }
    $(prefixId("subscribe-mail-new", prefix)).hideB();
    let otherNots = Object.values(notificationsMap).filter(n => n["user_wide"] !== true).filter(n => n["notification_type"] === "mail").filter(n => !notifications.some(v => v["management_code"] === n["management_code"]))
    let tableEntries = ``;
    otherNots.forEach(function (n) {
        let management_code = n["management_code"];
        let tokens = `<span class="badge badge-pill badge-primary">${(n["subscribed_tokens"] || []).length}</span>`;
        let notification_classes_icons = getNotificationClassIcon(n, "AT_creations") +
            getNotificationClassIcon(n, "subtoken_creations") +
            getNotificationClassIcon(n, "setting_changes") +
            getNotificationClassIcon(n, "security") +
            getNotificationClassIcon(n, "expiration");
        let notification_classes_str = n["notification_classes"].join(", ");
        let notification_classes_html = `<span data-toggle="tooltip"
              data-placement="bottom" title=""
              data-original-title="${notification_classes_str}">${notification_classes_icons}</span>`
        let checker = `<input class="form-check-input notifications-checkbox ml-0" type="checkbox" value="${management_code}" instance-prefix="${prefix}">`;
        tableEntries += `<tr management-code="${management_code}"><td>${checker}</td><td>${notification_classes_html}</td><td>${tokens}</td></tr>`;
    })
    $(prefixId("subscribe-mail-existing-notifications", prefix)).html(tableEntries);
    if (otherNots.length === 0) {
        if ($(prefixId('subscribe-mail-new', prefix)).hasClass("d-none")) {
            toggle_subscribe_notification_content(prefix);
        }
        $(`.${prefix}toggle-subscribe-notification-content-btn`).hideB();
    } else if ($(prefixId('subscribe-mail-new', prefix)).hasClass("d-none")) {
        $(prefixId("switch-to-new-mail-notification-btn", prefix)).showB();
    } else {
        $(prefixId("switch-to-existing-mail-notification-btn", prefix)).showB();
    }

    $('[data-toggle="tooltip"]').tooltip();
}

function addToExistingNotifications(prefix = "") {
    let selected_codes = $(prefixId("subscribe-mail-existing-notifications", prefix)).find(".notifications-checkbox:checked").map((_, v) => $(v).val()).get();
    let ajax_promises = [];
    let mom_id = $notificationMOMID.val();
    let data = {
        "mom_id": mom_id,
        "include_children": $(prefixId("subscribe-mail-include-children", prefix)).prop("checked")
    };
    data = JSON.stringify(data);
    selected_codes.forEach(function (mc) {
        ajax_promises.push(
            $.ajax({
                type: "POST",
                data: data,
                dataType: "json",
                contentType: "application/json",
                url: `${storageGet('notifications_endpoint')}/${mc}/token`
            }));
    })
    $.when(...ajax_promises).then(function () {
        updateNotificationsInfoPartInModal(mom_id, prefix);
    }, function (errRes) {
        $errorModalMsg.text(getErrorMessage(errRes));
        $errorModal.modal();
    })
}

function removeTokenFromNotificationFromModal(management_code, prefix = "") {
    let mom_id = $notificationMOMID.val();
    let data = {
        "mom_id": mom_id
    };
    data = JSON.stringify(data);
    $.ajax({
        type: "DELETE",
        data: data,
        dataType: "json",
        contentType: "application/json",
        url: `${storageGet('notifications_endpoint')}/${management_code}/token`,
        success: function () {
            updateNotificationsInfoPartInModal(mom_id, prefix);
        },
        error: function (errRes) {
            $errorModalMsg.text(getErrorMessage(errRes));
            $errorModal.modal();
        }
    });
}

function fillNotificationAndCalendarInfo(cals, notifications, $container, prefix = "") {
    let calsSet = cals !== undefined && cals.length > 0;
    let notificationsSet = notifications !== undefined && notifications.length > 0;
    if (!calsSet && !notificationsSet) {
        $container.hideB();
    } else {
        $container.showB();
    }
    fillCalendarInfo(cals, calsSet, prefix);
    fillNotificationInfo(notifications, notificationsSet, prefix);
}


function notificationModal(doesExpire) {
    let id = this.id.replace("notify-", "");
    $notificationMOMID.val(id);
    let cals = momIDCalendarsMap[id];
    let nots = momIDNotificationsMap[id] || [];
    let user_nots = momIDNotificationsMap["user"] || [];
    nots = nots.concat(user_nots);
    fillNotificationAndCalendarInfo(cals, nots, $(prefixId("not-info-body", notificationPrefix)), notificationPrefix);

    if (!doesExpire || !calendar_notifications_supported) {
        $('#notifications-calendarinfo-alert').hideB();
    } else {
        $('#notifications-calendarinfo-alert').showB();
        notificationModalInitSubscribeCalendars(cals);
        if (email_notifications_supported) {
            notificationModalCheckEmailOK();
        }
    }

    $notificationsModal.modal();
}

function notificationModalInitSubscribeCalendars(cals) {
    let all_cals = cachedNotificationsTypeData["calendars"];
    let options = ``;
    let filtered = cals === undefined ? all_cals : all_cals.filter(c => !cals.some(cc => c["name"] === cc["name"]));
    filtered.forEach(function (c) {
        let name = c["name"];
        options += `<option value="${name}">${name}</option>`;
    });
    options += `<option value="${new_calendar_option_value}" class="text-secondary">New calendar ...</option>`;
    $calendarSelector.html(options);
    $calendarSelector.trigger('change');
}

function notificationModalCheckEmailOK(nextMailOK = function () {
}, nextMailNotOK = function () {
}) {
    $('.notify-entry-content-element').hideB();
    let email_ok = cachedNotificationsTypeData["email_ok"];
    if (email_ok !== undefined && email_ok !== null && email_ok) {
        $('.notify-entry-content-normal').showB();
        return;
    }
    $.ajax({
        type: "GET",
        url: storageGet('usersettings_endpoint') + "/email",
        success: function (res) {
            let email = res["email_address"];
            let email_verified = res["email_verified"];
            if (email === undefined || email === null || email === "") {
                $('.email-not-set-hint').showB();
                cachedNotificationsTypeData["email_ok"] = false;
                nextMailNotOK();
                return;
            }
            if (email_verified === undefined || !email_verified) {
                $('.email-not-verified-hint').showB();
                cachedNotificationsTypeData["email_ok"] = false;
                nextMailNotOK();
                return;
            }
            $('.notify-entry-content-normal').showB();
            cachedNotificationsTypeData["email_ok"] = true;
            nextMailOK();
        },
        error: function (errRes) {
            $errorModalMsg.text(getErrorMessage(errRes));
            $errorModal.modal();
        },
    });

}

const $calendarSelector = $('#notify-calendar-selector');
const $calendarURL = $('#notify-calendar-url');

let calendarURLs = {};

let cachedNotificationsTypeData = {};

function getCalendars(callback = undefined, ...next) {
    if (!calendar_notifications_supported) {
        doNext(...next);
        return;
    }
    $.ajax({
        type: "GET",
        url: storageGet('notifications_endpoint') + "/calendars",
        success: function (res) {
            let cals = res["calendars"] || [];
            cals.forEach(function (c) {
                let tokens = c["subscribed_tokens"] || [];
                tokens.forEach(momid => {
                    addToArrayMap(momIDCalendarsMap, momid, c, (a, b) => a["ics_path"] === b["ics_path"])
                });
                calendarURLs[c["name"]] = c["ics_path"];
            })
            cachedNotificationsTypeData["calendars"] = cals;
            if (callback !== undefined && callback !== null) {
                next.unshift(function () {
                    callback(cals);
                });
            }
            doNext(...next);
        },
        error: function (errRes) {
            $settingsErrorModalMsg.text(getErrorMessage(errRes));
            $settingsErrorModal.modal();
        },
    });
}

function sendCalendarInviteFromModal(prefix = "") {
    let data = {
        "mom_id": $notificationMOMID.val(),
        "notification_type": "ics_invite"
    };
    let comment = $(prefixId('notify-entry-comment', prefix)).val();
    if (comment !== undefined && comment !== null && comment !== "") {
        data["comment"] = comment;
    }
    data = JSON.stringify(data);
    $.ajax({
        type: "POST",
        data: data,
        dataType: "json",
        contentType: "application/json",
        url: storageGet('notifications_endpoint'),
        success: function () {
            $(prefixId("notify-entry-content-normal", prefix)).hideB();
            $(prefixId("calendar-invite-successfully-sent", prefix)).showB();
        },
        error: function (errRes) {
            $errorModalMsg.text(getErrorMessage(errRes));
            $errorModal.modal();
        },
    });
}

const new_calendar_option_value = "$$$new-calendar$$$";

$calendarSelector.on('change', function () {
    let cal = $(this).val();
    if (cal === new_calendar_option_value) {
        $newCalendarContent(notificationPrefix).showB();
        $calendarURLContainer(notificationPrefix).hideB();
        return;
    }
    $newCalendarContent(notificationPrefix).hideB();
    let url = calendarURLs[cal];
    if (url === undefined) {
        $calendarURLContainer(notificationPrefix).hideB();
        return;
    }
    $calendarURLContainer(notificationPrefix).showB();
    $calendarURL.text(url);
    $calendarURL.attr("href", url);
})

function createCalendarFromNotifications() {
    let name = $newCalendarInput(notificationPrefix).val();
    let data = JSON.stringify({"name": name});
    $.ajax({
        type: "POST",
        data: data,
        dataType: "json",
        contentType: "application/json",
        url: storageGet('notifications_endpoint') + "/calendars",
        success: function () {
            getCalendars(function () {
                let cals = momIDCalendarsMap[$notificationMOMID.val()];
                notificationModalInitSubscribeCalendars(cals);
                $calendarSelector.val(name);
            });
        },
        error: function (errRes) {
            $errorModalMsg.text(getErrorMessage(errRes));
            $errorModal.modal();
        },
    });
}

$('#sent-calendar-add').on('click', function () {
    let momID = $notificationMOMID.val();
    let data = {"mom_id": momID};
    let comment = $('#notify-calendar-comment').val();
    if (comment !== undefined && comment !== null && comment !== "") {
        data["comment"] = comment;
    }
    let calendar = $calendarSelector.val();
    data = JSON.stringify(data);
    $.ajax({
        type: "POST",
        data: data,
        dataType: "json",
        contentType: "application/json",
        url: storageGet('notifications_endpoint') + "/calendars/" + calendar,
        success: function () {
            getCalendars(function () {
                let cals = momIDCalendarsMap[momID];
                fillCalendarInfo(cals, cals !== undefined && cals.length > 0, notificationPrefix);
                notificationModalInitSubscribeCalendars(cals);
            });
        },
        error: function (errRes) {
            $errorModalMsg.text(getErrorMessage(errRes));
            $errorModal.modal();
        },
    });
})

function updateNotificationsInfoPartInModal(mom_id, prefix = "") {
    listNotifications(function () {
        let nots = momIDNotificationsMap[mom_id];
        let user_nots = momIDNotificationsMap["user"] || [];
        nots = nots.concat(user_nots);
        fillNotificationInfo(nots, true, prefix);
        $(prefixId("subscribe-mail-content", prefix)).collapse('hide');
    });
}

function sendNewEmailNotificationFromModal(prefix = "") {
    let mom_id = $notificationMOMID.val();
    let data = {
        "mom_id": mom_id,
        "notification_type": "mail",
        "notification_classes": getCheckedCapabilities(prefix),
        "include_children": $(prefixId("subscribe-mail-include-children", prefix)).prop("checked")
    };
    data = JSON.stringify(data);
    $.ajax({
        type: "POST",
        data: data,
        dataType: "json",
        contentType: "application/json",
        url: storageGet('notifications_endpoint'),
        success: function () {
            updateNotificationsInfoPartInModal(mom_id, prefix);
        },
        error: function (errRes) {
            $errorModalMsg.text(getErrorMessage(errRes));
            $errorModal.modal();
        },
    });
}

function toggle_subscribe_notification_content(prefix = "") {
    let selectors = [
        prefixId('send-new-mail-notification-btn', prefix),
        prefixId('send-existing-mail-notification-btn', prefix),
        prefixId('subscribe-mail-new', prefix),
        prefixId('subscribe-mail-existing', prefix),
        `.${prefix}toggle-subscribe-notification-content-btn`
    ];
    $(selectors.join(",")).toggleClass('d-none');
}

enableSaveNotificationClassesButton();

function enableSaveNotificationClassesButton() {
    $('#btn-save-notification-classes').off('click').on('click', function () {
        let mc = $managementCodeInput.val();
        let data = {"notification_classes": getCheckedCapabilities(notificationListPrefix)};
        data = JSON.stringify(data);
        $.ajax({
            type: "PUT",
            data: data,
            dataType: "json",
            contentType: "application/json",
            url: `${storageGet('notifications_endpoint')}/${mc}/nc`,
            success: function () {
                listNotifications();
            },
            error: function (errRes) {
                $errorModalMsg.text(getErrorMessage(errRes));
                $errorModal.modal();
            }
        });
    })
}

function newNotificationModal() {
    $newNotificationContent.showB();
    notificationModalCheckEmailOK(function () {
        notificationAddAllTokenList($('#notifications-all-tokens-to-subscribe-table'));
    });
    $onlyAddTokensContent.hideB();
    $newNotificationModal.modal()
}


$('#btn-add-token-to-notification').on('click', function () {
    let tokens = $(this).closest('div').find('tbody tr').map((_, v) => $(v).attr('mom-id')).get();
    notificationAddAllTokenList($('#notifications-all-tokens-to-subscribe-table'), tokens);
    $onlyAddTokensContent.showB();
    $newNotificationContent.hideB();
    $newNotificationModal.modal()
});

$('.new-notification-save-btn').on('click', function () {
    if (!$onlyAddTokensContent.hasClass('d-none')) {
        addTokensToNotification();
    } else {
        saveNewNotification();
    }
})

function saveNewNotification() {
    let user_wide = $newNotificationUserWideInput.prop("checked");
    let data = {
        "user_wide": user_wide,
        "notification_type": "mail",
        "notification_classes": getCheckedCapabilities("new-notification-modal-"),
    };
    let token_data = getSelectedTokensForNotification();
    if (!user_wide && token_data.length > 0) {
        Object.assign(data, token_data[0]);
    }
    data = JSON.stringify(data);

    function end() {
        listNotifications(function () {
            $newNotificationModal.modal('hide');
        })
    }

    $.ajax({
        type: "POST",
        data: data,
        dataType: "json",
        contentType: "application/json",
        url: storageGet('notifications_endpoint'),
        success: function (res) {
            if (user_wide || token_data.length <= 1) {
                end();
                return;
            }
            let ajax_promises = [];
            let mc = res["management_code"];
            token_data.slice(1).forEach(function (data) {
                data = JSON.stringify(data);
                ajax_promises.push(
                    $.ajax({
                        type: "POST",
                        data: data,
                        dataType: "json",
                        contentType: "application/json",
                        url: `${storageGet('notifications_endpoint')}/${mc}/token`
                    }));
            })
            $.when(...ajax_promises).then(end, standardErrorHandler)
        },
        error: standardErrorHandler,
    });
}

function addTokensToNotification(callback = undefined) {
    let datas = getSelectedTokensForNotification();
    let ajax_promises = [];
    let mc = $managementCodeInput.val();
    datas.forEach(function (data) {
        data = JSON.stringify(data);
        ajax_promises.push(
            $.ajax({
                type: "POST",
                data: data,
                dataType: "json",
                contentType: "application/json",
                url: `${storageGet('notifications_endpoint')}/${mc}/token`
            }));
    })
    if (callback === undefined) {
        callback = function () {
            listNotifications(function () {
                $('#notifications-msg').find(`tr[management-code=${mc}] td[role=button]`)[0].click();
                $newNotificationModal.modal('hide');
            })
        };
    }
    $.when(...ajax_promises).then(callback, standardErrorHandler);
}

$('#new-notification-user-wide-input').on('change', function () {
    $('.user-wide-toggle-effected').toggleClass('d-none');
})