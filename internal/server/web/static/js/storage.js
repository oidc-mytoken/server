
function _storage() {
    return window.sessionStorage
}

function storageGet(key) {
    return _storage().getItem(key)
}

function storageSet(key, value) {
    return _storage().setItem(key, value)
}
