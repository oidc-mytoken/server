const $emailInput = $('#email-input');
const $emailVerifiedIcon = $('#email-verified-icon');
const $emailUnverifiedIcon = $('#email-unverified-icon');
const $preferredMimeTypeHTML = $('#preferred_mimetype_html');
const $preferredMimeTypeText = $('#preferred_mimetype_txt');
const $editMailBtn = $('#edit-mail-btn');
const $saveMailBtn = $('#save-mail-btn');


function $calendarTable(prefix = "") {
    return $(prefixId('calendars', prefix));
}

function $noCalendarsEntry(prefix = "") {
    return $(prefixId('noCalendars', prefix));
}

let settingsStatus = {"email_data_obtained": false, "calendars_loaded": false};

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


$('#calendar-trigger-btn').on('click', function () {
    if (!settingsStatus["calendars_loaded"]) {
        loadCalendars(settingsPrefix);
    }
    $('#notifications').toggleClass("break-out");
})

function loadCalendars(prefix = "") {
    clearCalendarTable(prefix);
    $.ajax({
        type: "GET",
        url: storageGet('notifications_endpoint') + "/calendars",
        success: function (res) {
            let cals = res["calendars"];
            if (cals === undefined || cals === null) {
                $noCalendarsEntry(prefix).showB();
            } else {
                cals.forEach(function (cal) {
                    addCalendarToTable(cal, prefix);
                })
                settingsStatus["calendars_loaded"] = true;
            }
        },
        error: function (errRes) {
            $settingsErrorModalMsg.text(getErrorMessage(errRes));
            $settingsErrorModal.modal();
        },
    });
}

function actionsCalendarHtml(prefix = "", calID = "", description = "") {
    return `<td>
                <button class="btn" role="button" onclick="deleteCalendar(this, '${prefix}')"><i class="fas fa-trash-alt text-danger"></i></button>
            </td>`;
}

function clearCalendarTable(prefix = "") {
    $calendarTable(prefix).find('tr.calendar-entry').remove();
}

function calendarIDFromICSPath(ics_path) {
    if (!ics_path) {
        return "";
    }
    let parts = ics_path.split('/');
    return parts[parts.length - 1];
}

function addCalendarToTable(cal, prefix = "", editable = true) {
    $noCalendarsEntry(prefix).hideB();
    let tags = cal['tags'];
    let ics_path = cal['ics_path'];
    let description = cal['description'] || '';
    let calID = calendarIDFromICSPath(ics_path);
    let viewCalendarHtml = `<td><a href="${ics_path}/view"><i class="fas fa-calendar-alt"></i></a></td>`;
    let tagsHtml;
    let descriptionHtml;
    if (editable) {
        tagsHtml = `${createTags(tags, true, "removeTagFromCalendar", `, '${calID}', '${prefix}'`)} <span class="badge badge-pill badge-success tag"><button class="btn tag-btn" type="button" onclick="addTagToCalendar('${calID}', '${prefix}')"><i class="fas fa-plus-circle"></i></button></span>`;
        descriptionHtml = `<td class="cal-desc">
            <div class="input-group">
                <textarea class="form-control desc-input" id="desc-input-${calID}" rows="2" disabled>${escapeHTML(description)}</textarea>
                <div class="input-group-append">
                    <button class="btn btn-outline-secondary edit-desc-btn" id="edit-desc-btn-${calID}" type="button" onclick="editCalendarDescription('${calID}', '${prefix}')"><i class="fas fa-pencil-alt"></i></button>
                    <button class="btn btn-outline-secondary save-desc-btn d-none" id="save-desc-btn-${calID}" type="button" onclick="saveCalendarDescription('${calID}', '${prefix}')"><i class="fas fa-save"></i></button>
                </div>
            </div>
        </td>`;
    } else {
        // Read-only view: just display tags and description without edit controls
        tagsHtml = createTags(tags, false) || '<span class="text-muted">-</span>';
        descriptionHtml = `<td class="cal-desc">${escapeHTML(description) || '<span class="text-muted">-</span>'}</td>`;
    }
    const actionsHtml = editable ? actionsCalendarHtml(prefix, calID, description) : "";
    const html = `<tr class="calendar-entry" data-cal-id="${calID}">${viewCalendarHtml}<td><a href="${ics_path}" target="_blank" rel="noopener noreferrer">${ics_path}</a></td>${descriptionHtml}<td class="cal-tags">${tagsHtml}</td>${actionsHtml}</tr>`;
    $calendarTable(prefix).prepend(html);
}

function escapeHTML(s) {
    return $('<div>').text(s || '').html();
}

function escapeQuotes(s) {
    return (s || '').replaceAll('"', '&quot;').replaceAll("'", "&#39;");
}

function addTagToCalendar(calID, prefix = "") {
    const openModal = function () {
        showAddTagModal(function (tag) {
            if (tag && tag.trim() !== '') {
                sendAddTagToCalendarRequest(calID, tag, prefix);
            }
        });
    }
    getTagList(openModal);
}

function sendAddTagToCalendarRequest(calID, tag, prefix = "") {
    let data = {
        "tag": tag
    };
    data = JSON.stringify(data);

    $.ajax({
        type: "POST",
        data: data,
        dataType: "json",
        contentType: "application/json",
        url: storageGet('notifications_endpoint') + "/calendars/" + calID + "/tags",
        success: function () {
            $('#add-tag-modal').modal('hide');
            loadCalendars(prefix);
        },
        error: function (errRes) {
            $settingsErrorModalMsg.text(getErrorMessage(errRes));
            $settingsErrorModal.modal();
        },
    });
}

function removeTagFromCalendar(tag, calID, prefix = "") {
    let data = {
        "tag": tag
    };
    data = JSON.stringify(data);

    $.ajax({
        type: "DELETE",
        data: data,
        dataType: "json",
        contentType: "application/json",
        url: storageGet('notifications_endpoint') + "/calendars/" + calID + "/tags",
        success: function () {
            loadCalendars(prefix);
        },
        error: function (errRes) {
            $settingsErrorModalMsg.text(getErrorMessage(errRes));
            $settingsErrorModal.modal();
        },
    });
}

function editCalendarDescription(calID, prefix = "") {
    $(`#edit-desc-btn-${calID}`).addClass('d-none');
    $(`#save-desc-btn-${calID}`).removeClass('d-none');
    $(`#desc-input-${calID}`).prop("disabled", false);
    $(`#desc-input-${calID}`).focus();
}

function saveCalendarDescription(calID, prefix = "") {
    const description = $(`#desc-input-${calID}`).val();

    let data = {
        "description": description
    };
    data = JSON.stringify(data);

    $.ajax({
        type: "PUT",
        data: data,
        dataType: "json",
        contentType: "application/json",
        url: storageGet('notifications_endpoint') + "/calendars/" + calID,
        success: function (res) {
            $(`#save-desc-btn-${calID}`).addClass('d-none');
            $(`#edit-desc-btn-${calID}`).removeClass('d-none');
            $(`#desc-input-${calID}`).prop("disabled", true);
        },
        error: function (errRes) {
            $settingsErrorModalMsg.text(getErrorMessage(errRes));
            $settingsErrorModal.modal();
        },
    });
}

// Add this to handle Ctrl+Enter key press in description textarea
$(document).on('keydown', '.desc-input', function (e) {
    if (e.ctrlKey && e.keyCode === 13) { // Ctrl+Enter
        e.preventDefault();
        const calID = $(this).closest('tr.calendar-entry').data('cal-id');
        saveCalendarDescription(calID, '');
    }
});

function deleteCalendar(el, prefix = "") {
    const $tr = $(el).closest('tr.calendar-entry');
    const calID = $tr.data('cal-id');
    const icsLink = $tr.find('td:nth-child(2) a').attr('href') || calID;
    $(
        `<div class="modal fade" tabindex="-1" role="dialog">
            <div class="modal-dialog modal-dialog-centered modal-lg" role="document">
                <div class="modal-content">
                    <div class="modal-header">
                        <h5 class="modal-title">Delete Calendar</h5>
                        <button type="button" class="close" data-dismiss="modal" aria-label="Close">
                            <span aria-hidden="true">&times;</span>
                        </button>
                    </div>
                    <div class="modal-body">Confirm to delete the calendar '${icsLink}'.</div>
                    <div class="modal-footer">
                        <button type="button" class="btn btn-secondary" data-dismiss="modal">Cancel</button>
                        <button type="button" class="btn btn-danger" data-dismiss="modal" onclick="sendDeleteCalendarRequest('${calID}', '${prefix}')">Delete</button>
                    </div>
                </div>
            </div>
        </div>`
    ).modal();
}

function sendDeleteCalendarRequest(calID, prefix = "") {
    $.ajax({
        type: "DELETE",
        url: storageGet('notifications_endpoint') + "/calendars/" + calID,
        success: function () {
            loadCalendars(prefix);
        },
        error: function (errRes) {
            $settingsErrorModalMsg.text(getErrorMessage(errRes));
            $settingsErrorModal.modal();
        },
    });
}

$('.new-calendar-btn').on('click', function () {
    addCalendar(extractPrefix('new-calendar-btn', $(this).attr('id')));
});


function newCalendarTagSelectChange() {
    let tag = $('#add-calendar-tag-selector').val();
    if (tag === new_tag_option_value) {
        $('#add-calendar-new-tag-content').showB();
    } else {
        $('#add-calendar-new-tag-content').hideB();
    }
}

let newCalendarTags = [];
let reloadTags = false;
let $addCalendarModal;

function addCalendar(prefix = "") {
    newCalendarTags = [];
    let proceed = function () {
        if ($addCalendarModal) {
            $addCalendarModal.remove();
        }
        $addCalendarModal = $(
            `<div class="modal fade" tabindex="-1" role="dialog" id="new-calendar-modal">
            <div class="modal-dialog modal-dialog-centered modal-lg" role="document">
                <div class="modal-content">
                    <div class="modal-header">
                        <h5 class="modal-title">Create New Calendar</h5>
                        <button type="button" class="close" data-dismiss="modal" aria-label="Close">
                            <span aria-hidden="true">&times;</span>
                        </button>
                    </div>
                    <div class="modal-body">
                        <div class="form-group">
                            <label for="calendar-desc-input">Description</label>
                            <textarea class="form-control" id="calendar-desc-input" rows="3" placeholder="Calendar description (optional)"></textarea>
                        </div>
                        <div class="form-group">
                            <label>Tags</label>
                            <div id="new-calendar-tags"></div>
                            <div>
                            <small>Mytokens with a common tag are automatically added to the calendar.</small>
                            </div>
                            
                              <select class="form-control custom-select form-inline" id="add-calendar-tag-selector" onchange="newCalendarTagSelectChange()">
${loadedTags.reduce((acc, tag) => acc + `<option value="${tag.tag}">${tag.tag}</option>`, "")}
                            <option class="text-secondary" value="${new_tag_option_value}">New Tag</option>
                        </select>
                        <div id="add-calendar-new-tag-content" class="input-group d-none mt-1">
                            <input class="form-control" type="text" placeholder="Tag" id="add-calendar-new-tag-input">
                        </div>
                            <button class="btn btn-sm btn-success mt-1" type="button" onclick="addTagToNewCalendar()">Add Tag <i class="fas fa-tag"></i></button>
                        </div>
                    </div>
                    <div class="modal-footer">
                        <button type="button" class="btn btn-secondary" data-dismiss="modal">Cancel</button>
                        <button type="button" class="btn btn-success" data-dismiss="modal" onclick="sendCreateCalendarRequest('${prefix}')">Create <i class="fas fa-calendar-alt"></i></button>
                    </div>
                </div>
            </div>
        </div>`
        ).modal();
        updateNewCalendarTagsDisplay();
    }
    if (reloadTags || !loadedTags || loadedTags.length === 0) {
        getTagList(proceed);
    } else {
        proceed();
    }
}

function updateNewCalendarTagsDisplay() {
    $('#new-calendar-tags').html(createTags(newCalendarTags, true, 'removeTagFromNewCalendar'));
}

function addTagToNewCalendar() {
    let tag = $('#add-calendar-tag-selector').val();
    if (tag === new_tag_option_value) {
        tag = $('#add-calendar-new-tag-input').val();
        reloadTags = true;
    }
    if (!newCalendarTags.find(t => t.tag === tag)) {
        let info = loadedTags.find(t => t.tag === tag) || {tag: tag, color: '888888'};
        newCalendarTags.push(info);
        updateNewCalendarTagsDisplay();
    }

}

function removeTagFromNewCalendar(tag) {
    newCalendarTags = newCalendarTags.filter(t => t.tag !== tag);
    updateNewCalendarTagsDisplay();
}

function sendCreateCalendarRequest(prefix = "") {
    const description = $('#calendar-desc-input').val();
    const tags = newCalendarTags.map(t => t.tag);
    let data = JSON.stringify({"description": description, "tags": tags});
    $.ajax({
        type: "POST",
        data: data,
        dataType: "json",
        contentType: "application/json",
        url: storageGet('notifications_endpoint') + "/calendars",
        success: function () {
            loadCalendars(prefix);
        },
        error: function (errRes) {
            $settingsErrorModalMsg.text(getErrorMessage(errRes));
            $settingsErrorModal.modal();
        },
    });
}
