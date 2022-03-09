const isNil = require("lodash/isNil");

const VERBOSE_ARG = "--verbose";
const CONFIG_ARG = "--config";

function createNullableArg(value) {
    return {
        isProvided: !isNil(value),
        value
    };
}

const processedArgs = {
    verbose: false,
    config: createNullableArg()
};

function findCliArg(key) {
    const index = process.argv.indexOf(key);

    return {
        key,
        index,
        isProvided: index !== -1
    };
}

function getCliArgWithValue(key) {
    const result = (isProvided = false, value, index) => ({
        key,
        isProvided,
        value,
        index
    });

    const arg = findCliArg(key);

    if (arg.isProvided) {
        const valueIndex = arg.index + 1;

        if (process.argv[valueIndex]) {
            return result(true, process.argv[valueIndex], arg.index);
        }
    }

    return result();
}

const jestConflictingArg = (key, hasValue = false) => ({
    key,
    hasValue
});

// prettier-ignore
const jestConflictingArgs = [
    jestConflictingArg(CONFIG_ARG, true),
];

function removeJestConflictingCustomCliArgs() {
    jestConflictingArgs.forEach(x => {
        const arg = findCliArg(x.key);

        if (arg.isProvided) {
            process.argv.splice(arg.index, x.hasValue ? 2 : 1);
        }
    });
}

function findArgsFromCli() {
    const verbose = findCliArg(VERBOSE_ARG);
    const config = getCliArgWithValue(CONFIG_ARG);

    removeJestConflictingCustomCliArgs();

    const values = {
        verbose: verbose.isProvided ? true : undefined,
        config: config.isProvided ? config.value : undefined
    };

    setArgs(values);
}

function setArgs(values) {
    if (!isNil(values)) {
        if (!isNil(values.verbose)) {
            processedArgs.verbose = values.verbose;
        }

        if (!isNil(values.config)) {
            processedArgs.config = createNullableArg(values.config);
        }
    }
}

function getArgs() {
    return processedArgs;
}

module.exports = {
    getArgs,
    setArgs,
    findArgsFromCli
};
