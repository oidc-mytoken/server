function createTags(tags) {
    return tags.map(function (tag) {
        return getTagPill(tag);
    }).join("")
}

function getTagPill(tag) {
    return `<span class="badge badge-pill ${textClassForBackgroundColor(tag.color)} tag" style="background-color: #${tag.color};">${tag.tag}</span>`;
}

function textClassForBackgroundColor(backgroundColor) {
    const r = parseInt(backgroundColor.substring(0, 2), 16);
    const g = parseInt(backgroundColor.substring(2, 4), 16);
    const b = parseInt(backgroundColor.substring(4, 6), 16);

    // Convert to relative luminance (per WCAG)
    function toLuminance(c) {
        let s = c / 255;
        return s <= 0.03928 ? s / 12.92 : Math.pow((s + 0.055) / 1.055, 2.4);
    }

    let L = 0.2126 * toLuminance(r) + 0.7152 * toLuminance(g) + 0.0722 * toLuminance(b);

    return L > 0.5 ? "text-black" : "text-white";
}