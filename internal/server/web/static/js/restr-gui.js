const usagesAT = $('#usages_AT');
const usagesOther = $('#usages_other');
const restrClauses = $('#restr-clauses');

$(function(){
    let date_input=$('.datetimepicker-input');
    let options={
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
    date_input.datetimepicker(options);

    $(".nbf").on("change.datetimepicker", function (e) {
        $('#'+$(this).attr("id").replace("nbf", "exp")).datetimepicker("minDate", e.date);
        GUIToRestr_Nbf();
    });
    $(".exp").on("change.datetimepicker", function (e) {
        $('#'+$(this).attr("id").replace("exp", "nbf")).datetimepicker("maxDate", e.date);
        GUIToRestr_Exp();
    });

    for (const c in countries) {
        let opt = document.createElement('option');
        opt.value = c;
        opt.innerHTML = c;
        $('.country-select').append(opt);
    }
    $(".country-select").prop("selectedIndex", -1);

    const scopes = typeof(supported_scopes)!=='undefined' ? supported_scopes : getSupportedScopesFromStorage();
    for (const scope of scopes) {
        let html = `<tr>
                            <td><span class="table-item">`+scope+`</span></td>
                            <td>
                                <i class="fas fa-check-circle text-success scope-active"></i>
                                <i class="fas fa-times-circle text-danger scope-inactive d-none"></i>
                            </td>
                            <td><input class="form-check-input scope-checkbox" id="scope_`+scope+`" type="checkbox" value="`+scope+`"></td>
                        </tr>`;
        $('#scopeTableBody').append(html);
    }

    $('.scope-checkbox').on("click", function (){
        let allScopesInactive = $('.scope-inactive');
        let allScopesActive = $('.scope-active');
        let checkedScopeBoxes = $('.scope-checkbox:checked');
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
        GUIToRestr_Scopes();
    })

    RestrToGUI();
})

function getSupportedScopesFromStorage() {
    const providers = storageGet("providers_supported");
    const iss = typeof(issuer)!=='undefined' ? issuer : storageGet("oidc_issuer");
    return  providers.find(x => x.issuer === iss).scopes_supported;
}


$('.btn-add-list-item').on("click", function (){
    let input = $(this).parents('tr').find('input,select');
    let v = input.val();
    if (v==="") {
        return;
    }
    let tableBody = $(this).parents('table').find('.list-table')
    _addListItem(v, tableBody);
})

$('.add-list-input').on('keyup', function(e){
    if (e.keyCode === 13) { // Enter
        e.preventDefault();
        $(this).parents('tr').find('.btn-add-list-item').click();
    }
})

function GUIGetActiveClause() {
    return Number($('.active-restr-clause').text())-1;
}

function GUISetRestr(key, value) {
    restrictions[GUIGetActiveClause()][key]=value
    updateRestrIcons();
}
function GUIDelRestr(key) {
    delete restrictions[GUIGetActiveClause()][key]
    updateRestrIcons();
}

function _guiToRestr_Time(id, restrKey) {
    let date = $(id).datetimepicker('date');
    if (date === null) {
        GUIDelRestr(restrKey);
        return;
    }
    let unix;
    if (date._isAMomentObject) {
        unix = date.unix();
    } else {
        unix = date.getTime() / 1000;
    }
    GUISetRestr(restrKey, unix);
}

function _guiToRestr_Table(id, restrKey) {
    let body = $(id);
    let items = body.find('.table-item')
    if (items.length === 0) {
        GUIDelRestr(restrKey);
        return;
    }
    let parseCountry = restrKey.startsWith("geoip");
    let values = [];
    items.each(function (){
        let v = $(this).text();
        if (parseCountry) {
            v = countries[v].code
        }
        values.push(v);
    })
    GUISetRestr(restrKey, values);
}

function GUIToRestr_Nbf() {
    _guiToRestr_Time('#nbf', 'nbf');
}

function GUIToRestr_Exp() {
    _guiToRestr_Time('#exp', 'exp');
}

function GUIToRestr_Scopes() {
    let checkedScopeBoxes = $('.scope-checkbox:checked');
    if (checkedScopeBoxes.length === 0) {
        GUIDelRestr('scope');
        return;
    }
    let scopes = []
    checkedScopeBoxes.each(function(i){
        scopes.push($(this).val());
    })
    GUISetRestr('scope', scopes.join(' '));
}

function GUIToRestr_IPs() {
    _guiToRestr_Table('#ipTableBody', "ip");
}

function GUIToRestr_GeoIPAllow() {
    _guiToRestr_Table('#geoip-allowTableBody', "geoip_allow");
}

function GUIToRestr_GeoIPDisallow() {
    _guiToRestr_Table('#geoip-disallowTableBody', "geoip_disallow");
}

function GUIToRestr_UsagesAT() {
    let usages = usagesAT.val()
    if (usages === "") {
        GUIDelRestr('usages_AT');
    } else {
        GUISetRestr('usages_AT', Number(usages));
    }
}

function GUIToRestr_UsagesOther() {
    let usages = usagesOther.val()
    if (usages === "") {
        GUIDelRestr('usages_other');
    } else {
        GUISetRestr('usages_other', Number(usages));
    }
}

usagesAT.on('change', GUIToRestr_UsagesAT);
usagesOther.on('change', GUIToRestr_UsagesOther);

function _addListItem(value, tableBody) {
    let html = `<tr><td class="align-middle"><span class="table-item">`+value+`</span></td><td class="align-middle"><button class="btn btn-small btn-delete-list-item"><i class="fas fa-minus"></i></button></td></tr>`;
    tableBody.append(html);
    $('.btn-delete-list-item').off("click").on("click", function (){
        let tablebodyId = $(this).parents('.list-table').attr("id");
        $(this).parents("tr").remove();
        _guiToRestr_Table("#"+tablebodyId, tablebodyId.split('TableBody')[0]);
    })
    let tbodyId = tableBody.attr("id");
    _guiToRestr_Table("#"+tbodyId, tbodyId.split('TableBody')[0]);
}

function restrClauseToGUI() {
    let restr = restrictions[GUIGetActiveClause()];

    let tmp = restr.scope;
    $('.scope-checkbox:checked').click();
    if (tmp){
        restr.scope = tmp;
    }

    $('.list-table').html("");

    let nbf = restr.nbf;
    let exp = restr.exp;

    $('#nbf').datetimepicker('date', nbf?nbf.toString():null);
    $('#exp').datetimepicker('date', exp?exp.toString():null);

    if (restr.scope) {
        for (const s of restr.scope.split(' ')) {
            $('#scope_'+escapeSelector(s)).click();
        }
    }
    if (restr.ip) {
        for (const ip of restr.ip) {
            _addListItem(ip, $('#ipTableBody'));
        }
    }
    if (restr.geoip_allow) {
        for (const code of restr.geoip_allow) {
            let country = countriesByCode[code.toUpperCase()];
            _addListItem(country, $('#geoip_allowTableBody'));
        }
    }
    if (restr.geoip_disallow) {
        for (const code of restr.geoip_disallow) {
            let country = countriesByCode[code.toUpperCase()];
            _addListItem(country, $('#geoip_disallowTableBody'));
        }
    }
    usagesAT.val(restr.usages_AT);
    usagesOther.val(restr.usages_other);
}

function newRestrClauseBtn(index) {
    let btn = `<button type="button" id="restr-clause-`+index+`" class="btn btn-info restr-btn">`+index+`</button>`;
    restrClauses.append(btn);
    $('#restr-clause-'+index).on('click', function (){
        let i = Number($(this).text())-1;
        GUIMarkActiveClause(i);
        restrClauseToGUI();
    })
}

function drawRestrictionClauseBtns() {
    restrClauses.find('button.restr-btn').remove();
    for (let i=0; i<restrictions.length; i++) {
        let index = i+1;
        newRestrClauseBtn(index);
    }
}

function RestrToGUI() {
    drawRestrictionClauseBtns();
    GUIMarkActiveClause(0);
    restrClauseToGUI();
}

function GUIMarkActiveClause(index) {
    let guiIndex = index+1;
    $('.active-restr-clause').removeClass('active-restr-clause active');
    $('#restr-clause-'+guiIndex).addClass('active-restr-clause active');
}

$('#new-restr-clause').on('click', function (){
    restrictions.push({});
    let guiIndex = restrictions.length;
    newRestrClauseBtn(guiIndex);
    $('#restr-clause-'+guiIndex).click();
    updateRestrIcons();
})

$('#del-restr-clause').on('click', function (){
    let index = GUIGetActiveClause();
    restrictions.splice(index, 1);
    updateRestrIcons();
    drawRestrictionClauseBtns();
    let newGuiIndex = index===0 ? 1 : index; // Activate the previous clause or the first one if no we deleted the first one
    $('#restr-clause-'+newGuiIndex).click();
})

$('#select-ip-based-restr').on('change', function (){
    $('#restr-ip').hideB();
    $('#restr-geoip_allow').hideB();
    $('#restr-geoip_disallow').hideB();
    $('#restr-'+$(this).val()).showB();
})