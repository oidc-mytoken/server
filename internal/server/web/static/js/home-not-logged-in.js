let loggedIn = false;

$(function () {
    chainFunctions(
        initCreateMT,
        initTokeninfo,
    );
    openCorrectTab();
})
