
function _tokeninfo(action, successFnc, errorFnc, token=undefined) {
    let data = {
        'action':action
    };
    if (token!=undefined) {
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
    e.preventDefault();
    _tokeninfo('introspect',
        function(res){
            $('#session-token-info-msg').text(JSON.stringify(res,null,4));
            $('#session-token-info-msg').removeClass('text-danger');
            $('#session-copy').removeClass('d-none');
        },
        function (errRes) {
            console.log(errRes);
            $('#session-token-info-msg').text(getErrorMessage(errRes));
            $('#session-token-info-msg').addClass('text-danger');
            $('#session-copy').removeClass('d-none');
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
    let table = '<table class="table table-hover table-grey">' +
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
    return table;
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
    if (children != undefined) {
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
    let table = '<table class="table table-hover table-grey">' +
        '<thead><tr>' +
        '<th>Token Name</th>' +
        '<th>Created</th>' +
        '<th>Created from IP</th>' +
        '</tr></thead>' +
        '<tbody>' +
        tableEntries.join('') +
        '</tbody></table>';
    return table;
}

$('#get-history').on('click', function(e){
    e.preventDefault();
    _tokeninfo('event_history',
        function(res){
            $('#history-msg').html(historyToHTML(res['events']));
            $('#history-msg').removeClass('text-danger');
            $('#history-copy').addClass('d-none');
        },
        function (errRes) {
            console.log(errRes);
            $('#history-msg').text(getErrorMessage(errRes));
            $('#history-msg').addClass('text-danger');
            $('#history-copy').removeClass('d-none');
        })
    return false;
})

$('#get-tree').on('click', function(e){
    e.preventDefault();
    _tokeninfo('subtoken_tree',
        function(res){
        console.log(res)
            $('#tree-msg').html(tokenlistToHTML([res['mytokens']]));
            $('#tree-msg').removeClass('text-danger');
            $('#tree-copy').addClass('d-none');
        },
        function (errRes) {
            console.log(errRes);
            $('#tree-msg').text(getErrorMessage(errRes));
            $('#tree-msg').addClass('text-danger');
            $('#tree-copy').removeClass('d-none');
        })
    return false;
})

$('#get-list').on('click', function(e){
    e.preventDefault();
    getMT(
        function (res) {
            const mToken = res['mytoken']
            _tokeninfo('list_mytokens',
                function(res){
                    console.log(res)
                    $('#list-msg').html(tokenlistToHTML(res['mytokens']));
                    $('#list-msg').removeClass('text-danger');
                    $('#list-copy').addClass('d-none');
                },
                function (errRes) {
                    console.log(errRes);
                    $('#list-msg').text(getErrorMessage(errRes));
                    $('#list-msg').addClass('text-danger');
                    $('#list-copy').removeClass('d-none');
                }, mToken) ;
        },
        function () {
            $('#list-msg').text(getErrorMessage(errRes));
            $('#list-msg').addClass('text-danger');
            $('#list-copy').removeClass('d-none');
        },
        "list_mytokens"
    );
    return false;
})
