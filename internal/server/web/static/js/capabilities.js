const $createMTCaps = $('.capability-check[value=create_mytoken]');
const $allCapRWModes = $('.rw-cap-mode');

function capabilityChecks(prefix = "") {
    return $('.capability-check[instance-prefix="' + prefix + '"]');
}

function subtokenCapabilityChecks(prefix = "") {
    return $('.subtoken-capability-check[instance-prefix="' + prefix + '"]');
}

function capRWModes(prefix = "") {
    return $allCapRWModes.filter('[instance-prefix="' + prefix + '"]');
}

function capabilityCreateMytoken(prefix = "") {
    return $('#' + prefix + 'cp-create_mytoken');
}

function capabilityAT(prefix = "") {
    return $('#' + prefix + 'cp-AT');
}

function capSummaryAT(prefix = "") {
    return $('#' + prefix + 'cap-summary-AT');
}

function capSummaryMT(prefix = "") {
    return $('#' + prefix + 'cap-summary-MT');
}

function capSummaryInfo(prefix = "") {
    return $('#' + prefix + 'cap-summary-info');
}

function capSummarySettings(prefix = "") {
    return $('#' + prefix + 'cap-summary-settings');
}

function capSummaryHowManyGreen(prefix = "") {
    return $('#' + prefix + 'cap-summary-count-green');
}

function capSummaryHowManyYellow(prefix = "") {
    return $('#' + prefix + 'cap-summary-count-yellow');
}

function capSummaryHowManyRed(prefix = "") {
    return $('#' + prefix + 'cap-summary-count-red');
}

function subtokenCapabilities(prefix = "") {
    return $('#' + prefix + 'subtokenCapabilities');
}


const rPrefix = "read@";

$('.capability-check').click(function () {
    checkThisCapability.call(this);
    updateCapSummary(this.getAttribute('instance-prefix'));
})
$('.subtoken-capability-check').click(function () {
    checkThisSubCapability.call(this);
    updateCapSummary(this.getAttribute('instance-prefix'));
})


$createMTCaps.on("click", function () {
    let enabled = $(this).prop("checked");
    let prefix = extractPrefix("cp-create_mytoken", this.id);
    let $subtokenCapabilities = subtokenCapabilities(prefix);
    $subtokenCapabilities.toggleClass('d-none');
    let $capabilityCheck = $subtokenCapabilities.find('.subtoken-capability-check');
    $capabilityCheck.prop("disabled", !enabled);
});

function checkThisSubCapability() {
    _checkThisCapability.call(this, "subtoken-capability");
}

function checkThisCapability() {
    _checkThisCapability.call(this, "capability");
}

function _checkThisCapability(type_prefix) {
    let activated = $(this).prop('checked');
    let classCheck = '.' + type_prefix + '-check'
    $(this).closest('li.list-group-item').find(classCheck).prop('checked', activated);
    if (!activated) {
        $(this).parents('li.list-group-item').children('div').children('div').children(classCheck).prop('checked', false);
    }
}

function initCapabilities(prefix) {
    capabilityChecks(prefix).each(checkThisCapability);
    subtokenCapabilityChecks(prefix).each(checkThisSubCapability);
    if (!capabilityCreateMytoken(prefix).prop("checked")) {
        subtokenCapabilities(prefix).hideB();
    } else {
        subtokenCapabilities(prefix).showB();
    }
    updateCapSummary(prefix);
    capRWModes(prefix).trigger('update-change');
}

function getCheckedCapabilities(prefix = "") {
    return _getCheckedCapabilities(capabilityChecks(prefix), 'cp', prefix);
}

function getCheckedSubtokenCapabilities(prefix = "") {
    if (!capabilityCreateMytoken(prefix).prop("checked")) {
        return [];
    }
    return _getCheckedCapabilities(subtokenCapabilityChecks(prefix), 'sub-cp', prefix);
}

function _getCheckedCapabilities($checks, idPrefix, preprefix = "") {
    let caps = $checks.filter(':checked').map(function (_, el) {
        let v = $(el).val();
        let $rw = $('#' + escapeSelector(preprefix + idPrefix + '-' + rPrefix + v + '-mode'));
        if ($rw.length && !$rw.prop('checked')) {
            v = rPrefix + v;
        }
        return v;
    }).get();
    caps = caps.filter(filterCaps);
    return caps;
}

function filterCaps(c, i, caps) {
    for (let j = 0; j < caps.length; j++) {
        if (i === j) {
            continue;
        }
        let cc = caps[j];
        if (isChildCapability(c, cc)) {
            return false;
        }
    }
    return true;
}

function isChildCapability(a, b) {
    let aReadOnly = a.startsWith(rPrefix);
    let bReadOnly = b.startsWith(rPrefix);
    if (aReadOnly) {
        a = a.substring(rPrefix.length);
    }
    if (bReadOnly) {
        b = b.substring(rPrefix.length);
    }
    let aParts = a.split(':');
    let bParts = b.split(':');
    if (bReadOnly && !aReadOnly) {
        return false;
    }
    if (bParts.length > aParts.length) {
        return false;
    }
    for (let i = 0; i < bParts.length; i++) {
        if (aParts[i] !== bParts[i]) {
            return false;
        }
    }
    return true;
}

function getCheckedCapabilitesAndSubtokencapabilities(prefix = "") {
    return getCheckedCapabilities(prefix).concat(getCheckedSubtokenCapabilities(prefix));
}

function searchAllChecked(str, prefix = "") {
    let read = "read@" + str
    for (const c of getCheckedCapabilitesAndSubtokencapabilities(prefix)) {
        if (c === str || c === read) {
            return true;
        }
        if (c.startsWith(str + ":") || c.startsWith(read + ":")) {
            return true
        }
    }
    return false;
}

function updateCapSummary(prefix = "") {
    let at = capabilityAT(prefix).prop("checked") || $('#sub-cp-AT').prop("checked");
    let mt = capabilityCreateMytoken(prefix).prop("checked");
    let info = searchAllChecked("tokeninfo", prefix);
    let settings = searchAllChecked("settings", prefix);

    let all = [];
    $.merge(all, capabilityChecks(prefix));
    if (capabilityCreateMytoken(prefix).prop('checked')) {
        $.merge(all, subtokenCapabilityChecks(prefix));
    }
    let counter = {
        'green': {},
        'yellow': {},
        'red': {}
    }
    for (const c of all) {
        if (!c.checked) {
            continue;
        }
        let name = $(c).val();
        let $icon = $($(c).closest('li.list-group-item').find('i.fa-exclamation-circle').not('.d-none')[0]);
        if ($icon.hasClass('text-success')) {
            counter['green'][name] = 1;
        }
        if ($icon.hasClass('text-warning')) {
            counter['yellow'][name] = 1;
        }
        if ($icon.hasClass('text-danger')) {
            counter['red'][name] = 1;
        }
    }
    let greens = Object.keys(counter['green']).length;
    let yellows = Object.keys(counter['yellow']).length;
    let reds = Object.keys(counter['red']).length;

    capSummaryHowManyGreen(prefix).text(greens);
    capSummaryHowManyYellow(prefix).text(yellows);
    capSummaryHowManyRed(prefix).text(reds);
    capSummaryHowManyGreen(prefix).attr('data-original-title', `This mytoken has ${greens} normal capabilities.`);
    capSummaryHowManyYellow(prefix).attr('data-original-title', `This mytoken has ${yellows} powerful capabilities.`);
    capSummaryHowManyRed(prefix).attr('data-original-title', `This mytoken has ${reds} very powerful capabilities.`);

    capSummaryAT(prefix).removeClass("text-success");
    capSummaryMT(prefix).removeClass("text-success");
    capSummaryInfo(prefix).removeClass("text-success");
    capSummarySettings(prefix).removeClass("text-success");
    if (at) {
        capSummaryAT(prefix).addClass("text-success");
        capSummaryAT(prefix).attr('data-original-title', "This mytoken can be used to obtain OIDC Access Tokens.");
    } else {
        capSummaryAT(prefix).attr('data-original-title', "This mytoken cannot be used to obtain OIDC Access Tokens.");
    }
    if (mt) {
        capSummaryMT(prefix).addClass("text-success");
        capSummaryMT(prefix).attr('data-original-title', "This mytoken can be used to create sub-mytokens.");
    } else {
        capSummaryMT(prefix).attr('data-original-title', "This mytoken cannot be used to create sub-mytokens.");
    }
    if (info) {
        capSummaryInfo(prefix).addClass("text-success");
        capSummaryInfo(prefix).attr('data-original-title', "This mytoken can be used to obtain tokeninfo about itself.");
    } else {
        capSummaryInfo(prefix).attr('data-original-title', "This mytoken cannot be used to obtain tokeninfo about itself.");
    }
    if (settings) {
        capSummarySettings(prefix).addClass("text-success");
        capSummarySettings(prefix).attr('data-original-title', "This mytoken can be used to change settings.");
    } else {
        capSummarySettings(prefix).attr('data-original-title', "This mytoken cannot be used to change settings.");
    }
}

$allCapRWModes.on('update-change', function () {
    let write = $(this).prop('checked');
    if (write) {
        $(this).closest('span').attr('data-original-title', `Allows full access. Click to only allow read access.`);
        $(this).closest('div.flex-fill').find('.rw-cap-read').hideB();
        $(this).closest('div.flex-fill').find('.rw-cap-write').showB();
    } else {
        $(this).closest('span').attr('data-original-title', `Allows only read access. Click to allow full access.`);
        $(this).closest('div.flex-fill').find('.rw-cap-write').hideB();
        $(this).closest('div.flex-fill').find('.rw-cap-read').showB();
    }
});

$allCapRWModes.change(function () {
    let write = $(this).prop('checked');
    $(this).trigger('update-change');
    let $modes = $(this).closest('li.list-group-item').find('.rw-cap-mode');
    $modes.bootstrapToggle(write ? 'on' : 'off', true);
    $modes.trigger('update-change');
    if (!write) {
        let $p = $(this).parents('li.list-group-item').children('div').find('.rw-cap-mode');
        $p.bootstrapToggle('off', true);
        $p.trigger('update-change');
    }
    updateCapSummary(this.getAttribute("instance-prefix"));
});

function checkCapability(cap, typePrefix, prefix = "") {
    let rCap = cap.startsWith(rPrefix);
    if (rCap) {
        cap = cap.substring(rPrefix.length);
    }
    $('#' + prefix + typePrefix + '-' + escapeSelector(cap)).prop("checked", true);
    let $mode = $('#' + prefix + typePrefix + '-' + escapeSelector(rPrefix + cap) + '-mode');
    let disabled = $mode.prop('disabled');
    $mode.prop('disabled', false);
    $mode.bootstrapToggle(rCap ? 'off' : 'on');
    $mode.prop('disabled', disabled);
}