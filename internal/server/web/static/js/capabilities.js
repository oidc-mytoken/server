
$('#cp-create_mytoken').click(function() {
    let enabled = $(this).prop("checked");
    let capabilityCheck = $('.subtoken-capability-check');
    capabilityCheck.prop("checked", enabled);
    capabilityCheck.prop("disabled", !enabled);
});
