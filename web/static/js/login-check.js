
function checkIfLoggedIn() {
    $.ajax({
        type: "POST",
        url: "/api/v0/test", //TODO
        success: function(res){
            if (window.location.pathname == "/") {
            window.location.href = "/home";
            }
        },
        error: function (res) {
            console.log(res);
            if (window.location.pathname != "/") {
                window.location.href = "/";
            }
        },
        contentType : "application/json"
    });
}

checkIfLoggedIn()
