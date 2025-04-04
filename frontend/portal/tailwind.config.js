import { baseConfig, packageTailwindFiles } from '@doota/tailwindcss-config'

/** @type {import('tailwindcss').Config} */
module.exports = {
  ...baseConfig,
  content: {
    relative: true,
    files: ['./src/**/*.{ts,tsx}', ...packageTailwindFiles('packages/ui-core')]
  }
}
