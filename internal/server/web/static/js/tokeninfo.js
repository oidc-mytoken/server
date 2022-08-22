const $tokenInput = $('#tokeninfo-token');

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
        url: storageGet('tokeninfo_endpoint'),
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
        subtokenCapabilityChecks(tokeninfoPrefix).prop("checked", false);
        capabilityChecks(tokeninfoPrefix).closest('.capability').hideB();
        subtokenCapabilityChecks(tokeninfoPrefix).closest('.capability').hideB();
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
    subtokenCapabilityChecks(tokeninfoPrefix).prop("checked", false);
    for (let c of tokenPayload['capabilities']) {
        checkCapability(c, 'cp', tokeninfoPrefix);
    }
    if (tokenPayload['subtoken_capabilities'] !== undefined) {
        for (let c of tokenPayload['subtoken_capabilities']) {
            checkCapability(c, 'sub-cp', tokeninfoPrefix);
        }
    }
    capabilityChecks(tokeninfoPrefix).not(":checked").closest('.capability').hideB();
    subtokenCapabilityChecks(tokeninfoPrefix).not(":checked").closest('.capability').hideB();
    capabilityChecks(tokeninfoPrefix).filter(":checked").parents('.capability').showB();
    subtokenCapabilityChecks(tokeninfoPrefix).filter(":checked").parents('.capability').showB();

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
    tableEntries = `<tr id="${thisID}" parent-id="${parentID}" class="${depth > 0 ? 'd-none' : ''}"><td class="${hasChildren ? 'token-fold' : ''}${nameClass}"><span style="margin-right: ${1.5 * depth}rem;"></span><i class="mr-2 fas fa-caret-right${hasChildren ? "" : " d-none"}"></i>${name}</td><td>${time}</td><td>${token['ip']}</td><td><button id="${token['revocation_id']}" class="btn ${deleteClass}" type="button" onclick="startRevocateID.call(this)"><i class="fas fa-trash"></i></button></td></tr>` + tableEntries;
    return tableEntries
}

function tokenlistToHTML(tokenTrees, deleteClass) {
    let tableEntries = "";
    tokenTrees.forEach(function (tokenTree) {
        tableEntries = _tokenTreeToHTML(tokenTree, deleteClass, 0) + tableEntries;
    });
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
            msg.html(tokenlistToHTML([res['mytokens']], revocationClassFromSubtokens));
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
    getMT(
        function (res) {
            _getListTokenInfo(res['mytoken']);
        },
        function (errRes) {
            listMsg.text(getErrorMessage(errRes));
            listMsg.addClass('text-danger');
            listCopy.removeClass('d-none');
        },
        "list_mytokens"
    );
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