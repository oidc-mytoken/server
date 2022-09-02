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
