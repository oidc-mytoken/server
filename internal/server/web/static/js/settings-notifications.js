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

function deleteCalendarHtml(prefix = "") {
    return `<td><a href="#" role="button" onclick="deleteCalendar(this, '${prefix}')"><i class="fas fa-trash-alt text-danger"></i></a></td>`;
}

function clearCalendarTable(prefix = "") {
    $calendarTable(prefix).find('tr.calendar-entry').remove();
}

function addCalendarToTable(cal, prefix = "", with_delete = true) {
    $noCalendarsEntry(prefix).hideB();
    let name = cal['name'];
    let ics_path = cal['ics_path'];
    let viewCalendarHtml = `<td><a href="${ics_path}/view"><i class="fas fa-calendar-alt"></i></a></td>`;
    const html = `<tr class="calendar-entry"><td>${name}</td>${viewCalendarHtml}<td><a href="${ics_path}" target="_blank" rel="noopener noreferrer">${ics_path}</a></td>${with_delete ? deleteCalendarHtml(prefix) : ""}</tr>`;
    $calendarTable(prefix).prepend(html);
}

function deleteCalendar(el, prefix = "") {
    const name = $(el).parent().siblings()[0].innerHTML;
    $(`
        <div class="modal fade" tabindex="-1" role="dialog">
            <div class="modal-dialog modal-dialog-centered modal-lg" role="document">
                <div class="modal-content">
                    <div class="modal-header">
                        <h5 class="modal-title">Delete Calendar</h5>
                        <button type="button" class="close" data-dismiss="modal" aria-label="Close">
                            <span aria-hidden="true">&times;</span>
                        </button>
                    </div>
                    <div class="modal-body">Confirm to delete the calendar '${name}'.</div>
                    <div class="modal-footer">
                        <button type="button" class="btn btn-secondary" data-dismiss="modal">Cancel</button>
                        <button type="button" class="btn btn-danger" data-dismiss="modal" onclick="sendDeleteCalendarRequest('${name}', '${prefix}')">Delete</button>
                    </div>
                </div>
            </div>
        </div>
   `).modal();
}

function sendDeleteCalendarRequest(name, prefix = "") {
    $.ajax({
        type: "DELETE",
        url: storageGet('notifications_endpoint') + "/calendars/" + name,
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

function addCalendar(prefix = "") {
    $(`
        <div class="modal fade" tabindex="-1" role="dialog">
            <div class="modal-dialog modal-dialog-centered modal-lg" role="document">
                <div class="modal-content">
                    <div class="modal-header">
                        <h5 class="modal-title">New Calendar</h5>
                        <button type="button" class="close" data-dismiss="modal" aria-label="Close">
                            <span aria-hidden="true">&times;</span>
                        </button>
                    </div>
                    <div class="modal-body">
                    Create a new calendar with the name:
                    <input type="text" class="form-control" id="calendar-name-input">
                    </div>
                    <div class="modal-footer">
                        <button type="button" class="btn btn-secondary" data-dismiss="modal">Cancel</button>
                        <button type="button" class="btn btn-success" data-dismiss="modal" onclick="sendCreateCalendarRequest('${prefix}')">Create <i class="fas fa-calendar-alt"></i></button>
                    </div>
                </div>
            </div>
        </div>
   `).modal();
}

function sendCreateCalendarRequest(prefix = "") {
    let data = JSON.stringify({"name": $('#calendar-name-input').val()});
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
