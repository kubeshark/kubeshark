const mergeWith = require("lodash/mergeWith");

function isFunction(value) {
    return typeof value === "function";
}

function isArray(value) {
    return Array.isArray(value);
}

function isString(value) {
    return typeof value === "string";
}

function isBoolean(value) {
    return typeof value === "boolean";
}

function deepMergeWithArray(...rest) {
    return mergeWith(...rest, (x, y) => {
        if (isArray(x)) {
            return x.concat(y);
        }
    });
}

module.exports = {
    isFunction,
    isArray,
    isString,
    isBoolean,
    deepMergeWithArray
};
