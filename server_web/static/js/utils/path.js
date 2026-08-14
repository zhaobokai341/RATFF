const resolvePath = function(base, target) {
    if (!target) return base;

    var isWindows = target.includes(':') || (base && base.includes(':'));
    var sep = isWindows ? '\\' : '/';

    if (isWindows && target.includes(':') && target[1] === ':') {
        var resolved = normalizePathParts(target.split(/[\/\\]/), sep);
        return resolved.join(sep);
    }

    if (target.startsWith('/') || (isWindows && target.startsWith('\\'))) {
        var resolved = normalizePathParts(target.split(/[\/\\]/), sep);
        return isWindows ? resolved.join(sep) : sep + resolved.join(sep);
    }

    var basePath = base || '';
    var baseParts = basePath.split(/[\/\\]/).filter(function(p) { return p; });
    var targetParts = target.split(/[\/\\]/).filter(function(p) { return p; });
    var allParts = baseParts.concat(targetParts);
    var resolved = normalizePathParts(allParts, sep);

    if (!isWindows && basePath.startsWith('/')) {
        return sep + resolved.join(sep);
    }
    if (isWindows && basePath.includes(':')) {
        var drive = basePath.substring(0, 2);
        return drive + sep + resolved.join(sep);
    }
    return resolved.join(sep);
};

const normalizePathParts = function(parts, sep) {
    var stack = [];
    for (var i = 0; i < parts.length; i++) {
        var part = parts[i];
        if (part === '.' || part === '') {
            continue;
        } else if (part === '..') {
            if (stack.length > 0) {
                stack.pop();
            }
        } else {
            stack.push(part);
        }
