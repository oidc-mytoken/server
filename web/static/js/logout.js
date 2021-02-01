
function logout() {
    revokeST(function () {
        window.location.href = "/"
    })
}