import createMuiTheme, { Theme, ThemeOptions } from "@material-ui/core/styles/createMuiTheme";
import { Palette } from "@material-ui/core/styles/createPalette";

interface IColor {
    main: string;
    light: string;
    dark: string;
    contrastText: string;
}

interface IPalette extends Palette {
    primaryBackground: IColor
    secondaryBackground: IColor
}

export interface ITheme extends Theme {
    palette: IPalette;
}

interface IThemeOptions extends ThemeOptions {
    palette: IPalette;
    overrides: any;
}

const theme = createMuiTheme({
    "palette": {
        "common": {
            "black": "#000",
            "white": "#fff"
        },
        "type": "dark",
        "primary": {
            "main": "#627ef7",
            "light": "#a0b2ff",
            "dark": "#2b3560",
            "contrastText": "#090b14"
        },
        "primaryBackground": {
            "main": "#171c30",
            "light": "#252c47",
            "dark": "#090b14",
            "contrastText": "#fff"
        },
        "secondary": {
            "main": "rgba(255, 255, 255, 0.75)",
            "light": "#fff",
            "dark": "rgba(255, 255, 255, 0.5)",
            "contrastText": "#000"
        },
        "secondaryBackground": {
            "main": "rgba(255, 255, 255, 0.12)",
            "light": "rgba(255, 255, 255, 0.25)",
            "dark": "rgba(255, 255, 255, 0.06)",
            "contrastText": "#fff"
        },
        "error": {
            "main": "#cb5411",
            "light": "#ff3a30",
            "dark": "rgba(255, 58, 48, 0.35)",
            "contrastText": "#fff"
        },
        "warning": {
            "light": "#ffb74d",
            "main": "#ff9800",
            "dark": "#f57c00",
            "contrastText": "#344073"
        },
        "info": {
            "light": "#64b5f6",
            "main": "#2196f3",
            "dark": "#1976d2",
            "contrastText": "#fff"
        },
        "success": {
            "light": "#3eb545",
            "main": "rgba(62, 181, 69, 0.5)",
            "dark": "rgba(62, 181, 69, 0.35)",
            "contrastText": "#fff"
        },
        "grey": {
            "50": "#fafafa",
            "100": "#f5f5f5",
            "200": "#eeeeee",
            "300": "#e0e0e0",
            "400": "#bdbdbd",
            "500": "#9e9e9e",
            "600": "#757575",
            "700": "#616161",
            "800": "#424242",
            "900": "#212121",
            "A100": "#d5d5d5",
            "A200": "#aaaaaa",
            "A400": "#303030",
            "A700": "#616161"
        },
        "contrastThreshold": 3,
        "tonalOffset": 0.2,
        "text": {
            "primary": "#fff",
            "secondary": "rgba(255, 255, 255, 0.85)",
            "disabled": "rgba(255, 255, 255, 0.75)",
            "hint": "rgba(255, 255, 255, 0.65)",
        },
        "divider": "rgba(255, 255, 255, 0.12)",
        "background": {
            "paper": "#344073",
            "default": "#252c47",
        },
        "action": {
            "active": "#fff",
            "hover": "rgba(255, 255, 255, 0.08)",
            "hoverOpacity": 0.08,
            "selected": "rgba(255, 255, 255, 0.16)",
            "selectedOpacity": 0.16,
            "disabled": "rgba(255, 255, 255, 0.3)",
            "disabledBackground": "rgba(255, 255, 255, 0.12)",
            "disabledOpacity": 0.38,
            "focus": "rgba(255, 255, 255, 0.12)",
            "focusOpacity": 0.12,
            "activatedOpacity": 0.24
        }
    },
    "props": {},
    "shadows": [
        "none",
        "0px 2px 1px -1px rgba(0,0,0,0.2),0px 1px 1px 0px rgba(0,0,0,0.14),0px 1px 3px 0px rgba(0,0,0,0.12)",
        "0px 3px 1px -2px rgba(0,0,0,0.2),0px 2px 2px 0px rgba(0,0,0,0.14),0px 1px 5px 0px rgba(0,0,0,0.12)",
        "0px 3px 3px -2px rgba(0,0,0,0.2),0px 3px 4px 0px rgba(0,0,0,0.14),0px 1px 8px 0px rgba(0,0,0,0.12)",
        "0px 2px 4px -1px rgba(0,0,0,0.2),0px 4px 5px 0px rgba(0,0,0,0.14),0px 1px 10px 0px rgba(0,0,0,0.12)",
        "0px 3px 5px -1px rgba(0,0,0,0.2),0px 5px 8px 0px rgba(0,0,0,0.14),0px 1px 14px 0px rgba(0,0,0,0.12)",
        "0px 3px 5px -1px rgba(0,0,0,0.2),0px 6px 10px 0px rgba(0,0,0,0.14),0px 1px 18px 0px rgba(0,0,0,0.12)",
        "0px 4px 5px -2px rgba(0,0,0,0.2),0px 7px 10px 1px rgba(0,0,0,0.14),0px 2px 16px 1px rgba(0,0,0,0.12)",
        "0px 5px 5px -3px rgba(0,0,0,0.2),0px 8px 10px 1px rgba(0,0,0,0.14),0px 3px 14px 2px rgba(0,0,0,0.12)",
        "0px 5px 6px -3px rgba(0,0,0,0.2),0px 9px 12px 1px rgba(0,0,0,0.14),0px 3px 16px 2px rgba(0,0,0,0.12)",
        "0px 6px 6px -3px rgba(0,0,0,0.2),0px 10px 14px 1px rgba(0,0,0,0.14),0px 4px 18px 3px rgba(0,0,0,0.12)",
        "0px 6px 7px -4px rgba(0,0,0,0.2),0px 11px 15px 1px rgba(0,0,0,0.14),0px 4px 20px 3px rgba(0,0,0,0.12)",
        "0px 7px 8px -4px rgba(0,0,0,0.2),0px 12px 17px 2px rgba(0,0,0,0.14),0px 5px 22px 4px rgba(0,0,0,0.12)",
        "0px 7px 8px -4px rgba(0,0,0,0.2),0px 13px 19px 2px rgba(0,0,0,0.14),0px 5px 24px 4px rgba(0,0,0,0.12)",
        "0px 7px 9px -4px rgba(0,0,0,0.2),0px 14px 21px 2px rgba(0,0,0,0.14),0px 5px 26px 4px rgba(0,0,0,0.12)",
        "0px 8px 9px -5px rgba(0,0,0,0.2),0px 15px 22px 2px rgba(0,0,0,0.14),0px 6px 28px 5px rgba(0,0,0,0.12)",
        "0px 8px 10px -5px rgba(0,0,0,0.2),0px 16px 24px 2px rgba(0,0,0,0.14),0px 6px 30px 5px rgba(0,0,0,0.12)",
        "0px 8px 11px -5px rgba(0,0,0,0.2),0px 17px 26px 2px rgba(0,0,0,0.14),0px 6px 32px 5px rgba(0,0,0,0.12)",
        "0px 9px 11px -5px rgba(0,0,0,0.2),0px 18px 28px 2px rgba(0,0,0,0.14),0px 7px 34px 6px rgba(0,0,0,0.12)",
        "0px 9px 12px -6px rgba(0,0,0,0.2),0px 19px 29px 2px rgba(0,0,0,0.14),0px 7px 36px 6px rgba(0,0,0,0.12)",
        "0px 10px 13px -6px rgba(0,0,0,0.2),0px 20px 31px 3px rgba(0,0,0,0.14),0px 8px 38px 7px rgba(0,0,0,0.12)",
        "0px 10px 13px -6px rgba(0,0,0,0.2),0px 21px 33px 3px rgba(0,0,0,0.14),0px 8px 40px 7px rgba(0,0,0,0.12)",
        "0px 10px 14px -6px rgba(0,0,0,0.2),0px 22px 35px 3px rgba(0,0,0,0.14),0px 8px 42px 7px rgba(0,0,0,0.12)",
        "0px 11px 14px -7px rgba(0,0,0,0.2),0px 23px 36px 3px rgba(0,0,0,0.14),0px 9px 44px 8px rgba(0,0,0,0.12)",
        "0px 11px 15px -7px rgba(0,0,0,0.2),0px 24px 38px 3px rgba(0,0,0,0.14),0px 9px 46px 8px rgba(0,0,0,0.12)"
    ],
    "typography": {
        "htmlFontSize": 16,
        "fontFamily": "'Source Sans Pro', 'Lucida Grande', 'Lucida Sans Unicode', 'Geneva', 'Verdana', sans-serif",
        "fontSize": 14,
        "fontWeightLight": 300,
        "fontWeightRegular": 400,
        "fontWeightMedium": 500,
        "fontWeightBold": 600,
        "h1": {
            "fontFamily": "'Source Sans Pro', 'Lucida Grande', 'Lucida Sans Unicode', 'Geneva', 'Verdana', sans-serif",
            "fontWeight": 300,
            "fontSize": "6rem",
            "lineHeight": 1.167,
            "letterSpacing": "-0.01562em"
        },
        "h2": {
            "fontFamily": "'Source Sans Pro', 'Lucida Grande', 'Lucida Sans Unicode', 'Geneva', 'Verdana', sans-serif",
            "fontWeight": 300,
            "fontSize": "3.75rem",
            "lineHeight": 1.2,
            "letterSpacing": "-0.00833em"
        },
        "h3": {
            "fontFamily": "'Source Sans Pro', 'Lucida Grande', 'Lucida Sans Unicode', 'Geneva', 'Verdana', sans-serif",
            "fontWeight": 400,
            "fontSize": "3rem",
            "lineHeight": 1.167,
            "letterSpacing": "0em"
        },
        "h4": {
            "fontFamily": "'Source Sans Pro', 'Lucida Grande', 'Lucida Sans Unicode', 'Geneva', 'Verdana', sans-serif",
            "fontWeight": 400,
            "fontSize": "2.125rem",
            "lineHeight": 1.235,
            "letterSpacing": "0.00735em"
        },
        "h5": {
            "fontFamily": "'Source Sans Pro', 'Lucida Grande', 'Lucida Sans Unicode', 'Geneva', 'Verdana', sans-serif",
            "fontWeight": 400,
            "fontSize": "1.2rem",
            "lineHeight": 1.334,
            "letterSpacing": "0.02em"
        },
        "h6": {
            "fontFamily": "'Source Sans Pro', 'Lucida Grande', 'Lucida Sans Unicode', 'Geneva', 'Verdana', sans-serif",
            "fontWeight": 400,
            "fontSize": "1rem",
            "lineHeight": 1.6,
            "letterSpacing": "0.0075em",
            "textTransform": "uppercase"
        },
        "subtitle1": {
            "fontFamily": "'Source Sans Pro', 'Lucida Grande', 'Lucida Sans Unicode', 'Geneva', 'Verdana', sans-serif",
            "fontWeight": 400,
            "fontSize": "1rem",
            "lineHeight": 1.75,
            "letterSpacing": "0.00938em"
        },
        "subtitle2": {
            "fontFamily": "'Source Sans Pro', 'Lucida Grande', 'Lucida Sans Unicode', 'Geneva', 'Verdana', sans-serif",
            "fontWeight": 500,
            "fontSize": "0.875rem",
            "lineHeight": 1.57,
            "letterSpacing": "0.00714em"
        },
        "body1": {
            "fontFamily": "'Source Sans Pro', 'Lucida Grande', 'Lucida Sans Unicode', 'Geneva', 'Verdana', sans-serif",
            "fontWeight": 400,
            "fontSize": "1rem",
            "lineHeight": 1.5,
            "letterSpacing": "0.00938em"
        },
        "body2": {
            "fontFamily": "'Source Sans Pro', 'Lucida Grande', 'Lucida Sans Unicode', 'Geneva', 'Verdana', sans-serif",
            "fontWeight": 400,
            "fontSize": "0.875rem",
            "lineHeight": 1.43,
            "letterSpacing": "0.01071em"
        },
        "button": {
            "fontFamily": "'Source Sans Pro', 'Lucida Grande', 'Lucida Sans Unicode', 'Geneva', 'Verdana', sans-serif",
            "fontWeight": 500,
            "fontSize": "0.875rem",
            "lineHeight": 1.75,
            "letterSpacing": "0.02857em",
            "textTransform": "uppercase"
        },
        "caption": {
            "fontFamily": "'Source Sans Pro', 'Lucida Grande', 'Lucida Sans Unicode', 'Geneva', 'Verdana', sans-serif",
            "fontWeight": 400,
            "fontSize": "0.8rem",
            "lineHeight": 1.66,
            "letterSpacing": "0.03333em"
        },
        "overline": {
            "fontFamily": "'Source Sans Pro', 'Lucida Grande', 'Lucida Sans Unicode', 'Geneva', 'Verdana', sans-serif",
            "fontWeight": 400,
            "fontSize": "0.75rem",
            "lineHeight": 2.66,
            "letterSpacing": "0.08333em",
            "textTransform": "uppercase"
        }
    },
    "shape": {
        "borderRadius": 16
    },
    "transitions": {
        "easing": {
            "easeInOut": "cubic-bezier(0.4, 0, 0.2, 1)",
            "easeOut": "cubic-bezier(0.0, 0, 0.2, 1)",
            "easeIn": "cubic-bezier(0.4, 0, 1, 1)",
            "sharp": "cubic-bezier(0.4, 0, 0.6, 1)"
        },
        "duration": {
            "shortest": 150,
            "shorter": 200,
            "short": 250,
            "standard": 300,
            "complex": 375,
            "enteringScreen": 225,
            "leavingScreen": 195
        }
    },
    "zIndex": {
        "mobileStepper": 1000,
        "speedDial": 1050,
        "appBar": 1100,
        "drawer": 1200,
        "modal": 1300,
        "snackbar": 1400,
        "tooltip": 1500
    }
} as IThemeOptions) as ITheme;

theme.overrides = {

    ...theme.overrides,

    MuiTooltip: {
        tooltip: {
            backgroundColor: "#2a335a",
            fontSize: "14px",
            borderRadius: "8px",
        }
    },

    "MuiButton": {
        label: {
            fontWeight: "bold",
        },
        "contained": {
            color: '#F7F9FC',
            padding: "7px 18px",
        },
        "containedPrimary": {

            color: '#F7F9FC',

            "&:disabled": {
                backgroundColor: "rgba(0,0,0,0.15)",
                color: "rgba(255, 255, 255, 0.8)"
            }
        },
        "outlinedPrimary": {

            padding: "8px 18px",
            "&:hover": {
                backgroundColor: "rgba(255, 255, 255, 0.06)"
            },

            "&:disabled": {
                backgroundColor: "rgba(255, 255, 255, 0.06)",
                color: "rgba(255, 255, 255, 0.25)"
            },

        }
    },

    "MuiSwitch": {
        root: {
            padding: 11,
            width: 44,
            height: 29,
        },
        switchBase: {
            '&$checked': {
                transform: "translateX(10px)",
                "& $thumb": {
                    backgroundColor: theme.palette.common.white,
                },
                "&.MuiSwitch-colorPrimary + $track": {
                    backgroundColor: theme.palette.secondaryBackground.dark,
                    border: "3px solid " + theme.palette.primary.light,
                    opacity: 1,
                },
            }
        },
        thumb: {
            width: 8,
            height: 8,
            border: "3px solid " + theme.palette.primary.main,
            backgroundColor: theme.palette.primary.dark,
            boxShadow: 'none',
        },
        track: {
            width: 14,
            height: 4,
            backgroundColor: theme.palette.secondaryBackground.dark,
            border: "3px solid " + theme.palette.primary.light,
            opacity: 1,
        }
    }
};

export default theme;

