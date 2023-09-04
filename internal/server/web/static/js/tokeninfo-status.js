import * as jose from './lib/jose/index.js';

const $tokeninfoBadgeName = $('#tokeninfo-token-name');
const $tokeninfoBadgeTypeShort = $('#tokeninfo-token-type-short');
const $tokeninfoBadgeTypeJWTValid = $('#tokeninfo-token-type-JWT-valid');
const $tokeninfoBadgeTypeJWTInvalid = $('#tokeninfo-token-type-JWT-invalid');
const $tokeninfoBadgeValid = $('#tokeninfo-token-valid');
const $tokeninfoBadgeInvalid = $('#tokeninfo-token-invalid');
const $tokeninfoBadgeMytokenIss = $('#tokeninfo-token-mytoken-iss');
const $tokeninfoBadgeOIDCIss = $('#tokeninfo-token-oidc-iss');
const $tokeninfoBadgeIat = $('#tokeninfo-token-iat');
const $tokeninfoBadgeExp = $('#tokeninfo-token-exp');
const $tokeninfoBadgeIatDate = $('#tokeninfo-token-iat-date');
const $tokeninfoBadgeExpDate = $('#tokeninfo-token-exp-date');
const $tokeninfoTypeBadges = $('.tokeninfo-token-type');
const $tokeninfoTokenGoneWarningMsg = $('#token-gone-warning');
const $tokeninfoActionButtons = $('#token-action-buttons');

$('#tokeninfo-token-copy').on('click', function () {
    if (!$tokeninfoTokenGoneWarningMsg.hasClass('d-none')) {
        $tokeninfoTokenGoneWarningMsg.alert('close');
    }
});

async function update_tokeninfo() {
    let token = storagePop('tokeninfo_token', true);
    if (token === "") {
        token = $tokenInput.val();
        $tokeninfoTokenGoneWarningMsg.hideB();
    } else {
        $tokenInput.val(token);
        $tokeninfoTokenGoneWarningMsg.showB();
    }
    let payload = {};
    if (token === "") {
        $tokeninfoTypeBadges.hideB();
        $tokeninfoBadgeInvalid.hideB();
        $tokeninfoBadgeValid.hideB();
        $tokeninfoBadgeOIDCIss.text("");
        $tokeninfoBadgeMytokenIss.text("");
        $tokeninfoBadgeName.text("");
        $tokeninfoBadgeExp.hideB();
        $tokeninfoBadgeIat.hideB();
        fillTokenInfo(payload);
        $tokeninfoActionButtons.hideB();
        return;
    }
    $tokeninfoActionButtons.showB();
    tokeninfoEndpointToUse = storageGet('tokeninfo_endpoint');
    let jwksUri = storageGet('jwks_uri');
    transferEndpoint = "";
    try {
        payload = jose.decodeJwt(token);
        let mytokenIss = payload['iss'];
        $tokeninfoTypeBadges.hideB();
        $tokeninfoBadgeTypeJWTInvalid.showB();
        if (mytokenIss.endsWith("/")) {
            mytokenIss = mytokenIss.substring(0, mytokenIss.length - 1);
        }
        if (!mytokenIss.startsWith(window.location.href)) {
            let url = mytokenIss + "/.well-known/mytoken-configuration";
            await fetch(url).then(function (res) {
                return res.json();
            }).then(function (data) {
                tokeninfoEndpointToUse = data['tokeninfo_endpoint'];
                jwksUri = data['jwks_uri'];
                transferEndpoint = data['token_transfer_endpoint'];
            }).catch(function (e) {
                console.error(e);
            });
        }

        let jwks = await $.ajax({url: jwksUri, type: "GET"});
        const pubKey = await jose.importJWK(jwks['keys'][0]);
        await jose.jwtVerify(token, pubKey);
        $tokeninfoBadgeTypeJWTInvalid.hideB();
        $tokeninfoBadgeTypeJWTValid.showB();
    } catch (e) {
        if (!(e instanceof jose.errors.JWSSignatureVerificationFailed)) {
            if (e instanceof jose.errors.JWTInvalid) {
                $tokeninfoTypeBadges.hideB();
                $tokeninfoBadgeTypeShort.showB();
            } else {
                console.error(e);
            }
        }
    }
    try {
        await $.ajax({
            type: "POST",
            url: tokeninfoEndpointToUse,
            data: JSON.stringify({
                'action': 'introspect',
                'mytoken': token,
            }),
            dataType: "json",
            contentType: "application/json",
            success: function (res) {
                payload = res['token'];
                if (res['valid']) {
                    $tokeninfoBadgeValid.showB();
                    $tokeninfoBadgeInvalid.hideB();
                } else {
                    $tokeninfoBadgeValid.hideB();
                    $tokeninfoBadgeInvalid.showB();
                }
            },
            error: function (errRes) {
                $tokeninfoBadgeValid.hideB();
                $tokeninfoBadgeInvalid.hideB();
                if (errRes.responseJSON['error'] === 'insufficient_capabilities') {
                    $tokeninfoBadgeValid.showB();
                } else {
                    $tokeninfoBadgeInvalid.showB();
                }
            }
        });
    } catch (e) {
        console.error(e);
    }

    let oidcIss = payload['oidc_iss'];
    let mytokenIss = payload['iss'];
    let name = payload['name'];
    $tokeninfoBadgeOIDCIss.text(oidcIss !== undefined ? oidcIss : "");
    $tokeninfoBadgeMytokenIss.text(mytokenIss !== undefined ? mytokenIss : "");
    $tokeninfoBadgeName.text(name !== undefined ? name : "");
    let exp = payload['exp'];
    if (exp === undefined) {
        $tokeninfoBadgeExp.hideB();
    } else {
        $tokeninfoBadgeExpDate.text(formatTime(exp));
        $tokeninfoBadgeExp.showB();
    }
    let iat = payload['iat'];
    if (iat === undefined) {
        $tokeninfoBadgeIat.hideB();
    } else {
        $tokeninfoBadgeIatDate.text(formatTime(iat));
        $tokeninfoBadgeIat.showB();
    }
    fillTokenInfo(payload);
    $('#introspect-tab').tab('show');
}

$tokenInput.on('change', update_tokeninfo);
$('#info-reload').on('click', update_tokeninfo);
$('#info-tab').on('shown.bs.tab', update_tokeninfo);

$('#recreate-mt').on('click', function () {
    const introspect_content = $('#tokeninfo-token-content');
    const payload = JSON.parse(introspect_content.text());

    const iat = payload['iat'];
    const now = Math.floor(new Date().getTime() / 1000);
    const now_iat_diff = now - iat;
    let restr = payload['restrictions'];
    if (restr) {
        restr.forEach(function (r, i, arr) {
            let nbf = r['nbf'];
            let exp = r['exp'];
            if (nbf) {
                r['nbf'] = now_iat_diff + nbf;
            }
            if (exp) {
                r['exp'] = now_iat_diff + exp;
            }
            arr[i] = r;
        });
    }

    let data = {
        "name": payload['name'],
        "oidc_issuer": payload['oidc_iss'],
        "grant_type": "oidc_flow",
        "oidc_flow": "authorization_code",
        "redirect_type": "native",
        "restrictions": restr,
        "capabilities": payload['capabilities'],
        "application_name": "mytoken webinterface",
    };
    if (!$tokeninfoBadgeTypeShort.hasClass('d-none')) {
        data['response_type'] = 'short_token';
    }
    let rot = payload['rotation'];
    if (rot) {
        data["rotation"] = rot;
    }
    data = JSON.stringify(data);
    $.ajax({
        type: "POST",
        url: storageGet('mytoken_endpoint'),
        data: data,
        success: function (res) {
            let url = res['consent_uri'];
            let code = res['polling_code'];
            let interval = res['interval'];
            authURL(tokeninfoPrefix).attr("href", url);
            authURL(tokeninfoPrefix).text(url);
            mtInstructions(tokeninfoPrefix).showB();
            polling_recreate_mytoken(code, interval);
            window.open(url, '_blank');
        },
        error: function (errRes) {
            let errMsg = getErrorMessage(errRes);
            mtShowError(errMsg, tokeninfoPrefix);
        },
        dataType: "json",
        contentType: "application/json"
    });
    mtShowPending(tokeninfoPrefix);
    $('#recreate-mt-modal').modal();
});

function polling_recreate_mytoken(code, interval) {
    polling_with_callback(code, interval, function (res) {
        let token_type = res['mytoken_type'];
        let token = res['mytoken'];
        if (token_type === "transfer_code") {
            token = res['transfer_code'];
        }
        storageSet("tokeninfo_token", token);
        mtInstructions(tokeninfoPrefix).hideB();
        mtSuccessHeading(tokeninfoPrefix).text("Success! We opened the tokeninfo in a new tab.");
        mtShowSuccess("", tokeninfoPrefix);
        window.open(window.location.href, '_blank');
    }, function (errRes) {
        let error = errRes.responseJSON['error'];
        let message;
        switch (error) {
            case "authorization_pending":
                // message = "Authorization still pending.";
                mtShowPending(tokeninfoPrefix);
                return true;
            case "access_denied":
                message = "You denied the authorization request.";
                break;
            case "expired_token":
                message = "Code expired. You might want to restart the flow.";
                break;
            case "invalid_grant":
            case "invalid_token":
                message = "Code already used.";
                break;
            case "undefined":
                message = "No response from server";
                break;
            default:
                message = getErrorMessage(errRes);
                break;
        }
        mtShowError(message, tokeninfoPrefix);
        return false;
    });
}