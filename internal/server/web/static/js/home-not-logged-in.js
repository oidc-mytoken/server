let loggedIn = false;

$(function () {
    chainFunctions(
        discovery,
        initCreateMT,
        initTokeninfo,
    );
    openCorrectTab();
})
