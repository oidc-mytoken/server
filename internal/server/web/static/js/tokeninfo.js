const $tokenInput = $('#tokeninfo-token');
const $errorModal = $('#error-modal');
const $errorModalMsg = $('#error-modal-msg')

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
    initCapabilities(tokeninfoPrefix);
    capabilityChecks(tokeninfoPrefix).not(":checked").closest('.capability').hideB();
    subtokenCapabilityChecks(tokeninfoPrefix).not(":checked").closest('.capability').hideB();
    capabilityChecks(tokeninfoPrefix).filter(":checked").closest('.capability').showB();
    subtokenCapabilityChecks(tokeninfoPrefix).filter(":checked").closest('.capability').showB();

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
    let icons = FaUserAgent.faUserAgent(userAgent);
    return '<span class="user-agent" data-toggle="tooltip" data-placement="bottom" title="' + userAgent + '">' + icons.browser.html + '</i>' + icons.os.html + '</i>' + icons.platform.html + '</i></span>';
}

function historyToHTML(events) {
    let tableEntries = [];
    events.forEach(function (event) {
        let comment = event['comment'] || '';
        let time = formatTime(event['time']);
        let agentIcons = userAgentToHTMLIcons(event['user_agent']);
        let entry = '<tr>' +
            '<td>' + event['event'] + '</td>' +
            '<td>' + comment + '</td>' +
            '<td>' + time + '</td>' +
            '<td>' + event['ip'] + '</td>' +
            '<td>' + agentIcons + '</td>' +
            '</tr>';
        tableEntries.push(entry);
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

const arrowI = '<i class="fas fa-chevron-circle-right" style="padding-right: 1em;"></i>';
const lastArrowI = '<i class="fas fa-arrow-circle-right" style="padding-right: 1em;"></i>';

function _tokenTreeToHTML(tree, depth) {
    let token = tree['token'];
    let name = token['name'] || '';
    if (depth > 0) {
        name = arrowI.repeat(depth - 1) + lastArrowI + name
    }
    let time = formatTime(token['created']);
    let tableEntries = ['<tr>' +
    '<td>' + name + '</td>' +
    '<td>' + time + '</td>' +
    '<td>' + token['ip'] + '</td>' +
    '</tr>'];
    let children = tree['children'];
    if (children !== undefined) {
        children.forEach(function (child) {
            tableEntries = tableEntries.concat(_tokenTreeToHTML(child, depth + 1));
        })
    }
    return tableEntries
}

function tokenlistToHTML(tokenTrees) {
    let tableEntries = [];
    tokenTrees.forEach(function (tokenTree) {
        tableEntries = tableEntries.concat(_tokenTreeToHTML(tokenTree, 0));
    });
    return '<table class="table table-hover table-grey">' +
        '<thead><tr>' +
        '<th>Token Name</th>' +
        '<th>Created</th>' +
        '<th>Created from IP</th>' +
        '</tr></thead>' +
        '<tbody>' +
        tableEntries.join('') +
        '</tbody></table>';
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

function getTreeTokenInfo(e) {
    e.preventDefault();
    let msg = $('#tree-msg');
    let copy = $('#tree-copy');
    _tokeninfo('subtokens',
        function (res) {
            msg.html(tokenlistToHTML([res['mytokens']]));
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

function getListTokenInfo(e) {
    e.preventDefault();
    let msg = $('#list-msg');
    let copy = $('#list-copy');
    getMT(
        function (res) {
            const mToken = res['mytoken']
            _tokeninfo('list_mytokens',
                function (infoRes) {
                    msg.html(tokenlistToHTML(infoRes['mytokens']));
                    msg.removeClass('text-danger');
                    copy.addClass('d-none');
                },
                function (errRes) {
                    msg.text(getErrorMessage(errRes));
                    msg.addClass('text-danger');
                    copy.removeClass('d-none');
                }, mToken);
        },
        function (errRes) {
            msg.text(getErrorMessage(errRes));
            msg.addClass('text-danger');
            copy.removeClass('d-none');
        },
        "list_mytokens"
    );
    return false;
}

$('#history-tab').on('shown.bs.tab', getHistoryTokenInfo)
$('#history-reload').on('click', getHistoryTokenInfo)
$('#tree-tab').on('shown.bs.tab', getTreeTokenInfo)
$('#tree-reload').on('click', getTreeTokenInfo)

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