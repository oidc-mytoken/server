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

function useRevocationToken(callback) {
    let tok = storageGet("revocation_mytoken")
    _tokeninfo(
        'introspect',
        function () {
            callback(tok);
        },
        function () {
            requestMT(
                {
                    "name": "mytoken-web MT for revocations",
                    "grant_type": "mytoken",
                    "capabilities": ["tokeninfo:introspect", "revoke_any_token", "list_mytokens"],
                    "restrictions": [
                        {
                            "exp": Math.floor(Date.now() / 1000) + 300,
                            "ip": ["this"],
                            "usages_AT": 0,
                            "usages_other": 60,
                        }
                    ]
                },
                function (res) {
                    let token = res['mytoken'];
                    storageSet('revocation_mytoken', token, true);
                    callback(token);
                },
                function (errRes) {
                    $errorModalMsg.text(getErrorMessage(errRes));
                    $errorModal.modal();
                }
            );
        },
        tok);
}

function revokeToken(token, recursive, okCallback) {
    _revoke({
        "token": token,
        "recursive": recursive,
    }, okCallback);
}

function revokeTokenID(id, recursive, okCallback) {
    useRevocationToken(function (token) {
        _revoke({
            "token": token,
            "revocation_id": id,
            "recursive": recursive,
        }, function () {
            okCallback(token);
        })
    })
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