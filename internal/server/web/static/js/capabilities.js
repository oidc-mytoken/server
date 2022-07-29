const $capabilityChecks = $('.capability-check');
const $subtokenCapabilityChecks = $('.subtoken-capability-check');
const $capabilityCreateMytoken = $('#cp-create_mytoken');
const $capabilityAT = $('#cp-AT');
const $capSummaryAT = $('#cap-summary-AT');
const $capSummaryMT = $('#cap-summary-MT');
const $capSummaryInfo = $('#cap-summary-info');
const $capSummarySettings = $('#cap-summary-settings');
const $capSummaryHowManyGreen = $('#cap-summary-count-green');
const $capSummaryHowManyYellow = $('#cap-summary-count-yellow');
const $capSummaryHowManyRed = $('#cap-summary-count-red');
const $capRWModes = $('.rw-cap-mode');

const rPrefix = "read@";

$capabilityChecks.click(function () {
    checkThisCapability.call(this);
    updateCapSummary();
})
$subtokenCapabilityChecks.click(function () {
    checkThisSubCapability.call(this);
    updateCapSummary();
})


$capabilityCreateMytoken.on("click", function () {
    let enabled = $(this).prop("checked");
    let $capabilityCheck = $('.subtoken-capability-check');
    $capabilityCheck.prop("disabled", !enabled);
    $('#subtokenCapabilities').toggleClass('d-none');
});

function checkThisSubCapability() {
    _checkThisCapability.call(this, "subtoken-capability");
}

function checkThisCapability() {
    _checkThisCapability.call(this, "capability");
}

function _checkThisCapability(prefix) {
    let activated = $(this).prop('checked');
    let classCheck = '.' + prefix + '-check'
    $(this).closest('li.list-group-item').find(classCheck).prop('checked', activated);
    if (!activated) {
        $(this).parents('li.list-group-item').children('div').children('div').children(classCheck).prop('checked', false);
    }
}

$(document).ready(function () {
    $capabilityChecks.each(checkThisCapability);
    $subtokenCapabilityChecks.each(checkThisSubCapability);
    if (!$capabilityCreateMytoken.prop("checked")) {
        $('#subtokenCapabilities').hideB();
    }
    updateCapSummary();
    $capRWModes.trigger('update-change');
})

function getCheckedCapabilities() {
    return _getCheckedCapabilities($capabilityChecks, 'cp');
}

function getCheckedSubtokenCapabilities() {
    if (!$capabilityCreateMytoken.prop("checked")) {
        return [];
    }
    return _getCheckedCapabilities($subtokenCapabilityChecks, 'sub-cp');
}

function _getCheckedCapabilities($checks, idPrefix) {
    let caps = $checks.filter(':checked').map(function (_, el) {
        let v = $(el).val();
        let $rw = $('#' + escapeSelector(idPrefix + '-' + rPrefix + v + '-mode'));
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

function getCheckedCapabilitesAndSubtokencapabilities() {
    return getCheckedCapabilities().concat(getCheckedSubtokenCapabilities());
}

function searchAllChecked(str) {
    let read = "read@" + str
    for (const c of getCheckedCapabilitesAndSubtokencapabilities()) {
        if (c === str || c === read) {
            return true;
        }
        if (c.startsWith(str + ":") || c.startsWith(read + ":")) {
            return true
        }
    }
    return false;
}

function updateCapSummary() {
    let at = $capabilityAT.prop("checked") || $('#sub-cp-AT').prop("checked");
    let mt = $capabilityCreateMytoken.prop("checked");
    let info = searchAllChecked("tokeninfo");
    let settings = searchAllChecked("settings");

    let all = [];
    $.merge(all, $capabilityChecks);
    if ($capabilityCreateMytoken.prop('checked')) {
        $.merge(all, $subtokenCapabilityChecks);
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

    $capSummaryHowManyGreen.text(greens);
    $capSummaryHowManyYellow.text(yellows);
    $capSummaryHowManyRed.text(reds);
    $capSummaryHowManyGreen.attr('data-original-title', `This mytoken has ${greens} normal capabilities.`);
    $capSummaryHowManyYellow.attr('data-original-title', `This mytoken has ${yellows} powerful capabilities.`);
    $capSummaryHowManyRed.attr('data-original-title', `This mytoken has ${reds} very powerful capabilities.`);

    $capSummaryAT.removeClass("text-success");
    $capSummaryMT.removeClass("text-success");
    $capSummaryInfo.removeClass("text-success");
    $capSummarySettings.removeClass("text-success");
    if (at) {
        $capSummaryAT.addClass("text-success");
        $capSummaryAT.attr('data-original-title', "This mytoken can be used to obtain OIDC Access Tokens.");
    } else {
        $capSummaryAT.attr('data-original-title', "This mytoken cannot be used to obtain OIDC Access Tokens.");
    }
    if (mt) {
        $capSummaryMT.addClass("text-success");
        $capSummaryMT.attr('data-original-title', "This mytoken can be used to create sub-mytokens.");
    } else {
        $capSummaryMT.attr('data-original-title', "This mytoken cannot be used to create sub-mytokens.");
    }
    if (info) {
        $capSummaryInfo.addClass("text-success");
        $capSummaryInfo.attr('data-original-title', "This mytoken can be used to obtain tokeninfo about itself.");
    } else {
        $capSummaryInfo.attr('data-original-title', "This mytoken cannot be used to obtain tokeninfo about itself.");
    }
    if (settings) {
        $capSummarySettings.addClass("text-success");
        $capSummarySettings.attr('data-original-title', "This mytoken can be used to change settings.");
    } else {
        $capSummarySettings.attr('data-original-title', "This mytoken cannot be used to change settings.");
    }
}

$capRWModes.on('update-change', function () {
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

$capRWModes.change(function () {
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
    updateCapSummary();
});