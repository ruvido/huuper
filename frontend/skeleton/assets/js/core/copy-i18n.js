// Translation layer for window.appCopy. Copy files stay English (they are the
// source of truth); a translation file overlays a localized version of the same
// keys on top of a given subtree, e.g.:
//
//   window.appCopyI18n.translate("retreats.public", { brand: "...", nav: {...} });
//
// Keys missing from the translation keep their English value, so a partial
// translation degrades gracefully instead of rendering blanks.
window.appCopyI18n = (() => {
  function isPlainObject(value) {
    return !!value && typeof value === "object" && !Array.isArray(value);
  }

  function deepMerge(base, overrides) {
    const out = Object.assign({}, base);
    Object.keys(overrides || {}).forEach((key) => {
      const next = overrides[key];
      out[key] = isPlainObject(next) && isPlainObject(base && base[key])
        ? deepMerge(base[key], next)
        : next;
    });
    return out;
  }

  // Replaces the subtree at `path` (dotted) with the English copy merged
  // under the given translations.
  function translate(path, translations) {
    const parts = String(path || "").split(".").filter(Boolean);
    if (parts.length === 0) return;
    window.appCopy = window.appCopy || {};

    let node = window.appCopy;
    for (const part of parts.slice(0, -1)) {
      if (!isPlainObject(node[part])) node[part] = {};
      node = node[part];
    }
    const leaf = parts[parts.length - 1];
    node[leaf] = deepMerge(node[leaf] || {}, translations || {});
  }

  return { translate };
})();
