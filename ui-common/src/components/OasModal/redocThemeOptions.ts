const fontFamilyVar = "Source Sans Pro, Lucida Grande, Tahoma, sans-serif"

export const redocThemeOptions = {
  theme: {
    codeBlock: {
      backgroundColor: "#14161c",
    },
    components: {
      buttons: {
        fontFamily: fontFamilyVar,
      },
      httpBadges: {
        fontFamily: fontFamilyVar,
      }
    },
    colors: {
      responses: {
        error: {
          tabTextColor: "#1b1b29"
        },
        info: {
          tabTextColor: "#1b1b29",
        },
        success: {
          tabTextColor: "#0c0b1a"
        },
      },
      text: {
        primary: "#1b1b29",
        secondary: "#4d4d4d"
      }
    },
    rightPanel: {
      backgroundColor: "#0D0B1D",
    },
    sidebar: {
      backgroundColor: "#ffffff"
    },
    typography: {
      code: {
        color: "#0c0b1a",
        fontFamily: fontFamilyVar
      },
      fontFamily: fontFamilyVar,
      fontSize: "90%",
      fontWieght: "normal",
      headings: {
        fontFamily: fontFamilyVar
      }
    }
  }
}