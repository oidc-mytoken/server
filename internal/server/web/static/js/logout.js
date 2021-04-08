
function logout() {
    revokeMT(function () {
        window.location.href = "/"
    })
}