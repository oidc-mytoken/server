const $tokenInput = $('#tokeninfo-token');

let tokeninfoEndpointToUse;

function _tokeninfo(action, successFnc, errorFnc, token = undefined) {
    let data = {
        'action': action
    };
    if (token !== undefined) {
        data['mytoken'] = token;
    }
    data = JSON.stringify(data);
    $.ajax({
        type: "POST",
        url: tokeninfoEndpointToUse,
        data: data,
        success: successFnc,
        error: errorFnc,
        dataType: "json",
        contentType: "application/json"
    });
}

function fillTokenInfo(tokenPayload) {
    // introspection
    let msg = $('#tokeninfo-token-content');
    let copy = $('#info-copy');
    if (tokenPayload === undefined || $.isEmptyObject(tokenPayload)) {
        msg.text("We do not have any information about this token's content.");
        msg.addClass('text-danger');
        copy.addClass('d-none');

        capabilityChecks(tokeninfoPrefix).prop("checked", false);
        capabilityChecks(tokeninfoPrefix).closest('.capability').hideB();
        updateCapSummary(tokeninfoPrefix);

        rotationAT(tokeninfoPrefix).prop('checked', false);
        rotationOther(tokeninfoPrefix).prop('checked', false);
        rotationLifetime(tokeninfoPrefix).val(0);
        rotationAutoRevoke(tokeninfoPrefix).prop('checked', false);
        updateRotationIcon(tokeninfoPrefix);

        setRestrictionsData([{}], tokeninfoPrefix);
        scopeTableBody(tokeninfoPrefix).html("");
        RestrToGUI(tokeninfoPrefix);
        return;
    }
    msg.text(JSON.stringify(tokenPayload, null, 4));
    msg.removeClass('text-danger');
    copy.removeClass('d-none');

    // capabilities
    initCapabilities(tokeninfoPrefix);
    capabilityChecks(tokeninfoPrefix).prop("checked", false);
    for (let c of tokenPayload['capabilities']) {
        checkCapability(c, tokeninfoPrefix);
    }
    capabilityChecks(tokeninfoPrefix).not(":checked").closest('.capability').hideB();
    capabilityChecks(tokeninfoPrefix).filter(":checked").parents('.capability').showB();

    // rotation
    let rot = tokenPayload['rotation'] || {};
    rotationAT(tokeninfoPrefix).prop('checked', rot['on_AT'] || false);
    rotationOther(tokeninfoPrefix).prop('checked', rot['on_other'] || false);
    rotationLifetime(tokeninfoPrefix).val(rot['lifetime'] || 0);
    rotationAutoRevoke(tokeninfoPrefix).prop('checked', rot['auto_revoke'] || false);
    updateRotationIcon(tokeninfoPrefix);

    // restrictions
    setRestrictionsData(tokenPayload['restrictions'] || [{}], tokeninfoPrefix);
    scopeTableBody(tokeninfoPrefix).html("");
    const scopes = getSupportedScopesFromStorage(tokenPayload['oidc_iss']);
    for (const scope of scopes) {
        _addScopeValueToGUI(scope, scopeTableBody(tokeninfoPrefix), "restr", tokeninfoPrefix);
    }
    RestrToGUI(tokeninfoPrefix);
}

function userAgentToHTMLIcons(userAgent) {
    if (userAgent === "") {
        return "";
    }
    if (userAgent.startsWith("oidc-agent")) {
        return `<span class="user-agent" data-toggle="tooltip" data-placement="bottom" title="${userAgent}"><svg xmlns="http://www.w3.org/2000/svg" class="svg-icon" aria-hidden="true" focusable="false" viewBox="0 0 79.378 70.108" xmlns:v="https://vecta.io/nano"><defs><filter id="A" x="0" width="1" y="0" height="1" color-interpolation-filters="sRGB"><feGaussianBlur stdDeviation=".002"/></filter></defs><path transform="matrix(.264583 0 0 .264583 -92.607809 -71.45579)" d="M500.02 270.074l-136.424.902c-5.112 1.186-8.552 3.813-11.311 8.637-2.02 3.532-2.235 5.262-2.25 18.072-.011 10.41.316 14.301 1.234 14.672 1.801.727 295.276.781 297.168.055 1.336-.513 1.582-2.734 1.582-14.287 0-12.217-.241-14.154-2.25-18.086-2.65-5.187-5.931-7.811-11.328-9.062-2.595-.602-69.508-.902-136.422-.902zM435.693 381.57c.458.013.924.078 1.393.196 2.706.682 45.825 43.726 45.819 45.739-.002.818-10.317 11.078-22.921 22.799-23.859 22.188-26.081 23.616-30.098 19.339-4.253-4.528-2.811-6.691 16.014-24.011 9.826-9.041 18.143-17.135 18.483-17.988.397-.996-5.823-7.864-17.381-19.189-16.683-16.347-17.975-17.899-17.645-21.187.342-3.411 3.13-5.791 6.337-5.699zm49.046 74.949h44.917 44.917l1.565 3.889c1.405 3.489 1.379 4.286-.248 7.75l-1.816 3.86h-44.419-44.421l-1.814-3.86c-1.628-3.463-1.653-4.261-.248-7.75zm15.28-121.488l-148.418.565c-1.428.519-1.582 10.037-1.582 98.034v97.462l2.646 1.973 2.648 1.973 145.248-.246 147.352-2.053v.002l2.105-1.805v-97.385c0-87.926-.154-97.437-1.582-97.956-1.034-.376-74.726-.565-148.418-.565z" fill="#fff" filter="url(#A)"/></svg></span>`;
    }
    if (userAgent.startsWith("mytoken")) {
        return `<span class="user-agent" data-toggle="tooltip" data-placement="bottom" title="${userAgent}"><svg xmlns="http://www.w3.org/2000/svg" class="svg-icon" aria-hidden="true" focusable="false" viewBox="-0.5 -0.5 374 160" xmlns:v="https://vecta.io/nano"><defs><style>@font-face{font-family:"Lato";src:url(data:application/font-woff;charset=utf-8;base64,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) format("woff"); font-weight:bold;font-style:normal;}</style></defs><text x="202.5" y="61.5" fill="#ccc" font-family="Lato" font-size="50" text-anchor="end" font-weight="bold">my</text><g stroke="#fff"><ellipse cx="60" cy="80" rx="50" ry="70" fill="transparent" stroke-width="20" pointer-events="all"/><g fill="#fff" stroke-width="8"><path pointer-events="all" d="M110 70h260v20H110z"/><path pointer-events="all" d="M340 80h10v60h-10zm20 0h10v50h-10zm-40 0h10v60h-10zm-20 0h10v50h-10zm-20 0h10v60h-10z"/></g></g><text x="206.5" y="61.5" fill="#fff" font-family="Lato" font-size="50" font-weight="bold">token</text><path d="M58 91h17v-2H58zM43.33 73.43l10.08 10.08-10.08 10.08 1.41 1.42 11.49-11.5-11.49-11.49zM37 59h2v-2h-2zm4 0h2v-2h-2zm4 0h2v-2h-2zm-11 2v-6h52v6zm-2-8v54h34v-2H34V63h52v42h-7v2h9V53z" fill="#fff" pointer-events="all"/></svg></span>`;
    }
    let icons = FaUserAgent.faUserAgent(userAgent);
    return `<span class="user-agent" data-toggle="tooltip" data-placement="bottom" title="${userAgent}">${icons.browser.html}</i>${icons.os.html}</i>${icons.platform.html}</i></span>`;
}

function historyToHTML(events) {
    let tableEntries = [];
    events.forEach(function (event) {
        let comment = event['comment'] || '';
        let time = formatTime(event['time']);
        let agentIcons = userAgentToHTMLIcons(event['user_agent'] || "");
        let entry = '<tr>' +
            '<td>' + event['event'] + '</td>' +
            '<td>' + comment + '</td>' +
            '<td>' + time + '</td>' +
            '<td>' + event['ip'] + '</td>' +
            '<td class="text-center">' + agentIcons + '</td>' +
            '</tr>';
        tableEntries.unshift(entry);
    });
    return '<table class="table table-hover table-grey">' +
        '<thead><tr>' +
        '<th>Event</th>' +
        '<th>Comment</th>' +
        '<th>Time</th>' +
        '<th>IP</th>' +
        '<th>User Agent</th>' +
        '</tr></thead>' +
        '<tbody>' +
        tableEntries.join('') +
        '</tbody></table>';
}


let tokenTreeIDCounter = 1;

function _tokenTreeToHTML(tree, deleteClass, depth, parentID = 0) {
    let token = tree['token'];
    let name = token['name'] || 'unnamed token';
    let nameClass = name === 'unnamed token' ? ' text-muted' : '';
    let thisID = `token-tree-el-${tokenTreeIDCounter++}`
    let time = formatTime(token['created']);
    let tableEntries = "";
    let children = tree['children'];
    let hasChildren = false;
    if (children !== undefined) {
        children.forEach(function (child) {
            tableEntries = _tokenTreeToHTML(child, deleteClass, depth + 1, thisID) + tableEntries;
            hasChildren = true;
        })
    }
    let deleteBtn = `<button id="${token['mom_id']}" class="btn ${deleteClass}" type="button" onclick="startRevocateID.call(this)" ${loggedIn ? "" : "disabled"} data-toggle="tooltip" data-placement="right" title="${loggedIn ? 'Revoke Token' : 'Sign in to revoke token.'}"><i class="fas fa-trash"></i></button>`;
    tableEntries = `<tr id="${thisID}" parent-id="${parentID}" class="${depth > 0 ? 'd-none' : ''}"><td class="${hasChildren ? 'token-fold' : ''}${nameClass}"><span style="margin-right: ${1.5 * depth}rem;"></span><i class="mr-2 fas fa-caret-right${hasChildren ? "" : " d-none"}"></i>${name}</td><td>${time}</td><td>${token['ip']}</td><td>${deleteBtn}</td></tr>` + tableEntries;
    return tableEntries
}

function tokenlistToHTML(tokenTrees, deleteClass) {
    let tableEntries = "";
    tokenTrees.forEach(function (tokenTree) {
        tableEntries = _tokenTreeToHTML(tokenTree, deleteClass, 0) + tableEntries;
    });
    if (tableEntries === "") {
        tableEntries = `<tr><td colSpan="4" class="text-muted text-center">No subtokens</td></tr>`;
    }
    return '<table class="table table-hover table-grey">' +
        '<thead><tr>' +
        '<th style="min-width: 50%;">Token Name</th>' +
        '<th>Created</th>' +
        '<th>Created from IP</th>' +
        '<th></th>' +
        '</tr></thead>' +
        '<tbody>' +
        tableEntries +
        '</tbody></table>';
}

function activateTokenList() {
    $('.token-fold').off('click').on('click', function () {
        let $caret = $(this).find('i');
        $caret.toggleClass('fa-caret-right');
        $caret.toggleClass('fa-caret-down');
        let $tr = $(this).parent();
        let trID = $tr.attr('id');
        let $childTrs = $tr.parent().find(`tr[parent-id="${trID}"]`);
        $childTrs.toggleClass('d-none');
        let $tdsWithOpenChilds = $childTrs.find('td.token-fold').has('i.fa-caret-down');
        $tdsWithOpenChilds.click();
    });
}

function getHistoryTokenInfo(e) {
    e.preventDefault();
    let msg = $('#history-msg');
    let copy = $('#history-copy');
    _tokeninfo('event_history',
        function (res) {
            msg.html(historyToHTML(res['events']));
            msg.removeClass('text-danger');
            copy.addClass('d-none');
        },
        function (errRes) {
            msg.text(getErrorMessage(errRes));
            msg.addClass('text-danger');
            copy.removeClass('d-none');
        }, $tokenInput.val())
    return false;
}

function getSubtokensInfo(e) {
    e.preventDefault();
    _getSubtokensInfo();
}

function _getSubtokensInfo() {
    let msg = $('#tree-msg');
    let copy = $('#tree-copy');
    _tokeninfo('subtokens',
        function (res) {
            let token = res['mytokens'];
            let subtokens = [];
            if ('children' in token) {
                subtokens = token['children'];
            }
            msg.html(tokenlistToHTML(subtokens, revocationClassFromSubtokens));
            activateTokenList();
            msg.removeClass('text-danger');
            copy.addClass('d-none');
        },
        function (errRes) {
            msg.text(getErrorMessage(errRes));
            msg.addClass('text-danger');
            copy.removeClass('d-none');
        }, $tokenInput.val())
    return false;
}

const listMsg = $('#list-msg');
const listCopy = $('#list-copy');

function _getListTokenInfo(token) {
    tokeninfoEndpointToUse = storageGet("tokeninfo_endpoint");
    _tokeninfo('list_mytokens',
        function (infoRes) {
            listMsg.html(tokenlistToHTML(infoRes['mytokens'], revocationClassFromTokenList));
            activateTokenList();
            listMsg.removeClass('text-danger');
            listCopy.addClass('d-none');
        },
        function (errRes) {
            listMsg.text(getErrorMessage(errRes));
            listMsg.addClass('text-danger');
            listCopy.removeClass('d-none');
        }, token);
}

function getListTokenInfo(e) {
    e.preventDefault();
    _getListTokenInfo();
    return false;
}

$('#history-tab').on('shown.bs.tab', getHistoryTokenInfo)
$('#history-reload').on('click', getHistoryTokenInfo)
$('#tree-tab').on('shown.bs.tab', getSubtokensInfo)
$('#tree-reload').on('click', getSubtokensInfo)

$('#list-mts-tab').on('shown.bs.tab', getListTokenInfo)
$('#list-reload').on('click', getListTokenInfo)

const tokeninfoPrefix = "tokeninfo-";

function initTokeninfo(...next) {
    initCapabilities(tokeninfoPrefix);
    updateRotationIcon(tokeninfoPrefix);
    initRestr(tokeninfoPrefix);
    $(prefixId("nbf", tokeninfoPrefix)).datetimepicker("minDate", null);
    $(prefixId("exp", tokeninfoPrefix)).datetimepicker("minDate", null);
    doNext(...next);
}

let transferEndpoint = "";
$('#create-tc').on('click', function () {
    createTransferCode($tokenInput.val(),
        function (tc, expiresIn) {
            let now = new Date();
            let expiresAt = formatTime(now.setSeconds(now.getSeconds() + expiresIn) / 1000);
            $('.insert-tc').text(tc);
            $('#tc-expires').text(expiresAt);
            $('#tc-modal').modal();
        },
        function (errMsg) {
            $errorModalMsg.text(errMsg);
            $errorModal.modal();
        },
        transferEndpoint);
});

$('#revoke-tokeninfo').on('click', function () {
    for (const c of revocationClasses) {
        $revocationFormID.removeClass(c);
    }
    $revocationFormID.addClass(revocationClassFromTokeninfo);
    $revocationModal.modal();
})