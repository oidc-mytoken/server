// Tags for Create Mytoken form

// Store selected tags for the create mytoken form
let createMTSelectedTags = [];

function $createMTTagsContainer(prefix = "") {
    return $(prefixId('create-mt-tags-container', prefix));
}

function $noCreateMTTags(prefix = "") {
    return $(prefixId('no-create-mt-tags', prefix));
}

function $addCreateMTTagBtn(prefix = "") {
    return $(prefixId('add-create-mt-tag-btn', prefix));
}

function initCreateMTTags(prefix = "") {
    createMTSelectedTags = [];
    renderCreateMTTags(prefix);

    $addCreateMTTagBtn(prefix).off('click').on('click', function () {
        showAddTagModalForCreateMT(prefix);
    });
}

function renderCreateMTTags(prefix = "") {
    const $container = $createMTTagsContainer(prefix);
    const $noTags = $noCreateMTTags(prefix);

    // Remove existing tag pills (but keep the "no tags" message)
    $container.find('.create-mt-tag-pill').remove();

    if (createMTSelectedTags.length === 0) {
        $noTags.showB();
    } else {
        $noTags.hideB();
        createMTSelectedTags.forEach(function (tagInfo, index) {
            const pill = createMTTagPill(tagInfo, index, prefix);
            $container.append(pill);
        });
    }
}

function createMTTagPill(tagInfo, index, prefix = "") {
    const tag = tagInfo.tag;
    const includeChildren = tagInfo.include_children;
    const tagData = loadedTags.find(t => t.tag === tag) || {tag: tag, color: generateTagColor(tag)};
    const textClass = textClassForBackgroundColor(tagData.color);

    let childrenBadge = "";
    if (includeChildren) {
        childrenBadge = `<i class="fas fa-sitemap ml-1" title="Includes children"></i>`;
    }

    return `<span class="badge badge-pill ${textClass} tag mr-1 create-mt-tag-pill" 
                data-tag="${tag}" 
                data-index="${index}"
                style="background-color: #${tagData.color};">
                ${tag}${childrenBadge}
                <button class="btn tag-btn ${textClass}" type="button" 
                    onclick="removeCreateMTTag(${index}, '${prefix}')">
                    <i class="fas fa-times"></i>
                </button>
            </span>`;
}

function removeCreateMTTag(index, prefix = "") {
    createMTSelectedTags.splice(index, 1);
    renderCreateMTTags(prefix);
}

function addCreateMTTag(tag, includeChildren = false, prefix = "") {
    // Check if tag already exists
    if (createMTSelectedTags.some(t => t.tag === tag)) {
        return;
    }

    createMTSelectedTags.push({
        tag: tag,
        include_children: includeChildren
    });
    renderCreateMTTags(prefix);
}

function getCreateMTTags() {
    return createMTSelectedTags;
}

function clearCreateMTTags(prefix = "") {
    createMTSelectedTags = [];
    renderCreateMTTags(prefix);
}

// Generate a deterministic color for a tag based on its name
function generateTagColor(tagName) {
    let hash = 0;
    for (let i = 0; i < tagName.length; i++) {
        hash = tagName.charCodeAt(i) + ((hash << 5) - hash);
    }
    // Convert to 6 digit hex
    let color = '';
    for (let i = 0; i < 3; i++) {
        const value = (hash >> (i * 8)) & 0xFF;
        color += ('00' + value.toString(16)).slice(-2);
    }
    return color;
}

function showAddTagModalForCreateMT(prefix = "") {
    // Remove existing modal if any
    $('#create-mt-add-tag-modal').remove();

    // Build options from loadedTags
    let tagOptions = loadedTags.map(function (tag) {
        // Skip already selected tags
        if (createMTSelectedTags.some(t => t.tag === tag.tag)) {
            return '';
        }
        return `<option value="${tag.tag}">${tag.tag}</option>`;
    }).join('');

    $('body').append(`
        <div class="modal fade" id="create-mt-add-tag-modal" tabindex="-1" role="dialog">
            <div class="modal-dialog modal-dialog-centered modal-md" role="document">
                <div class="modal-content bg-my_grey">
                    <div class="modal-header">
                        <h5 class="modal-title">Add Tag to Mytoken</h5>
                        <button type="button" class="close" data-dismiss="modal" aria-label="Close">
                            <span aria-hidden="true">&times;</span>
                        </button>
                    </div>
                    <div class="modal-body">
                        <div class="form-group">
                            <label for="create-mt-tag-selector">Select Tag</label>
                            <select class="form-control custom-select" id="create-mt-tag-selector">
                                ${tagOptions}
                                <option class="text-secondary" value="${new_tag_option_value}">+ Create New Tag</option>
                            </select>
                        </div>
                        <div id="create-mt-new-tag-content" class="form-group d-none">
                            <label for="create-mt-new-tag-input">New Tag Name</label>
                            <input class="form-control" type="text" placeholder="Enter tag name" id="create-mt-new-tag-input">
                        </div>
                        <div class="form-group form-check">
                            <input type="checkbox" class="form-check-input" id="create-mt-tag-include-children">
                            <label class="form-check-label" for="create-mt-tag-include-children">
                                Include children <i class="fas fa-sitemap"></i>
                            </label>
                            <small class="form-text text-muted">If checked, subtokens created from this mytoken will inherit this tag.</small>
                        </div>
                    </div>
                    <div class="modal-footer">
                        <button type="button" class="btn btn-secondary" data-dismiss="modal">Cancel</button>
                        <button type="button" class="btn btn-success" id="create-mt-add-tag-submit">
                            Add Tag <i class="fas fa-tag"></i>
                        </button>
                    </div>
                </div>
            </div>
        </div>`);

    const $modal = $('#create-mt-add-tag-modal');
    const $selector = $('#create-mt-tag-selector');
    const $newTagContent = $('#create-mt-new-tag-content');
    const $newTagInput = $('#create-mt-new-tag-input');
    const $includeChildren = $('#create-mt-tag-include-children');
    const $submitBtn = $('#create-mt-add-tag-submit');

    // Handle selector change
    $selector.on('change', function () {
        if ($(this).val() === new_tag_option_value) {
            $newTagContent.showB();
        } else {
            $newTagContent.hideB();
        }
    });

    // Handle submit
    $submitBtn.on('click', function () {
        let tag = $selector.val();
        if (tag === new_tag_option_value) {
            tag = $newTagInput.val().trim();
            if (tag === '') {
                return;
            }
        }
        const includeChildren = $includeChildren.is(':checked');
        addCreateMTTag(tag, includeChildren, prefix);
        $modal.modal('hide');
    });

    $modal.modal('show');

    // Trigger change to set initial state
    $selector.trigger('change');
}

// Set tags from external data (e.g., when filling from profile)
function setCreateMTTagsFromData(tags, prefix = "") {
    createMTSelectedTags = [];
    if (tags && Array.isArray(tags)) {
        tags.forEach(function (tag) {
            if (typeof tag === 'string') {
                createMTSelectedTags.push({tag: tag, include_children: false});
            } else if (typeof tag === 'object' && tag.tag) {
                createMTSelectedTags.push({
                    tag: tag.tag,
                    include_children: tag.include_children || false
                });
            }
        });
    }
    renderCreateMTTags(prefix);
}
