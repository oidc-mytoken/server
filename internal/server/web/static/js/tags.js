function createTags(tags, includeDeleteBtn = false, deleteCallbackStr = "noop", additionalParamsStr = "") {
    if (!tags) {
        return "";
    }
    return tags.map(function (tag) {
        return getTagPill(tag, includeDeleteBtn, deleteCallbackStr, additionalParamsStr);
    }).join("")
}

function noop() {
}

function getTagPill(tag, includeDeleteBtn = false, deleteCallbackStr = "", additionalParamsStr = "") {
    const textClass = textClassForBackgroundColor(tag.color);
    let deleteBtn = "";
    if (includeDeleteBtn && loggedIn) {
        deleteBtn = `<button class="btn tag-btn ${textClass}" type="button" onclick="${deleteCallbackStr}('${tag.tag}'${additionalParamsStr})"><i class="fas fa-trash"></i></button>`;
    }
    return `<span 
                class="badge badge-pill ${textClass} tag mr-1" 
                data-tag="${tag.tag}"
                style="background-color: #${tag.color};">
                ${tag.tag}
                ${deleteBtn}
            </span>`;
}

function textClassForBackgroundColor(backgroundColor) {
    const r = parseInt(backgroundColor.substring(0, 2), 16);
    const g = parseInt(backgroundColor.substring(2, 4), 16);
    const b = parseInt(backgroundColor.substring(4, 6), 16);

    // Convert to relative luminance (per WCAG)
    function toLuminance(c) {
        let s = c / 255;
        return s <= 0.03928 ? s / 12.92 : Math.pow((s + 0.055) / 1.055, 2.4);
    }

    let L = 0.2126 * toLuminance(r) + 0.7152 * toLuminance(g) + 0.0722 * toLuminance(b);

    return L > 0.5 ? "text-black" : "text-white";
}

const new_tag_option_value = "$$$new-tag$$$";

function newTagSelectChange() {
    let tag = $('#add-tag-selector').val();
    if (tag === new_tag_option_value) {
        $('#add-new-tag-content').showB();
    } else {
        $('#add-new-tag-content').hideB();
    }
}

function newTagSelectSubmit(callback) {
    let tag = $('#add-tag-selector').val();
    if (tag === new_tag_option_value) {
        tag = $('#add-new-tag-input').val();
    }
    callback(tag);
}

function showAddTagModal(callback) {
    if ($('#add-tag-modal').length === 0) {
        $('body').append(`
        <div class="modal fade" id="add-tag-modal" tabindex="-1" role="dialog"
            aria-labelledby="new-notifications-modal-title"
            aria-hidden="true">
            <div class="modal-dialog modal-dialog-centered modal-md" role="document">
                <div class="modal-content bg-my_grey">
                    <div class="modal-header">
                        <h5 class="modal-title" id="new-notifications-modal-title">
                            <span>Add Tag</span>
                        </h5>
                        <button type="button" class="close" data-dismiss="modal" aria-label="Close">
                            <span aria-hidden="true">&times;</span>
                        </button>
                    </div>
                    <div class="modal-body">
                        <select class="form-control custom-select form-inline" id="add-tag-selector">
${loadedTags.reduce((acc, tag) => acc + `<option value="${tag.tag}">${tag.tag}</option>`, "")}
                            <option class="text-secondary" value="${new_tag_option_value}">New Tag</option>
                        </select>
                        <div id="add-new-tag-content" class="input-group d-none mt-1">
                            <input class="form-control" type="text" placeholder="Tag" id="add-new-tag-input">
                        </div>
                        <button class="btn btn-success mt-2 add-tag-submit-btn" id="add-tag-submit-btn">
                            Add Tag <i class="fas fa-tag"></i>
                        </button>
                    </div>
                </div>
            </div>
        </div>`);
    }

    $('#add-tag-modal').modal('show');

    $(document).off('change', '#add-tag-selector').on('change', '#add-tag-selector', newTagSelectChange);
    $(document).off('click', '#add-tag-submit-btn').on('click', '#add-tag-submit-btn', function () {
        newTagSelectSubmit(callback);
    });

    newTagSelectChange();
}

let loadedTags = [];

function getTagList(...next) {
    $.ajax({
        type: "GET",
        url: `${storageGet('usersettings_endpoint')}/tags`,
        success: function (res) {
            let tags = res["tags"];
            if (tags === undefined || tags === null) {
                loadedTags = [];
            } else {
                loadedTags = tags;
            }
            doNext(...next);
        },
        error: function (errRes) {
            $settingsErrorModalMsg.text(getErrorMessage(errRes));
            $settingsErrorModal.modal();
        },
    });
}
