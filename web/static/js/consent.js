
function parseRestriction() {
    let howManyClausesRestrictIP = 0;
    let howManyClausesRestrictScope = 0;
    let howManyClausesRestrictAud = 0;
    let howManyClausesRestrictUsages = 0;
    let expires = 0;
    let doesNotExpire = false;
    console.log(restrictions);
    restrictions.forEach(function (r) {
        if (r['scope'] != undefined) {
            howManyClausesRestrictScope++;
        }
        let aud = r['audience'];
        if (aud != undefined && aud.length > 0) {
            howManyClausesRestrictAud++;
        }
        let ip = r['ip'];
        let ipW = r['geoip_white'];
        let ipB = r['geoip_black'];
        if ((ip != undefined && ip.length > 0) ||
            (ipW != undefined && ipW.length > 0) ||
            (ipB != undefined && ipB.length > 0)) {
            howManyClausesRestrictIP++;
        }
        if (r['usages_other']!=undefined || r['usages_AT']!=undefined) {
            howManyClausesRestrictUsages++;
        }
        let exp = r['exp'];
        if (exp==undefined || exp==0) {
           doesNotExpire = true
        } else if (exp>expires) {
            expires=exp;
        }
    })
    if (doesNotExpire) {
        expires = 0;
    }
    if (howManyClausesRestrictIP==restrictions.length) {
        $('#r-icon-ip').addClass( 'text-success');
        $('#r-icon-ip').removeClass( 'text-warning');
        $('#r-icon-ip').removeClass( 'text-danger');
        $('#r-icon-ip').attr('data-original-title', "The IPs from which this token can be used are restricted.");
    } else {
        $('#r-icon-ip').addClass( 'text-warning');
        $('#r-icon-ip').removeClass( 'text-success');
        $('#r-icon-ip').removeClass( 'text-danger');
        $('#r-icon-ip').attr('data-original-title', "This token can be used from any IP.");
    }
    if (howManyClausesRestrictScope==restrictions.length) {
        $('#r-icon-scope').addClass( 'text-success');
        $('#r-icon-scope').removeClass( 'text-warning');
        $('#r-icon-scope').removeClass( 'text-danger');
        $('#r-icon-scope').attr('data-original-title', "This token has restrictions for scopes.");
    } else {
        $('#r-icon-scope').addClass( 'text-warning');
        $('#r-icon-scope').removeClass( 'text-success');
        $('#r-icon-scope').removeClass( 'text-danger');
        $('#r-icon-scope').attr('data-original-title', "This token can use all configured scopes.");
    }
    if (howManyClausesRestrictAud==restrictions.length) {
        $('#r-icon-aud').addClass( 'text-success');
        $('#r-icon-aud').removeClass( 'text-warning');
        $('#r-icon-aud').removeClass( 'text-danger');
        $('#r-icon-aud').attr('data-original-title', "This token can only obtain access tokens with restricted audiences.");
    } else {
        $('#r-icon-aud').addClass( 'text-warning');
        $('#r-icon-aud').removeClass( 'text-success');
        $('#r-icon-aud').removeClass( 'text-danger');
        $('#r-icon-aud').attr('data-original-title', "This token can obtain access tokens with any audiences.");
    }
    if (howManyClausesRestrictUsages==restrictions.length) {
        $('#r-icon-usages').addClass( 'text-success');
        $('#r-icon-usages').removeClass( 'text-warning');
        $('#r-icon-usages').removeClass( 'text-danger');
        $('#r-icon-usages').attr('data-original-title', "This token can only be used a limited number of times.");
    } else {
        $('#r-icon-usages').addClass( 'text-warning');
        $('#r-icon-usages').removeClass( 'text-success');
        $('#r-icon-usages').removeClass( 'text-danger');
        $('#r-icon-usages').attr('data-original-title', "This token can be used an infinite number of times.");
    }
    if (expires==0) {
        $('#r-icon-time').addClass( 'text-danger');
        $('#r-icon-time').removeClass( 'text-success');
        $('#r-icon-time').removeClass( 'text-warning');
        $('#r-icon-time').attr('data-original-title', "This token does not expire!");
    } else if ((expires - Date.now()/1000)> 3*24*3600) {
        $('#r-icon-time').addClass( 'text-warning');
        $('#r-icon-time').removeClass( 'text-success');
        $('#r-icon-time').removeClass( 'text-danger');
        $('#r-icon-time').attr('data-original-title', "This token is long-lived.");
    } else {
        $('#r-icon-time').addClass( 'text-success');
        $('#r-icon-time').removeClass( 'text-warning');
        $('#r-icon-time').removeClass( 'text-danger');
        $('#r-icon-time').attr('data-original-title', "This token expires within 3 days.");
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

parseRestriction();
newJSONEditor('restrictions');
$('#restrictions').text(JSON.stringify(restrictions, null, 4));

function updateIcons() {
    let r = [];
    try {
        r = JSON.parse($('#restrictions').val());
        $('#restrictions').removeClass('is-invalid');
        $('#restrictions').addClass('is-valid');
    } catch (e) {
        $('#restrictions').removeClass('is-valid');
        $('#restrictions').addClass('is-invalid');
       return;
    }
    restrictions = r;
    parseRestriction();
}

function approve() {
    fetch()
    $.ajax({
        type: "POST",
        url: window.location.href,
        // data: data,
        success: function (data){
            window.location.href = data['authorization_url'];
        },
        error: function(){alert('error');},
        dataType: "json",
        contentType : "application/json"
    });
}

function cancel() {
    //TODO POST cancel
    window.location.href = "/";
}