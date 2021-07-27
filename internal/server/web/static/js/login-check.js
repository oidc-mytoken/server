
function checkIfLoggedIn() {
    let data = {
     'action':'introspect'
    };
    data = JSON.stringify(data);
    $.ajax({
        type: "POST",
        url: storageGet('tokeninfo_endpoint'),
        data: data,
        success: function(res){
            if (window.location.pathname === "/") {
            window.location.href = "/home";
            }
        },
        error: function (res) {
            if (window.location.pathname !== "/") {
                window.location.href = "/";
            }
        },
        dataType: "json",
        contentType : "application/json"
    });
}

checkIfLoggedIn()
