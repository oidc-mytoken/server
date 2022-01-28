const $capabilityChecks = $('.capability-check');
const $subtokenCapabilityChecks = $('.subtoken-capability-check');
const $capabilityCreateMytoken = $('#cp-create_mytoken');

$capabilityChecks.click(function () {
    checkThisCapability.call(this)
})
$subtokenCapabilityChecks.click(function () {
    checkThisSubCapability.call(this)
})


$capabilityCreateMytoken.on("click", function () {
    let enabled = $(this).prop("checked");
    let $capabilityCheck = $('.subtoken-capability-check');
    $capabilityCheck.prop("disabled", !enabled);
    $('#subtokenCapabilities').toggleClass('d-none');
});

function checkThisSubCapability() {
    _checkThisCapability.call(this, "subtoken-capability")
}

function checkThisCapability() {
    _checkThisCapability.call(this, "capability")
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
