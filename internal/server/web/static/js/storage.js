
function _storage() {
    return window.sessionStorage
}

function storageGet(key) {
    return JSON.parse(_storage().getItem(key));
}

function storageSet(key, value) {
    return _storage().setItem(key, JSON.stringify(value));
}
