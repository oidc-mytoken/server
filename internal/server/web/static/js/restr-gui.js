function usagesAT(prefix = "") {
    return $('#' + prefix + 'usages_AT');
}

function usagesOther(prefix = "") {
    return $('#' + prefix + 'usages_other');
}

function restrClauses(prefix = "") {
    return $('#' + prefix + 'restr-clauses');
}

function scopeTableBody(prefix = "") {
    return $('#' + prefix + 'scopeTableBody');
}

function initDatetimePicker() {
    let $date_input = $('.datetimepicker-input');
    let options = {
        format: 'YYYY-MM-DD HH:mm:ss',
        extraFormats: ['X'],
        minDate: Date.now(),
        // keepInvalid: true,
        buttons: {
            showClear: true,
        },
        icons: {
            time: 'fas fa-clock',
            date: 'fas fa-calendar',
            up: 'fas fa-arrow-up',
            down: 'fas fa-arrow-down',
            previous: 'fas fa-chevron-left',
            next: 'fas fa-chevron-right',
            today: 'fas fa-calendar-check-o',
            clear: 'fas fa-trash',
            close: 'fas fa-times'
        }
    };
    $date_input.datetimepicker(options);

    $(".nbf").on("change.datetimepicker", function (e) {
        let id = this.id;
        let prefix = extractPrefix("nbf", id);
        $('#' + id.replace("nbf", "exp")).datetimepicker("minDate", e.date);
        GUIToRestr_Nbf(prefix);
    });
    $(".exp").on("change.datetimepicker", function (e) {
        let id = this.id;
        let prefix = extractPrefix("exp", id);
        $('#' + id.replace("exp", "nbf")).datetimepicker("maxDate", e.date);
        GUIToRestr_Exp(prefix);
    });
}

function initCountries() {
    for (const c in countries) {
        let opt = document.createElement('option');
        opt.value = c;
        opt.innerHTML = c;
        $('.country-select').append(opt);
    }
    $(".country-select").prop("selectedIndex", -1);
}

function _addScopeValueToGUI(scope, $htmlEl, prefix, prefixprefix = "") {
    let disabled = "";
    if ('restrictions' in getPrefixData(prefixprefix)) {
        disabled = getPrefixData(prefixprefix)['restrictions']['read-only'] ? " disabled" : "";
    }
    let html = `<tr>
                            <td><span class="table-item">${scope}</span></td>
                            <td>
                                <i class="fas fa-check-circle text-success scope-active"></i>
                                <i class="fas fa-times-circle text-danger scope-inactive d-none"></i>
                            </td>
                            <td><input class="form-check-input scope-checkbox" id="${prefixprefix}${prefix}_scope_${scope}" type="checkbox" value="${scope}" instance-prefix="${prefixprefix}"${disabled}></td>
                        </tr>`;
    $htmlEl.append(html);
}


$(document).ready(function () {
    initDatetimePicker();
    initCountries();
})

function initRestrGUI(prefix = "") {
    if (prefix !== tokeninfoPrefix) {
        const scopes = typeof (supported_scopes) !== 'undefined' ? supported_scopes : getSupportedScopesFromStorage();
        for (const scope of scopes) {
            _addScopeValueToGUI(scope, scopeTableBody(prefix), "restr", prefix);
        }
    }

    $('.scope-checkbox[instance-prefix="' + prefix + '"]').on("click", function () {
        let allScopesInactive = $(this).parents('tbody').find('.scope-inactive');
        let allScopesActive = $(this).parents('tbody').find('.scope-active');
        let checkedScopeBoxes = $(this).parents('tbody').find('.scope-checkbox:checked');
        let activeIcon = $(this).parents('tr').find('.scope-active');
        let inactiveIcon = $(this).parents('tr').find('.scope-inactive');
        let activated = $(this).prop('checked');
        if (activated) {
            if (checkedScopeBoxes.length === 1) { // There was no box checked before, but now one has been checked
                allScopesActive.hideB();
                allScopesInactive.showB();
            }
            inactiveIcon.hideB();
            activeIcon.showB();
        } else {
            activeIcon.hideB();
            inactiveIcon.showB();
        }
        if (checkedScopeBoxes.length === 0) {
            allScopesInactive.hideB();
            allScopesActive.showB();
        }
    })
    scopeTableBody(prefix).find('.scope-checkbox').on("click", function () {
        GUIToRestr_Scopes(prefix);
    })
    $(prefixId("select-ip-based-restr", prefix)).trigger('change');

    RestrToGUI(prefix);
}

function getSupportedScopesFromStorage(iss = "") {
    const providers = storageGet("providers_supported");
    if (iss === "") {
        iss = typeof (issuer) !== 'undefined' ? issuer : storageGet("oidc_issuer");
    }
    return providers.find(x => x.issuer === iss).scopes_supported;
}


$('.btn-add-list-item').on("click", function () {
    let input = $(this).parents('tr').find('input,select');
    let v = input.val();
    if (v === "") {
        return;
    }
    let tableBody = $(this).parents('table').find('.list-table')
    _addListItem(v, tableBody, $(this).attr("instance-prefix"));
})

$('.add-list-input').on('keyup', function (e) {
    if (e.keyCode === 13) { // Enter
        e.preventDefault();
        $(this).parents('tr').find('.btn-add-list-item').click();
    }
})

function GUIGetActiveClause(prefix = "") {
    return Number($('.active-restr-clause[instance-prefix="' + prefix + '"]').text()) - 1;
}

function GUISetRestr(key, value, prefix = "") {
    let restr = getRestrictionsData(prefix)
    restr[GUIGetActiveClause(prefix)][key] = value
    updateRestrIcons(prefix);
}

function GUIDelRestr(key, prefix = "") {
    let restr = getRestrictionsData(prefix)
    delete restr[GUIGetActiveClause(prefix)][key]
    updateRestrIcons(prefix);
}

function _guiToRestr_Time(id, restrKey, prefix = "") {
    let date = $(id).datetimepicker('date');
    if (date === null) {
        GUIDelRestr(restrKey, prefix);
        return;
    }
    let unix;
    if (date._isAMomentObject) {
        unix = date.unix();
    } else {
        unix = date.getTime() / 1000;
    }
    GUISetRestr(restrKey, unix, prefix);
}

function _guiToRestr_Table(id, restrKey, prefix = "") {
    let body = $(id);
    let items = body.find('.table-item')
    if (items.length === 0) {
        GUIDelRestr(restrKey, prefix);
        return;
    }
    let parseCountry = restrKey.startsWith("geoip");
    let values = [];
    items.each(function () {
        let v = $(this).text();
        if (parseCountry) {
            v = countries[v].code
        }
        values.push(v);
    })
    GUISetRestr(restrKey, values.filter(onlyUnique), prefix);
}

function GUIToRestr_Nbf(prefix = "") {
    _guiToRestr_Time(prefixId("nbf", prefix), 'nbf', prefix);
}

function GUIToRestr_Exp(prefix = "") {
    _guiToRestr_Time(prefixId("exp", prefix), 'exp', prefix);
}

function GUIToRestr_Scopes(prefix = "") {
    let checkedScopeBoxes = $(prefixId('scopeTableBody', prefix)).find('.scope-checkbox:checked');
    if (checkedScopeBoxes.length === 0) {
        GUIDelRestr('scope', prefix);
        return;
    }
    let scopes = []
    checkedScopeBoxes.each(function () {
        scopes.push($(this).val());
    })
    GUISetRestr('scope', scopes.join(' '), prefix);
}

//
// function GUIToRestr_IPs(prefix="") {
//     _guiToRestr_Table(prefixId('ipTableBody',prefix), "ip",prefix);
// }
//
// function GUIToRestr_GeoIPAllow(prefix="") {
//     _guiToRestr_Table(prefixId('geoip-allowTableBody',prefix), "geoip_allow",prefix);
// }
//
// function GUIToRestr_GeoIPDisallow(prefix="") {
//     _guiToRestr_Table(prefixId('geoip-disallowTableBody',prefix), "geoip_disallow",prefix);
// }

function GUIToRestr_UsagesAT(prefix = "") {
    let usages = usagesAT(prefix).val()
    if (usages === "") {
        GUIDelRestr('usages_AT', prefix);
    } else {
        GUISetRestr('usages_AT', Number(usages), prefix);
    }
}

function GUIToRestr_UsagesOther(prefix = "") {
    let usages = usagesOther(prefix).val()
    if (usages === "") {
        GUIDelRestr('usages_other', prefix);
    } else {
        GUISetRestr('usages_other', Number(usages), prefix);
    }
}

$('.usages_AT').on('change', function () {
    GUIToRestr_UsagesAT(extractPrefix('usages_AT', this.id));
});
$('.usages_other').on('change', function () {
    GUIToRestr_UsagesOther(extractPrefix('usages_other', this.id));
});

function _addListItem(value, tableBody, prefix = "") {
    let ro = getPrefixData(prefix)['restrictions']['read-only'];
    let disabled = ro ? " disabled" : "";
    let hide = ro ? " d-none" : "";
    let html = `<tr><td class="align-middle"><span class="table-item">${value}</span></td><td class="align-middle"><button class="btn btn-small btn-delete-list-item${hide}"${disabled}><i class="fas fa-minus"></i></button></td></tr>`;
    tableBody.append(html);
    $('.btn-delete-list-item').off("click").on("click", function () {
        let tablebodyId = $(this).parents('.list-table').attr("id");
        $(this).parents("tr").remove();
        if (tableBody.hasClass('restr')) {
            _guiToRestr_Table("#" + tablebodyId, tablebodyId.substring(prefix.length).split('TableBody')[0], prefix);
        }
    })
    let tbodyId = tableBody.attr("id");
    if (tableBody.hasClass('restr')) {
        _guiToRestr_Table("#" + tbodyId, tbodyId.substring(prefix.length).split('TableBody')[0], prefix);
    }
}

function restrClauseToGUI(prefix = "") {
    let restr = getRestrictionsData(prefix)[GUIGetActiveClause(prefix)];

    let tmp = restr.scope;
    for (let el of $('.scope-checkbox[instance-prefix="' + prefix + '"]:checked')) {
        let $s = $(el);
        let disabled = $s.attr('disabled');
        $s.attr('disabled', false);
        $s.click();
        $s.attr('disabled', disabled);
    }
    if (tmp) {
        restr.scope = tmp;
    }

    $('.list-table[instance-prefix="' + prefix + '"]').html("");

    let nbf = restr.nbf;
    let exp = restr.exp;

    $(prefixId('nbf', prefix)).datetimepicker('date', nbf ? nbf.toString() : null);
    $(prefixId('exp', prefix)).datetimepicker('date', exp ? exp.toString() : null);

    if (restr.scope) {
        for (const s of restr.scope.split(' ')) {
            let $s = $(prefixId('restr_scope_' + s, prefix));
            let disabled = $s.attr('disabled');
            $s.attr('disabled', false);
            $s.click();
            $s.attr('disabled', disabled);
        }
    }
    if (restr.audience) {
        for (const a of restr.audience) {
            _addListItem(a, $(prefixId('audienceTableBody', prefix)), prefix);
        }
    }
    if (restr.ip) {
        for (const ip of restr.ip) {
            _addListItem(ip, $(prefixId('ipTableBody', prefix)), prefix);
        }
    }
    if (restr.geoip_allow) {
        for (const code of restr.geoip_allow) {
            let country = countriesByCode[code.toUpperCase()];
            _addListItem(country, $(prefixId('geoip_allowTableBody', prefix)), prefix);
        }
    }
    if (restr.geoip_disallow) {
        for (const code of restr.geoip_disallow) {
            let country = countriesByCode[code.toUpperCase()];
            _addListItem(country, $(prefixId('geoip_disallowTableBody', prefix)), prefix);
        }
    }
    usagesAT(prefix).val(restr.usages_AT);
    usagesOther(prefix).val(restr.usages_other);
}

function newRestrClauseBtn(index, prefix = "") {
    let btn = `<button type="button" id="${prefix}restr-clause-${index}" class="btn btn-info restr-btn" instance-prefix="${prefix}">${index}</button>`;
    restrClauses(prefix).append(btn);
    $(prefixId('restr-clause-' + index, prefix)).on('click', function () {
        let i = Number($(this).text()) - 1;
        GUIMarkActiveClause(i, prefix);
        restrClauseToGUI(prefix);
    })
}

function drawRestrictionClauseBtns(prefix = "") {
    restrClauses(prefix).find('button.restr-btn').remove();
    for (let i = 0; i < getRestrictionsData(prefix).length; i++) {
        newRestrClauseBtn(i + 1, prefix);
    }
}

function RestrToGUI(prefix = "") {
    drawRestrictionClauseBtns(prefix);
    GUIMarkActiveClause(0, prefix);
    restrClauseToGUI(prefix);
}

function GUIMarkActiveClause(index, prefix) {
    let guiIndex = index + 1;
    $('.active-restr-clause[instance-prefix="' + prefix + '"]').removeClass('active-restr-clause active');
    $(prefixId('restr-clause-' + guiIndex, prefix)).addClass('active-restr-clause active');
}

function newRestrClause(prefix = "") {
    let restr = getRestrictionsData(prefix);
    restr.push({});
    let guiIndex = restr.length;
    newRestrClauseBtn(guiIndex);
    updateRestrIcons(prefix);
    drawRestrictionClauseBtns(prefix);
    $(prefixId('restr-clause-' + guiIndex, prefix)).click();
}

function delRestrClause(prefix = "") {
    let index = GUIGetActiveClause(prefix);
    let restr = getRestrictionsData(prefix);
    restr.splice(index, 1);
    updateRestrIcons(prefix);
    drawRestrictionClauseBtns(prefix);
    let newGuiIndex = index === 0 ? 1 : index; // Activate the previous clause or the first one if no we deleted the first one
    $(prefixId('restr-clause-' + newGuiIndex, prefix)).click();
}

function selectIPTable(prefix = "") {
    $(prefixId('restr-ip', prefix)).hideB();
    $(prefixId('restr-geoip_allow', prefix)).hideB();
    $(prefixId('restr-geoip_disallow', prefix)).hideB();
    $(prefixId('restr-' + $(this).val(), prefix)).showB();
}