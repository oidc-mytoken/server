function editorMode(prefix = "") {
    return $('#' + prefix + 'restr-editor-mode');
}

function restrictionsArea(prefix = "") {
    return $('#' + prefix + 'restrictionsArea');
}

function restrIconTime(prefix = "") {
    return $('#' + prefix + 'r-icon-time');
}

function restrIconIP(prefix = "") {
    return $('#' + prefix + 'r-icon-ip');
}

function restrIconScope(prefix = "") {
    return $('#' + prefix + 'r-icon-scope');
}

function restrIconAud(prefix = "") {
    return $('#' + prefix + 'r-icon-aud');
}

function restrIconUsages(prefix = "") {
    return $('#' + prefix + 'r-icon-usages');
}

function getRestrictionsData(prefix = "") {
    return getPrefixData(prefix)['restrictions']['restrictions'];
}

function setRestrictionsData(data, prefix = "") {
    getPrefixData(prefix)['restrictions']['restrictions'] = data;
}

function updateRestrIcons(prefix = "") {
    let howManyClausesRestrictIP = 0;
    let howManyClausesRestrictScope = 0;
    let howManyClausesRestrictAud = 0;
    let howManyClausesRestrictUsages = 0;
    let expires = 0;
    let doesNotExpire = false;
    getRestrictionsData(prefix).forEach(function (r) {
        if (r['scope'] !== undefined) {
            howManyClausesRestrictScope++;
        }
        let aud = r['audience'];
        if (aud !== undefined && aud.length > 0) {
            howManyClausesRestrictAud++;
        }
        let ip = r['ip'];
        let ipW = r['geoip_allow'];
        let ipB = r['geoip_disallow'];
        if ((ip !== undefined && ip.length > 0) ||
            (ipW !== undefined && ipW.length > 0) ||
            (ipB !== undefined && ipB.length > 0)) {
            howManyClausesRestrictIP++;
        }
        if (r['usages_other'] !== undefined || r['usages_AT'] !== undefined) {
            howManyClausesRestrictUsages++;
        }
        let exp = r['exp'];
        if (exp === undefined || exp === 0) {
            doesNotExpire = true
        } else if (exp > expires) {
            expires = exp;
        }
    })
    if (doesNotExpire) {
        expires = 0;
    }
    let restr = getRestrictionsData(prefix);
    if (howManyClausesRestrictIP === restr.length && restr.length > 0) {
        restrIconIP(prefix).addClass('text-success');
        restrIconIP(prefix).removeClass('text-warning');
        restrIconIP(prefix).removeClass('text-danger');
        restrIconIP(prefix).attr('data-original-title', "The IPs from which this token can be used are restricted.");
    } else {
        restrIconIP(prefix).addClass('text-warning');
        restrIconIP(prefix).removeClass('text-success');
        restrIconIP(prefix).removeClass('text-danger');
        restrIconIP(prefix).attr('data-original-title', "This token can be used from any IP.");
    }
    if (howManyClausesRestrictScope === restr.length && restr.length > 0) {
        restrIconScope(prefix).addClass('text-success');
        restrIconScope(prefix).removeClass('text-warning');
        restrIconScope(prefix).removeClass('text-danger');
        restrIconScope(prefix).attr('data-original-title', "This token has restrictions for scopes.");
    } else {
        restrIconScope(prefix).addClass('text-warning');
        restrIconScope(prefix).removeClass('text-success');
        restrIconScope(prefix).removeClass('text-danger');
        restrIconScope(prefix).attr('data-original-title', "This token can use all configured scopes.");
    }
    if (howManyClausesRestrictAud === restr.length && restr.length > 0) {
        restrIconAud(prefix).addClass('text-success');
        restrIconAud(prefix).removeClass('text-warning');
        restrIconAud(prefix).removeClass('text-danger');
        restrIconAud(prefix).attr('data-original-title', "This token can only obtain access tokens with restricted" +
            " audiences.");
    } else {
        restrIconAud(prefix).addClass('text-warning');
        restrIconAud(prefix).removeClass('text-success');
        restrIconAud(prefix).removeClass('text-danger');
        restrIconAud(prefix).attr('data-original-title', "This token can obtain access tokens with any audiences.");
    }
    if (howManyClausesRestrictUsages === restr.length && restr.length > 0) {
        restrIconUsages(prefix).addClass('text-success');
        restrIconUsages(prefix).removeClass('text-warning');
        restrIconUsages(prefix).removeClass('text-danger');
        restrIconUsages(prefix).attr('data-original-title', "This token can only be used a limited number of times.");
    } else {
        restrIconUsages(prefix).addClass('text-warning');
        restrIconUsages(prefix).removeClass('text-success');
        restrIconUsages(prefix).removeClass('text-danger');
        restrIconUsages(prefix).attr('data-original-title', "This token can be used an infinite number of times.");
    }
    if (expires === 0) {
        restrIconTime(prefix).addClass('text-danger');
        restrIconTime(prefix).removeClass('text-success');
        restrIconTime(prefix).removeClass('text-warning');
        restrIconTime(prefix).attr('data-original-title', "This token does not expire!");
    } else if ((expires - Date.now() / 1000) > 3 * 24 * 3600) {
        restrIconTime(prefix).addClass('text-warning');
        restrIconTime(prefix).removeClass('text-success');
        restrIconTime(prefix).removeClass('text-danger');
        restrIconTime(prefix).attr('data-original-title', "This token is long-lived.");
    } else {
        restrIconTime(prefix).addClass('text-success');
        restrIconTime(prefix).removeClass('text-warning');
        restrIconTime(prefix).removeClass('text-danger');
        restrIconTime(prefix).attr('data-original-title', "This token expires within 3 days.");
    }
}

function newJSONEditor(textareaID) {
    return new Behave({
        textarea: document.getElementById(textareaID),
        replaceTab: true,
        softTabs: true,
        tabSize: 4,
        autoOpen: true,
        overwrite: true,
        autoStrip: true,
        autoIndent: true
    });
}

function fillJSONEditor(prefix = "") {
    restrictionsArea(prefix).val(JSON.stringify(getRestrictionsData(prefix), null, 4));
}

function updateRestrFromJSONEditor(prefix) {
    let r = [];
    try {
        r = JSON.parse(restrictionsArea(prefix).val());
        restrictionsArea(prefix).removeClass('is-invalid');
        restrictionsArea(prefix).addClass('is-valid');
    } catch (e) {
        restrictionsArea(prefix).removeClass('is-valid');
        restrictionsArea(prefix).addClass('is-invalid');
        return;
    }
    setRestrictionsData(r, prefix);
    updateRestrIcons(prefix);
}


function initRestr(prefix = "", ...next) {
    newJSONEditor(prefix + 'restrictionsArea');
    editorMode(prefix).bootstrapToggle("on");
    editorMode(prefix).on('change', function () {
        let htmlEdit = $('#' + prefix + 'restr-easy-editor');
        let jsonEdit = $('#' + prefix + 'restr-json-editor');
        if ($(this).prop("checked")) { // easy editor
            RestrToGUI(prefix);
            jsonEdit.hideB();
            htmlEdit.showB();
        } else { // JSON editor
            fillJSONEditor(prefix);
            jsonEdit.showB();
            htmlEdit.hideB();
        }
    });
    updateRestrIcons(prefix);
    initRestrGUI(prefix);
    doNext(...next);
}

$('.restr-expand').on('click', function () {
    let prefix = this.getAttribute("instance-prefix");
    $('#' + prefix + 'restr-editor-wrap').toggleClass('d-none');
});