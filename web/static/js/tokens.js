
let clipboard = new ClipboardJS('.btn-copy');
clipboard.on('success', function (e) {
    e.clearSelection();
    let el = $(e.trigger);
    let originalText = el.attr('data-original-title');
    el.attr('data-original-title', 'Copied!').tooltip('show');
    el.attr('data-original-title', originalText);
});

$('#get-at').on('click', function(e){
    e.preventDefault();
    getST(
        function (res) {
           const superToken = res['super_token']
            getAT(
                function(res) {
                    $('#at-msg').text(res['access_token']);
                    $('#at-msg').removeClass('text-danger');
                    $('#at-copy').removeClass('d-none');
                },
                function (errRes) {
                    $('#at-msg').text(getErrorMessage(errRes));
                    $('#at-msg').addClass('text-danger');
                    $('#at-copy').removeClass('d-none');
                },
                superToken);
        },
        function () {
            $('#at-msg').text(getErrorMessage(errRes));
            $('#at-msg').addClass('text-danger');
            $('#at-copy').removeClass('d-none');
        }
    );
    return false;
});

function getAT(okCallback, errCallback, superToken) {
    let data = {
        "grant_type": "super_token",
        "comment": "from web interface"
    };
    if (superToken!=undefined) {
        data["super_token"]=superToken
    }

    data = JSON.stringify(data);
    $.ajax({
        type: "POST",
        url: storageGet('access_token_endpoint'),
        data: data,
        success: okCallback,
        error: errCallback,
        dataType: "json",
        contentType : "application/json"
    });
}

function getST(okCallback, errCallback, capability="AT") {
    let data = {
        "grant_type": "super_token",
        "capabilities": [capability],
        "restrictions": [
            {
                "exp":  Math.floor(Date.now() / 1000) + 60,
                "ip": ["this"],
                "usages_AT": capability=="AT" ? 1 : 0,
                "usages_other": capability=="AT" ? 0 : 1
            }
        ]
    };
    data = JSON.stringify(data);
    $.ajax({
        type: "POST",
        url: storageGet('super_token_endpoint'),
        data: data,
        success: okCallback,
        error: errCallback,
        dataType: "json",
        contentType : "application/json"
    });
}

function revokeST(callback, recursive=true) {
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
        contentType : "application/json"
    });
}
