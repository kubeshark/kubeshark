import {makeStyles} from "@material-ui/core";

// @ts-ignore
export const useCommonStyles = makeStyles(() => ({
    button: {
        backgroundColor: "#205cf5",
        color: "white",
        fontWeight: "600 !important",
        fontSize: 13,
        padding: "6px 12px",

        "&:hover": {
            backgroundColor: "#205cf5",
        },
    },
    outlinedButton: {
        backgroundColor: "transparent",
        color: "#205cf5",
        fontWeight: "600 !important",
        fontSize: 13,
        padding: "6px 12px",
        border: "1px #205cf5 solid",

        "&:hover": {
            backgroundColor: "transparent",
        },
    }
}));
