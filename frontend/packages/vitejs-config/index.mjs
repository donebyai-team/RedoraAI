import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react-swc'
import { resolve } from 'path'
import preserveDirectives from 'rollup-preserve-directives'

/**
 * @typedef {object} BaseConfigOptions
 * @property {string} rootDir
 * @property {string} basePath
 * @property {string} [srcDirRelativeToRoot]
 * @property {string[]} [publicDirRelativeToRoot]
 * @property {string[]} [outDirRelativeToRoot]
 * @property {import('vite').BuildOptions['lib'] | undefined} [buildLibConfig]
 * @property {import('vite').BuildOptions['rollupOptions']} [rollupOptions]
 * @property {(() => import('vite').PluginOption) | undefined} [watchRebuildPlugin]
 * @property {(() => import('vite').PluginOption) | undefined} [makeEntryPointPlugin]
 * @property {boolean} [enableReactPlugin]
 * @property {unknown[]} [extraPrePlugins]
 * @property {unknown[]} [extraPostPlugins]
 * @property {import('vite').UserConfig} [overrideConfig]
 */

/**
 * @param {BaseConfigOptions} options
 * @return {import('vite').UserConfig}
 */
export function baseConfig({
  rootDir,
  basePath,
  srcDirRelativeToRoot,
  publicDirRelativeToRoot,
  outDirRelativeToRoot,
  buildLibConfig,
  rollupOptions,
  watchRebuildPlugin,
  makeEntryPointPlugin,
  enableReactPlugin,
  extraPrePlugins,
  extraPostPlugins,
  overrideConfig
}) {
  const isDev = process.env.__DEV__ === 'true'
  const isProduction = !isDev

  /** @type {(_: string[] | undefined, _: string[]) => string} */
  const pathRelativeToRoot = (path, ...defaultPaths) => {
    return resolve(...[rootDir, ...(path ?? defaultPaths)])
  }

  const srcDir = pathRelativeToRoot(srcDirRelativeToRoot, 'src')

  return defineConfig({
    resolve: {
      alias: {
        '@src': srcDir
      }
    },
    base: basePath || '',
    plugins: [
      ...(extraPrePlugins ?? []),
      (enableReactPlugin ?? true) && react(),
      isDev && watchRebuildPlugin && watchRebuildPlugin(),
      isDev && makeEntryPointPlugin && makeEntryPointPlugin(),
      // Deals with source map error that are actual warnings since they deal only
      // with source map and from my experience, they were still good.
      //
      // See https://github.com/vitejs/vite/issues/15012#issuecomment-2143554173
      // for potential hypothesis and the actual Vite issue.
      preserveDirectives(),
      ...(extraPostPlugins ?? [])
    ],
    publicDir: pathRelativeToRoot(publicDirRelativeToRoot, 'public'),
    build: {
      outDir: pathRelativeToRoot(outDirRelativeToRoot),
      // We have multiple Turbo projects running concurrently on build, we cannot delete
      // the output directory because it will delete the other projects' output. The root command
      // to build clears the output directory manually since they control when the build starts.
      emptyOutDir: false,
      sourcemap: isDev,
      minify: isProduction,
      reportCompressedSize: isProduction,
      rollupOptions,
      lib: buildLibConfig
    },
    define: {
      // That is a pain because some dependencies also have `process.env.<name>` usage which
      // causes runtime error. We need to define them here so they get replaced across the board.
      // Locally you can run in 'frontend' root folder the following command:
      // `cat ./extension/chrome/dist/content-ui/index.iife.js| grep -oE "process.env.[a-zA-Z0-9_]+" | uniq`
      // and then check if you are missing some replacements.
      //
      // I tried 'vite-plugin-env-compatible' but it didn't work for me, the final build had
      // multiple `process.env.<name>` references.
      'process.env.NODE_ENV': isDev ? `"development"` : `"production"`,
      'process.env.NEXT_PUBLIC_API_URL': JSON.stringify(process.env.NEXT_PUBLIC_API_URL ?? 'http://localhost:8787'),
      'process.env.NEXT_PUBLIC_APP_URL': JSON.stringify(process.env.NEXT_PUBLIC_APP_URL ?? 'http://localhost:3000'),
      'process.env.BUF_BIGINT_DISABLE': undefined,
      'process.env.__NEXT_OPTIMIZE_FONTS': undefined,
      'process.env.NEXT_DEPLOYMENT_ID': undefined,
      'process.env.__NEXT_IMAGE_OPTS': undefined
    },
    ...overrideConfig
  })
}
