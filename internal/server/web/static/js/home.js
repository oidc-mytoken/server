$(function (){
    chainFunctions(
        checkIfLoggedIn,
        initAT,
        initRestrGUI,
    );
    // https://stackoverflow.com/a/17552459
    // Javascript to enable link to tab
    let url = document.location.toString();
    if (url.match('#')) {
        let hash = url.split('#')[1];
        if (['token-info', 'token-history', 'token-tree'].includes(hash)) {
            $('.nav-tabs a[href="#info"]').tab('show') ;
        }
        $('.nav-tabs a[href="#'+hash+'"]').tab('show') ;
    }
})


// With HTML5 history API, we can easily prevent scrolling!
$('.nav-tabs a').on('shown.bs.tab', function (e) {
    if (history.pushState) {
        history.pushState(null, null, e.target.hash);
    } else {
        window.location.hash = e.target.hash; //Polyfill for old browsers
    }
    let $found = $(this).parents('.card').find('.tab-pane.active .nav-tabs a.active');
    if ($found.attr('id') !== $(this).attr('id')) {
        $found.triggerHandler('shown.bs.tab');
    }
})


function getMT(okCallback, errCallback, capability="AT") {
    let data = {
        "name":"mytoken-web MT for "+capability,
        "grant_type": "mytoken",
        "capabilities": [capability],
        "restrictions": [
            {
                "exp":  Math.floor(Date.now() / 1000) + 60,
                "ip": ["this"],
                "usages_AT": capability==="AT" ? 1 : 0,
                "usages_other": capability==="AT" ? 0 : 1
            }
        ]
    };
    requestMT(data, okCallback, errCallback);
}
