import { baseConfig, packageTailwindFiles } from './tailwind.config'

/** @type {import('tailwindcss').Config} */
module.exports = {
  ...baseConfig,
  content: {
    relative: true,
    files: [
      // FIXME: Could we use PNPM workspaces to get the list of packages somehow?
      ...packageTailwindFiles('packages/ui-core'),
      ...packageTailwindFiles('extension/outlook')
    ]
  }
}
