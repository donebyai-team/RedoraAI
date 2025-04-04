import { join } from 'path'
import { readdirSync } from 'fs'

/** @type {import('tailwindcss').Config} */
module.exports.baseConfig = {
  /** shared theme configuration */
  theme: {
    extend: {
      fontFamily: {
        mono: ['geistmono-regular', 'monospace']
      },
      fontSize: {
        xs: '12px',
        sm: '14px',
        base: '16px',
        lg: '18px',
        xl: '20px'
      },
      colors: {
        gray: {
          1: '#F5F5F5',
          2: '#D4D4D4',
          3: '#737373'
        },
        purple: {
          1: '#412D80',
          2: '#6244BF',
          3: '#A25AFF',
          4: '#C1ACFF',
          5: '#E0D6FF',
          6: '#F2EEFF',
          7: '#F8F6FF'
        },
        error: '#FF6166'
      },
      spacing: {
        px: '1px',
        0: '0',
        1: '10px',
        2: '20px',
        3: '40px'
      },
      borderRadius: {
        none: '0',
        DEFAULT: '10px',
        full: '9999px'
      },
      boxShadow: {
        DEFAULT: '0px 5px 20px rgba(0, 0, 0, 0.06)'
      }
    }
  },
  /** shared plugins configuration */
  plugins: []
}

const frontendRoot = join(__dirname, '../..')

// The packageTailwindFiles function is used to generate the glob patterns of files to be processed by Tailwind CSS
// in a certain package we have in the workspace. It exists mainly to make it easy to include all files in a package
// expect the `node_modules` directory. We hit this problem because our source files are in the root of the package
// so blindly doing `./packages/ui-core/**/*.{ts,tsx}` would include all files in `node_modules` as well.
//
// The received 'packageDirectory' should be relative to the frontend root directory. The function will list
// glob patterns that should then be used by tailwindcss to process the files in the package.
module.exports.packageTailwindFiles = function (packageDirectory) {
  const packagePath = join(frontendRoot, packageDirectory)

  return readdirSync(packagePath)
    .filter(file => !file.match(/\..+$/))
    .map(directory => join(packagePath, directory))
    .map(directory => join(directory, '**', '*.{ts,tsx}'))
}
