function initScopesGUI(prefix = "") {
    $('.scope-checkbox[instance-prefix="' + prefix + '"]').on("click", function () {
        let allScopesInactive = $(this).parents('tbody').find('.scope-inactive');
        let allScopesActive = $(this).parents('tbody').find('.scope-active');
        let checkedScopeBoxes = $(this).parents('tbody').find('.scope-checkbox:checked');
        let activeIcon = $(this).parents('tr').find('.scope-active');
        let inactiveIcon = $(this).parents('tr').find('.scope-inactive');
        let activated = $(this).prop('checked');
        if (activated) {
            if (checkedScopeBoxes.length === 1) { // There was no box checked before, but now one has been checked
                allScopesActive.hideB();
                allScopesInactive.showB();
            }
            inactiveIcon.hideB();
            activeIcon.showB();
        } else {
            activeIcon.hideB();
            inactiveIcon.showB();
        }
        if (checkedScopeBoxes.length === 0) {
            allScopesInactive.hideB();
            allScopesActive.showB();
        }
    })

}
