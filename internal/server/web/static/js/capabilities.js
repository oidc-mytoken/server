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
    let classActive = '.' + prefix + '-active';
    let classInactive = '.' + prefix + '-inactive';
    let classCheck = '.' + prefix + '-check'
    if (activated) {
        let allSubCPInactive = $(this).closest('li.list-group-item').find(classInactive);
        let allSubCPActive = $(this).closest('li.list-group-item').find(classActive);
        allSubCPInactive.hideB();
        allSubCPActive.showB();
    } else {
        let $checkedParents = $(this).parents('li.list-group-item').children('div').children('div').children(classCheck + ':checked');
        if ($checkedParents.length === 0) {
            let $span = $(this).siblings('span');
            $span.find(classActive).hideB();
            $span.find(classInactive).showB();
        }
        let $allSubCPChecks = $(this).closest('li.list-group-item').find('ul.list-group li.list-group-item' +
            ' ' + classCheck);
        $allSubCPChecks.each(function () {
            _checkThisCapability.call(this, prefix);
        })
    }
}

$(document).ready(function () {
    $capabilityChecks.each(checkThisCapability);
    $subtokenCapabilityChecks.each(checkThisSubCapability);
    if (!$capabilityCreateMytoken.prop("checked")) {
        $('#subtokenCapabilities').hideB();
    }
    updateCapSummary();
})

function getCheckedCapabilities() {
    return _getCheckedCapabilities($capabilityChecks);
}

function getCheckedSubtokenCapabilities() {
    if (!$capabilityCreateMytoken.prop("checked")) {
        return [];
    }
    return _getCheckedCapabilities($subtokenCapabilityChecks);
}

function _getCheckedCapabilities($checks) {
    return $checks.filter(':checked').map(function (_, el) {
        return $(el).val();
    }).get();
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
    let howManyGreenCaps = 0;
    let howManyYellowCaps = 0;
    let howManyRedCaps = 0;
    let at = $capabilityAT.prop("checked") || $('#sub-cp-AT').prop("checked");
    let mt = $capabilityCreateMytoken.prop("checked");
    let info = searchAllChecked("tokeninfo");
    let settings = searchAllChecked("settings");

    for (const c of $.merge($capabilityChecks, $subtokenCapabilityChecks)) {
        if (!c.checked) {
            continue;
        }
        let $icon = $(c).closest('li.list-group-item').find('i.fa-exclamation-circle');
        if ($icon.hasClass('text-success')) {
            howManyGreenCaps++;
        }
        if ($icon.hasClass('text-warning')) {
            howManyYellowCaps++;
        }
        if ($icon.hasClass('text-danger')) {
            howManyRedCaps++;
        }
    }

    $capSummaryHowManyGreen.text(howManyGreenCaps);
    $capSummaryHowManyYellow.text(howManyYellowCaps);
    $capSummaryHowManyRed.text(howManyRedCaps);
    $capSummaryHowManyGreen.attr('data-original-title', `This mytoken has ${howManyGreenCaps} normal capabilities.`);
    $capSummaryHowManyYellow.attr('data-original-title', `This mytoken has ${howManyYellowCaps} powerful capabilities.`);
    $capSummaryHowManyRed.attr('data-original-title', `This mytoken has ${howManyRedCaps} very powerful capabilities.`);

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