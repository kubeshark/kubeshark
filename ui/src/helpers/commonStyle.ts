import {makeStyles} from "@material-ui/core";

// @ts-ignore
export const useCommonStyles = makeStyles(() => ({
    button: {
        backgroundColor: "#205cf5",
        color: "white",
        fontWeight: "600 !important",
        fontSize: 12,
        padding: "8px 12px",
        borderRadius: "6px ! important",

        "&:hover": {
            backgroundColor: "#205cf5",
        },
    },
    outlinedButton: {
        backgroundColor: "transparent",
        color: "#205cf5",
        fontWeight: "600 !important",
        fontSize: 12,
        padding: "8px 12px",
        border: "1px #205cf5 solid",
        borderRadius: "6px ! important",

        "&:hover": {
            backgroundColor: "transparent",
        },
    },
    textField: {
        outline: 0,
        background: "white",
        borderRadius: "4px",
        padding: "8px 10px",
        border: "1px #9D9D9D solid",
        fontSize: "14px",
        color: "#494677"
    },
    imagedButton: {
        padding: "0px 14px"
    }
}));
