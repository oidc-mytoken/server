
function parseRestriction() {
    let howManyClausesRestrictIP = 0;
    let howManyClausesRestrictScope = 0;
    let howManyClausesRestrictAud = 0;
    let howManyClausesRestrictUsages = 0;
    let expires = 0;
    let doesNotExpire = false;
    restrictions.forEach(function (r) {
        if (r['scope'] !== undefined) {
            howManyClausesRestrictScope++;
        }
        let aud = r['audience'];
        if (aud !== undefined && aud.length > 0) {
            howManyClausesRestrictAud++;
        }
        let ip = r['ip'];
        let ipW = r['geoip_white'];
        let ipB = r['geoip_black'];
        if ((ip !== undefined && ip.length > 0) ||
            (ipW !== undefined && ipW.length > 0) ||
            (ipB !== undefined && ipB.length > 0)) {
            howManyClausesRestrictIP++;
        }
        if (r['usages_other']!==undefined || r['usages_AT']!==undefined) {
            howManyClausesRestrictUsages++;
        }
        let exp = r['exp'];
        if (exp===undefined || exp===0) {
           doesNotExpire = true
        } else if (exp>expires) {
            expires=exp;
        }
    })
    if (doesNotExpire) {
        expires = 0;
    }
    let iconTime = $('#r-icon-time');
    let iconIP = $('#r-icon-ip');
    let iconScope = $('#r-icon-scope');
    let iconAud = $('#r-icon-aud');
    let iconUsages = $('#r-icon-usages');
    if (howManyClausesRestrictIP===restrictions.length) {
        iconIP.addClass( 'text-success');
        iconIP.removeClass( 'text-warning');
        iconIP.removeClass( 'text-danger');
        iconIP.attr('data-original-title', "The IPs from which this token can be used are restricted.");
    } else {
        iconIP.addClass( 'text-warning');
        iconIP.removeClass( 'text-success');
        iconIP.removeClass( 'text-danger');
        iconIP.attr('data-original-title', "This token can be used from any IP.");
    }
    if (howManyClausesRestrictScope===restrictions.length) {
        iconScope.addClass( 'text-success');
       iconScope.removeClass( 'text-warning');
       iconScope.removeClass( 'text-danger');
       iconScope.attr('data-original-title', "This token has restrictions for scopes.");
    } else {
       iconScope.addClass( 'text-warning');
       iconScope.removeClass( 'text-success');
       iconScope.removeClass( 'text-danger');
       iconScope.attr('data-original-title', "This token can use all configured scopes.");
    }
    if (howManyClausesRestrictAud===restrictions.length) {
        iconAud.addClass( 'text-success');
      iconAud.removeClass( 'text-warning');
      iconAud.removeClass( 'text-danger');
       iconAud.attr('data-original-title', "This token can only obtain access tokens with restricted audiences.");
    } else {
      iconAud.addClass( 'text-warning');
       iconAud.removeClass( 'text-success');
       iconAud.removeClass( 'text-danger');
       iconAud.attr('data-original-title', "This token can obtain access tokens with any audiences.");
    }
    if (howManyClausesRestrictUsages===restrictions.length) {
       iconUsages.addClass( 'text-success');
       iconUsages.removeClass( 'text-warning');
       iconUsages.removeClass( 'text-danger');
       iconUsages.attr('data-original-title', "This token can only be used a limited number of times.");
    } else {
      iconUsages.addClass( 'text-warning');
      iconUsages.removeClass( 'text-success');
      iconUsages.removeClass( 'text-danger');
      iconUsages.attr('data-original-title', "This token can be used an infinite number of times.");
    }
    if (expires===0) {
       iconTime.addClass( 'text-danger');
      iconTime.removeClass( 'text-success');
       iconTime.removeClass( 'text-warning');
       iconTime.attr('data-original-title', "This token does not expire!");
    } else if ((expires - Date.now()/1000)> 3*24*3600) {
       iconTime.addClass( 'text-warning');
       iconTime.removeClass( 'text-success');
       iconTime.removeClass( 'text-danger');
       iconTime.attr('data-original-title', "This token is long-lived.");
    } else {
       iconTime.addClass( 'text-success');
       iconTime.removeClass( 'text-warning');
       iconTime.removeClass( 'text-danger');
       iconTime.attr('data-original-title', "This token expires within 3 days.");
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
    let res = $('#restrictions');
    try {
        r = JSON.parse(res.val());
        res.removeClass('is-invalid');
        res.addClass('is-valid');
    } catch (e) {
        res.removeClass('is-valid');
        res.addClass('is-invalid');
       return;
    }
    restrictions = r;
    parseRestriction();
}

function approve() {
    let data = {
        "restrictions": restrictions,
        "capabilities": $('.capability-check:checked').map(function(_, el) {
            return $(el).val();
        }).get(),
        "subtoken_capabilities": $('.subtoken-capability-check:checked').map(function(_, el) {
            return $(el).val();
        }).get()
    };
    data = JSON.stringify(data);
    console.log(data);
    $.ajax({
        type: "POST",
        url: window.location.href,
        data: data,
        success: function (data){
            window.location.href = data['authorization_url'];
        },
        error: function(errRes){
            let errMsg = getErrorMessage(errRes);
            console.log(errMsg);
        },
        dataType: "json",
        contentType : "application/json"
    });
}

function cancel() {
    //TODO POST cancel
    window.location.href = "/";
}

$('#cp-create_super_token').click(function() {
    let enabled = $(this).prop("checked");
    $('.subtoken-capability-check').prop("checked", enabled);
    $('.subtoken-capability-check').prop("disabled", !enabled);
});
