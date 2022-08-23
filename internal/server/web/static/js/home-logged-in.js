let loggedIn = true;

$(function () {
    chainFunctions(
        checkIfLoggedIn,
        initAT,
        initCreateMT,
        initTokeninfo,
    );
    openCorrectTab();
})
