$.fn.serializeObject = function () {
    let o = {};
    let a = this.serializeArray();
    $.each(a, function () {
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

$.fn.showB = function () {
    this.removeClass('d-none');
}
$.fn.hideB = function () {
    this.addClass('d-none');
}

$('.my-expand').on('click', function () {
    const expand = "Expand";
    const collapse = "Collapse";
    if ($(this).text() === expand) {
        $(this).text(collapse);
    } else {
        $(this).text(expand);
    }
});

function getErrorMessage(e) {
    let errRes = e.responseJSON
    let err = errRes['error'];
    let desc = errRes['error_description'];
    if (desc) {
        err += ": " + desc;
    }
    let status = e.statusText
    return status + ": " + err
}

function noLandscape(prefix) {
    let landscaped = $('.' + prefix + '-landscape');
    landscaped.removeClass('col');
    landscaped.removeClass('row');
    landscaped.removeClass('form-row');
}

function escapeSelector(s) {
    return $.escapeSelector(s)
}

function doNext(...next) {
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

function onlyUnique(value, index, self) {
    return self.indexOf(value) === index;
}

function extractMaxScopesFromToken(token) {
    let restr = token['restrictions'];
    if (!restr) {
        return "";
    }
    let scopes = [];
    for (const r of restr) {
        let s = r['scope'];
        if (!s || s === "") { // if any restriction allows all scopes, return ""
            return "";
        }
        scopes.push(...s.split(' '));
    }
    return scopes.filter(onlyUnique).join(" ")
}

function formatTime(t) {
    return new Date(t * 1000).toLocaleString()
}