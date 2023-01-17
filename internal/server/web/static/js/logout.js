function logout() {
    revokeMT(function () {
        storageClear();
        window.location.href = "/"
    });
}