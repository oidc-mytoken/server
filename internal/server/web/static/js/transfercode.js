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