export const singleEntryToHAR = (entry) => {

    if (!entry) return null;

    const modifiedEntry = {
        ...entry,
        "startedDateTime": "2019-10-08T11:49:51.090+03:00",
        "cache": {},
        "timings": {
            "blocked": -1,
            "dns": -1,
            "connect": -1,
            "ssl": -1,
            "send": -1,
            "wait": -1,
            "receive": -1
        },
        "time": -1
    };

    const har = {
        log: {
            entries: [modifiedEntry],
            version: "1.2",
            creator: {
                "name": "Firefox",
                "version": "69.0.1"
            }
        }
    }

    return har;
};

export const formatSize = (n: number) => n > 1000 ? `${Math.round(n / 1000)}KB` : `${n} B`;
