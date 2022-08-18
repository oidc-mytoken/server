$(function () {
    chainFunctions(
        checkIfLoggedIn,
        initAT,
        initCreateMT,
        initTokeninfo,
    );
    openCorrectTab();
})
