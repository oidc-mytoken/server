let notificationsMap = {};

let momIDNotificationsMap = {};
let momIDCalendarsMap = {};

const notificationPrefix = "notifications-";


function listNotifications(...next) {
    $.ajax({
        type: "GET",
        url: storageGet('notifications_endpoint'),
        success: function (res) {
            let notifications = res["notifications"];
            notificationsMap = {};
            momIDNotificationsMap = {};
            notifications.forEach(function (n) {
                notificationsMap[n["management_code"]] = n;
                if (n["user_wide"]) {
                    addToArrayMap(momIDNotificationsMap, "user", n, (a, b) => a["notification_id"] === b["notification_id"])
                } else {
                    n["subscribed_tokens"].forEach(function (momid) {
                        addToArrayMap(momIDNotificationsMap, momid, n, (a, b) => a["notification_id"] === b["notification_id"])
                    });
                }
            })
            $('#notifications-msg').html(notificationsToTable(notifications));
            $('[data-toggle="tooltip"]').tooltip();
            doNext(...next);
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
        let details_html = details ? `role="button" onclick="toggleSubscriptionDetails('${management_code}')"` : '';
        tableEntries += `<tr management-code="${management_code}" ${details_html}><td>${typeIcon}</td><td>${notification_classes_html}</td><td>${tokens}</td><td>${last_td(n)}</td></tr>`;
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
    if (!notificationsSet) {
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
        let tokens = `<span class="badge badge-pill badge-primary">${n["subscribed_tokens"].length}</span>`;
        let notification_classes_icons = getNotificationClassIcon(n, "AT_creations") +
            getNotificationClassIcon(n, "subtoken_creations") +
            getNotificationClassIcon(n, "setting_changes") +
            getNotificationClassIcon(n, "security");
        let notification_classes_str = n["notification_classes"].join(", ");
        let notification_classes_html = `<span data-toggle="tooltip"
              data-placement="bottom" title=""
              data-original-title="${notification_classes_str}">${notification_classes_icons}</span>`
        let checker = `<input class="form-check-input notifications-checkbox ml-0" type="checkbox" value="${management_code}" instance-prefix="${prefix}">`;
        tableEntries += `<tr management-code="${management_code}"><td>${checker}</td><td>${notification_classes_html}</td><td>${tokens}</td></tr>`;
    })
    $(prefixId("subscribe-mail-existing-notifications", prefix)).html(tableEntries);

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
        return;
    }
    $container.showB();
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
    $tokeninfoCalendarListing(notificationPrefix).showB();
    fillNotificationAndCalendarInfo(cals, nots, $(prefixId("not-info-body", notificationPrefix)), notificationPrefix);

    if (!doesExpire) {
        $tokeninfoCalendarListing(notificationPrefix).hideB();
    } else {
        notificationModalInitSubscribeCalendars(cals);
        notificationModalInitEntryInvite();
    }

    $notificationsModal.modal();
}

function notificationModalInitSubscribeCalendars(cals) {
    let all_cals = cachedNotificationsTypeData["calendars"];
    let options = ``;
    let filtered = cals === undefined ? all_cals : all_cals.filter(c => !cals.some(cc => c["name"] === cc["name"]));
    filtered.forEach(function (c) {
        let name = c["name"];
        options += `<option value=${name}>${name}</option>`;
    });
    options += `<option value="${new_calendar_option_value}" class="text-secondary">New calendar ...</option>`;
    $calendarSelector.html(options);
    $calendarSelector.trigger('change');
}

function notificationModalInitEntryInvite() {
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
                return;
            }
            if (email_verified === undefined || !email_verified) {
                $('.email-not-verified-hint').showB();
                cachedNotificationsTypeData["email_ok"] = false;
                return;
            }
            $('.notify-entry-content-normal').showB();
            cachedNotificationsTypeData["email_ok"] = true;
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
    $.ajax({
        type: "GET",
        url: storageGet('notifications_endpoint') + "/calendars",
        success: function (res) {
            let cals = res["calendars"];
            cals.forEach(function (c) {
                let tokens = c["subscribed_tokens"] || [];
                tokens.forEach(function (momid) {
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
                fillCalendarInfo(cals, true, notificationPrefix);
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