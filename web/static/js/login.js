


$('#login-form').on('submit', function(e){
    e.preventDefault();
    let data = $(this).serializeObject()
    data['restrictions'] = [
        {
            "exp": Math.floor(Date.now() / 1000) + 3600*24, // TODO configurable
            "ip": ["this"],
            "usages_AT": 1,
            "usages_other": 50,
        }
    ]
    data['capabilities'] = [
        "create_mytoken",
        "tokeninfo_introspect",
        "tokeninfo_history",
        "tokeninfo_tree"
    ]
    data['subtoken_capabilities'] = [
        "AT",
        // "settings",
        "list_mytokens"
    ]
    data['name'] = "mytoken-web";
    data = JSON.stringify(data);
    $.ajax({
        type: "POST",
        url: storageGet("mytoken_endpoint"),
        data: data,
        success: function(res){
            window.location.href = res['authorization_url'];
        },
        dataType: "json",
        contentType : "application/json"
    });
    return false;
});
