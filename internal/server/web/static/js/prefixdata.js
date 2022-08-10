noPrefix = "no-prefix";

function initPrefixData(prefix = "") {
    if (typeof prefixData === 'undefined') {
        prefixData = {};
    }
    if (prefix === "") {
        prefix = noPrefix;
    }
    if (!(prefix in prefixData)) {
        prefixData[prefix] = {};
    }
}


function getPrefixData(prefix = "") {
    initPrefixData(prefix);
    return prefixData[prefix === "" ? noPrefix : prefix];
}
