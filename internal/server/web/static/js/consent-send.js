function approve(data) {
    data = JSON.stringify(data);
    $.ajax({
        type: "POST",
        url: window.location.href,
        data: data,
        headers: {
            Accept: "application/json"
        },
        success: function (res) {
            window.location.href = res['authorization_uri'];
        },
        error: function (errRes) {
            let errMsg = getErrorMessage(errRes);
            if (errRes.status === 404) {
                errMsg = "Expired. Please start the flow again.";
            }
            $('#error-modal-msg').text(errMsg);
            $('#error-modal').modal();
        },
        dataType: "json",
        contentType: "application/json"
    });
}

function cancel() {
    $.ajax({
        type: "POST",
        url: window.location.href,
        success: function (data) {
            window.location.href = data['url'];
        },
        error: function (errRes) {
            let errMsg = getErrorMessage(errRes);
            console.error(errMsg);
            window.location.href = errRes.responseJSON['url'];
        },
        dataType: "json",
        contentType: "application/json"
    });
}
