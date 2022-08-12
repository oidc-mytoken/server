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


async function update_tokeninfo() {
    let token = storagePop('tokeninfo_token', true);
    if (token === "") {
        token = $tokenInput.val();
    } else {
        $tokenInput.val(token);
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
        return;
    }
    let tokeninfoEndpoint = storageGet('tokeninfo_endpoint');
    let jwksUri = storageGet('jwks_uri');
    transferEndpoint = "";
    try {
        payload = jose.decodeJwt(token);
        let mytokenIss = payload['iss'];
        $tokeninfoTypeBadges.hideB();
        $tokeninfoBadgeTypeJWTInvalid.showB();
        if (!mytokenIss.startsWith(window.location.href)) {
            await fetch(mytokenIss + "/.well-known/mytoken-configuration").then(function (res) {
                return res.json();
            }).then(function (data) {
                tokeninfoEndpoint = data['tokeninfo_endpoint'];
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
            url: tokeninfoEndpoint,
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
}

$tokenInput.on('change', update_tokeninfo);
$('#info-reload').on('click', update_tokeninfo)
$('#info-tab').on('shown.bs.tab', update_tokeninfo)
