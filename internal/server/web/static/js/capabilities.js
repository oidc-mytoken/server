const $createMTCaps = $('.capability-check[value=create_mytoken]');
const $allCapRWModes = $('.rw-cap-mode');

function capabilityChecks(prefix = "") {
    return $('.capability-check[instance-prefix="' + prefix + '"]');
}


function capRWModes(prefix = "") {
    return $allCapRWModes.filter('[instance-prefix="' + prefix + '"]');
}

function capabilityCreateMytoken(prefix = "") {
    return $(prefixId('cp-create_mytoken', prefix));
}

function capabilityRevokeAnyToken(prefix = "") {
    return $(prefixId('cp-revoke_any_token', prefix));
}

function capabilityAT(prefix = "") {
    return $(prefixId('cp-AT', prefix));
}

function capSummaryAT(prefix = "") {
    return $(prefixId('cap-summary-AT', prefix));
}

function capSummaryMT(prefix = "") {
    return $(prefixId('cap-summary-MT', prefix));
}

function capSummaryInfo(prefix = "") {
    return $(prefixId('cap-summary-info', prefix));
}

function capSummaryMOM(prefix = "") {
    return $(prefixId('cap-summary-mom', prefix));
}

function capSummarySettings(prefix = "") {
    return $(prefixId('cap-summary-settings', prefix));
}

function capSummaryHowManyGreen(prefix = "") {
    return $(prefixId('cap-summary-count-green', prefix));
}

function capSummaryHowManyYellow(prefix = "") {
    return $(prefixId('cap-summary-count-yellow', prefix));
}

function capSummaryHowManyRed(prefix = "") {
    return $(prefixId('cap-summary-count-red', prefix));
}


function enableCapability(cap, prefix = "") {
    // This function should be called after initCapabilities to preselect / check capabilities
    let $c = $(prefixId(cap, prefix));
    let disabled = $c.prop('disabled');
    $c.prop('disabled', false);
    $c.prop("checked", true);
    checkThisCapability.call($c.get());
    updateCapSummary(prefix);
    $c.prop('disabled', disabled);
}


const rPrefix = "read@";

$('.capability-check').click(function () {
    checkThisCapability.call(this);
    updateCapSummary(this.getAttribute('instance-prefix'));
})

function checkThisCapability() {
    let activated = $(this).prop('checked');
    $(this).closest('li.list-group-item').find('.capability-check').prop('checked', activated);
    if (!activated) {
        $(this).parents('li.list-group-item').children('div').children('div').children('.capability-check').prop('checked', false);
    }
}

function initCapabilities(prefix) {
    capabilityChecks(prefix).each(checkThisCapability);
    updateCapSummary(prefix);
    capRWModes(prefix).trigger('update-change');
}

function getCheckedCapabilities(prefix = "") {
    let caps = capabilityChecks(prefix).filter(':checked').map(function (_, el) {
        let v = $(el).val();
        let $rw = $(prefixId('cp-' + rPrefix + v + '-mode', prefix));
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

function searchAllChecked(str, prefix = "") {
    let read = "read@" + str
    for (const c of getCheckedCapabilities(prefix)) {
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
    let at = capabilityAT(prefix).prop("checked");
    let mt = capabilityCreateMytoken(prefix).prop("checked");
    let info = searchAllChecked("tokeninfo", prefix);
    let mom = searchAllChecked("manage_mytokens", prefix);
    let settings = searchAllChecked("settings", prefix);

    let counter = {
        'green': {},
        'yellow': {},
        'red': {}
    }
    for (const c of capabilityChecks(prefix).filter(':checked')) {
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

    capSummaryAT(prefix).addClass("text-muted");
    capSummaryMT(prefix).addClass("text-muted");
    capSummaryInfo(prefix).addClass("text-muted");
    capSummaryMOM(prefix).addClass("text-muted");
    capSummarySettings(prefix).addClass("text-muted");
    if (at) {
        capSummaryAT(prefix).removeClass("text-muted");
        capSummaryAT(prefix).attr('data-original-title', "This mytoken can be used to obtain OIDC Access Tokens.");
    } else {
        capSummaryAT(prefix).attr('data-original-title', "This mytoken cannot be used to obtain OIDC Access Tokens.");
    }
    if (mt) {
        capSummaryMT(prefix).removeClass("text-muted");
        capSummaryMT(prefix).attr('data-original-title', "This mytoken can be used to create sub-mytokens.");
    } else {
        capSummaryMT(prefix).attr('data-original-title', "This mytoken cannot be used to create sub-mytokens.");
    }
    if (info) {
        capSummaryInfo(prefix).removeClass("text-muted");
        capSummaryInfo(prefix).attr('data-original-title', "This mytoken can be used to obtain tokeninfo about itself.");
    } else {
        capSummaryInfo(prefix).attr('data-original-title', "This mytoken cannot be used to obtain tokeninfo about itself.");
    }
    if (mom) {
        capSummaryMOM(prefix).removeClass("text-muted");
        capSummaryMOM(prefix).attr('data-original-title', "This mytoken can be used to manage other mytokens.");
    } else {
        capSummaryMOM(prefix).attr('data-original-title', "This mytoken cannot be used to manage other mytokens.");
    }
    if (settings) {
        capSummarySettings(prefix).removeClass("text-muted");
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

function checkCapability(cap, prefix = "") {
    let rCap = cap.startsWith(rPrefix);
    if (rCap) {
        cap = cap.substring(rPrefix.length);
    }
    enableCapability('cp-' + cap, prefix);
    let $mode = $(prefixId('cp-' + rPrefix + cap + '-mode', prefix));
    let disabled = $mode.prop('disabled');
    $mode.prop('disabled', false);
    $mode.bootstrapToggle(rCap ? 'off' : 'on');
    $mode.prop('disabled', disabled);
}

function set_capabilities_in_gui(capabilities, prefix = "") {
    if (capabilities === undefined) {
        return
    }
    capabilityChecks(prefix).prop("checked", false);
    capabilities.forEach(function (c) {
        checkCapability(c, prefix);
    });
}

function cap_enableProfileSupport(prefix = "") {
    _enableProfileSupport("cap", set_capabilities_in_gui, prefix);
}