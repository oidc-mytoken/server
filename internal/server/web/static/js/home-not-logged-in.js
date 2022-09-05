let loggedIn = false;

$(function () {
    chainFunctions(
        discovery,
        initCreateMT,
        initTokeninfo,
    );
    openCorrectTab();
})

$('#login-iss').on('change', function () {
    initRestrGUI(mtPrefix);
});