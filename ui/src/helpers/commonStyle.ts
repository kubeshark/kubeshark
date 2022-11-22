import makeStyles from '@mui/styles/makeStyles';

// @ts-ignore
export const useCommonStyles = makeStyles(() => ({
    button: {
        backgroundColor: "#326de6",
        color: "white",
        fontWeight: "600 !important",
        fontSize: 12,
        padding: "9px 12px",
        borderRadius: "6px ! important",
        "&:hover": {
            backgroundColor: "#326de6",
        },
        "&:disabled": {
            backgroundColor: "rgba(0, 0, 0, 0.26)"
        }
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
    clickedButton: {
        color: "white",
        backgroundColor: "#326de6",
        "&:hover": {
            backgroundColor: "#326de6",
        },
    },
    imagedButton: {
        padding: "1px 14px"
    },

    textField: {
        outline: 0,
        background: "white",
        borderRadius: "4px",
        padding: "8px 10px",
        border: "1px #9D9D9D solid",
        fontSize: "14px",
        color: "#494677",
        height: "30px",
        boxSizing: "border-box"
    },
    modal: {
        position: 'absolute',
        top: '40%',
        left: '50%',
        transform: 'translate(-50%, -40%)',
        width: "CLAMP(600px,50%, 800px)",
        bgcolor: 'background.paper',
        borderRadius: '5px',
        boxShadow: 24,
        outline: "none",
        p: 4,
        color: '#000',
    }
}));
