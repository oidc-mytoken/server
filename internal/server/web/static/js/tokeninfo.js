
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


$('#get-session-token-info').on('click', function(e){
    let msg = $('#session-token-info-msg');
    let copy = $('#session-copy');
    e.preventDefault();
    _tokeninfo('introspect',
        function(res){
            let iss = res['token']['oidc_iss'];
            if (iss) {
                storageSet('oidc_issuer', iss, true);
            }
            msg.text(JSON.stringify(res,null,4));
            msg.removeClass('text-danger');
            copy.removeClass('d-none');
        },
        function (errRes) {
            console.log(errRes);
            msg.text(getErrorMessage(errRes));
            msg.addClass('text-danger');
            copy.removeClass('d-none');
        })
    return false;
})

function userAgentToHTMLIcons(userAgent) {
    let icons = FaUserAgent.faUserAgent(userAgent);
    return '<span class="user-agent" data-toggle="tooltip" data-placement="bottom" title="'+userAgent+'">'+icons.browser.html + '</i>' + icons.os.html + '</i>' + icons.platform.html+ '</i></span>';
}

function historyToHTML(events) {
    let tableEntries = [];
    events.forEach(function (event) {
        let comment = event['comment'] || '';
        let time = new Date(event['time']*1000).toLocaleString();
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
    let time = new Date(token['created']*1000).toLocaleString();
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

$('#get-history').on('click', function(e){
    e.preventDefault();
    let msg = $('#history-msg');
    let copy = $('#history-copy');
    _tokeninfo('event_history',
        function(res){
            msg.html(historyToHTML(res['events']));
            msg.removeClass('text-danger');
            copy.addClass('d-none');
        },
        function (errRes) {
            console.log(errRes);
            msg.text(getErrorMessage(errRes));
            msg.addClass('text-danger');
            copy.removeClass('d-none');
        })
    return false;
})

$('#get-tree').on('click', function(e){
    e.preventDefault();
    let msg = $('#tree-msg');
    let copy = $('#tree-copy');
    _tokeninfo('subtoken_tree',
        function(res){
            msg.html(tokenlistToHTML([res['mytokens']]));
            msg.removeClass('text-danger');
            copy.addClass('d-none');
        },
        function (errRes) {
            console.log(errRes);
            msg.text(getErrorMessage(errRes));
            msg.addClass('text-danger');
            copy.removeClass('d-none');
        })
    return false;
})

$('#get-list').on('click', function(e){
    e.preventDefault();
    let msg = $('#list-msg');
    let copy = $('#list-copy');
    getMT(
        function (res) {
            const mToken = res['mytoken']
            _tokeninfo('list_mytokens',
                function(infoRes){
                    msg.html(tokenlistToHTML(infoRes['mytokens']));
                    msg.removeClass('text-danger');
                    copy.addClass('d-none');
                },
                function (errRes) {
                    console.log(errRes);
                    msg.text(getErrorMessage(errRes));
                    msg.addClass('text-danger');
                    copy.removeClass('d-none');
                }, mToken) ;
        },
        function (errRes) {
            msg.text(getErrorMessage(errRes));
            msg.addClass('text-danger');
            copy.removeClass('d-none');
        },
        "list_mytokens"
    );
    return false;
})
