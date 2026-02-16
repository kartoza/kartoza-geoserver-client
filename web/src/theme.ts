import { extendTheme, type ThemeConfig } from '@chakra-ui/react'

// Kartoza brand colors - official brand palette
// Primary colors: #417d9b (blue), #dea037 (gold), #8b8d8a (gray)
const colors = {
  // Primary blue - Kartoza brand blue #417d9b
  kartoza: {
    50: '#e8f2f6',
    100: '#c5dfe8',
    200: '#9ecbd9',
    300: '#6eb2c7',
    400: '#4f9bb3',
    500: '#417d9b', // Primary brand blue
    600: '#386d87',
    700: '#2d5a70',
    800: '#234859',
    900: '#193642',
  },
  // Gold/orange accent - Kartoza brand gold #dea037
  accent: {
    50: '#fdf6e8',
    100: '#fae9c5',
    200: '#f5d89e',
    300: '#efc777',
    400: '#e9b650',
    500: '#dea037', // Primary brand gold
    600: '#c78d2f',
    700: '#a67525',
    800: '#865d1c',
    900: '#664612',
  },
  // Grays - Kartoza brand gray #8b8d8a
  gray: {
    50: '#f5f5f5', // light-bg
    100: '#e8e9e8', // light-bg-alt
    200: '#d4d5d4',
    300: '#b8b9b8',
    400: '#9fa09f',
    500: '#8b8d8a', // Primary brand gray
    600: '#737573',
    700: '#5c5e5c',
    800: '#454745',
    900: '#2e302e', // text-dark
  },
}

// Kartoza shadows using brand blue #417d9b (65, 125, 155)
const shadows = {
  sm: '0 2px 8px rgba(65, 125, 155, 0.08)',
  md: '0 4px 16px rgba(65, 125, 155, 0.12)',
  lg: '0 8px 32px rgba(65, 125, 155, 0.16)',
  xl: '0 16px 48px rgba(65, 125, 155, 0.20)',
  kartoza: '0 4px 16px rgba(65, 125, 155, 0.10), 0 1px 4px rgba(0, 0, 0, 0.06)',
  kartozaHover: '0 8px 28px rgba(65, 125, 155, 0.16), 0 2px 8px rgba(0, 0, 0, 0.08)',
  accent: '0 4px 20px rgba(222, 160, 55, 0.4)', // Brand gold #dea037
  accentHover: '0 6px 28px rgba(222, 160, 55, 0.5)',
}

// Border radii matching Kartoza design
const radii = {
  sm: '8px',
  md: '12px',
  lg: '20px',
  xl: '32px',
}

const config: ThemeConfig = {
  initialColorMode: 'light',
  useSystemColorMode: false,
}

const theme = extendTheme({
  config,
  colors,
  shadows,
  radii,
  fonts: {
    heading: "'Roboto', -apple-system, BlinkMacSystemFont, sans-serif",
    body: "'Roboto', -apple-system, BlinkMacSystemFont, sans-serif",
  },
  styles: {
    global: {
      body: {
        bg: 'gray.50',
        color: 'gray.900',
      },
      // Custom CSS for Kartoza styling using brand colors
      '.kartoza-gradient': {
        background: 'linear-gradient(90deg, #dea037 0%, #417d9b 100%)',
      },
      '.kartoza-card': {
        borderRadius: '12px',
        boxShadow: '0 4px 16px rgba(65, 125, 155, 0.10), 0 1px 4px rgba(0, 0, 0, 0.06)',
        transition: 'box-shadow 0.3s ease, transform 0.3s ease',
        _hover: {
          boxShadow: '0 8px 28px rgba(65, 125, 155, 0.16), 0 2px 8px rgba(0, 0, 0, 0.08)',
          transform: 'translateY(-3px)',
        },
      },
    },
  },
  components: {
    Button: {
      baseStyle: {
        borderRadius: '10px',
        fontWeight: '600',
        transition: 'all 0.25s ease',
      },
      defaultProps: {
        colorScheme: 'kartoza',
      },
      variants: {
        solid: {
          bg: 'kartoza.500',
          color: 'white',
          boxShadow: '0 2px 8px rgba(65, 125, 155, 0.12)',
          _hover: {
            bg: 'kartoza.600',
            boxShadow: '0 4px 14px rgba(65, 125, 155, 0.20)',
            transform: 'translateY(-1px)',
          },
        },
        outline: {
          borderColor: 'kartoza.500',
          borderWidth: '2px',
          color: 'kartoza.500',
          _hover: {
            bg: 'kartoza.50',
            transform: 'translateY(-1px)',
          },
        },
        ghost: {
          color: 'kartoza.500',
          _hover: {
            bg: 'kartoza.50',
          },
        },
        accent: {
          bg: 'accent.500',
          color: 'white',
          boxShadow: '0 4px 20px rgba(222, 160, 55, 0.4)',
          _hover: {
            bg: 'accent.600',
            boxShadow: '0 6px 28px rgba(222, 160, 55, 0.5)',
            transform: 'translateY(-2px)',
          },
        },
        'accent-outline': {
          borderColor: 'accent.400',
          borderWidth: '2px',
          color: 'accent.400',
          _hover: {
            bg: 'accent.50',
          },
        },
      },
      sizes: {
        lg: {
          fontSize: 'md',
          px: 8,
          py: 6,
        },
        xl: {
          fontSize: 'lg',
          px: 10,
          py: 7,
          minW: '200px',
        },
      },
    },
    Link: {
      baseStyle: {
        color: 'kartoza.500',
        _hover: {
          textDecoration: 'underline',
          color: 'kartoza.600',
        },
      },
    },
    Heading: {
      baseStyle: {
        color: 'gray.900',
        fontWeight: '600',
      },
      variants: {
        brand: {
          color: 'kartoza.700',
        },
        accent: {
          color: 'accent.400',
        },
      },
    },
    Card: {
      baseStyle: {
        container: {
          borderRadius: '12px',
          boxShadow: '0 4px 16px rgba(65, 125, 155, 0.10), 0 1px 4px rgba(0, 0, 0, 0.06)',
          transition: 'box-shadow 0.3s ease, transform 0.3s ease',
          overflow: 'hidden',
          _hover: {
            boxShadow: '0 8px 28px rgba(65, 125, 155, 0.16), 0 2px 8px rgba(0, 0, 0, 0.08)',
          },
        },
      },
      variants: {
        elevated: {
          container: {
            _hover: {
              transform: 'translateY(-3px)',
            },
          },
        },
        feature: {
          container: {
            borderLeft: '4px solid',
            borderLeftColor: 'kartoza.500',
          },
        },
        accent: {
          container: {
            borderTop: '4px solid',
            borderTopColor: 'accent.400',
          },
        },
      },
    },
    Modal: {
      baseStyle: {
        dialog: {
          borderRadius: '12px',
          boxShadow: '0 16px 48px rgba(65, 125, 155, 0.20)',
        },
        header: {
          borderBottom: '1px solid',
          borderBottomColor: 'gray.100',
        },
        footer: {
          borderTop: '1px solid',
          borderTopColor: 'gray.100',
        },
      },
    },
    Input: {
      defaultProps: {
        focusBorderColor: 'kartoza.500',
      },
      baseStyle: {
        field: {
          borderRadius: '8px',
        },
      },
    },
    Select: {
      defaultProps: {
        focusBorderColor: 'kartoza.500',
      },
    },
    Checkbox: {
      defaultProps: {
        colorScheme: 'kartoza',
      },
    },
    Switch: {
      defaultProps: {
        colorScheme: 'kartoza',
      },
    },
    Progress: {
      defaultProps: {
        colorScheme: 'kartoza',
      },
    },
    Tabs: {
      defaultProps: {
        colorScheme: 'kartoza',
      },
    },
    Badge: {
      baseStyle: {
        borderRadius: '6px',
        fontWeight: '600',
      },
      variants: {
        subtle: {
          bg: 'kartoza.50',
          color: 'kartoza.700',
        },
        solid: {
          bg: 'kartoza.500',
          color: 'white',
        },
        accent: {
          bg: 'accent.400',
          color: 'white',
        },
      },
    },
    Stat: {
      baseStyle: {
        container: {
          p: 4,
        },
        label: {
          color: 'gray.600',
          fontSize: 'sm',
          fontWeight: '500',
        },
        number: {
          color: 'kartoza.700',
          fontWeight: '700',
        },
        helpText: {
          color: 'gray.500',
        },
      },
    },
    Divider: {
      baseStyle: {
        borderColor: 'gray.200',
      },
    },
  },
  semanticTokens: {
    colors: {
      primary: 'kartoza.500',
      'primary.dark': 'kartoza.700',
      'primary.light': 'kartoza.300',
      secondary: 'accent.400',
      success: '#4CAF50',
      error: '#E55B3C',
      warning: 'accent.400',
      info: 'kartoza.300',
    },
  },
})

export default theme
