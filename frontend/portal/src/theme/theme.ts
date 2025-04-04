import { Inter } from 'next/font/google'
import { themeFactory } from '@doota/mui-config'

const inter = Inter({
  weight: ['300', '400', '500', '600', '700'],
  subsets: ['latin'],
  display: 'swap'
})

export const theme = themeFactory(inter.style.fontFamily)
export default theme
