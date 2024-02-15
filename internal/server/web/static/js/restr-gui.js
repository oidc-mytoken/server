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

function initDatetimePicker(prefix = "") {
    let options = {
        format: 'YYYY-MM-DD HH:mm:ss',
        extraFormats: ['X'],
        minDate: Date.now(),
        // keepInvalid: true,
        buttons: {
            showClear: true,
        },
        useCurrent: false,
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
    let $nbf = $(prefixId('nbf', prefix));
    let $exp = $(prefixId('exp', prefix));
    $nbf.datetimepicker(options);
    $exp.datetimepicker(options);

    $nbf.on("change.datetimepicker", function (e) {
        let id = this.id;
        $('#' + id.replace("nbf", "exp")).datetimepicker("minDate", e.date || Date.now() / 1000);
        GUIToRestr_Nbf(e.date, prefix);
    });
    $exp.on("change.datetimepicker", function (e) {
        let id = this.id;
        $('#' + id.replace("exp", "nbf")).datetimepicker("maxDate", e.date);
        GUIToRestr_Exp(e.date, prefix);
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
                            <td><input class="form-check-input scope-checkbox any-restr-input" id="${prefixprefix}${prefix}_scope_${scope}" type="checkbox" value="${scope}" instance-prefix="${prefixprefix}"${disabled}></td>
                        </tr>`;
    $htmlEl.append(html);
}


$(document).ready(function () {
    initCountries();
})

function initRestrGUI(prefix = "") {
    initDatetimePicker(prefix);
    if (typeof (tokeninfoPrefix) === 'undefined' || prefix !== tokeninfoPrefix) {
        scopeTableBody(prefix).html("");
        const scopes = typeof (supported_scopes) !== 'undefined' ? supported_scopes : getSupportedScopesFromStorage();
        for (const scope of scopes) {
            _addScopeValueToGUI(scope, scopeTableBody(prefix), "restr", prefix);
        }
    }

    initScopesGUI(prefix);
    scopeTableBody(prefix).find('.scope-checkbox').on("click", function () {
        GUIToRestr_Scopes(prefix);
    })
    $(prefixId("select-ip-based-restr", prefix)).trigger('change');

    RestrToGUI(prefix);
}

function getSupportedScopesFromStorage(iss = "") {
    const providers = storageGet("providers_supported");
    if (iss === "") {
        if (typeof (issuer) !== 'undefined') {
            iss = issuer;
        } else if (typeof ($mtOIDCIss !== 'undefined')) {
            iss = $mtOIDCIss.val();
        } else {
            iss = storageGet("oidc_issuer");
        }
    }
    let p = providers.find(x => x.issuer === iss)
    if (p === undefined) {
        return [];
    }
    return p.scopes_supported;
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

function _guiToRestr_Time(date, restrKey, prefix = "") {
    if (!date) {
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

function GUIToRestr_Nbf(date, prefix = "") {
    _guiToRestr_Time(date, 'nbf', prefix);
}

function GUIToRestr_Exp(date, prefix = "") {
    _guiToRestr_Time(date, 'exp', prefix);
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
    let ro = false;
    if ('restrictions' in getPrefixData(prefix)) {
        ro = getPrefixData(prefix)['restrictions']['read-only'];
    }
    let disabled = ro ? " disabled" : "";
    let hide = ro ? " d-none" : "";
    let html = `<tr><td class="align-middle"><span class="table-item">${value}</span></td><td class="align-middle"><button class="btn btn-small btn-delete-list-item any-restr-input${hide}"${disabled} instance-prefix="${prefix}"><i class="fas fa-minus"></i></button></td></tr>`;
    tableBody.append(html);
    $('.btn-delete-list-item').off("click").on("click", function () {
        let tablebodyId = $(this).parents('.list-table').attr("id");
        $(this).parents("tr").remove();
        if (tableBody.hasClass('restr')) {
            _guiToRestr_Table("#" + tablebodyId, tablebodyId.substring(prefix.length).split('TableBody')[0], prefix);
        }
        if (restrictionProfileSupportEnableForPrefixes.includes(prefix)) {
            $(prefixId('restr-template', prefix)).val("");
            $(prefixId(`profile-template`, prefix)).val("");
        }
    })
    let tbodyId = tableBody.attr("id");
    if (tableBody.hasClass('restr')) {
        _guiToRestr_Table("#" + tbodyId, tbodyId.substring(prefix.length).split('TableBody')[0], prefix);
    }
}

let datetimepickerChangeTriggeredFromJS = false;

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

    datetimepickerChangeTriggeredFromJS = true;
    $(prefixId('nbf', prefix)).datetimepicker('date', nbf ? nbf.toString() : null);
    $(prefixId('exp', prefix)).datetimepicker('date', exp ? exp.toString() : null);
    datetimepickerChangeTriggeredFromJS = false;

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
    if (restr.hosts) {
        for (const host of restr.hosts) {
            _addListItem(host, $(prefixId('hostsTableBody', prefix)), prefix);
        }
    } else if (restr.ip) {
        for (const ip of restr.ip) {
            _addListItem(ip, $(prefixId('hostsTableBody', prefix)), prefix);
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
    $(prefixId('restr-hosts', prefix)).hideB();
    $(prefixId('restr-geoip_allow', prefix)).hideB();
    $(prefixId('restr-geoip_disallow', prefix)).hideB();
    $(prefixId('restr-' + $(this).val(), prefix)).showB();
}

function set_restrictions_in_gui(restrictions, prefix = "") {
    if (restrictions === undefined) {
        return;
    }
    restrictions.forEach(function (r, i, restrs) {
        // we delete the includes because they were already applied and keeping it can lead to confusion, especially
        // when restrictions should be removed, but then they are re-added because of the include
        delete r.include;
        restrs[i] = r;
    });
    setRestrictionsData(restrictions, prefix);
    if (editorMode(prefix).prop('checked')) {
        RestrToGUI(prefix);
    } else {
        fillJSONEditor(prefix);
    }
}

let restrictionProfileSupportEnableForPrefixes = [];

function restr_enableProfileSupport(prefix = "") {
    _enableProfileSupport("restr", set_restrictions_in_gui, prefix);
    restrictionProfileSupportEnableForPrefixes.push(prefix);
}