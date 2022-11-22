import makeStyles from '@mui/styles/makeStyles';

// @ts-ignore
export const useCommonStyles = makeStyles(() => ({
    button: {
        backgroundColor: "#326de6",
        color: "white",
        fontWeight: "600 !important",
        fontSize: 12,
        padding: "8px 12px",
        borderRadius: "6px ! important",

        "&:hover": {
            backgroundColor: "#326de6",
        },
    },
    outlinedButton: {
        backgroundColor: "transparent",
        color: "#326de6",
        fontWeight: "600 !important",
        fontSize: 12,
        padding: "8px 12px",
        border: "1px #326de6 solid",
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
