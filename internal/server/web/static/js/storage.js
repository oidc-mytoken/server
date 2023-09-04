function _sessionStorage() {
    return window.sessionStorage
}

function _localStorage() {
    return window.localStorage
}

function _storage(session = false) {
    return session ? _sessionStorage() : _localStorage();
}

function storageGet(key) {
    let v = _sessionStorage().getItem(key)
    if (!v) {
        return undefined;
    }
    return JSON.parse(v);
}

function storagePop(key) {
    let v = storageGet(key);
    storageSet(key, "");
    return v;
}

function storageSet(key, value) {
    return _storage(true).setItem(key, JSON.stringify(value));
}

function storageClear() {
    _storage(true).clear();
}
