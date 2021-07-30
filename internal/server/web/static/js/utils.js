

$.fn.serializeObject = function() {
    let o = {};
    let a = this.serializeArray();
    $.each(a, function() {
        if (o[this.name]) {
            if (!o[this.name].push) {
                o[this.name] = [o[this.name]];
            }
            o[this.name].push(this.value || '');
        } else {
            o[this.name] = this.value || '';
        }
    });
    return o;
};

$.fn.showB= function() {
    this.removeClass('d-none');
}
$.fn.hideB= function() {
    this.addClass('d-none');
}


function getErrorMessage(e) {
    let errRes = e.responseJSON
    let err = errRes['error'];
    let desc = errRes['error_description'];
    if (desc) {
        err += ": " + desc;
    }
    let status = e.statusText
    return status + ": "+ err
}

function noLandscape(prefix) {
    let landscaped = $('.'+prefix+'-landscape');
    landscaped.removeClass('col');
    landscaped.removeClass('row');
    landscaped.removeClass('form-row');
}

function escapeSelector(s){
    return s.replace( /(:|\.|\[|]|\/)/g, "\\$1" );
}

function doNext(...next) {
    console.log(next);
    switch (next.length) {
        case 0:
            return;
        case 1:
            return next[0]();
        default:
            let other = next.splice(1);
            return next[0](...other);
    }
}

function chainFunctions(...fncs) {
    switch (fncs.length) {
        case 0:
            return;
        case 1:
            return fncs[0]();
        default:
            let other = fncs.splice(1);
            return fncs[0](...other);
    }
}