function requestMT(data, okCallback, errCallback) {
    $.ajax({
        type: "POST",
        url: storageGet('mytoken_endpoint'),
        data: JSON.stringify(data),
        success: okCallback,
        error: errCallback,
        dataType: "json",
        contentType: "application/json"
    });
}

function revokeMT(callback, recursive = true) {
    let data = {
        "recursive": recursive
    };
    data = JSON.stringify(data);
    $.ajax({
        type: "POST",
        url: storageGet('revocation_endpoint'),
        data: data,
        success: callback,
        error: callback,
        dataType: "json",
        contentType: "application/json"
    });
}
