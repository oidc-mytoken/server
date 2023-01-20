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
    if ($(this).text().trim() === expand) {
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
    landscaped.removeClass('col-md');
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

function extractPrefix(normalID, prefixedID) {
    return prefixedID.substring(0, prefixedID.indexOf(normalID))
}

function prefixId(id, prefix = "") {
    return '#' + prefix + escapeSelector(id);
}

let clipboard = new ClipboardJS('.copier');
clipboard.on('success', function (e) {
    e.clearSelection();
    let el = $(e.trigger);
    let originalText = el.attr('data-original-title');
    el.attr('data-original-title', 'Copied!').tooltip('show');
    el.attr('data-original-title', originalText);
});

function _enableProfileSupport(template_prefix, set_payload_in_gui, prefix = "") {
    let $template_select = $(prefixId(`${template_prefix}-template`, prefix));
    $template_select.val("");
    $template_select.on('change', function () {
        const v = $(this).val();
        if (v === null || v === "") {
            return;
        }
        const payload = JSON.parse(v);
        set_payload_in_gui(payload, prefix);
    });
    let $checks = $(`.any-${template_prefix}-input[instance-prefix=${prefix}]`);
    if (template_prefix === "cap") {
        $checks = capabilityChecks(prefix);
    }
    $checks.on('change change.datetimepicker', function (e) {
        if ($(e.currentTarget).hasClass('datetimepicker-input') && datetimepickerChangeTriggeredFromJS) {
            return;
        }
        $template_select.val(""); // reset template selection to custom if something is changed
    });
}

function is_string(str) {
    return typeof str === 'string' || str instanceof String;
}

function select_set_value_by_option_name($select, option) {
    const options = Array.from($select.get()[0].options);
    const optionToSelect = options.find(item => item.text === option);
    const value = optionToSelect.value;
    $select.val(value);
}