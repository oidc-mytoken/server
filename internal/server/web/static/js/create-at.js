const atResult = $('#at-result');
const atResultColor = $('#at-result-color');
const atConfig = $('#at-config');
const atSuccessHeading = $('#at-result-heading-success');
const atErrorHeading = $('#at-result-heading-error');
const atPendingHeading = $('#at-result-heading-pending');
const atPendingSpinner = $('#at-pending-spinner');
const atCopyButton = $('#at-result-copy');
const atResultMsg = $('#at-result-msg');


$('#next-at').on('click', function () {
    atResult.hideB();
    atConfig.showB();
})


function atShowPending() {
    atPendingHeading.showB();
    atPendingSpinner.showB();
    atSuccessHeading.hideB();
    atErrorHeading.hideB();
    atCopyButton.hideB();
    atResultMsg.text('');
    atCopyButton.hideB();
    atResultColor.addClass('alert-warning');
    atResultColor.removeClass('alert-success');
    atResultColor.removeClass('alert-danger');
}

function atShowSuccess(msg) {
    atPendingHeading.hideB();
    atPendingSpinner.hideB();
    atSuccessHeading.showB();
    atErrorHeading.hideB();
    atResultMsg.text(msg);
    atCopyButton.showB();
    atResultColor.addClass('alert-success');
    atResultColor.removeClass('alert-danger');
    atResultColor.removeClass('alert-warning');
}

function atShowError(msg) {
    atPendingHeading.hideB();
    atPendingSpinner.hideB();
    atSuccessHeading.hideB();
    atErrorHeading.showB();
    atResultMsg.text(msg);
    atCopyButton.showB();
    atResultColor.addClass('alert-danger');
    atResultColor.removeClass('alert-success');
    atResultColor.removeClass('alert-warning');
}

function _getATScopesFromGUI() {
    let checkedScopeBoxes = $('#at-scopeTableBody').find('.scope-checkbox:checked');
    if (checkedScopeBoxes.length === 0) {
        return null;
    }
    let scopes = []
    checkedScopeBoxes.each(function (i) {
        scopes.push($(this).val());
    })
    return scopes.join(' ');
}

function _getATAudsFromGUI() {
    let body = $('#at-audience-table');
    let items = body.find('.table-item')
    if (items.length === 0) {
        return null;
    }
    let auds = [];
    items.each(function () {
        auds.push($(this).text());
    })
    return auds.filter(onlyUnique).join(' ')
}

function getAT(okCallback, errCallback, mToken) {
    let data = {
        "grant_type": "mytoken",
        "comment": "from web interface"
    };
    let scopes = _getATScopesFromGUI()
    if (scopes) {
        data["scope"] = scopes;
    }
    let auds = _getATAudsFromGUI()
    if (auds) {
        data["audience"] = auds;
    }
    if (mToken) {
        data["mytoken"] = mToken
    }

    data = JSON.stringify(data);
    $.ajax({
        type: "POST",
        url: storageGet('access_token_endpoint'),
        data: data,
        success: okCallback,
        error: errCallback,
        dataType: "json",
        contentType: "application/json"
    });
}

$('#get-at').on('click', function (e) {
    getMT(
        function (res) {
            const mToken = res['mytoken']
            getAT(
                function (tokenRes) {
                    atShowSuccess(tokenRes['access_token']);
                },
                function (errRes) {
                    let errMsg = getErrorMessage(errRes);
                    atShowError(errMsg);
                },
                mToken);
        },
        function (errRes) {
            let errMsg = getErrorMessage(errRes);
            atShowError(errMsg);
        }
    );
    atResult.showB();
    atConfig.hideB();
    atShowPending();
    return false;
});

function initAT(...next) {
    let scopes = storageGet("token_scopes");
    if (scopes === "") { // token not restricted with scopes
        scopes = getSupportedScopesFromStorage();
    } else {
        scopes = scopes.split(' ')
    }
    let $table = $('#at-scopeTableBody');
    for (const scope of scopes) {
        _addScopeValueToGUI(scope, $table, "at");
    }
    doNext(...next);
}