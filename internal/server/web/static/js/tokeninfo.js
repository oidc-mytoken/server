const $tokenInput = $('#tokeninfo-token');

function _tokeninfo(action, successFnc, errorFnc, token=undefined) {
    let data = {
        'action':action
    };
    if (token!==undefined) {
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
        contentType : "application/json"
    });
}

function getTokenInfo(e) {
    e.preventDefault();
    let msg = $('#tokeninfo-token-content');
    let copy = $('#info-copy');
    _tokeninfo('introspect',
        function (res) {
            let token = res['token'];
            let iss = token['oidc_iss'];
            if (iss) {
                storageSet('oidc_issuer', iss, true);
            }
            let scopes = extractMaxScopesFromToken(token);
            storageSet('token_scopes', scopes, true);
            msg.text(JSON.stringify(res, null, 4));
            msg.removeClass('text-danger');
            copy.removeClass('d-none');
        },
        function (errRes) {
            msg.text(getErrorMessage(errRes));
            msg.addClass('text-danger');
            copy.removeClass('d-none');
        }, $tokenInput.val())
    return false;
}

function userAgentToHTMLIcons(userAgent) {
    let icons = FaUserAgent.faUserAgent(userAgent);
    return '<span class="user-agent" data-toggle="tooltip" data-placement="bottom" title="'+userAgent+'">'+icons.browser.html + '</i>' + icons.os.html + '</i>' + icons.platform.html+ '</i></span>';
}

function historyToHTML(events) {
    let tableEntries = [];
    events.forEach(function (event) {
        let comment = event['comment'] || '';
        let time = formatTime(event['time']);
        let agentIcons = userAgentToHTMLIcons(event['user_agent']);
        let entry = '<tr>' +
            '<td>'+event['event']+'</td>' +
            '<td>'+comment+'</td>' +
            '<td>'+time+'</td>' +
            '<td>'+event['ip']+'</td>' +
            '<td>'+agentIcons+'</td>' +
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
        name = arrowI.repeat(depth-1) + lastArrowI + name
    }
    let time = formatTime(token['created']);
    let tableEntries = ['<tr>' +
        '<td>'+name+'</td>' +
        '<td>'+time+'</td>' +
        '<td>'+token['ip']+'</td>' +
        '</tr>'];
    let children = tree['children'];
    if (children !== undefined) {
        children.forEach(function (child) {
            tableEntries = tableEntries.concat(_tokenTreeToHTML(child, depth+1));
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

$('#info-tab').on('shown.bs.tab', getTokenInfo)
$('#info-reload').on('click', getTokenInfo)
$('#history-tab').on('shown.bs.tab', getHistoryTokenInfo)
$('#history-reload').on('click', getHistoryTokenInfo)
$('#tree-tab').on('shown.bs.tab', getTreeTokenInfo)
$('#tree-reload').on('click', getTreeTokenInfo)

$('#list-mts-tab').on('shown.bs.tab', getListTokenInfo)
$('#list-reload').on('click', getListTokenInfo)