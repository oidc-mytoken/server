function $tagTable(prefix = "") {
    return $(prefixId('tags', prefix));
}

function $noTagsEntry(prefix = "") {
    return $(prefixId('noTags', prefix));
}

function sendCreateTagRequest(prefix = "") {
    const tag = $(prefixId("tag-name-input", prefix)).val();
    $.ajax({
        type: "POST",
        contentType: "application/json",
        url: `${storageGet('usersettings_endpoint')}/tags/${tag}`,
        success: function () {
            settingsStatus["calendars_loaded"] = false;
            loadTags(prefix);
        },
        error: function (errRes) {
            $settingsErrorModalMsg.text(getErrorMessage(errRes));
            $settingsErrorModal.modal();
        },
    });
}

function sendDeleteTagRequest(tag, prefix = "") {
    $.ajax({
        type: "DELETE",
        contentType: "application/json",
        url: `${storageGet('usersettings_endpoint')}/tags/${tag}`,
        success: function () {
            settingsStatus["calendars_loaded"] = false;
            loadTags(prefix);
        },
        error: function (errRes) {
            $settingsErrorModalMsg.text(getErrorMessage(errRes));
            $settingsErrorModal.modal();
        },
    });
}

$('#tags-tab').on('shown.bs.tab', function (e) {
    e.preventDefault();
    loadTags(settingsPrefix);
    return false;
});

function loadTags(prefix = "") {
    clearTagTable(prefix);
    $.ajax({
        type: "GET",
        url: `${storageGet('usersettings_endpoint')}/tags`,
        success: function (res) {
            let tags = res["tags"];
            if (tags === undefined || tags === null) {
                $noTagsEntry(prefix).showB();
            } else {
                tags.forEach(function (tag) {
                    addTagToTable(tag, prefix);
                })
                settingsStatus["tags_loaded"] = true;
            }
        },
        error: function (errRes) {
            $settingsErrorModalMsg.text(getErrorMessage(errRes));
            $settingsErrorModal.modal();
        },
    });
}


function deleteTagHtml(tag, prefix = "") {
    return `<td><button class="btn" role="button" onclick="deleteTagModal('${tag}', '${prefix}')"><i class="fas fa-trash-alt text-danger"></i></button></td>`;
}

function clearTagTable(prefix = "") {
    $tagTable(prefix).find('tr.tag-entry').remove();
}

function addTagToTable(tag, prefix = "", with_delete = true) {
    $noTagsEntry(prefix).hideB();
    const html = `<tr class="tag-entry"><td>${getTagPill(tag)}</td><td>${tagNameInput(tag.tag, prefix)}</td><td>${tagColor(tag, prefix)}</td>${with_delete ? deleteTagHtml(tag.tag, prefix) : ""}</tr>`;
    $tagTable(prefix).prepend(html);
}

function tagColor(tag, prefix = "") {
    return `<input type="color" onchange="changeTagColor(this, '${tag.tag}', '${prefix}')" value="#${tag.color}">`
}

function tagNameInput(tag, prefix = "") {
    return `<input type="text" onChange="changeTagName(this, '${tag}', '${prefix}')" value="${tag}">`
}

function deleteTagModal(tag, prefix = "") {
    $(`
        <div class="modal fade" tabindex="-1" role="dialog">
            <div class="modal-dialog modal-dialog-centered modal-lg" role="document">
                <div class="modal-content">
                    <div class="modal-header">
                        <h5 class="modal-title">Delete Tag</h5>
                        <button type="button" class="close" data-dismiss="modal" aria-label="Close">
                            <span aria-hidden="true">&times;</span>
                        </button>
                    </div>
                    <div class="modal-body">Confirm to delete the tag '${tag}' and removing it from all mytokens / calendars / notifications.</div>
                    <div class="modal-footer">
                        <button type="button" class="btn btn-secondary" data-dismiss="modal">Cancel</button>
                        <button type="button" class="btn btn-danger" data-dismiss="modal" onclick="sendDeleteTagRequest('${tag}', '${prefix}')">Delete</button>
                    </div>
                </div>
            </div>
        </div>
   `).modal();
}

function changeTagName(el, oldName, prefix = "") {
    const newName = $(el).val();
    updateTag(oldName, {"tag": newName}, prefix);
}

function changeTagColor(el, name, prefix = "") {
    const color = $(el).val().slice(1);
    updateTag(name, {"color": color}, prefix);
}

function updateTag(tag, data, prefix = "") {
    $.ajax({
        type: "PUT",
        data: JSON.stringify(data),
        dataType: "json",
        contentType: "application/json",
        url: `${storageGet('usersettings_endpoint')}/tags/${tag}`,
        success: function () {
            settingsStatus["calendars_loaded"] = false;
            loadTags(prefix);
        },
        error: function (errRes) {
            $settingsErrorModalMsg.text(getErrorMessage(errRes));
            $settingsErrorModal.modal();
        },
    });
}