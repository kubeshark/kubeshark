function when(condition, fct, unmetValue) {
    if (condition) {
        return fct();
    }

    return unmetValue;
}

function whenDev(fct, unmetValue) {
    return when(process.env.NODE_ENV === "development", fct, unmetValue);
}

function whenProd(fct, unmetValue) {
    return when(process.env.NODE_ENV === "production", fct, unmetValue);
}

function whenTest(fct, unmetValue) {
    return when(process.env.NODE_ENV === "test", fct, unmetValue);
}

module.exports = {
    when,
    whenDev,
    whenProd,
    whenTest
};
