function createTransferCode(token, okCallback, errCallback, endpoint = "") {
    let data = {'mytoken': token};
    data = JSON.stringify(data);
    if (endpoint === "") {
        endpoint = storageGet('token_transfer_endpoint');
    }
    $.ajax({
        type: "POST",
        url: endpoint,
        data: data,
        success: function (data) {
            okCallback(data['transfer_code'], data['expires_in']);
        },
        error: function (errRes) {
            errCallback(getErrorMessage(errRes));
        },
        dataType: "json",
        contentType: "application/json"
    });
}

function exchangeTransferCode() {
    let tc = $('#tc-input').val();
    $.ajax({
        type: "POST",
        url: storageGet("mytoken_endpoint"),
        data: JSON.stringify({
            "grant_type": "transfer_code",
            "transfer_code": tc
        }),
        success: function (data) {
            let token = data['mytoken'];
            storageSet("tokeninfo_token", token);
            $('#info-tab').click();
        },
        error: function (errRes) {
            $('#tc-error-modal-msg').text(getErrorMessage(errRes));
            $('#tc-error-modal').modal();
        },
        dataType: "json",
        contentType: "application/json"
    });
}