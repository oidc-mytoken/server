let loggedIn = false;

$(function () {
    chainFunctions(
        initCreateMT,
        initTokeninfo,
    );
    openCorrectTab();
})

$('#login-iss').on('change', function () {
    initRestrGUI(mtPrefix);
});