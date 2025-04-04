declare module '@mui/material/styles' {
  interface PaletteOptions {
    customColors?: {
      borderColor?: string
      entity?: {
        matchedBg?: string
        matchedBorder?: string
        unmatchedBg?: string
        unmatchedBorder?: string
        label?: string
        value?: string
      }

      messageState: {
        unspecified: string
        unprocessed: string
        queued: string
        categorizing: string
        extracting: string
        processed: string
        archived: string
        failed: string
      }
    }
  }

  interface Palette {
    customColors: {
      borderColor: string
      entity: {
        matchedBg: string
        matchedBorder: string
        unmatchedBg: string
        unmatchedBorder: string
        label: string
        value: string
      }

      messageState: {
        unspecified: string
        unprocessed: string
        queued: string
        categorizing: string
        extracting: string
        processed: string
        archived: string
        failed: string
      }
    }
  }
}

export {}
