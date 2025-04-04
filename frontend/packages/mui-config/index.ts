import { createTheme } from '@mui/material/styles'
import { PaletteMode, Theme, ThemeOptions } from '@mui/material'
import { TypographyOptions } from '@mui/material/styles/createTypography'

type FontFamilyType = TypographyOptions['fontFamily']

const baseThemeFactory = (mode: PaletteMode, fontFamily: FontFamilyType): Theme =>
  createTheme({
    breakpoints: createBreakpoints(),
    spacing: (factor: number) => `${0.25 * factor}rem`,
    typography: createTypography(fontFamily),
    palette: createPalette(mode),
    shadows: createShadows(mode)
  })

export const themeFactory = (fontFamily: FontFamilyType, components?: Theme['components']): Theme => {
  const base = baseThemeFactory('light', fontFamily)

  return createTheme(base, {
    components: {
      MuiAlert: {
        styleOverrides: {
          root: ({ ownerState }) => ({
            ...(ownerState.severity === 'info' && {
              backgroundColor: '#60a5fa'
            })
          })
        }
      },

      MuiLink: {
        styleOverrides: {
          root: {
            color: base.palette.grey[800],
            textDecorationColor: base.palette.grey[800],
            letterSpacing: '0.2px',
            lineHeight: '1.5'
          }
        }
      },

      ...components
    }
  } as ThemeOptions)
}

const createTypography = (fontFamily: FontFamilyType): TypographyOptions => {
  return {
    fontFamily,
    h1: {
      fontWeight: 500,
      letterSpacing: '-1.5px'
    },
    h2: {
      fontWeight: 500,
      letterSpacing: '-0.5px'
    },
    h3: {
      fontWeight: 500,
      letterSpacing: 0
    },
    h4: {
      fontWeight: 500,
      letterSpacing: '0.25px'
    },
    h5: {
      fontSize: '1.125rem',
      fontStyle: 'normal',
      fontWeight: 600,
      lineHeight: '1',
      marginBottom: '0.5rem'
    },
    h6: {
      fontSize: '1rem',
      fontStyle: 'normal',
      fontWeight: 600,
      lineHeight: '0.8',
      marginBottom: 16
    },
    subtitle1: {
      letterSpacing: '0.15px'
    },
    subtitle2: {
      letterSpacing: '0.1px'
    },
    body1: {
      fontSize: '14px',
      letterSpacing: '0.15px',
      lineHeight: '20px'
    },
    body2: {
      lineHeight: 1.5,
      letterSpacing: '0.15px'
    },
    button: {
      letterSpacing: '0.3px'
    },
    caption: {
      letterSpacing: '0.4px',
      fontSize: '0.875rem'
    },
    overline: {
      letterSpacing: '1px'
    }
  }
}

const createPalette = (mode: PaletteMode): ThemeOptions['palette'] => {
  const whiteColor = '#FFF'
  const lightColor = '0, 0, 0'
  const darkColor = '231, 227, 252'
  const mainColor = mode === 'light' ? lightColor : darkColor

  const defaultBgColor = () => {
    if (mode === 'light') {
      return '#ffffff'
    } else return '#28243D'
  }

  return {
    mode: 'light',
    //
    // @ts-ignore The customColors is not part of the default palette typings in every package where Material UI is used for now.
    // for example, if type is checked in portal we are good but not when checked in ui-core, so we simply ignore it for now
    customColors: {
      borderColor: 'rgba(0, 0, 0, 0.1)',
      entity: {
        matchedBg: '#EDEEFF',
        matchedBorder: '#B8B6FF',
        unmatchedBg: '#F8F8F8',
        unmatchedBorder: '#B0B0B0',
        label: '#717171',
        value: '#222222'
      },

      messageState: {
        unspecified: '#E0E0E0',
        unprocessed: '#777B82',
        queued: '#777B82',
        categorizing: '#777B82',
        extracting: '#777B82',
        processed: '#825AFF',
        failed: '#FF6166',
        archived: '#FFCA64'
      }
    },
    common: {
      black: '#000',
      white: whiteColor
    },
    primary: {
      light: '#E0D6FF',
      main: '#825AFF',
      dark: '#6244BF'
      // contrastText: whiteColor
    },
    secondary: {
      // light: '#9C9FA4',
      main: '#8A8D93'
      // dark: '#777B82',
      // contrastText: whiteColor
    },
    error: {
      // light: '#FF6166',
      main: '#FF4C51'
      // dark: '#E04347',
      // contrastText: whiteColor
    },
    warning: {
      // light: '#FFCA64',
      main: '#FFB400'
      // dark: '#E09E00',
      // contrastText: whiteColor
    },
    info: {
      // light: '#32BAFF',
      main: '#16B1FF'
      // dark: '#139CE0',
      // contrastText: whiteColor
    },
    success: {
      // light: '#6AD01F',
      main: '#56CA00'
      // dark: '#4CB200',
      // contrastText: whiteColor
    },

    grey: {
      50: '#FAFAFA',
      100: '#F5F5F5',
      200: '#EEEEEE',
      300: '#E0E0E0',
      400: '#BDBDBD',
      500: '#8B8B8B',
      600: '#757575',
      700: '#616161',
      800: '#424242',
      A100: '#F5F5F5',
      A200: '#EEEEEE',
      A400: '#BDBDBD',
      A700: '#616161'
    },
    text: {
      primary: `rgba(${mainColor}, 0.87)`,
      secondary: `rgba(${mainColor}, 0.6)`,
      disabled: `rgba(${mainColor}, 0.38)`
    },
    divider: `rgba(${mainColor}, 0.12)`,
    background: {
      paper: mode === 'light' ? whiteColor : '#312D4B',
      default: defaultBgColor()
    },
    action: {
      active: `rgba(${mainColor}, 0.54)`,
      hover: `rgba(${mainColor}, 0.04)`,
      selected: `rgba(${mainColor}, 0.08)`,
      disabled: `rgba(${mainColor}, 0.26)`,
      disabledBackground: `rgba(${mainColor}, 0.12)`,
      focus: `rgba(${mainColor}, 0.12)`
    }
  }
}

const createBreakpoints = () => {
  return {
    values: {
      xs: 0,
      sm: 600,
      md: 900,
      lg: 1200,
      xl: 1536
    }
  }
}

const createShadows = (mode: PaletteMode): ThemeOptions['shadows'] => {
  if (mode === 'light') {
    return [
      'none',
      '0px 2px 1px -1px rgba(58, 53, 65, 0.2), 0px 1px 1px 0px rgba(58, 53, 65, 0.14), 0px 1px 3px 0px rgba(58, 53, 65, 0.12)',
      '0px 3px 1px -2px rgba(58, 53, 65, 0.2), 0px 2px 2px 0px rgba(58, 53, 65, 0.14), 0px 1px 5px 0px rgba(58, 53, 65, 0.12)',
      '0px 4px 8px -4px rgba(58, 53, 65, 0.42)',
      '0px 6px 18px -8px rgba(58, 53, 65, 0.56)',
      '0px 3px 5px -1px rgba(58, 53, 65, 0.2), 0px 5px 8px 0px rgba(58, 53, 65, 0.14), 0px 1px 14px 0px rgba(58, 53, 65, 0.12)',
      '0px 2px 10px 0px rgba(58, 53, 65, 0.1)',
      '0px 4px 5px -2px rgba(58, 53, 65, 0.2), 0px 7px 10px 1px rgba(58, 53, 65, 0.14), 0px 2px 16px 1px rgba(58, 53, 65, 0.12)',
      '0px 5px 5px -3px rgba(58, 53, 65, 0.2), 0px 8px 10px 1px rgba(58, 53, 65, 0.14), 0px 3px 14px 2px rgba(58, 53, 65, 0.12)',
      '0px 5px 6px -3px rgba(58, 53, 65, 0.2), 0px 9px 12px 1px rgba(58, 53, 65, 0.14), 0px 3px 16px 2px rgba(58, 53, 65, 0.12)',
      '0px 6px 6px -3px rgba(58, 53, 65, 0.2), 0px 10px 14px 1px rgba(58, 53, 65, 0.14), 0px 4px 18px 3px rgba(58, 53, 65, 0.12)',
      '0px 6px 7px -4px rgba(58, 53, 65, 0.2), 0px 11px 15px 1px rgba(58, 53, 65, 0.14), 0px 4px 20px 3px rgba(58, 53, 65, 0.12)',
      '0px 7px 8px -4px rgba(58, 53, 65, 0.2), 0px 12px 17px 2px rgba(58, 53, 65, 0.14), 0px 5px 22px 4px rgba(58, 53, 65, 0.12)',
      '0px 7px 8px -4px rgba(58, 53, 65, 0.2), 0px 13px 19px 2px rgba(58, 53, 65, 0.14), 0px 5px 24px 4px rgba(58, 53, 65, 0.12)',
      '0px 7px 9px -4px rgba(58, 53, 65, 0.2), 0px 14px 21px 2px rgba(58, 53, 65, 0.14), 0px 5px 26px 4px rgba(58, 53, 65, 0.12)',
      '0px 8px 9px -5px rgba(58, 53, 65, 0.2), 0px 15px 22px 2px rgba(58, 53, 65, 0.14), 0px 6px 28px 5px rgba(58, 53, 65, 0.12)',
      '0px 8px 10px -5px rgba(58, 53, 65, 0.2), 0px 16px 24px 2px rgba(58, 53, 65, 0.14), 0px 6px 30px 5px rgba(58, 53, 65, 0.12)',
      '0px 8px 11px -5px rgba(58, 53, 65, 0.2), 0px 17px 26px 2px rgba(58, 53, 65, 0.14), 0px 6px 32px 5px rgba(58, 53, 65, 0.12)',
      '0px 9px 11px -5px rgba(58, 53, 65, 0.2), 0px 18px 28px 2px rgba(58, 53, 65, 0.14), 0px 7px 34px 6px rgba(58, 53, 65, 0.12)',
      '0px 9px 12px -6px rgba(58, 53, 65, 0.2), 0px 19px 29px 2px rgba(58, 53, 65, 0.14), 0px 7px 36px 6px rgba(58, 53, 65, 0.12)',
      '0px 10px 13px -6px rgba(58, 53, 65, 0.2), 0px 20px 31px 3px rgba(58, 53, 65, 0.14), 0px 8px 38px 7px rgba(58, 53, 65, 0.12)',
      '0px 10px 13px -6px rgba(58, 53, 65, 0.2), 0px 21px 33px 3px rgba(58, 53, 65, 0.14), 0px 8px 40px 7px rgba(58, 53, 65, 0.12)',
      '0px 10px 14px -6px rgba(58, 53, 65, 0.2), 0px 22px 35px 3px rgba(58, 53, 65, 0.14), 0px 8px 42px 7px rgba(58, 53, 65, 0.12)',
      '0px 11px 14px -7px rgba(58, 53, 65, 0.2), 0px 23px 36px 3px rgba(58, 53, 65, 0.14), 0px 9px 44px 8px rgba(58, 53, 65, 0.12)',
      '0px 11px 15px -7px rgba(58, 53, 65, 0.2), 0px 24px 38px 3px rgba(58, 53, 65, 0.14), 0px 9px 46px 8px rgba(58, 53, 65, 0.12)'
    ]
  } else {
    return [
      'none',
      '0px 2px 1px -1px rgba(19, 17, 32, 0.2), 0px 1px 1px 0px rgba(19, 17, 32, 0.14), 0px 1px 3px 0px rgba(19, 17, 32, 0.12)',
      '0px 3px 1px -2px rgba(19, 17, 32, 0.2), 0px 2px 2px 0px rgba(19, 17, 32, 0.14), 0px 1px 5px 0px rgba(19, 17, 32, 0.12)',
      '0px 4px 8px -4px rgba(19, 17, 32, 0.42)',
      '0px 6px 18px -8px rgba(19, 17, 32, 0.56)',
      '0px 3px 5px -1px rgba(19, 17, 32, 0.2), 0px 5px 8px rgba(19, 17, 32, 0.14), 0px 1px 14px rgba(19, 17, 32, 0.12)',
      '0px 2px 10px 0px rgba(19, 17, 32, 0.1)',
      '0px 4px 5px -2px rgba(19, 17, 32, 0.2), 0px 7px 10px 1px rgba(19, 17, 32, 0.14), 0px 2px 16px 1px rgba(19, 17, 32, 0.12)',
      '0px 5px 5px -3px rgba(19, 17, 32, 0.2), 0px 8px 10px 1px rgba(19, 17, 32, 0.14), 0px 3px 14px 2px rgba(19, 17, 32, 0.12)',
      '0px 5px 6px -3px rgba(19, 17, 32, 0.2), 0px 9px 12px 1px rgba(19, 17, 32, 0.14), 0px 3px 16px 2px rgba(19, 17, 32, 0.12)',
      '0px 6px 6px -3px rgba(19, 17, 32, 0.2), 0px 10px 14px 1px rgba(19, 17, 32, 0.14), 0px 4px 18px 3px rgba(19, 17, 32, 0.12)',
      '0px 6px 7px -4px rgba(19, 17, 32, 0.2), 0px 11px 15px 1px rgba(19, 17, 32, 0.14), 0px 4px 20px 3px rgba(19, 17, 32, 0.12)',
      '0px 7px 8px -4px rgba(19, 17, 32, 0.2), 0px 12px 17px 2px rgba(19, 17, 32, 0.14), 0px 5px 22px 4px rgba(19, 17, 32, 0.12)',
      '0px 7px 8px -4px rgba(19, 17, 32, 0.2), 0px 13px 19px 2px rgba(19, 17, 32, 0.14), 0px 5px 24px 4px rgba(19, 17, 32, 0.12)',
      '0px 7px 9px -4px rgba(19, 17, 32, 0.2), 0px 14px 21px 2px rgba(19, 17, 32, 0.14), 0px 5px 26px 4px rgba(19, 17, 32, 0.12)',
      '0px 8px 9px -5px rgba(19, 17, 32, 0.2), 0px 15px 22px 2px rgba(19, 17, 32, 0.14), 0px 6px 28px 5px rgba(19, 17, 32, 0.12)',
      '0px 8px 10px -5px rgba(19, 17, 32, 0.2), 0px 16px 24px 2px rgba(19, 17, 32, 0.14), 0px 6px 30px 5px rgba(19, 17, 32, 0.12)',
      '0px 8px 11px -5px rgba(19, 17, 32, 0.2), 0px 17px 26px 2px rgba(19, 17, 32, 0.14), 0px 6px 32px 5px rgba(19, 17, 32, 0.12)',
      '0px 9px 11px -5px rgba(19, 17, 32, 0.2), 0px 18px 28px 2px rgba(19, 17, 32, 0.14), 0px 7px 34px 6px rgba(19, 17, 32, 0.12)',
      '0px 9px 12px -6px rgba(19, 17, 32, 0.2), 0px 19px 29px 2px rgba(19, 17, 32, 0.14), 0px 7px 36px 6px rgba(19, 17, 32, 0.12)',
      '0px 10px 13px -6px rgba(19, 17, 32, 0.2), 0px 20px 31px 3px rgba(19, 17, 32, 0.14), 0px 8px 38px 7px rgba(19, 17, 32, 0.12)',
      '0px 10px 13px -6px rgba(19, 17, 32, 0.2), 0px 21px 33px 3px rgba(19, 17, 32, 0.14), 0px 8px 40px 7px rgba(19, 17, 32, 0.12)',
      '0px 10px 14px -6px rgba(19, 17, 32, 0.2), 0px 22px 35px 3px rgba(19, 17, 32, 0.14), 0px 8px 42px 7px rgba(19, 17, 32, 0.12)',
      '0px 11px 14px -7px rgba(19, 17, 32, 0.2), 0px 23px 36px 3px rgba(19, 17, 32, 0.14), 0px 9px 44px 8px rgba(19, 17, 32, 0.12)',
      '0px 11px 15px -7px rgba(19, 17, 32, 0.2), 0px 24px 38px 3px rgba(19, 17, 32, 0.14), 0px 9px 46px 8px rgba(19, 17, 32, 0.12)'
    ]
  }
}
