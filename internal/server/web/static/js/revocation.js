const $revocationModal = $('#revoke-id-modal');
const $revocationFormID = $('#revoke-id');
const $revocationFormRecursive = $('#revoke-recursive');

const revocationClassFromSubtokens = "revoke-from-subtokens";
const revocationClassFromTokenList = "revoke-from-list";
const revocationClassFromTokeninfo = "revoke-from-tokeninfo";

const revocationClasses = [
    revocationClassFromSubtokens,
    revocationClassFromTokenList,
    revocationClassFromTokeninfo,
]

function revokeToken(token, recursive, okCallback) {
    _revoke({
        "token": token,
        "recursive": recursive,
    }, okCallback);
}

function revokeTokenID(id, recursive, okCallback) {
    _revoke({
        "mom_id": id,
        "recursive": recursive,
    }, okCallback);
}

function revokeTokenFromList(id, recursive) {
    revokeTokenID(id, recursive, _getListTokenInfo);
}

function revokeTokenFromSubtokens(id, recursive) {
    revokeTokenID(id, recursive, _getSubtokensInfo);
}

function startRevocateID() {
    let id = this.id;
    $revocationFormID.val(id);
    for (const c of revocationClasses) {
        if ($(this).hasClass(c)) {
            $revocationFormID.addClass(c);
        } else {
            $revocationFormID.removeClass(c);
        }
    }
    $revocationModal.modal();
}

function revokeID() {
    let id = $revocationFormID.val();
    let recursive = $revocationFormRecursive.is(':checked');
    if ($revocationFormID.hasClass(revocationClassFromSubtokens)) {
        revokeTokenFromSubtokens(id, recursive);
    } else if ($revocationFormID.hasClass(revocationClassFromTokenList)) {
        revokeTokenFromList(id, recursive)
    } else if ($revocationFormID.hasClass(revocationClassFromTokeninfo)) {
        revokeToken($tokenInput.val(), recursive, function () {
            $tokenInput.val("");
            $tokenInput.trigger("change");
        })
    } else {
        revokeTokenID(id, recursive, function () {
        });
    }
}

function _revoke(data, okCallback) {
    data = JSON.stringify(data);
    $.ajax({
        type: "POST",
        data: data,
        dataType: "json",
        contentType: "application/json",
        url: storageGet('revocation_endpoint'),
        success: function () {
            okCallback();
        },
        error: function (errRes) {
            $errorModalMsg.text(getErrorMessage(errRes));
            $errorModal.modal();
        },
    });
}