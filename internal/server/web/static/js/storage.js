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
        v = _localStorage().getItem(key);
    }
    return JSON.parse(v);
}

function storagePop(key, session = false) {
    let v = storageGet(key);
    storageSet(key, "", session);
    return v;
}

function storageSet(key, value, session = false) {
    return _storage(session).setItem(key, JSON.stringify(value));
}
